package service_test

import (
	"testing"
	"time"

	"aquaculture-water-feeding-control/backend/internal/constants"
	"aquaculture-water-feeding-control/backend/internal/database"
	"aquaculture-water-feeding-control/backend/internal/model"
	"aquaculture-water-feeding-control/backend/internal/repository"
	"aquaculture-water-feeding-control/backend/internal/service"
)

func newRevokeService(t *testing.T, withExecution bool) (*service.PlanService, uint) {
	t.Helper()
	db, err := database.Open("file:revokescoring?mode=memory&cache=shared", "sqlite", "production")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	now := time.Now().UTC()
	pond := model.Pond{Code: "P-RV01", Name: "撤销塘", Species: "草鱼", AreaSquareMeters: 1000, CapacityKg: 5000, GrowthStage: "成长期", Status: constants.PondStatusActive, Manager: "张三"}
	db.Create(&pond)
	plan := model.FeedingPlan{PondID: pond.ID, Name: "撤销计划", Version: 1, DailyAmountKg: 100, FrequencyPerDay: 2, FeedType: "颗粒料", TargetGrowthStage: "成长期", MinOxygen: 4, StartDate: now.Add(-24*time.Hour), EndDate: now.Add(24*time.Hour), Status: constants.PlanStatusApproved, CreatedBy: "admin"}
	db.Create(&plan)
	if withExecution {
		exec := model.ControlExecution{PondID: pond.ID, FeedingPlanID: plan.ID, ScheduledAt: now, PlannedAmountKg: 50, Status: constants.ExecutionScheduled, Operator: "admin"}
		db.Create(&exec)
	}
	planRepo := repository.NewPlanRepository(db)
	pondRepo := repository.NewPondRepository(db)
	readingRepo := repository.NewReadingRepository(db)
	auditService := service.NewAuditService(repository.NewAuditRepository(db))
	return service.NewPlanService(planRepo, pondRepo, readingRepo, auditService), plan.ID
}

func TestRevokePlanWithExecutionReturnsConflict(t *testing.T) {
	svc, planID := newRevokeService(t, true)
	actor := service.Actor{UserID: 1, Username: "admin", DisplayName: "管理员", Role: "admin"}
	_, err := svc.Revoke(planID, "尝试撤销", actor)
	if err == nil {
		t.Fatalf("revoke plan with execution should return conflict error, got nil")
	}
	appErr, ok := service.AsAppError(err)
	if !ok || appErr.Code != service.CodeConflict {
		t.Fatalf("revoke error code = %v (type %T), want CONFLICT", err, err)
	}
}

func TestRevokePlanWithoutExecutionSucceeds(t *testing.T) {
	svc, planID := newRevokeService(t, false)
	actor := service.Actor{UserID: 1, Username: "admin", DisplayName: "管理员", Role: "admin"}
	_, err := svc.Revoke(planID, "正常撤销", actor)
	if err != nil {
		t.Fatalf("revoke plan without execution should succeed, got %v", err)
	}
}
