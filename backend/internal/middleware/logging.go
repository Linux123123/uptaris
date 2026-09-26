package middleware

import (
	"io"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uptaris/uptaris/backend/internal/response"
)

func RequestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		started := time.Now()
		c.Next()
		logger.InfoContext(c.Request.Context(), "http request", "method", c.Request.Method, "route", c.FullPath(), "status", c.Writer.Status(), "duration", time.Since(started))
	}
}

func Recovery() gin.HandlerFunc {
	return gin.CustomRecoveryWithWriter(io.Discard, func(c *gin.Context, _ any) { response.Fail(c, 500, "internal_error", "unexpected server error") })
}
