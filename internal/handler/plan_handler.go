package handler

import (
	"aquaculture-water-feeding-control/backend/internal/dto"
	"aquaculture-water-feeding-control/backend/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PlanHandler struct {
	service *service.PlanService
}

func NewPlanHandler(plans *service.PlanService) *PlanHandler {
	return &PlanHandler{service: plans}
}

func (h *PlanHandler) List(c *gin.Context) {
	query, ok := bindPageQuery(c)
	if !ok {
		return
	}
	pondID, _ := strconv.ParseUint(c.Query("pondId"), 10, 64)
	result, err := h.service.List(query, uint(pondID))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *PlanHandler) Recommendation(c *gin.Context) {
	pondID, err := strconv.ParseUint(c.Query("pondId"), 10, 64)
	if err != nil || pondID == 0 {
		respondError(c, service.NewError(service.CodeValidation, "请选择养殖池"))
		return
	}
	result, err := h.service.Recommendation(uint(pondID), c.Query("weather"))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *PlanHandler) Get(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	result, err := h.service.Get(id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *PlanHandler) Create(c *gin.Context) {
	var input dto.FeedingPlanInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, service.NewError(service.CodeValidation, "投喂计划参数不完整或格式无效"))
		return
	}
	result, err := h.service.Create(input, actorFromContext(c))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": result})
}

func (h *PlanHandler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var input dto.FeedingPlanInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, service.NewError(service.CodeValidation, "投喂计划参数不完整或格式无效"))
		return
	}
	result, err := h.service.Update(id, input, actorFromContext(c))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *PlanHandler) Submit(c *gin.Context) {
	h.executeTransition(c, "submit")
}

func (h *PlanHandler) Approve(c *gin.Context) {
	h.executeTransition(c, "approve")
}

func (h *PlanHandler) Revoke(c *gin.Context) {
	h.executeTransition(c, "revoke")
}

func (h *PlanHandler) executeTransition(c *gin.Context, action string) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var input dto.TransitionPlanInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, service.NewError(service.CodeValidation, "请填写状态变更原因"))
		return
	}
	var result any
	var err error
	switch action {
	case "submit":
		result, err = h.service.Submit(id, input.Reason, actorFromContext(c))
	case "approve":
		result, err = h.service.Approve(id, input.Reason, actorFromContext(c))
	case "revoke":
		result, err = h.service.Revoke(id, input.Reason, actorFromContext(c))
	}
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *PlanHandler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.service.Delete(id, actorFromContext(c)); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
