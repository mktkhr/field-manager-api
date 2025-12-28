// Package middleware はHTTPミドルウェアを提供する
package middleware

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/mktkhr/field-manager-api/internal/logger"
)

// NewRequestLoggerFromGin はGin ContextからRequestLoggerを作成する
func NewRequestLoggerFromGin(c *gin.Context) *slog.Logger {
	requestID := ""
	if id, exists := c.Get("requestID"); exists {
		if reqID, ok := id.(string); ok {
			requestID = reqID
		}
	}

	return logger.NewRequestLogger(
		requestID,
		c.Request.UserAgent(),
		c.ClientIP(),
	)
}
