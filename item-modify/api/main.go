package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	cosCfg, err := loadCOSConfig()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	store, err := NewStore(cosCfg)
	if err != nil {
		log.Fatalf("COS store error: %v", err)
	}

	h := NewHandler(store)

	r := gin.Default()

	// CORS — allow the UI (any origin in dev) to call the API.
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// REST API
	items := r.Group("/api/items")
	{
		items.POST("", h.Create)
		items.GET("", h.List)
		items.GET("/:id", h.Get)
		items.PUT("/:id", h.Update)
		items.DELETE("/:id", h.Delete)
	}

	// UI_DIR defaults to ../ui (local dev); set to /ui inside the container.
	uiDir := os.Getenv("UI_DIR")
	if uiDir == "" {
		uiDir = "../ui"
	}
	r.StaticFile("/", uiDir+"/index.html")
	r.Static("/ui", uiDir)

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

// loadCOSConfig resolves COS credentials in priority order:
//  1. IBM Secrets Manager (when SM_INSTANCE_URL, SM_API_KEY, SM_SECRET_ID are set)
//     — fetches the IAM API key from the iam_credentials secret; the remaining
//       COS vars (COS_INSTANCE_CRN, COS_ENDPOINT, COS_BUCKET) are still read
//       from the environment.
//  2. Direct environment variables (COS_API_KEY, COS_INSTANCE_CRN, COS_ENDPOINT, COS_BUCKET)
func loadCOSConfig() (COSConfig, error) {
	smCfg, ok, err := SecretsManagerConfigFromEnv()
	if err != nil {
		return COSConfig{}, err
	}
	if ok {
		log.Println("loading COS IAM API key from IBM Secrets Manager")
		apiKey, err := IAMAPIKeyFromSecretsManager(smCfg)
		if err != nil {
			return COSConfig{}, err
		}
		cosCfg, err := COSInfraConfigFromEnv()
		if err != nil {
			return COSConfig{}, err
		}
		cosCfg.APIKey = apiKey
		return cosCfg, nil
	}

	log.Println("loading COS credentials from environment variables")
	return COSConfigFromEnv()
}
