package service

import (
	"aquaculture-water-feeding-control/backend/internal/constants"
	"aquaculture-water-feeding-control/backend/internal/dto"
	"aquaculture-water-feeding-control/backend/internal/model"
	"aquaculture-water-feeding-control/backend/internal/repository"
	"fmt"
	"math"
	"strings"
	"time"

	"gorm.io/gorm"
)

type PlanService struct {
	repo          *repository.PlanRepository
	ponds         *repository.PondRepository
	readings      *repository.ReadingRepository
	audit         *AuditService
	transactional bool
}

func (s *PlanService) withinTransaction(fn func(*PlanService) error) error {
	return s.audit.WithinTransaction(func(tx *gorm.DB, audit *AuditService) error {
		scoped := &PlanService{
			repo: repository.NewPlanRepository(tx), ponds: repository.NewPondRepository(tx),
			readings: repository.NewReadingRepository(tx), audit: audit, transactional: true,
		}
		return fn(scoped)
	})
}

func (s *PlanService) Recommendation(pondID uint, weather string) (dto.FeedingRecommendation, error) {
	pond, err := s.ponds.Get(pondID)
	if err == gorm.ErrRecordNotFound {
		return dto.FeedingRecommendation{}, NewError(CodeNotFound, "养殖池不存在")
	}
	if err != nil {
		return dto.FeedingRecommendation{}, WrapError(CodeInternal, "查询养殖池失败", err)
	}
	if pond.Status != constants.PondStatusActive {
		return dto.FeedingRecommendation{}, NewError(CodeConflict, "只能为运行中养殖池生成投喂建议")
	}
	plan, err := s.repo.LatestApprovedForPond(pondID)
	if err == gorm.ErrRecordNotFound {
		return dto.FeedingRecommendation{}, NewError(CodeConflict, "当前养殖池没有已批准的投喂计划")
	}
	if err != nil {
		return dto.FeedingRecommendation{}, WrapError(CodeInternal, "查询已批准计划失败", err)
	}
	reading, err := s.readings.LatestForPond(pondID)
	if err == gorm.ErrRecordNotFound {
		return dto.FeedingRecommendation{}, NewError(CodeConflict, "生成建议前必须有水质读数")
	}
	if err != nil {
		return dto.FeedingRecommendation{}, WrapError(CodeInternal, "查询最新水质读数失败", err)
	}
	if time.Since(reading.MeasuredAt) > 24*time.Hour {
		return dto.FeedingRecommendation{}, NewError(CodeConflict, "最新水质读数已超过 24 小时，请先采集新读数")
	}

	factor := 1.0
	reasons := make([]string, 0, 16)
	reasons = append(reasons, "以已批准计划 v"+fmt.Sprint(plan.Version)+" 为基准")
	action := "feed"
	notes := make([]string, 0, 8)
	notes = append(notes, fmt.Sprintf("养殖池 %d 建议基于最新水质生成", pondID))
	if reading.RiskLevel == constants.RiskCritical || reading.DissolvedOxygen < plan.MinOxygen {
		factor = 0
		action = "hold"
		reasons = append(reasons, "水质严重异常或溶解氧低于计划阈值，暂停投喂")
	} else {
		if reading.RiskLevel == constants.RiskWarning {
			factor *= 0.7
			action = "reduce"
			reasons = append(reasons, "存在水质预警，建议减量 30%")
		}
		if reading.Temperature < 20 || reading.Temperature > 31 {
			factor *= 0.8
			action = "reduce"
			reasons = append(reasons, "水温不在最佳摄食区间，追加减量 20%")
		}
		weatherText := strings.ToLower(strings.TrimSpace(weather))
		if strings.Contains(weatherText, "storm") || strings.Contains(weather, "暴雨") || strings.Contains(weather, "雷雨") {
			factor = 0
			action = "hold"
			reasons = append(reasons, "强对流天气窗口不适合投喂")
		} else if strings.Contains(weatherText, "rain") || strings.Contains(weather, "小雨") || strings.Contains(weather, "大风") {
			factor *= 0.85
			action = "reduce"
			reasons = append(reasons, "天气窗口不稳定，追加减量 15%")
		}
		if strings.Contains(pond.GrowthStage, "幼") {
			factor *= 0.9
			reasons = append(reasons, "幼体阶段采用少量多餐系数")
		}
	}
	dailyAmount := math.Round(plan.DailyAmountKg*factor*100) / 100
	perFeeding := 0.0
	if plan.FrequencyPerDay > 0 {
		perFeeding = math.Round(dailyAmount/float64(plan.FrequencyPerDay)*100) / 100
	}
	hints := make([]string, 0, 8)
	if action == "reduce" {
		hints = append(hints, "建议分批少量投喂")
	} else if action == "hold" {
		hints = append(hints, "暂停投喂")
	}
	return dto.FeedingRecommendation{
		PondID: pondID, PlanID: plan.ID, PlanVersion: plan.Version, GeneratedAt: time.Now().UTC(),
		ReadingMeasuredAt: reading.MeasuredAt, Weather: weather, Action: action, DailyAmountKg: dailyAmount,
		AmountPerFeedingKg: perFeeding, FrequencyPerDay: plan.FrequencyPerDay,
		AdjustmentPercent: math.Round((factor-1)*10000) / 100, Reasons: reasons, Hints: hints, Notes: notes,
	}, nil
}

