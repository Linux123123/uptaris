package response

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/uptaris/uptaris/backend/internal/accounts"
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
	if accountErr, ok := errors.AsType[*accounts.Error](err); ok {
		status := 422
		if accountErr.Code == "password_processing_failed" {
			status = 500
		} else if accountErr.Code == "invalid_current_password" {
			status = 409
		}

		Fail(c, status, accountErr.Code, accountErr.Message)
		return
	}

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
		case "email_exists", "oauth_identity_in_use", "oauth_provider_already_linked", "last_signin_method":
			status = 409
		case "two_factor_unavailable", "oauth_provider_unavailable":
			status = 503
		case "internal_error":
			status = 500
		case "two_factor_locked":
			status = 429
		case "invalid_two_factor_code", "invalid_two_factor_challenge", "invalid_enrollment", "invalid_passkey_ceremony", "invalid_passkey_response":
			status = 422
		case "two_factor_enabled", "two_factor_disabled", "invalid_current_password":
			status = 409
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
