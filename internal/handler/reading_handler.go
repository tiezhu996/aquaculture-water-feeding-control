package handler

import (
	"aquaculture-water-feeding-control/backend/internal/dto"
	"aquaculture-water-feeding-control/backend/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ReadingHandler struct {
	service *service.ReadingService
}

func NewReadingHandler(readings *service.ReadingService) *ReadingHandler {
	return &ReadingHandler{service: readings}
}

func (h *ReadingHandler) List(c *gin.Context) {
	query, ok := bindPageQuery(c)
	if !ok {
		return
	}
	pondID, _ := strconv.ParseUint(c.Query("pondId"), 10, 64)
	unconfirmed, _ := strconv.ParseBool(c.Query("unconfirmed"))
	result, err := h.service.List(query, uint(pondID), unconfirmed)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *ReadingHandler) Get(c *gin.Context) {
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

func (h *ReadingHandler) Create(c *gin.Context) {
	var input dto.WaterReadingInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, service.NewError(service.CodeValidation, "水质读数参数不完整或超出合理范围"))
		return
	}
	result, err := h.service.Create(input, actorFromContext(c))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": result})
}

func (h *ReadingHandler) Confirm(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var input dto.ConfirmReadingInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, service.NewError(service.CodeValidation, "请填写异常确认说明"))
		return
	}
	result, err := h.service.Confirm(id, input.Note, actorFromContext(c))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *ReadingHandler) Delete(c *gin.Context) {
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
