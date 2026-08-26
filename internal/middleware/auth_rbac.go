package middleware

import (
	"aquaculture-water-feeding-control/backend/internal/service"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthRequired(auth *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := strings.TrimSpace(c.GetHeader("Authorization"))
		if !strings.HasPrefix(header, "Bearer ") {
			abortAuth(c, http.StatusUnauthorized, "UNAUTHORIZED", "缺少访问令牌")
			return
		}
		raw := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
		claims, err := auth.ParseToken(raw)
		if err != nil {
			message := "访问令牌无效或已过期"
			if appErr, ok := service.AsAppError(err); ok {
				message = appErr.Message
			}
			abortAuth(c, http.StatusUnauthorized, "UNAUTHORIZED", message)
			return
		}
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("displayName", claims.DisplayName)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func RequireRoles(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		roleText, _ := role.(string)
		if _, ok := allowed[roleText]; !ok {
			abortAuth(c, http.StatusForbidden, "FORBIDDEN", "当前角色无权执行此操作")
			return
		}
		c.Next()
	}
}

func abortAuth(c *gin.Context, status int, code, message string) {
	requestID, _ := c.Get("requestID")
	c.AbortWithStatusJSON(status, gin.H{
		"error": gin.H{"code": code, "message": message, "requestId": requestID},
	})
}
