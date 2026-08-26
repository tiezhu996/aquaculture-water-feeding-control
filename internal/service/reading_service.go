package service

import (
	"aquaculture-water-feeding-control/backend/internal/constants"
	"aquaculture-water-feeding-control/backend/internal/dto"
	"aquaculture-water-feeding-control/backend/internal/model"
	"aquaculture-water-feeding-control/backend/internal/repository"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

type ReadingService struct {
	repo          *repository.ReadingRepository
	ponds         *repository.PondRepository
	audit         *AuditService
	transactional bool
}

func (s *ReadingService) withinTransaction(fn func(*ReadingService) error) error {
	return s.audit.WithinTransaction(func(tx *gorm.DB, audit *AuditService) error {
		scoped := &ReadingService{repo: repository.NewReadingRepository(tx), ponds: repository.NewPondRepository(tx), audit: audit, transactional: true}
		return fn(scoped)
	})
}

func NewReadingService(repo *repository.ReadingRepository, ponds *repository.PondRepository, audit *AuditService) *ReadingService {
	return &ReadingService{repo: repo, ponds: ponds, audit: audit}
}

func (s *ReadingService) List(query dto.PageQuery, pondID uint, unconfirmed bool) (dto.PageResult[model.WaterReading], error) {
	query.Normalize()
	items, total, err := s.repo.List(query, pondID, unconfirmed)
	if err != nil {
		return dto.PageResult[model.WaterReading]{}, WrapError(CodeInternal, "查询水质读数失败", err)
	}
	return dto.PageResult[model.WaterReading]{Items: items, Total: total, Page: query.Page, PageSize: query.PageSize}, nil
}

func (s *ReadingService) Get(id uint) (model.WaterReading, error) {
	var reading model.WaterReading
	var err error
	if s.transactional {
		reading, err = s.repo.GetForUpdate(id)
	} else {
		reading, err = s.repo.Get(id)
	}
	if err == gorm.ErrRecordNotFound {
		return model.WaterReading{}, NewError(CodeNotFound, "水质读数不存在")
	}
	if err != nil {
		return model.WaterReading{}, WrapError(CodeInternal, "查询水质读数失败", err)
	}
	return reading, nil
}

func (s *ReadingService) Create(input dto.WaterReadingInput, actor Actor) (model.WaterReading, error) {
	if !s.transactional {
		var result model.WaterReading
		err := s.withinTransaction(func(scoped *ReadingService) error {
			var inner error
			result, inner = scoped.Create(input, actor)
			return inner
		})
		return result, err
	}
	pond, err := s.ponds.GetForUpdate(input.PondID)
	if err == gorm.ErrRecordNotFound {
		return model.WaterReading{}, NewError(CodeValidation, "养殖池不存在")
	}
	if err != nil {
		return model.WaterReading{}, WrapError(CodeInternal, "查询养殖池失败", err)
	}
	if pond.Status == constants.PondStatusClosed {
		return model.WaterReading{}, NewError(CodeConflict, "已关闭养殖池不能新增读数")
	}
	if input.MeasuredAt.After(time.Now().Add(10 * time.Minute)) {
		return model.WaterReading{}, NewError(CodeValidation, "测量时间不能晚于当前时间")
	}
	risk, message := assessWaterRisk(input)
	reading := model.WaterReading{
		PondID: input.PondID, DissolvedOxygen: input.DissolvedOxygen, Temperature: input.Temperature,
		PH: input.PH, Ammonia: input.Ammonia, Turbidity: input.Turbidity, MeasuredAt: input.MeasuredAt.UTC(),
		Source: input.Source, RiskLevel: risk, AlertMessage: message,
	}
	if err := s.repo.Create(&reading); err != nil {
		return model.WaterReading{}, WrapError(CodeInternal, "创建水质读数失败", err)
	}
	reading.Pond = &pond
	if err := s.audit.Record(actor, "create", "water_reading", reading.ID, nil, reading, message); err != nil {
		return model.WaterReading{}, err
	}
	return reading, nil
}

func (s *ReadingService) Confirm(id uint, note string, actor Actor) (model.WaterReading, error) {
	if !s.transactional {
		var result model.WaterReading
		err := s.withinTransaction(func(scoped *ReadingService) error {
			var inner error
			result, inner = scoped.Confirm(id, note, actor)
			return inner
		})
		return result, err
	}
	reading, err := s.Get(id)
	if err != nil {
		return model.WaterReading{}, err
	}
	if reading.RiskLevel == constants.RiskNormal {
		return model.WaterReading{}, NewError(CodeConflict, "正常读数无需异常确认")
	}
	if reading.Confirmed {
		return model.WaterReading{}, NewError(CodeConflict, "该异常读数已确认")
	}
	if reading.Pond.Status == constants.PondStatusClosed {
		return model.WaterReading{}, NewError(CodeConflict, "已关闭养殖池不能确认异常读数")
	}
	before := reading
	now := time.Now().UTC()
	reading.Confirmed = true
	reading.ConfirmedBy = actor.DisplayName
	reading.ConfirmedAt = &now
	reading.ConfirmationNote = strings.TrimSpace(note)
	if err := s.repo.Save(&reading); err != nil {
		return model.WaterReading{}, WrapError(CodeInternal, "确认异常读数失败", err)
	}
	if err := s.audit.Record(actor, "confirm", "water_reading", reading.ID, before, reading, note); err != nil {
		return model.WaterReading{}, err
	}
	return reading, nil
}

func (s *ReadingService) Delete(id uint, actor Actor) error {
	if !s.transactional {
		return s.withinTransaction(func(scoped *ReadingService) error { return scoped.Delete(id, actor) })
	}
	reading, err := s.Get(id)
	if err != nil {
		return err
	}
	if reading.Source != "manual" {
		return NewError(CodeConflict, "只允许删除手工录入的读数")
	}
	if reading.Confirmed {
		return NewError(CodeConflict, "已确认的异常读数不能删除")
	}
	if reading.Pond.Status == constants.PondStatusClosed {
		return NewError(CodeConflict, "已关闭养殖池不能删除读数")
	}
	if err := s.repo.Delete(&reading); err != nil {
		return WrapError(CodeInternal, "删除水质读数失败", err)
	}
	return s.audit.Record(actor, "delete", "water_reading", reading.ID, reading, nil, "删除手工录入读数")
}

func assessWaterRisk(input dto.WaterReadingInput) (constants.RiskLevel, string) {
	critical := make([]string, 0)
	warnings := make([]string, 0)
	if input.DissolvedOxygen < 3 {
		critical = append(critical, fmt.Sprintf("溶解氧 %.1f mg/L 严重偏低", input.DissolvedOxygen))
	} else if input.DissolvedOxygen < 5 {
		warnings = append(warnings, fmt.Sprintf("溶解氧 %.1f mg/L 偏低", input.DissolvedOxygen))
	}
	if input.PH < 5.5 || input.PH > 10 {
		critical = append(critical, fmt.Sprintf("pH %.1f 超出安全范围", input.PH))
	} else if input.PH < 6.5 || input.PH > 9 {
		warnings = append(warnings, fmt.Sprintf("pH %.1f 偏离建议范围", input.PH))
	}
	if input.Ammonia > 1 {
		critical = append(critical, fmt.Sprintf("氨氮 %.2f mg/L 严重超标", input.Ammonia))
	} else if input.Ammonia > 0.3 {
		warnings = append(warnings, fmt.Sprintf("氨氮 %.2f mg/L 偏高", input.Ammonia))
	}
	if input.Temperature < 10 || input.Temperature > 36 {
		critical = append(critical, fmt.Sprintf("水温 %.1f℃ 超出安全范围", input.Temperature))
	} else if input.Temperature < 18 || input.Temperature > 32 {
		warnings = append(warnings, fmt.Sprintf("水温 %.1f℃ 需关注", input.Temperature))
	}
	if input.Turbidity > 100 {
		warnings = append(warnings, fmt.Sprintf("浊度 %.0f NTU 偏高", input.Turbidity))
	}
	if len(critical) > 0 {
		return constants.RiskCritical, strings.Join(append(critical, warnings...), "；")
	}
	if len(warnings) > 0 {
		return constants.RiskWarning, strings.Join(warnings, "；")
	}
	return constants.RiskNormal, "各项水质指标在控制范围内"
}
