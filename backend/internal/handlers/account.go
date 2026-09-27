package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/uptaris/uptaris/backend/internal/middleware"
	"github.com/uptaris/uptaris/backend/internal/request"
	"github.com/uptaris/uptaris/backend/internal/response"
)

// @Summary Get account sign-in settings
// @Tags account
// @Security bearerauth
// @Success 200 {object} response.AccountSettingsResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /account [get]
func (h *Handlers) AccountSettings(c *gin.Context) {
	identity, ok := middleware.Require(c, h.auth, "viewer", "operator", "admin")
	if !ok {
		return
	}

	settings, err := h.accounts.Settings(c.Request.Context(), identity.ID, h.auth.OAuthProviders())
	if err != nil {
		response.Error(c, err)
		return
	}

	c.JSON(http.StatusOK, settings)
}

// @Summary Set or change account password
// @Description Current password required when account already has password. Setting first password adds another sign-in method.
// @Tags account
// @Security bearerauth
// @Accept json
// @Param body body request.PasswordChange true "Current password and new password"
// @Success 204
// @Failure 422 {object} response.ErrorResponse "validation_failed: password policy requirements not met"
// @Failure 409 {object} response.ErrorResponse "invalid_current_password"
// @Router /account/password [put]
func (h *Handlers) SetPassword(c *gin.Context) {
	identity, ok := middleware.Require(c, h.auth, "viewer", "operator", "admin")
	if !ok {
		return
	}

	var input request.PasswordChange

	if !request.JSON(c, &input) || !request.Valid(c, request.Validate(&input)) {
		return
	}

	if err := h.accounts.SetPassword(c.Request.Context(), identity.ID, identity.SessionID, input.CurrentPassword, input.NewPassword); err != nil {
		response.Error(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
