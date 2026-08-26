package router_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"aquaculture-water-feeding-control/backend/internal/config"
	"aquaculture-water-feeding-control/backend/internal/database"
	"aquaculture-water-feeding-control/backend/internal/handler"
	"aquaculture-water-feeding-control/backend/internal/repository"
	"aquaculture-water-feeding-control/backend/internal/router"
	"aquaculture-water-feeding-control/backend/internal/service"
)

func newScoringEngine(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	db, err := database.Open("file:rlscoring?mode=memory&cache=shared", "sqlite", "production")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	userRepo := repository.NewUserRepository(db)
	pondRepo := repository.NewPondRepository(db)
	readingRepo := repository.NewReadingRepository(db)
	planRepo := repository.NewPlanRepository(db)
	executionRepo := repository.NewExecutionRepository(db)
	auditRepo := repository.NewAuditRepository(db)
	auditService := service.NewAuditService(auditRepo)
	authService := service.NewAuthService(userRepo, "scoring-secret-2026-change-me-ok", 8*time.Hour)
	handlers := router.Handlers{
		Auth:       handler.NewAuthHandler(authService),
		Health:     handler.NewHealthHandler(db, client),
		Ponds:      handler.NewPondHandler(service.NewPondService(pondRepo, auditService)),
		Readings:   handler.NewReadingHandler(service.NewReadingService(readingRepo, pondRepo, auditService)),
		Plans:      handler.NewPlanHandler(service.NewPlanService(planRepo, pondRepo, readingRepo, auditService)),
		Executions: handler.NewExecutionHandler(service.NewExecutionService(executionRepo, planRepo, pondRepo, readingRepo, auditService)),
		Audit:      handler.NewAuditHandler(auditService),
	}
	cfg := config.Config{
		Environment: "test", HTTPAddr: ":0", DatabaseURL: "file:rlscoring?mode=memory&cache=shared",
		DBDriver: "sqlite", RedisAddr: mr.Addr(), JWTSecret: "scoring-secret-2026-change-me-ok",
		TokenTTL: 8 * time.Hour, CORSOrigins: []string{"*"}, RateLimit: 100000,
	}
	return router.New(cfg, client, authService, handlers)
}

func TestRateLimitConcurrentScopes(t *testing.T) {
	engine := newScoringEngine(t)
	const workers = 40
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			body := bytes.NewBufferString(`{"username":"admin","password":"admin123"}`)
			req := httptest.NewRequest(http.MethodPost, "/api/auth/login", body)
			req.Header.Set("Content-Type", "application/json")
			req.RemoteAddr = "10.10.10.10:40000"
			rec := httptest.NewRecorder()
			engine.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK && rec.Code != http.StatusTooManyRequests && rec.Code != http.StatusUnauthorized {
				t.Errorf("unexpected status = %d", rec.Code)
			}
		}()
	}
	close(start)
	wg.Wait()
}
