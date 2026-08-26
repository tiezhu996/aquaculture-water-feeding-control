package handler

import (
	"aquaculture-water-feeding-control/backend/internal/dto"
	"aquaculture-water-feeding-control/backend/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ExecutionHandler struct {
	service *service.ExecutionService
}

func NewExecutionHandler(executions *service.ExecutionService) *ExecutionHandler {
	return &ExecutionHandler{service: executions}
}

func (h *ExecutionHandler) List(c *gin.Context) {
	query, ok := bindPageQuery(c)
	if !ok {
		return
	}
	pondID, _ := strconv.ParseUint(c.Query("pondId"), 10, 64)
	planID, _ := strconv.ParseUint(c.Query("planId"), 10, 64)
	result, err := h.service.List(query, uint(pondID), uint(planID))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *ExecutionHandler) Get(c *gin.Context) {
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

func (h *ExecutionHandler) Create(c *gin.Context) {
	var input dto.ExecutionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, service.NewError(service.CodeValidation, "执行安排参数不完整或格式无效"))
		return
	}
	result, err := h.service.Create(input, actorFromContext(c))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": result})
}

func (h *ExecutionHandler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var input dto.UpdateExecutionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, service.NewError(service.CodeValidation, "执行记录参数不完整或格式无效"))
		return
	}
	result, err := h.service.Update(id, input, actorFromContext(c))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *ExecutionHandler) Complete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var input dto.CompleteExecutionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, service.NewError(service.CodeValidation, "执行反馈参数不完整或格式无效"))
		return
	}
	result, err := h.service.Complete(id, input, actorFromContext(c))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *ExecutionHandler) Delete(c *gin.Context) {
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
