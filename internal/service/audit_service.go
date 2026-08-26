package service

import (
	"aquaculture-water-feeding-control/backend/internal/dto"
	"aquaculture-water-feeding-control/backend/internal/model"
	"aquaculture-water-feeding-control/backend/internal/repository"
	"encoding/json"

	"gorm.io/gorm"
)

type AuditService struct {
	repo *repository.AuditRepository
}

func (s *AuditService) WithinTransaction(fn func(*gorm.DB, *AuditService) error) error {
	return s.repo.Transaction(func(tx *gorm.DB) error {
		return fn(tx, NewAuditService(repository.NewAuditRepository(tx)))
	})
}

func NewAuditService(repo *repository.AuditRepository) *AuditService {
	return &AuditService{repo: repo}
}

func (s *AuditService) Record(actor Actor, action, entityType string, entityID uint, before, after any, reason string) error {
	fromState, err := encodeSnapshot(before)
	if err != nil {
		return WrapError(CodeInternal, "无法序列化审计前状态", err)
	}
	toState, err := encodeSnapshot(after)
	if err != nil {
		return WrapError(CodeInternal, "无法序列化审计后状态", err)
	}
	entry := model.AuditLog{
		Actor:      actor.DisplayName,
		Role:       actor.Role,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		FromState:  fromState,
		ToState:    toState,
		Reason:     reason,
		RequestID:  actor.RequestID,
	}
	if entry.Actor == "" {
		entry.Actor = actor.Username
	}
	if err := s.repo.Create(&entry); err != nil {
		return WrapError(CodeInternal, "写入审计记录失败", err)
	}
	return nil
}

func (s *AuditService) List(query dto.PageQuery, entityType, action string) (dto.PageResult[model.AuditLog], error) {
	query.Normalize()
	items, total, err := s.repo.List(query, entityType, action)
	if err != nil {
		return dto.PageResult[model.AuditLog]{}, WrapError(CodeInternal, "查询审计记录失败", err)
	}
	return dto.PageResult[model.AuditLog]{Items: items, Total: total, Page: query.Page, PageSize: query.PageSize}, nil
}

func encodeSnapshot(value any) (string, error) {
	if value == nil {
		return "", nil
	}
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
