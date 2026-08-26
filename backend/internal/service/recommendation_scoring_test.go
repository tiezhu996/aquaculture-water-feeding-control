package service_test

import (
	"fmt"
	"testing"
	"time"

	"aquaculture-water-feeding-control/backend/internal/constants"
	"aquaculture-water-feeding-control/backend/internal/database"
	"aquaculture-water-feeding-control/backend/internal/model"
	"aquaculture-water-feeding-control/backend/internal/repository"
	"aquaculture-water-feeding-control/backend/internal/service"

	"gorm.io/gorm"
)

var s005seq int

func open005(t *testing.T) *gorm.DB {
	t.Helper()
	s005seq++
	db, err := database.Open("file:scoring005-"+itoa05(s005seq)+"?mode=memory&cache=shared", "sqlite", "production")
	if err != nil { t.Fatalf("open db: %v", err) }
	return db
}

func newPlanSvc005(t *testing.T) (*service.PlanService, uint, uint) {
	t.Helper()
	db := open005(t)
	now := time.Now().UTC()
	pond1 := model.Pond{Code: "P-REC05", Name: "塘一", Species: "草鱼", AreaSquareMeters: 1000, CapacityKg: 5000, GrowthStage: "成长期", Status: constants.PondStatusActive}
	db.Create(&pond1)
	pond2 := model.Pond{Code: "P-REC06", Name: "塘二", Species: "草鱼", AreaSquareMeters: 1000, CapacityKg: 5000, GrowthStage: "成长期", Status: constants.PondStatusActive}
	db.Create(&pond2)
	plan1 := model.FeedingPlan{PondID: pond1.ID, Name: "计划一", Version: 1, DailyAmountKg: 100, FrequencyPerDay: 2, FeedType: "颗粒料", TargetGrowthStage: "成长期", MinOxygen: 4, StartDate: now.Add(-24*time.Hour), EndDate: now.Add(24*time.Hour), Status: constants.PlanStatusApproved, CreatedBy: "admin"}
	db.Create(&plan1)
	plan2 := model.FeedingPlan{PondID: pond2.ID, Name: "计划二", Version: 1, DailyAmountKg: 100, FrequencyPerDay: 2, FeedType: "颗粒料", TargetGrowthStage: "成长期", MinOxygen: 4, StartDate: now.Add(-24*time.Hour), EndDate: now.Add(24*time.Hour), Status: constants.PlanStatusApproved, CreatedBy: "admin"}
	db.Create(&plan2)
	reading1 := model.WaterReading{PondID: pond1.ID, DissolvedOxygen: 4.2, Temperature: 26, PH: 7.5, Ammonia: 0.1, Turbidity: 30, MeasuredAt: now, Source: "manual", RiskLevel: constants.RiskWarning}
	db.Create(&reading1)
	reading2 := model.WaterReading{PondID: pond2.ID, DissolvedOxygen: 6, Temperature: 26, PH: 7.5, Ammonia: 0.1, Turbidity: 30, MeasuredAt: now, Source: "manual", RiskLevel: constants.RiskNormal}
	db.Create(&reading2)
	return service.NewPlanService(repository.NewPlanRepository(db), repository.NewPondRepository(db), repository.NewReadingRepository(db), service.NewAuditService(repository.NewAuditRepository(db))), pond1.ID, pond2.ID
}

func TestRecommendationReasonsIndependent(t *testing.T) {
	svc, pond1, pond2 := newPlanSvc005(t)
	rec1, err := svc.Recommendation(pond1, "晴朗")
	if err != nil { t.Fatalf("rec1: %v", err) }
	_, err = svc.Recommendation(pond2, "暴雨")
	if err != nil { t.Fatalf("rec2: %v", err) }
	if len(rec1.Reasons) < 2 {
		t.Fatalf("rec1 reasons too few: %v", rec1.Reasons)
	}
	if rec1.Reasons[1] != "存在水质预警，建议减量 30%" {
		t.Fatalf("rec1 reasons corrupted: %v", rec1.Reasons)
	}
	if len(rec1.Hints) == 0 || rec1.Hints[0] != "建议分批少量投喂" {
		t.Fatalf("rec1 hints corrupted: %v", rec1.Hints)
	}
	wantNote := fmt.Sprintf("养殖池 %d 建议基于最新水质生成", pond1)
	if len(rec1.Notes) == 0 || rec1.Notes[0] != wantNote {
		t.Fatalf("rec1 notes corrupted: %v (want %s)", rec1.Notes, wantNote)
	}
}

func itoa05(v int) string {
	if v == 0 { return "0" }
	var b [20]byte
	i := len(b)
	for v > 0 { i--; b[i] = byte('0'+v%10); v /= 10 }
	return string(b[i:])
}
