package repository

import (
	"aquaculture-water-feeding-control/backend/internal/dto"
	"aquaculture-water-feeding-control/backend/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ReadingRepository struct {
	db *gorm.DB
}

func NewReadingRepository(db *gorm.DB) *ReadingRepository {
	return &ReadingRepository{db: db}
}

func (r *ReadingRepository) List(query dto.PageQuery, pondID uint, unconfirmed bool) ([]model.WaterReading, int64, error) {
	base := r.db.Model(&model.WaterReading{})
	if pondID > 0 {
		base = base.Where("pond_id = ?", pondID)
	}
	if query.Status != "" {
		base = base.Where("risk_level = ?", query.Status)
	}
	if unconfirmed {
		base = base.Where("confirmed = ? AND risk_level <> ?", false, "normal")
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var readings []model.WaterReading
	err := base.Preload("Pond").Order("measured_at DESC").Offset((query.Page - 1) * query.PageSize).Limit(query.PageSize).Find(&readings).Error
	return readings, total, err
}

func (r *ReadingRepository) Get(id uint) (model.WaterReading, error) {
	var reading model.WaterReading
	err := r.db.Preload("Pond").First(&reading, id).Error
	return reading, err
}

func (r *ReadingRepository) GetForUpdate(id uint) (model.WaterReading, error) {
	var reading model.WaterReading
	err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).Preload("Pond").First(&reading, id).Error
	return reading, err
}

func (r *ReadingRepository) Create(reading *model.WaterReading) error {
	return r.db.Create(reading).Error
}

func (r *ReadingRepository) Save(reading *model.WaterReading) error {
	return r.db.Save(reading).Error
}

func (r *ReadingRepository) Delete(reading *model.WaterReading) error {
	return r.db.Delete(reading).Error
}

func (r *ReadingRepository) LatestForPond(pondID uint) (model.WaterReading, error) {
	var reading model.WaterReading
	err := r.db.Where("pond_id = ?", pondID).Order("measured_at DESC").First(&reading).Error
	return reading, err
}
