package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/uptaris/uptaris/backend/internal/middleware"
	"github.com/uptaris/uptaris/backend/internal/request"
	"github.com/uptaris/uptaris/backend/internal/response"
)

func (h *Handlers) challengeCookie(c *gin.Context, value string, maxAge int) {
	// Use the session cookie policy, including cross-site frontend/API deployments.
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "uptaris_two_factor",
		Value:    value,
		Path:     "/api/v1/auth/two-factor",
		HttpOnly: true,
		Secure:   h.cfg.CookieSecure,
		SameSite: h.authCookieSameSite(),
		MaxAge:   maxAge,
	})
}

// @Summary Complete two-factor sign-in
// @Description Requires short-lived HttpOnly uptaris_two_factor cookie from password or OAuth sign-in. Accepts authenticator or single-use backup code. Five attempts maximum.
// @Tags auth
// @Accept json
// @Produce json
// @Param body body request.TwoFactorVerification true "Second factor"
// @Success 200 {object} response.AuthResponse
// @Failure 422 {object} response.ErrorResponse
// @Failure 429 {object} response.ErrorResponse
// @Router /auth/two-factor/verify [post]
func (h *Handlers) VerifyTwoFactor(c *gin.Context) {
	var input request.TwoFactorVerification

	if !request.JSON(c, &input) || !request.Valid(c, request.Validate(&input)) {
		return
	}

	token, err := c.Cookie("uptaris_two_factor")
	if err != nil || token == "" {
		response.Fail(c, 422, "invalid_two_factor_challenge", "sign-in expired; sign in again")
		return
	}

	session, err := h.auth.CompleteTwoFactor(c.Request.Context(), token, input.Code, input.BackupCode)
	if err != nil {
		response.Error(c, err)
		return
	}

	h.session(c, session)
}

// @Summary Cancel pending two-factor sign-in
// @Tags auth
// @Success 204
// @Router /auth/two-factor/cancel [post]
func (h *Handlers) CancelTwoFactor(c *gin.Context) {
	token, _ := c.Cookie("uptaris_two_factor")
	if err := h.auth.CancelTwoFactor(c.Request.Context(), token); err != nil {
		response.Error(c, err)
		return
	}

	h.challengeCookie(c, "", -1)
	c.Status(204)
}

// @Summary Start authenticator enrollment
// @Description Current password required when account has a password. Secret and otpauth URI expire after ten minutes; QR generated locally by frontend.
// @Tags account
// @Security bearerauth
// @Accept json
// @Produce json
// @Param body body request.TwoFactorSetup true "Current password"
// @Success 200 {object} response.TwoFactorEnrollmentResponse
// @Failure 409 {object} response.ErrorResponse
// @Router /account/two-factor/setup [post]
func (h *Handlers) SetupTwoFactor(c *gin.Context) {
	identity, ok := middleware.Require(c, h.auth, "viewer", "operator", "admin")
	if !ok {
		return
	}

	var input request.TwoFactorSetup

	if !request.JSON(c, &input) || !request.Valid(c, request.Validate(&input)) {
		return
	}

	result, err := h.auth.SetupTwoFactor(c.Request.Context(), identity.ID, input.CurrentPassword)
	if err != nil {
		response.Error(c, err)
		return
	}

	c.JSON(200, result)
}

// @Summary Confirm authenticator enrollment
// @Description Activates 2FA and revokes other sessions. Returns ten backup codes once; store them safely.
// @Tags account
// @Security bearerauth
// @Accept json
// @Produce json
// @Param body body request.TwoFactorVerification true "Authenticator code (backupCode must be false)"
// @Success 200 {object} response.BackupCodesResponse
// @Failure 422 {object} response.ErrorResponse
// @Router /account/two-factor/confirm [post]
func (h *Handlers) ConfirmTwoFactor(c *gin.Context) {
	identity, ok := middleware.Require(c, h.auth, "viewer", "operator", "admin")
	if !ok {
		return
	}

	var input request.TwoFactorVerification

	if !request.JSON(c, &input) || !request.Valid(c, request.Validate(&input)) {
		return
	}

	if input.BackupCode {
		response.Fail(c, 422, "invalid_two_factor_code", "use an authenticator code to confirm enrollment")
		return
	}

	codes, err := h.auth.ConfirmTwoFactor(c.Request.Context(), identity.ID, identity.SessionID, input.Code)
	if err != nil {
		response.Error(c, err)
		return
	}

	c.JSON(200, response.BackupCodesResponse{BackupCodes: codes})
}

// @Summary Disable two-factor authentication
// @Tags account
// @Security bearerauth
// @Accept json
// @Param body body request.TwoFactorManagement true "Current password (if present) and second factor"
// @Success 204
// @Failure 422 {object} response.ErrorResponse
// @Router /account/two-factor/disable [post]
func (h *Handlers) DisableTwoFactor(c *gin.Context) { h.manageTwoFactor(c, true) }

// @Summary Refresh backup codes
// @Description Invalidates all previous backup codes; returns ten new codes once.
// @Tags account
// @Security bearerauth
// @Accept json
// @Produce json
// @Param body body request.TwoFactorManagement true "Current password (if present) and second factor"
// @Success 200 {object} response.BackupCodesResponse
// @Failure 422 {object} response.ErrorResponse
// @Router /account/two-factor/backup-codes [post]
func (h *Handlers) RegenerateBackupCodes(c *gin.Context) { h.manageTwoFactor(c, false) }

func (h *Handlers) manageTwoFactor(c *gin.Context, disable bool) {
	identity, ok := middleware.Require(c, h.auth, "viewer", "operator", "admin")
	if !ok {
		return
	}

	var input request.TwoFactorManagement

	if !request.JSON(c, &input) || !request.Valid(c, request.Validate(&input)) {
		return
	}

	codes, err := h.auth.ManageTwoFactor(c.Request.Context(), identity.ID, identity.SessionID, input.CurrentPassword, input.Code, input.BackupCode, disable)
	if err != nil {
		response.Error(c, err)
		return
	}

	if disable {
		c.Status(204)
		return
	}

	c.JSON(200, response.BackupCodesResponse{BackupCodes: codes})
}
