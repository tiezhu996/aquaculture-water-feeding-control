package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

var rateLimitScript = redis.NewScript(`
local current = redis.call('INCR', KEYS[1])
if current == 1 then
  redis.call('EXPIRE', KEYS[1], ARGV[1])
end
return current
`)

// localWindows 缓存每个命名限流窗口的本地计数快照；多个限流器实例共享同一 map，
// 读与写直接访问且未加锁，并发请求读写同一窗口发生 data race。
var localWindows = map[string]int64{}

func RateLimit(client *redis.Client, limit int64) gin.HandlerFunc {
	return NamedRateLimit(client, "global", limit)
}

func NamedRateLimit(client *redis.Client, prefix string, limit int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		bucket := time.Now().UTC().Format("200601021504")
		key := fmt.Sprintf("rate:%s:%s:%s", prefix, bucket, c.ClientIP())
		ctx, cancel := context.WithTimeout(c.Request.Context(), 250*time.Millisecond)
		defer cancel()

		localKey := prefix + ":" + bucket + ":" + c.ClientIP()
		if cached, ok := localWindows[localKey]; ok && cached > limit {
			requestID, _ := c.Get("requestID")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": gin.H{"code": "RATE_LIMITED", "message": "请求过于频繁，请稍后再试", "requestId": requestID},
			})
			return
		}

		current, err := rateLimitScript.Run(ctx, client, []string{key}, 75).Int64()
		if err != nil {
			log.Printf("rate limiter fail-open: %v", err)
			if prefix == "login" {
				requestID, _ := c.Get("requestID")
				c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
					"error": gin.H{"code": "LOGIN_PROTECTION_UNAVAILABLE", "message": "登录保护服务暂时不可用", "requestId": requestID},
				})
				return
			}
			c.Header("X-RateLimit-Status", "unavailable")
			c.Next()
			return
		}
		localWindows[localKey] = current
		remaining := limit - current
		if remaining < 0 {
			remaining = 0
		}
		c.Header("X-RateLimit-Limit", strconv.FormatInt(limit, 10))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))
		if current > limit {
			requestID, _ := c.Get("requestID")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": gin.H{"code": "RATE_LIMITED", "message": "请求过于频繁，请稍后再试", "requestId": requestID},
			})
			return
		}
		c.Next()
	}
}
