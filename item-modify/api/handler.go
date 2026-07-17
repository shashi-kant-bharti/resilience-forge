package main

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Handler holds the HTTP handlers for the Item resource.
type Handler struct {
	store *Store
}

// NewHandler wires a Store into a Handler.
func NewHandler(store *Store) *Handler {
	return &Handler{store: store}
}

// --- helpers -----------------------------------------------------------------

func newID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// --- request / response bodies -----------------------------------------------

type itemRequest struct {
	Name  string `json:"name"  binding:"required"`
	Value string `json:"value"`
}

// --- handlers ----------------------------------------------------------------

// Create POST /items
func (h *Handler) Create(c *gin.Context) {
	var req itemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now().UTC()
	item := &Item{
		ID:        newID(),
		Name:      req.Name,
		Value:     req.Value,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := h.store.Create(item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, item)
}

// List GET /items
func (h *Handler) List(c *gin.Context) {
	items, err := h.store.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

// Get GET /items/:id
func (h *Handler) Get(c *gin.Context) {
	id := c.Param("id")
	item, err := h.store.GetByID(id)
	if errors.Is(err, ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("item %q not found", id)})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)
}

// Update PUT /items/:id
func (h *Handler) Update(c *gin.Context) {
	id := c.Param("id")

	existing, err := h.store.GetByID(id)
	if errors.Is(err, ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("item %q not found", id)})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var req itemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated := &Item{
		ID:        existing.ID,
		Name:      req.Name,
		Value:     req.Value,
		CreatedAt: existing.CreatedAt,
		UpdatedAt: time.Now().UTC(),
	}
	if err := h.store.Update(updated); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updated)
}

// Delete DELETE /items/:id
func (h *Handler) Delete(c *gin.Context) {
	id := c.Param("id")
	err := h.store.Delete(id)
	if errors.Is(err, ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("item %q not found", id)})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