func NewPlanService(repo *repository.PlanRepository, ponds *repository.PondRepository, readings *repository.ReadingRepository, audit *AuditService) *PlanService {
	return &PlanService{repo: repo, ponds: ponds, readings: readings, audit: audit}
}

func (s *PlanService) List(query dto.PageQuery, pondID uint) (dto.PageResult[model.FeedingPlan], error) {
	query.Normalize()
	items, total, err := s.repo.List(query, pondID)
	if err != nil {
		return dto.PageResult[model.FeedingPlan]{}, WrapError(CodeInternal, "查询投喂计划失败", err)
	}
	return dto.PageResult[model.FeedingPlan]{Items: items, Total: total, Page: query.Page, PageSize: query.PageSize}, nil
}

func (s *PlanService) Get(id uint) (model.FeedingPlan, error) {
	var plan model.FeedingPlan
	var err error
	if s.transactional {
		plan, err = s.repo.GetForUpdate(id)
	} else {
		plan, err = s.repo.Get(id)
	}
	if err == gorm.ErrRecordNotFound {
		return model.FeedingPlan{}, NewError(CodeNotFound, "投喂计划不存在")
	}
	if err != nil {
		return model.FeedingPlan{}, WrapError(CodeInternal, "查询投喂计划失败", err)
	}
	return plan, nil
}

func (s *PlanService) Create(input dto.FeedingPlanInput, actor Actor) (model.FeedingPlan, error) {
	if !s.transactional {
		var result model.FeedingPlan
		err := s.withinTransaction(func(scoped *PlanService) error {
			var inner error
			result, inner = scoped.Create(input, actor)
			return inner
		})
		return result, err
	}
	pond, err := s.validateInput(input)
	if err != nil {
		return model.FeedingPlan{}, err
	}
	plan := model.FeedingPlan{
		PondID: input.PondID, Name: input.Name, Version: 1, DailyAmountKg: input.DailyAmountKg,
		FrequencyPerDay: input.FrequencyPerDay, FeedType: input.FeedType, TargetGrowthStage: input.TargetGrowthStage,
		MinOxygen: input.MinOxygen, StartDate: input.StartDate.UTC(), EndDate: input.EndDate.UTC(),
		Status: constants.PlanStatusDraft, Rationale: input.Rationale, CreatedBy: actor.DisplayName,
	}
	if plan.CreatedBy == "" {
		plan.CreatedBy = actor.Username
	}
	if err := s.repo.Create(&plan); err != nil {
		return model.FeedingPlan{}, WrapError(CodeInternal, "创建投喂计划失败", err)
	}
	plan.Pond = &pond
	if err := s.audit.Record(actor, "create", "feeding_plan", plan.ID, nil, plan, input.Rationale); err != nil {
		return model.FeedingPlan{}, err
	}
	return plan, nil
}

