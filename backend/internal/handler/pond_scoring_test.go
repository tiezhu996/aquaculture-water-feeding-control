package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"aquaculture-water-feeding-control/backend/internal/database"
	"aquaculture-water-feeding-control/backend/internal/handler"
	"aquaculture-water-feeding-control/backend/internal/middleware"
	"aquaculture-water-feeding-control/backend/internal/repository"
	"aquaculture-water-feeding-control/backend/internal/service"
)

func newPondEngine(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, err := database.Open("file:pondscoring?mode=memory&cache=shared", "sqlite", "production")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	pondRepo := repository.NewPondRepository(db)
	userRepo := repository.NewUserRepository(db)
	auditService := service.NewAuditService(repository.NewAuditRepository(db))
	authService := service.NewAuthService(userRepo, "pond-scoring-secret-2026-change-me", 8*time.Hour)
	pondService := service.NewPondService(pondRepo, auditService)
	authHandler := handler.NewAuthHandler(authService)
	pondHandler := handler.NewPondHandler(pondService)

	engine := gin.New()
	engine.Use(middleware.RequestID(), middleware.Recovery())
	api := engine.Group("/api")
	api.POST("/auth/login", authHandler.Login)
	protected := api.Group("")
	protected.Use(middleware.AuthRequired(authService))
	protected.GET("/ponds/:id", pondHandler.Get)
	pondWrite := protected.Group("/ponds")
	pondWrite.Use(middleware.RequireRoles("admin", "manager"))
	pondWrite.POST("", pondHandler.Create)
	return engine
}

func pondLogin(t *testing.T, engine *gin.Engine) string {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"username": "admin", "password": "admin123"})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("login status=%d body=%s", w.Code, w.Body.String())
	}
	var payload struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &payload)
	return payload.Data.Token
}

func pondDoJSON(t *testing.T, engine *gin.Engine, method, path, token string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var reader *bytes.Reader
	if body != nil {
		raw, _ := json.Marshal(body)
		reader = bytes.NewReader(raw)
	} else {
		reader = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	return w
}

func TestGetMissingPondReturns404(t *testing.T) {
	engine := newPondEngine(t)
	token := pondLogin(t, engine)
	w := pondDoJSON(t, engine, http.MethodGet, "/api/ponds/999999", token, nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("missing pond status = %d, want 404 (body %s)", w.Code, w.Body.String())
	}
}

func TestCreateDuplicatePondReturns409(t *testing.T) {
	engine := newPondEngine(t)
	token := pondLogin(t, engine)
	first := pondDoJSON(t, engine, http.MethodPost, "/api/ponds", token, map[string]any{
		"code": "P-DUP01", "name": "重复塘", "species": "草鱼", "areaSquareMeters": 1000,
		"capacityKg": 5000, "growthStage": "成长期", "status": "active", "manager": "张三",
	})
	if first.Code != http.StatusCreated {
		t.Fatalf("first create status = %d (body %s)", first.Code, first.Body.String())
	}
	second := pondDoJSON(t, engine, http.MethodPost, "/api/ponds", token, map[string]any{
		"code": "P-DUP01", "name": "重复塘二号", "species": "草鱼", "areaSquareMeters": 1000,
		"capacityKg": 5000, "growthStage": "成长期", "status": "active", "manager": "张三",
	})
	if second.Code != http.StatusConflict {
		t.Fatalf("duplicate pond status = %d, want 409 (body %s)", second.Code, second.Body.String())
	}
}
