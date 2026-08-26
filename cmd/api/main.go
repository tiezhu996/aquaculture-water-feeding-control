package main

import (
	"aquaculture-water-feeding-control/backend/internal/config"
	"aquaculture-water-feeding-control/backend/internal/database"
	"aquaculture-water-feeding-control/backend/internal/handler"
	"aquaculture-water-feeding-control/backend/internal/repository"
	"aquaculture-water-feeding-control/backend/internal/router"
	"aquaculture-water-feeding-control/backend/internal/service"
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load configuration: %v", err)
	}
	db, err := database.Open(cfg.DatabaseURL, cfg.DBDriver, cfg.Environment)
	if err != nil {
		log.Fatalf("initialize database: %v", err)
	}
	redisClient := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr, Password: cfg.RedisPassword})
	defer redisClient.Close()

	userRepo := repository.NewUserRepository(db)
	pondRepo := repository.NewPondRepository(db)
	readingRepo := repository.NewReadingRepository(db)
	planRepo := repository.NewPlanRepository(db)
	executionRepo := repository.NewExecutionRepository(db)
	auditRepo := repository.NewAuditRepository(db)

	auditService := service.NewAuditService(auditRepo)
	authService := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.TokenTTL)
	pondService := service.NewPondService(pondRepo, auditService)
	readingService := service.NewReadingService(readingRepo, pondRepo, auditService)
	planService := service.NewPlanService(planRepo, pondRepo, readingRepo, auditService)
	executionService := service.NewExecutionService(executionRepo, planRepo, pondRepo, readingRepo, auditService)

	handlers := router.Handlers{
		Auth:       handler.NewAuthHandler(authService),
		Health:     handler.NewHealthHandler(db, redisClient),
		Ponds:      handler.NewPondHandler(pondService),
		Readings:   handler.NewReadingHandler(readingService),
		Plans:      handler.NewPlanHandler(planService),
		Executions: handler.NewExecutionHandler(executionService),
		Audit:      handler.NewAuditHandler(auditService),
	}
	engine := router.New(cfg, redisClient, authService, handlers)
	server := &http.Server{
		Addr: cfg.HTTPAddr, Handler: engine, ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout: 15 * time.Second, WriteTimeout: 20 * time.Second, IdleTimeout: 60 * time.Second,
	}
	go func() {
		log.Printf("aquaculture control API listening on %s", cfg.HTTPAddr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("serve API: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}
