package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequestIDRejectsLogInjectionCharacters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID())
	router.GET("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", "unsafe request id")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	got := response.Header().Get("X-Request-ID")
	if got == "" || got == req.Header.Get("X-Request-ID") || !safeRequestID.MatchString(got) {
		t.Fatalf("unsafe request ID was not replaced: %q", got)
	}
}
