package request

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/uptaris/uptaris/backend/internal/response"
)

func Valid(c *gin.Context, err error) bool {
	if err != nil {
		if _, ok := err.(*ValidationError); !ok {
			slog.ErrorContext(c.Request.Context(), "validation failed", "error", err)
			response.Fail(c, 500, "internal_error", "request validation unavailable")
			return false
		}
		response.Error(c, err)
		return false
	}
	return true
}
