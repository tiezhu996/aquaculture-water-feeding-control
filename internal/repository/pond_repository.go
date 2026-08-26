package repository

import (
	"aquaculture-water-feeding-control/backend/internal/dto"
	"aquaculture-water-feeding-control/backend/internal/model"
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PondRepository struct {
	db *gorm.DB
}

func NewPondRepository(db *gorm.DB) *PondRepository {
	return &PondRepository{db: db}
}

func (r *PondRepository) List(query dto.PageQuery) ([]model.Pond, int64, error) {
	base := r.db.Model(&model.Pond{})
	if search := strings.TrimSpace(query.Search); search != "" {
		like := "%" + strings.ToLower(search) + "%"
		base = base.Where("LOWER(code) LIKE ? OR LOWER(name) LIKE ? OR LOWER(species) LIKE ?", like, like, like)
	}
	if query.Status != "" {
		base = base.Where("status = ?", query.Status)
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var ponds []model.Pond
	err := base.Order("updated_at DESC").Offset((query.Page - 1) * query.PageSize).Limit(query.PageSize).Find(&ponds).Error
	return ponds, total, err
}

func (r *PondRepository) Get(id uint) (model.Pond, error) {
	var pond model.Pond
	err := r.db.First(&pond, id).Error
	if err != nil {
		return pond, fmt.Errorf("query pond %d: %v", id, err)
	}
	return pond, nil
}

func (r *PondRepository) GetForUpdate(id uint) (model.Pond, error) {
	var pond model.Pond
	err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).First(&pond, id).Error
	return pond, err
}

func (r *PondRepository) Create(pond *model.Pond) error {
	return r.db.Create(pond).Error
}

func (r *PondRepository) Save(pond *model.Pond) error {
	return r.db.Save(pond).Error
}

func (r *PondRepository) Delete(pond *model.Pond) error {
	return r.db.Delete(pond).Error
}

func (r *PondRepository) DependencyCount(id uint) (int64, error) {
	var readings, plans, executions int64
	if err := r.db.Model(&model.WaterReading{}).Where("pond_id = ?", id).Count(&readings).Error; err != nil {
		return 0, err
	}
	if err := r.db.Model(&model.FeedingPlan{}).Where("pond_id = ?", id).Count(&plans).Error; err != nil {
		return 0, err
	}
	if err := r.db.Model(&model.ControlExecution{}).Where("pond_id = ?", id).Count(&executions).Error; err != nil {
		return 0, err
	}
	return readings + plans + executions, nil
}
