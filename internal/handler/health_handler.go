package handler

import (
	"context"
	"net/http"
	"time"

	"aquaculture-water-feeding-control/backend/internal/database"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewHealthHandler(db *gorm.DB, redisClient *redis.Client) *HealthHandler {
	return &HealthHandler{db: db, redis: redisClient}
}

func (h *HealthHandler) Health(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	checks := gin.H{"database": "ok", "redis": "ok"}
	httpStatus := http.StatusOK
	if err := database.Ping(ctx, h.db); err != nil {
		checks["database"] = "unavailable"
		httpStatus = http.StatusServiceUnavailable
	}
	redisCtx, redisCancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer redisCancel()
	if err := h.redis.Ping(redisCtx).Err(); err != nil {
		checks["redis"] = "unavailable"
	}
	status := "ok"
	if httpStatus != http.StatusOK || checks["redis"] == "unavailable" {
		status = "degraded"
	}
	c.JSON(httpStatus, gin.H{"status": status, "checks": checks, "time": time.Now().UTC()})
}
