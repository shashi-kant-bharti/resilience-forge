package main

import (
	"fmt"
	"os"
)

// COSConfig holds the IBM Cloud Object Storage connection settings.
type COSConfig struct {
	APIKey      string
	InstanceCRN string
	Endpoint    string
	Bucket      string
}

// COSConfigFromEnv reads all four COS settings from environment variables.
// Use this when NOT using Secrets Manager.
//
// Required environment variables:
//
//	COS_API_KEY       – IBM Cloud IAM API key
//	COS_INSTANCE_CRN  – COS service instance CRN
//	COS_ENDPOINT      – regional endpoint (e.g. s3.us-south.cloud-object-storage.appdomain.cloud)
//	COS_BUCKET        – bucket name
func COSConfigFromEnv() (COSConfig, error) {
	cfg := COSConfig{
		APIKey:      os.Getenv("COS_API_KEY"),
		InstanceCRN: os.Getenv("COS_INSTANCE_CRN"),
		Endpoint:    os.Getenv("COS_ENDPOINT"),
		Bucket:      os.Getenv("COS_BUCKET"),
	}
	for name, val := range map[string]string{
		"COS_API_KEY":      cfg.APIKey,
		"COS_INSTANCE_CRN": cfg.InstanceCRN,
		"COS_ENDPOINT":     cfg.Endpoint,
		"COS_BUCKET":       cfg.Bucket,
	} {
		if val == "" {
			return COSConfig{}, fmt.Errorf("environment variable %s is required", name)
		}
	}
	return cfg, nil
}

// COSInfraConfigFromEnv reads only the three infrastructure vars (CRN, endpoint,
// bucket). Used when the API key is sourced from Secrets Manager instead.
//
// Required environment variables:
//
//	COS_INSTANCE_CRN  – COS service instance CRN
//	COS_ENDPOINT      – regional endpoint
//	COS_BUCKET        – bucket name
func COSInfraConfigFromEnv() (COSConfig, error) {
	cfg := COSConfig{
		InstanceCRN: os.Getenv("COS_INSTANCE_CRN"),
		Endpoint:    os.Getenv("COS_ENDPOINT"),
		Bucket:      os.Getenv("COS_BUCKET"),
	}
	for name, val := range map[string]string{
		"COS_INSTANCE_CRN": cfg.InstanceCRN,
		"COS_ENDPOINT":     cfg.Endpoint,
		"COS_BUCKET":       cfg.Bucket,
	} {
		if val == "" {
			return COSConfig{}, fmt.Errorf("environment variable %s is required", name)
		}
	}
	return cfg, nil
}

// SecretsManagerConfig holds the connection settings for IBM Secrets Manager.
//
// When these three env vars are set, the service fetches the COS IAM API key
// from Secrets Manager instead of COS_API_KEY. The remaining COS vars
// (COS_INSTANCE_CRN, COS_ENDPOINT, COS_BUCKET) are still required.
//
// Required environment variables:
//
//	SM_INSTANCE_URL  – SM instance URL
//	                   e.g. https://d53e26db-b7c1-461d-94e6-8de136174d04.eu-gb.secrets-manager.appdomain.cloud
//	SM_API_KEY       – IAM API key that has SecretsReader access on the SM instance
//	SM_SECRET_ID     – ID of the iam_credentials secret (e.g. 2c54b838-360b-fc07-cff9-3eb9a13075e1)
type SecretsManagerConfig struct {
	InstanceURL string
	APIKey      string
	SecretID    string
}

// SecretsManagerConfigFromEnv reads SM settings from environment variables.
//   - Returns (cfg, true, nil)  when all three SM vars are present.
//   - Returns (zero, false, nil) when none are set → caller falls back to COS_API_KEY env var.
//   - Returns an error when only some are set (misconfiguration).
func SecretsManagerConfigFromEnv() (SecretsManagerConfig, bool, error) {
	url := os.Getenv("SM_INSTANCE_URL")
	key := os.Getenv("SM_API_KEY")
	id := os.Getenv("SM_SECRET_ID")

	if url == "" && key == "" && id == "" {
		return SecretsManagerConfig{}, false, nil
	}

	for name, val := range map[string]string{
		"SM_INSTANCE_URL": url,
		"SM_API_KEY":      key,
		"SM_SECRET_ID":    id,
	} {
		if val == "" {
			return SecretsManagerConfig{}, false,
				fmt.Errorf("environment variable %s is required when using Secrets Manager", name)
		}
	}

	return SecretsManagerConfig{InstanceURL: url, APIKey: key, SecretID: id}, true, nil
}
