package repository

import (
	"aquaculture-water-feeding-control/backend/internal/dto"
	"aquaculture-water-feeding-control/backend/internal/model"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PlanRepository struct {
	db *gorm.DB
}

func NewPlanRepository(db *gorm.DB) *PlanRepository {
	return &PlanRepository{db: db}
}

func (r *PlanRepository) List(query dto.PageQuery, pondID uint) ([]model.FeedingPlan, int64, error) {
	base := r.db.Model(&model.FeedingPlan{})
	if pondID > 0 {
		base = base.Where("pond_id = ?", pondID)
	}
	if query.Status != "" {
		base = base.Where("status = ?", query.Status)
	}
	if search := strings.TrimSpace(query.Search); search != "" {
		base = base.Where("LOWER(name) LIKE ?", "%"+strings.ToLower(search)+"%")
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var plans []model.FeedingPlan
	err := base.Preload("Pond").Order("updated_at DESC").Offset((query.Page - 1) * query.PageSize).Limit(query.PageSize).Find(&plans).Error
	return plans, total, err
}

func (r *PlanRepository) Get(id uint) (model.FeedingPlan, error) {
	var plan model.FeedingPlan
	err := r.db.Preload("Pond").First(&plan, id).Error
	return plan, err
}

func (r *PlanRepository) GetForUpdate(id uint) (model.FeedingPlan, error) {
	var plan model.FeedingPlan
	err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).Preload("Pond").First(&plan, id).Error
	return plan, err
}

func (r *PlanRepository) Create(plan *model.FeedingPlan) error {
	return r.db.Create(plan).Error
}

func (r *PlanRepository) Save(plan *model.FeedingPlan) error {
	return r.db.Save(plan).Error
}

func (r *PlanRepository) Delete(plan *model.FeedingPlan) error {
	return r.db.Delete(plan).Error
}

func (r *PlanRepository) ExecutionCount(id uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.ControlExecution{}).Where("feeding_plan_id = ?", id).Count(&count).Error
	return count, err
}

func (r *PlanRepository) LatestApprovedForPond(pondID uint) (model.FeedingPlan, error) {
	var plan model.FeedingPlan
	err := r.db.Preload("Pond").Where("pond_id = ? AND status = ?", pondID, "approved").Order("version DESC, updated_at DESC").First(&plan).Error
	return plan, err
}
