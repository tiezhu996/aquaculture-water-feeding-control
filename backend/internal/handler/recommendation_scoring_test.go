package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"aquaculture-water-feeding-control/backend/internal/constants"
	"aquaculture-water-feeding-control/backend/internal/database"
	"aquaculture-water-feeding-control/backend/internal/handler"
	"aquaculture-water-feeding-control/backend/internal/middleware"
	"aquaculture-water-feeding-control/backend/internal/model"
	"aquaculture-water-feeding-control/backend/internal/repository"
	"aquaculture-water-feeding-control/backend/internal/service"
)

var s005hseq int

func newRecommendationEngine(t *testing.T) (*gin.Engine, uint) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	s005hseq++
	db, err := database.Open("file:rechandler005-"+itoa05h(s005hseq)+"?mode=memory&cache=shared", "sqlite", "production")
	if err != nil { t.Fatalf("open db: %v", err) }
	now := time.Now().UTC()
	pond := model.Pond{Code: "P-RH05", Name: "塘", Species: "草鱼", AreaSquareMeters: 1000, CapacityKg: 5000, GrowthStage: "成长期", Status: constants.PondStatusActive}
	db.Create(&pond)
	plan := model.FeedingPlan{PondID: pond.ID, Name: "计划", Version: 1, DailyAmountKg: 100, FrequencyPerDay: 2, FeedType: "颗粒料", TargetGrowthStage: "成长期", MinOxygen: 4, StartDate: now.Add(-24*time.Hour), EndDate: now.Add(24*time.Hour), Status: constants.PlanStatusApproved, CreatedBy: "admin"}
	db.Create(&plan)
	reading := model.WaterReading{PondID: pond.ID, DissolvedOxygen: 4.2, Temperature: 26, PH: 7.5, Ammonia: 0.1, Turbidity: 30, MeasuredAt: now, Source: "manual", RiskLevel: constants.RiskWarning}
	db.Create(&reading)

	planRepo := repository.NewPlanRepository(db)
	pondRepo := repository.NewPondRepository(db)
	readingRepo := repository.NewReadingRepository(db)
	userRepo := repository.NewUserRepository(db)
	auditService := service.NewAuditService(repository.NewAuditRepository(db))
	authService := service.NewAuthService(userRepo, "rec-scoring-secret-2026-change-me", 8*time.Hour)
	planService := service.NewPlanService(planRepo, pondRepo, readingRepo, auditService)
	authHandler := handler.NewAuthHandler(authService)
	planHandler := handler.NewPlanHandler(planService)

	engine := gin.New()
	engine.Use(middleware.RequestID(), middleware.Recovery())
	api := engine.Group("/api")
	api.POST("/auth/login", authHandler.Login)
	protected := api.Group("")
	protected.Use(middleware.AuthRequired(authService))
	protected.GET("/plans/recommendation", planHandler.Recommendation)
	return engine, pond.ID
}

func recLogin(t *testing.T, engine *gin.Engine) string {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"username": "admin", "password": "admin123"})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	var payload struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &payload)
	return payload.Data.Token
}

func recDo(t *testing.T, engine *gin.Engine, path, token string) []string {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	var payload struct {
		WeatherTags  []string `json:"weatherTags"`
		WeatherCodes []string `json:"weatherCodes"`
	}
	json.Unmarshal(w.Body.Bytes(), &payload)
	return append(payload.WeatherTags, payload.WeatherCodes...)
}

func TestRecommendationWeatherTagsIndependent(t *testing.T) {
	engine, pondID := newRecommendationEngine(t)
	token := recLogin(t, engine)
	_ = recDo(t, engine, "/api/plans/recommendation?pondId="+itoa05h(int(pondID))+"&weather=sunny", token)
	got := recDo(t, engine, "/api/plans/recommendation?pondId="+itoa05h(int(pondID))+"&weather=storm", token)
	if len(got) != 2 || got[0] != "storm" || got[1] != "STORM" {
		t.Fatalf("second call weather tags/codes corrupted: %v", got)
	}
}

func itoa05h(v int) string {
	if v == 0 { return "0" }
	var b [20]byte
	i := len(b)
	for v > 0 { i--; b[i] = byte('0'+v%10); v /= 10 }
	return string(b[i:])
}
