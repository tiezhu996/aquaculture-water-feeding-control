package repository

import (
	"aquaculture-water-feeding-control/backend/internal/dto"
	"aquaculture-water-feeding-control/backend/internal/model"

	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ExecutionRepository struct {
	db *gorm.DB
}

func NewExecutionRepository(db *gorm.DB) *ExecutionRepository {
	return &ExecutionRepository{db: db}
}

func (r *ExecutionRepository) List(query dto.PageQuery, pondID, planID uint) ([]model.ControlExecution, int64, error) {
	base := r.db.Model(&model.ControlExecution{})
	if pondID > 0 {
		base = base.Where("pond_id = ?", pondID)
	}
	if planID > 0 {
		base = base.Where("feeding_plan_id = ?", planID)
	}
	if query.Status != "" {
		base = base.Where("status = ?", query.Status)
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var executions []model.ControlExecution
	err := base.Preload("Pond").Order("scheduled_at DESC").Offset((query.Page - 1) * query.PageSize).Limit(query.PageSize).Find(&executions).Error
	return executions, total, err
}

func (r *ExecutionRepository) Get(id uint) (model.ControlExecution, error) {
	var execution model.ControlExecution
	err := r.db.Preload("Pond").First(&execution, id).Error
	return execution, err
}

func (r *ExecutionRepository) GetForUpdate(id uint) (model.ControlExecution, error) {
	var execution model.ControlExecution
	err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).Preload("Pond").First(&execution, id).Error
	return execution, err
}

func (r *ExecutionRepository) CountOpenForPlanExcluding(planID, excludedID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.ControlExecution{}).
		Where("feeding_plan_id = ? AND id <> ? AND status IN ?", planID, excludedID, []string{"scheduled", "running"}).Count(&count).Error
	return count, err
}

func (r *ExecutionRepository) PlannedAmountForDay(pondID uint, from, until time.Time, excludedID uint) (float64, error) {
	var total float64
	err := r.db.Model(&model.ControlExecution{}).
		Where("pond_id = ? AND id <> ? AND scheduled_at >= ? AND scheduled_at < ? AND status <> ?", pondID, excludedID, from, until, "cancelled").
		Select("COALESCE(SUM(planned_amount_kg), 0)").Scan(&total).Error
	return total, err
}

func (r *ExecutionRepository) Create(execution *model.ControlExecution) error {
	return r.db.Create(execution).Error
}

func (r *ExecutionRepository) Save(execution *model.ControlExecution) error {
	return r.db.Save(execution).Error
}

func (r *ExecutionRepository) Delete(execution *model.ControlExecution) error {
	return r.db.Delete(execution).Error
}
