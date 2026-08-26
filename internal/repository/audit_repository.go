package repository

import (
	"aquaculture-water-feeding-control/backend/internal/dto"
	"aquaculture-water-feeding-control/backend/internal/model"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type AuditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) Transaction(fn func(*gorm.DB) error) error {
	var err error
	for attempt := 0; attempt < 3; attempt++ {
		err = r.db.Transaction(fn, &sql.TxOptions{Isolation: sql.LevelSerializable})
		if !retryableTransactionError(err) {
			return err
		}
	}
	return err
}

func retryableTransactionError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && (pgErr.Code == "40001" || pgErr.Code == "40P01")
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
