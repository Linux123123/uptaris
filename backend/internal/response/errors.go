package response

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/uptaris/uptaris/backend/internal/auth"
	"github.com/uptaris/uptaris/backend/internal/users"
	"gorm.io/gorm"
)

func Fail(c *gin.Context, status int, code, msg string) {
	body := ErrorResponse{}
	body.Error.Code = code
	body.Error.Message = msg
	c.AbortWithStatusJSON(status, body)
}

func DatabaseError(c *gin.Context, err error) {
	// Avoid exporting raw database errors containing SQL values or credentials.
	slog.ErrorContext(c.Request.Context(), "database operation failed", "error_type", fmt.Sprintf("%T", err), "route", c.FullPath())
	Fail(c, 503, "service_unavailable", "service temporarily unavailable")
}

func Error(c *gin.Context, err error) {
	if invalid, ok := errors.AsType[interface {
		error
		ValidationMessage() string
	}](err); ok {
		Fail(c, 422, "validation_failed", invalid.ValidationMessage())
		return
	}
	if authentication, ok := errors.AsType[*auth.Error](err); ok {
		status := 401
		switch authentication.Code {
		case "email_exists":
			status = 409
		case "internal_error":
			status = 500
		case "forbidden":
			status = 403
		}
		if status == 401 {
			c.Header("WWW-Authenticate", "Bearer")
		}
		Fail(c, status, authentication.Code, authentication.Message)
		return
	}
	switch {
	case errors.Is(err, users.ErrRole):
		Fail(c, 422, "validation_failed", err.Error())
	case errors.Is(err, users.ErrOwnRole):
		Fail(c, 409, "self_role_change", err.Error())
	case errors.Is(err, users.ErrOwnDelete):
		Fail(c, 409, "self_delete", err.Error())
	case errors.Is(err, gorm.ErrRecordNotFound):
		Fail(c, 404, "not_found", "resource not found")
	default:
		DatabaseError(c, err)
	}
}

func ResourceError(c *gin.Context, err error, resource string) bool {
	if err == nil {
		return true
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		Fail(c, 404, "not_found", resource+" not found")
	} else {
		Error(c, err)
	}
	return false
}
