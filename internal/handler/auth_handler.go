package handler

import (
	"aquaculture-water-feeding-control/backend/internal/dto"
	"aquaculture-water-feeding-control/backend/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{service: auth}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var input dto.LoginRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, service.NewError(service.CodeValidation, "请输入有效的用户名和密码"))
		return
	}
	result, err := h.service.Login(input)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *AuthHandler) Me(c *gin.Context) {
	actor := actorFromContext(c)
	result, err := h.service.Me(actor.UserID)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func actorFromContext(c *gin.Context) service.Actor {
	userID, _ := c.Get("userID")
	username, _ := c.Get("username")
	displayName, _ := c.Get("displayName")
	role, _ := c.Get("role")
	requestID, _ := c.Get("requestID")
	actor := service.Actor{}
	if value, ok := userID.(uint); ok {
		actor.UserID = value
	}
	actor.Username, _ = username.(string)
	actor.DisplayName, _ = displayName.(string)
	actor.Role, _ = role.(string)
	actor.RequestID, _ = requestID.(string)
	return actor
}

func parseID(c *gin.Context) (uint, bool) {
	value, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || value == 0 {
		respondError(c, service.NewError(service.CodeValidation, "资源 ID 无效"))
		return 0, false
	}
	return uint(value), true
}

func respondError(c *gin.Context, err error) {
	requestID, _ := c.Get("requestID")
	requestIDText, _ := requestID.(string)
	status := http.StatusInternalServerError
	code := service.CodeInternal
	message := "服务器内部错误"
	if appErr, ok := service.AsAppError(err); ok {
		code = appErr.Code
		message = appErr.Message
		switch appErr.Code {
		case service.CodeValidation:
			status = http.StatusBadRequest
		case service.CodeUnauthorized:
			status = http.StatusUnauthorized
		case service.CodeForbidden:
			status = http.StatusForbidden
		case service.CodeNotFound:
			status = http.StatusNotFound
		case service.CodeConflict:
			status = http.StatusConflict
		}
	}
	c.JSON(status, gin.H{"error": gin.H{"code": code, "message": message, "requestId": requestIDText}})
}

func bindPageQuery(c *gin.Context) (dto.PageQuery, bool) {
	var query dto.PageQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		respondError(c, service.NewError(service.CodeValidation, "分页查询参数无效"))
		return dto.PageQuery{}, false
	}
	return query, true
}
