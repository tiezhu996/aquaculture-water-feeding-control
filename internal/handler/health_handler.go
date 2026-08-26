package handler

import (
	"context"
	"net/http"
	"time"

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
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()
	checks := gin.H{"database": "ok", "redis": "ok"}
	httpStatus := http.StatusOK
	sqlDB, err := h.db.DB()
	if err != nil || sqlDB.PingContext(ctx) != nil {
		checks["database"] = "unavailable"
		httpStatus = http.StatusServiceUnavailable
	}
	if err := h.redis.Ping(ctx).Err(); err != nil {
		checks["redis"] = "unavailable"
	}
	status := "ok"
	if httpStatus != http.StatusOK || checks["redis"] == "unavailable" {
		status = "degraded"
	}
	c.JSON(httpStatus, gin.H{"status": status, "checks": checks, "time": time.Now().UTC()})
}