func (s *PlanService) Update(id uint, input dto.FeedingPlanInput, actor Actor) (model.FeedingPlan, error) {
	if !s.transactional {
		var result model.FeedingPlan
		err := s.withinTransaction(func(scoped *PlanService) error {
			var inner error
			result, inner = scoped.Update(id, input, actor)
			return inner
		})
		return result, err
	}
	plan, err := s.Get(id)
	if err != nil {
		return model.FeedingPlan{}, err
	}
	if plan.Status != constants.PlanStatusDraft {
		return model.FeedingPlan{}, NewError(CodeConflict, "只有草稿计划可以编辑")
	}
	pond, err := s.validateInput(input)
	if err != nil {
		return model.FeedingPlan{}, err
	}
	before := plan
	plan.PondID = input.PondID
	plan.Name = input.Name
	plan.Version++
	plan.DailyAmountKg = input.DailyAmountKg
	plan.FrequencyPerDay = input.FrequencyPerDay
	plan.FeedType = input.FeedType
	plan.TargetGrowthStage = input.TargetGrowthStage
	plan.MinOxygen = input.MinOxygen
	plan.StartDate = input.StartDate.UTC()
	plan.EndDate = input.EndDate.UTC()
	plan.Rationale = input.Rationale
	if err := s.repo.Save(&plan); err != nil {
		return model.FeedingPlan{}, WrapError(CodeInternal, "更新投喂计划失败", err)
	}
	plan.Pond = &pond
	if err := s.audit.Record(actor, "revise", "feeding_plan", plan.ID, before, plan, input.Rationale); err != nil {
		return model.FeedingPlan{}, err
	}
	return plan, nil
}

func (s *PlanService) Submit(id uint, reason string, actor Actor) (model.FeedingPlan, error) {
	return s.transition(id, constants.PlanStatusDraft, constants.PlanStatusPending, "submit", reason, actor)
}

func (s *PlanService) Approve(id uint, reason string, actor Actor) (model.FeedingPlan, error) {
	if !s.transactional {
		var result model.FeedingPlan
		err := s.withinTransaction(func(scoped *PlanService) error {
			var inner error
			result, inner = scoped.Approve(id, reason, actor)
			return inner
		})
		return result, err
	}
	plan, err := s.Get(id)
	if err != nil {
		return model.FeedingPlan{}, err
	}
	if plan.Status != constants.PlanStatusPending {
		return model.FeedingPlan{}, NewError(CodeConflict, "只有待审核计划可以批准")
	}
	pond, err := s.ponds.GetForUpdate(plan.PondID)
	if err != nil {
		return model.FeedingPlan{}, WrapError(CodeInternal, "查询养殖池失败", err)
	}
	if pond.Status != constants.PondStatusActive {
		return model.FeedingPlan{}, NewError(CodeConflict, "只有运行中养殖池的计划可批准")
	}
	latest, err := s.readings.LatestForPond(plan.PondID)
	if err == gorm.ErrRecordNotFound {
		return model.FeedingPlan{}, NewError(CodeConflict, "批准前必须有最新水质读数")
	}
	if err != nil {
		return model.FeedingPlan{}, WrapError(CodeInternal, "查询最新水质读数失败", err)
	}
	if time.Since(latest.MeasuredAt) > 24*time.Hour {
		return model.FeedingPlan{}, NewError(CodeConflict, "最新水质读数已超过 24 小时，不能批准")
	}
	if latest.DissolvedOxygen < plan.MinOxygen {
		return model.FeedingPlan{}, NewError(CodeConflict, "最新溶解氧低于计划阈值，不能批准")
	}
	if latest.RiskLevel == constants.RiskCritical {
		return model.FeedingPlan{}, NewError(CodeConflict, "存在严重水质异常，确认留痕后仍不可批准投喂")
	}
	if latest.RiskLevel != constants.RiskNormal && !latest.Confirmed {
		return model.FeedingPlan{}, NewError(CodeConflict, "存在未确认的水质异常")
	}
	before := plan
	now := time.Now().UTC()
	plan.Status = constants.PlanStatusApproved
	plan.ReviewedBy = actor.DisplayName
	plan.ReviewedAt = &now
	if err := s.repo.Save(&plan); err != nil {
		return model.FeedingPlan{}, WrapError(CodeInternal, "批准投喂计划失败", err)
	}
	return plan, s.audit.Record(actor, "approve", "feeding_plan", plan.ID, before, plan, reason)
}

