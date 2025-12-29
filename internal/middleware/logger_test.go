package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestNewRequestLoggerFromGin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("リクエストIDが設定されている場合", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
		c.Request.Header.Set("User-Agent", "TestAgent/1.0")
		c.Set("requestID", "test-request-id-123")

		logger := NewRequestLoggerFromGin(c)
		if logger == nil {
			t.Error("NewRequestLoggerFromGin()はnilを返すべきではない")
		}
	})

	t.Run("リクエストIDが設定されていない場合", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

		logger := NewRequestLoggerFromGin(c)
		if logger == nil {
			t.Error("NewRequestLoggerFromGin()はnilを返すべきではない")
		}
	})

	t.Run("リクエストIDが文字列以外の場合", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
		c.Set("requestID", 12345) // 文字列ではなく整数

		logger := NewRequestLoggerFromGin(c)
		if logger == nil {
			t.Error("NewRequestLoggerFromGin()はnilを返すべきではない")
		}
	})

	t.Run("UserAgentとClientIPが取得される", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
		c.Request.Header.Set("User-Agent", "Mozilla/5.0")
		c.Request.Header.Set("X-Forwarded-For", "192.168.1.100")
		c.Set("requestID", "ua-test-id")

		logger := NewRequestLoggerFromGin(c)
		if logger == nil {
			t.Error("NewRequestLoggerFromGin()はnilを返すべきではない")
		}
	})
}
