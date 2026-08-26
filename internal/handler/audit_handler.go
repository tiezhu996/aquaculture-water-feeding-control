package handler

import (
	"aquaculture-water-feeding-control/backend/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuditHandler struct {
	service *service.AuditService
}

func NewAuditHandler(audit *service.AuditService) *AuditHandler {
	return &AuditHandler{service: audit}
}

func (h *AuditHandler) List(c *gin.Context) {
	query, ok := bindPageQuery(c)
	if !ok {
		return
	}
	result, err := h.service.List(query, c.Query("entityType"), c.Query("action"))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}
