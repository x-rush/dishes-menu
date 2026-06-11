package api

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type HealthHandler struct {
	db *sqlx.DB
}

func NewHealthHandler(db *sqlx.DB) *HealthHandler { return &HealthHandler{db: db} }

func (h *HealthHandler) Health(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 1*time.Second)
	defer cancel()
	dbStatus := "ok"
	if err := h.db.PingContext(ctx); err != nil {
		dbStatus = "down: " + err.Error()
	}
	c.JSON(200, gin.H{
		"status":  "ok",
		"db":      dbStatus,
		"version": "1.0.0",
	})
}
