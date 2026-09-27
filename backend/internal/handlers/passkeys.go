package handlers

import (
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/uptaris/uptaris/backend/internal/middleware"
	"github.com/uptaris/uptaris/backend/internal/request"
	"github.com/uptaris/uptaris/backend/internal/response"
)

// @Summary Begin passkey sign-in
// @Tags auth
// @Success 200 {object} response.PasskeyOptionsResponse
// @Router /auth/passkeys/login/options [post]
func (h *Handlers) BeginPasskeyLogin(c *gin.Context) {
	result, err := h.auth.BeginPasskeyLogin(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// @Summary Finish passkey sign-in
// @Tags auth
// @Accept json
// @Param body body request.PasskeyVerification true "Passkey assertion and one-time ceremony token"
// @Success 200 {object} response.AuthResponse
// @Router /auth/passkeys/login/verify [post]
func (h *Handlers) FinishPasskeyLogin(c *gin.Context) {
	var input request.PasskeyVerification

	if !request.JSON(c, &input) || !validPasskeyInput(c, input) {
		return
	}

	session, err := h.auth.FinishPasskeyLogin(c.Request.Context(), input.CeremonyToken, input.Credential)
	if err != nil {
		response.Error(c, err)
		return
	}

	h.session(c, session)
}

// @Summary List account passkeys
// @Tags account
// @Security bearerauth
// @Success 200 {array} response.PasskeyResponse
// @Router /account/passkeys [get]
func (h *Handlers) ListPasskeys(c *gin.Context) {
	identity, ok := middleware.Require(c, h.auth, "viewer", "operator", "admin")
	if !ok {
		return
	}

	items, err := h.auth.ListPasskeys(c.Request.Context(), identity.ID)
	if err != nil {
		response.Error(c, err)
		return
	}

	c.JSON(http.StatusOK, items)
}

// @Summary Begin passkey enrollment
// @Tags account
// @Security bearerauth
// @Success 200 {object} response.PasskeyOptionsResponse
// @Router /account/passkeys/options [post]
func (h *Handlers) BeginPasskeyRegistration(c *gin.Context) {
	identity, ok := middleware.Require(c, h.auth, "viewer", "operator", "admin")
	if !ok {
		return
	}

	result, err := h.auth.BeginPasskeyRegistration(c.Request.Context(), identity.ID, identity.SessionID)
	if err != nil {
		response.Error(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// @Summary Finish passkey enrollment
// @Tags account
// @Security bearerauth
// @Accept json
// @Param body body request.PasskeyVerification true "Passkey attestation and one-time ceremony token"
// @Success 201 {object} response.PasskeyResponse
// @Router /account/passkeys/verify [post]
func (h *Handlers) FinishPasskeyRegistration(c *gin.Context) {
	identity, ok := middleware.Require(c, h.auth, "viewer", "operator", "admin")
	if !ok {
		return
	}

	var input request.PasskeyVerification

	if !request.JSON(c, &input) || !validPasskeyInput(c, input) {
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	if utf8.RuneCountInString(input.Name) > 80 {
		response.Fail(c, 422, "validation_failed", "passkey name must be at most 80 characters")
		return
	}

	item, err := h.auth.FinishPasskeyRegistration(c.Request.Context(), identity.ID, identity.SessionID, input.CeremonyToken, input.Name, input.Credential)
	if err != nil {
		response.Error(c, err)
		return
	}

	c.JSON(http.StatusCreated, item)
}

// @Summary Remove account passkey
// @Tags account
// @Security bearerauth
// @Param id path int true "Passkey ID"
// @Success 204
// @Router /account/passkeys/{id} [delete]
func (h *Handlers) DeletePasskey(c *gin.Context) {
	identity, ok := middleware.Require(c, h.auth, "viewer", "operator", "admin")
	if !ok {
		return
	}

	id, ok := request.ID(c, "id")
	if !ok {
		return
	}

	if err := h.auth.DeletePasskey(c.Request.Context(), identity.ID, id); err != nil {
		response.Error(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func validPasskeyInput(c *gin.Context, input request.PasskeyVerification) bool {
	if len(input.CeremonyToken) < 32 || len(input.CeremonyToken) > 256 || len(input.Credential) == 0 {
		response.Fail(c, 422, "validation_failed", "ceremonyToken and credential are required")
		return false
	}

	return true
}
