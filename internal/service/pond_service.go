package service

import (
	"aquaculture-water-feeding-control/backend/internal/dto"
	"aquaculture-water-feeding-control/backend/internal/model"
	"aquaculture-water-feeding-control/backend/internal/repository"
	"strings"

	"gorm.io/gorm"
)

type PondService struct {
	repo          *repository.PondRepository
	audit         *AuditService
	transactional bool
}

func (s *PondService) withinTransaction(fn func(*PondService) error) error {
	return s.audit.WithinTransaction(func(tx *gorm.DB, audit *AuditService) error {
		return fn(&PondService{repo: repository.NewPondRepository(tx), audit: audit, transactional: true})
	})
}

func NewPondService(repo *repository.PondRepository, audit *AuditService) *PondService {
	return &PondService{repo: repo, audit: audit}
}

func (s *PondService) List(query dto.PageQuery) (dto.PageResult[model.Pond], error) {
	query.Normalize()
	items, total, err := s.repo.List(query)
	if err != nil {
		return dto.PageResult[model.Pond]{}, WrapError(CodeInternal, "查询养殖池失败", err)
	}
	return dto.PageResult[model.Pond]{Items: items, Total: total, Page: query.Page, PageSize: query.PageSize}, nil
}

func (s *PondService) Get(id uint) (model.Pond, error) {
	var pond model.Pond
	var err error
	if s.transactional {
		pond, err = s.repo.GetForUpdate(id)
	} else {
		pond, err = s.repo.Get(id)
	}
	if err == gorm.ErrRecordNotFound {
		return model.Pond{}, NewError(CodeNotFound, "养殖池不存在")
	}
	if err != nil {
		return model.Pond{}, WrapError(CodeInternal, "查询养殖池失败", err)
	}
	return pond, nil
}

func (s *PondService) Create(input dto.PondInput, actor Actor) (model.Pond, error) {
	if !s.transactional {
		var result model.Pond
		err := s.withinTransaction(func(scoped *PondService) error {
			var inner error
			result, inner = scoped.Create(input, actor)
			return inner
		})
		return result, err
	}
	if !input.Status.Valid() {
		return model.Pond{}, NewError(CodeValidation, "养殖池状态无效")
	}
	pond := model.Pond{
		Code: strings.ToUpper(strings.TrimSpace(input.Code)), Name: strings.TrimSpace(input.Name), Species: strings.TrimSpace(input.Species),
		AreaSquareMeters: input.AreaSquareMeters, CapacityKg: input.CapacityKg, GrowthStage: input.GrowthStage,
		Status: input.Status, Manager: input.Manager, Notes: input.Notes,
	}
	if err := s.repo.Create(&pond); err != nil {
		if IsDuplicatedKey(err) {
			return model.Pond{}, NewError(CodeConflict, "养殖池编码已存在")
		}
		return model.Pond{}, WrapError(CodeInternal, "创建养殖池失败", err)
	}
	if err := s.audit.Record(actor, "create", "pond", pond.ID, nil, pond, "创建养殖池"); err != nil {
		return model.Pond{}, err
	}
	return pond, nil
}

func (s *PondService) Update(id uint, input dto.PondInput, actor Actor) (model.Pond, error) {
	if !s.transactional {
		var result model.Pond
		err := s.withinTransaction(func(scoped *PondService) error {
			var inner error
			result, inner = scoped.Update(id, input, actor)
			return inner
		})
		return result, err
	}
	pond, err := s.Get(id)
	if err != nil {
		return model.Pond{}, err
	}
	if !input.Status.Valid() {
		return model.Pond{}, NewError(CodeValidation, "养殖池状态无效")
	}
	before := pond
	pond.Code = strings.ToUpper(strings.TrimSpace(input.Code))
	pond.Name = strings.TrimSpace(input.Name)
	pond.Species = strings.TrimSpace(input.Species)
	pond.AreaSquareMeters = input.AreaSquareMeters
	pond.CapacityKg = input.CapacityKg
	pond.GrowthStage = input.GrowthStage
	pond.Status = input.Status
	pond.Manager = input.Manager
	pond.Notes = input.Notes
	if err := s.repo.Save(&pond); err != nil {
		if IsDuplicatedKey(err) {
			return model.Pond{}, NewError(CodeConflict, "养殖池编码已存在")
		}
		return model.Pond{}, WrapError(CodeInternal, "更新养殖池失败", err)
	}
	if err := s.audit.Record(actor, "update", "pond", pond.ID, before, pond, "更新基础信息或状态"); err != nil {
		return model.Pond{}, err
	}
	return pond, nil
}

func (s *PondService) Delete(id uint, actor Actor) error {
	if !s.transactional {
		return s.withinTransaction(func(scoped *PondService) error { return scoped.Delete(id, actor) })
	}
	pond, err := s.Get(id)
	if err != nil {
		return err
	}
	dependencies, err := s.repo.DependencyCount(id)
	if err != nil {
		return WrapError(CodeInternal, "检查养殖池关联数据失败", err)
	}
	if dependencies > 0 {
		return NewError(CodeConflict, "养殖池已有水质、计划或执行记录，不能删除")
	}
	if err := s.repo.Delete(&pond); err != nil {
		return WrapError(CodeInternal, "删除养殖池失败", err)
	}
	return s.audit.Record(actor, "delete", "pond", pond.ID, pond, nil, "删除无关联数据的养殖池")
}