func (s *PlanService) Revoke(id uint, reason string, actor Actor) (model.FeedingPlan, error) {
	if !s.transactional {
		var result model.FeedingPlan
		err := s.withinTransaction(func(scoped *PlanService) error {
			var inner error
			result, inner = scoped.Revoke(id, reason, actor)
			return inner
		})
		return result, err
	}
	plan, err := s.Get(id)
	if err != nil {
		return model.FeedingPlan{}, err
	}
	if plan.Status != constants.PlanStatusPending && plan.Status != constants.PlanStatusApproved {
		return model.FeedingPlan{}, NewError(CodeConflict, "只有待审核或已批准计划可撤销")
	}
	count, err := s.repo.ExecutionCount(id)
	if err != nil {
		return model.FeedingPlan{}, WrapError(CodeInternal, "检查执行记录失败", err)
	}
	if count > 0 {
		return model.FeedingPlan{}, NewError(CodeConflict, "计划已生成执行记录，不能撤销")
	}
	before := plan
	plan.Status = constants.PlanStatusDraft
	plan.Version++
	plan.ReviewedBy = ""
	plan.ReviewedAt = nil
	if err := s.repo.Save(&plan); err != nil {
		return model.FeedingPlan{}, WrapError(CodeInternal, "撤销投喂计划失败", err)
	}
	return plan, s.audit.Record(actor, "revoke", "feeding_plan", plan.ID, before, plan, reason)
}

func (s *PlanService) Delete(id uint, actor Actor) error {
	if !s.transactional {
		return s.withinTransaction(func(scoped *PlanService) error { return scoped.Delete(id, actor) })
	}
	plan, err := s.Get(id)
	if err != nil {
		return err
	}
	if plan.Status != constants.PlanStatusDraft {
		return NewError(CodeConflict, "只有草稿计划可以删除")
	}
	count, err := s.repo.ExecutionCount(id)
	if err != nil {
		return WrapError(CodeInternal, "检查执行记录失败", err)
	}
	if count > 0 {
		return NewError(CodeConflict, "计划已有执行记录，不能删除")
	}
	if err := s.repo.Delete(&plan); err != nil {
		return WrapError(CodeInternal, "删除投喂计划失败", err)
	}
	return s.audit.Record(actor, "delete", "feeding_plan", plan.ID, plan, nil, "删除草稿计划")
}

func (s *PlanService) transition(id uint, from, to constants.PlanStatus, action, reason string, actor Actor) (model.FeedingPlan, error) {
	if !s.transactional {
		var result model.FeedingPlan
		err := s.withinTransaction(func(scoped *PlanService) error {
			var inner error
			result, inner = scoped.transition(id, from, to, action, reason, actor)
			return inner
		})
		return result, err
	}
	plan, err := s.Get(id)
	if err != nil {
		return model.FeedingPlan{}, err
	}
	if plan.Status != from {
		return model.FeedingPlan{}, NewError(CodeConflict, "计划当前状态不允许此操作")
	}
	before := plan
	plan.Status = to
	if err := s.repo.Save(&plan); err != nil {
		return model.FeedingPlan{}, WrapError(CodeInternal, "更新投喂计划状态失败", err)
	}
	return plan, s.audit.Record(actor, action, "feeding_plan", plan.ID, before, plan, reason)
}

func (s *PlanService) validateInput(input dto.FeedingPlanInput) (model.Pond, error) {
	if !input.EndDate.After(input.StartDate) {
		return model.Pond{}, NewError(CodeValidation, "计划结束日期必须晚于开始日期")
	}
	if input.EndDate.Sub(input.StartDate) > 180*24*time.Hour {
		return model.Pond{}, NewError(CodeValidation, "单个计划周期不能超过 180 天")
	}
	var pond model.Pond
	var err error
	if s.transactional {
		pond, err = s.ponds.GetForUpdate(input.PondID)
	} else {
		pond, err = s.ponds.Get(input.PondID)
	}
	if err == gorm.ErrRecordNotFound {
		return model.Pond{}, NewError(CodeValidation, "养殖池不存在")
	}
	if err != nil {
		return model.Pond{}, WrapError(CodeInternal, "查询养殖池失败", err)
	}
	if pond.Status == constants.PondStatusClosed {
		return model.Pond{}, NewError(CodeConflict, "已关闭养殖池不能创建计划")
	}
	if input.DailyAmountKg > pond.CapacityKg*0.12 {
		return model.Pond{}, NewError(CodeValidation, "日投喂量不能超过养殖容量的 12%")
	}
	return pond, nil
}
