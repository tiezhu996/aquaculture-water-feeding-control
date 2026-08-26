package repository

import (
	"aquaculture-water-feeding-control/backend/internal/dto"
	"aquaculture-water-feeding-control/backend/internal/model"

	"gorm.io/gorm"
)

type AuditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) Transaction(fn func(*gorm.DB) error) (err error) {
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
		if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit().Error
		}
	}()
	err = fn(tx)
	return err
}

func (r *AuditRepository) Create(log *model.AuditLog) error {
	return r.db.Create(log).Error
}

func (r *AuditRepository) List(query dto.PageQuery, entityType, action string) ([]model.AuditLog, int64, error) {
	base := r.db.Model(&model.AuditLog{})
	if entityType != "" {
		base = base.Where("entity_type = ?", entityType)
	}
	if action != "" {
		base = base.Where("action = ?", action)
	}
	if query.Search != "" {
		base = base.Where("actor ILIKE ? OR reason ILIKE ?", "%"+query.Search+"%", "%"+query.Search+"%")
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var logs []model.AuditLog
	err := base.Order("created_at DESC").Offset((query.Page - 1) * query.PageSize).Limit(query.PageSize).Find(&logs).Error
	return logs, total, err
}
