package router

import (
	"aquaculture-water-feeding-control/backend/internal/config"
	"aquaculture-water-feeding-control/backend/internal/handler"
	"aquaculture-water-feeding-control/backend/internal/middleware"
	"aquaculture-water-feeding-control/backend/internal/service"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type Handlers struct {
	Auth       *handler.AuthHandler
	Health     *handler.HealthHandler
	Ponds      *handler.PondHandler
	Readings   *handler.ReadingHandler
	Plans      *handler.PlanHandler
	Executions *handler.ExecutionHandler
	Audit      *handler.AuditHandler
}

func New(cfg config.Config, redisClient *redis.Client, auth *service.AuthService, h Handlers) *gin.Engine {
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()
	engine.Use(middleware.RequestID(), middleware.AccessLog(), middleware.Recovery())
	engine.Use(cors.New(cors.Config{
		AllowOrigins:  cfg.CORSOrigins,
		AllowMethods:  []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowHeaders:  []string{"Origin", "Content-Type", "Authorization", "X-Request-ID"},
		ExposeHeaders: []string{"X-Request-ID", "X-RateLimit-Limit", "X-RateLimit-Remaining"},
		MaxAge:        12 * time.Hour,
	}))
	engine.GET("/healthz", h.Health.Health)
	engine.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "aquaculture-water-feeding-control", "docs": "/healthz"})
	})

	api := engine.Group("/api")
	api.Use(middleware.RateLimit(redisClient, cfg.RateLimit))
	api.POST("/auth/login", middleware.NamedRateLimit(redisClient, "login", 10), h.Auth.Login)
	protected := api.Group("")
	protected.Use(middleware.AuthRequired(auth))
	protected.GET("/auth/me", h.Auth.Me)

	protected.GET("/ponds", h.Ponds.List)
	protected.GET("/ponds/:id", h.Ponds.Get)
	pondWrite := protected.Group("/ponds")
	pondWrite.Use(middleware.RequireRoles("admin", "manager"))
	pondWrite.POST("", h.Ponds.Create)
	pondWrite.PUT("/:id", h.Ponds.Update)
	pondWrite.DELETE("/:id", h.Ponds.Delete)

	protected.GET("/readings", h.Readings.List)
	protected.GET("/readings/:id", h.Readings.Get)
	readingWrite := protected.Group("/readings")
	readingWrite.Use(middleware.RequireRoles("admin", "manager", "operator"))
	readingWrite.POST("", h.Readings.Create)
	readingWrite.PATCH("/:id/confirm", h.Readings.Confirm)
	readingDelete := protected.Group("/readings")
	readingDelete.Use(middleware.RequireRoles("admin", "manager"))
	readingDelete.DELETE("/:id", h.Readings.Delete)

	protected.GET("/plans", h.Plans.List)
	protected.GET("/plans/recommendation", h.Plans.Recommendation)
	protected.GET("/plans/:id", h.Plans.Get)
	planWrite := protected.Group("/plans")
	planWrite.Use(middleware.RequireRoles("admin", "manager", "operator"))
	planWrite.POST("", h.Plans.Create)
	planWrite.PUT("/:id", h.Plans.Update)
	planWrite.PATCH("/:id/submit", h.Plans.Submit)
	planReview := protected.Group("/plans")
	planReview.Use(middleware.RequireRoles("admin", "manager"))
	planReview.PATCH("/:id/approve", h.Plans.Approve)
	planReview.PATCH("/:id/revoke", h.Plans.Revoke)
	planReview.DELETE("/:id", h.Plans.Delete)

	protected.GET("/executions", h.Executions.List)
	protected.GET("/executions/:id", h.Executions.Get)
	executionWrite := protected.Group("/executions")
	executionWrite.Use(middleware.RequireRoles("admin", "manager", "operator"))
	executionWrite.POST("", h.Executions.Create)
	executionWrite.PUT("/:id", h.Executions.Update)
	executionWrite.PATCH("/:id/complete", h.Executions.Complete)
	executionWrite.DELETE("/:id", h.Executions.Delete)

	audit := protected.Group("/audit")
	audit.Use(middleware.RequireRoles("admin", "manager"))
	audit.GET("", h.Audit.List)
	return engine
}
