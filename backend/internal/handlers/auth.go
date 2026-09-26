package handlers

import (
	"net/http"
	"time"

	"github.com/uptaris/uptaris/backend/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/uptaris/uptaris/backend/internal/auth"
	"github.com/uptaris/uptaris/backend/internal/request"
	"github.com/uptaris/uptaris/backend/internal/response"
)

// @Summary Register account
// @Tags auth
// @Accept json
// @Param body body request.Credentials true "Email and password"
// @Success 201 {object} response.CreateUserResponse
// @Produce json
// @Header 201 {string} Location "Created resource URI"
// @Failure 400 {object} response.ErrorResponse "invalid_json: malformed JSON body"
// @Failure 403 {object} response.ErrorResponse "origin_not_allowed: request origin not allowed"
// @Failure 409 {object} response.ErrorResponse "email_exists: email already registered"
// @Failure 413 {object} response.ErrorResponse "body_too_large: request body exceeds 64 KiB"
// @Failure 415 {object} response.ErrorResponse "unsupported_media_type: non-JSON request body"
// @Failure 422 {object} response.ErrorResponse "validation_failed: invalid request field"
// @Failure 429 {object} response.ErrorResponse "rate_limited: too many authentication attempts"
// @Header 429 {string} Retry-After "Seconds before retry"
// @Failure 500 {object} response.ErrorResponse "internal_error: unexpected server error"
// @Failure 503 {object} response.ErrorResponse "service_unavailable: database operation failed"
// @Router /auth/register [post]
func (h *Handlers) Register(c *gin.Context) {
	var input request.Credentials
	if !request.JSON(c, &input) {
		return
	}
	if !request.Valid(c, request.Validate(&input)) {
		return
	}
	user, err := h.auth.Register(c.Request.Context(), input.Email, input.Password)
	if err != nil {
		response.Error(c, err)
		return
	}
	c.Header("Location", "/api/v1/auth/me")
	c.JSON(201, response.User(user))
}

// @Summary Log in
// @Tags auth
// @Accept json
// @Param body body request.Credentials true "Email and password"
// @Success 200 {object} response.AuthResponse
// @Produce json
// @Header 200 {string} Set-Cookie "HttpOnly refresh cookie"
// @Failure 400 {object} response.ErrorResponse "invalid_json: malformed JSON body"
// @Failure 401 {object} response.ErrorResponse "invalid_credentials: email or password invalid"
// @Failure 403 {object} response.ErrorResponse "origin_not_allowed: request origin not allowed"
// @Failure 413 {object} response.ErrorResponse "body_too_large: request body exceeds 64 KiB"
// @Failure 415 {object} response.ErrorResponse "unsupported_media_type: non-JSON request body"
// @Failure 422 {object} response.ErrorResponse "validation_failed: invalid request field"
// @Failure 429 {object} response.ErrorResponse "rate_limited: too many authentication attempts"
// @Header 429 {string} Retry-After "Seconds before retry"
// @Failure 500 {object} response.ErrorResponse "internal_error: unexpected server error"
// @Failure 503 {object} response.ErrorResponse "service_unavailable: database operation failed"
// @Router /auth/login [post]
func (h *Handlers) Login(c *gin.Context) {
	var input request.Credentials
	if !request.JSON(c, &input) {
		return
	}
	if !request.Valid(c, request.Validate(&input)) {
		return
	}
	session, err := h.auth.Login(c.Request.Context(), input.Email, input.Password)
	if err != nil {
		response.Error(c, err)
		return
	}
	h.session(c, session)
}

// @Summary Refresh access token
// @Description Requires uptaris_refresh HttpOnly cookie from login or previous refresh. Rotates it atomically within the existing session.
// @Tags auth
// @Success 200 {object} response.AuthResponse
// @Produce json
// @Header 200 {string} Set-Cookie "HttpOnly refresh cookie"
// @Failure 401 {object} response.ErrorResponse "refresh_required or invalid_refresh: refresh cookie missing or invalid"
// @Failure 403 {object} response.ErrorResponse "origin_not_allowed: request origin not allowed"
// @Failure 413 {object} response.ErrorResponse "body_too_large: request body exceeds 64 KiB"
// @Failure 415 {object} response.ErrorResponse "unsupported_media_type: non-JSON request body"
// @Failure 500 {object} response.ErrorResponse "internal_error: unexpected server error"
// @Failure 503 {object} response.ErrorResponse "service_unavailable: database operation failed"
// @Router /auth/refresh [post]
func (h *Handlers) Refresh(c *gin.Context) {
	cookie, err := c.Request.Cookie("uptaris_refresh")
	if err != nil {
		response.Fail(c, 401, "refresh_required", "refresh token required")
		return
	}
	session, err := h.auth.Refresh(c.Request.Context(), cookie.Value)
	if err != nil {
		response.Error(c, err)
		return
	}
	h.session(c, session)
}

// @Summary Log out
// @Description Revokes the current session, including all access tokens issued through refresh.
// @Tags auth
// @Security bearerauth
// @Success 204
// @Produce json
// @Header 204 {string} Set-Cookie "Expired refresh cookie"
// @Failure 401 {object} response.ErrorResponse "authentication_required, invalid_token, revoked_token, or stale_token"
// @Failure 403 {object} response.ErrorResponse "origin_not_allowed: request origin not allowed"
// @Failure 413 {object} response.ErrorResponse "body_too_large: request body exceeds 64 KiB"
// @Failure 415 {object} response.ErrorResponse "unsupported_media_type: non-JSON request body"
// @Failure 500 {object} response.ErrorResponse "internal_error: unexpected server error"
// @Failure 503 {object} response.ErrorResponse "service_unavailable: database operation failed"
// @Router /auth/logout [post]
func (h *Handlers) Logout(c *gin.Context) {
	identity, ok := middleware.Require(c, h.auth, "viewer", "operator", "admin")
	if !ok {
		return
	}
	if err := h.auth.Logout(c.Request.Context(), identity); err != nil {
		response.Error(c, err)
		return
	}
	h.cookie(c, "", -1)
	c.Status(204)
}

// @Summary Get authenticated user
// @Tags auth
// @Security bearerauth
// @Success 200 {object} response.UserResponse
// @Produce json
// @Failure 401 {object} response.ErrorResponse "authentication_required, invalid_token, revoked_token, or stale_token"
// @Failure 413 {object} response.ErrorResponse "body_too_large: request body exceeds 64 KiB"
// @Failure 500 {object} response.ErrorResponse "internal_error: unexpected server error"
// @Failure 503 {object} response.ErrorResponse "service_unavailable: database operation failed"
// @Router /auth/me [get]
func (h *Handlers) Me(c *gin.Context) {
	identity, ok := middleware.Require(c, h.auth, "viewer", "operator", "admin")
	if !ok {
		return
	}
	user, err := h.auth.User(c.Request.Context(), identity.ID)
	if err != nil {
		response.Error(c, err)
		return
	}
	c.JSON(200, response.User(user))
}

func (h *Handlers) cookie(c *gin.Context, value string, maxAge int) {
	sameSite := http.SameSiteLaxMode
	if h.cfg.CookieSameSite == "strict" {
		sameSite = http.SameSiteStrictMode
	}
	if h.cfg.CookieSameSite == "none" {
		sameSite = http.SameSiteNoneMode
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "uptaris_refresh",
		Value:    value,
		Path:     "/api/v1/auth",
		HttpOnly: true,
		Secure:   h.cfg.CookieSecure,
		SameSite: sameSite,
		MaxAge:   maxAge,
	})
}

func (h *Handlers) session(c *gin.Context, session *auth.Session) {
	h.cookie(c, session.RefreshToken, max(1, int(time.Until(session.ExpiresAt).Seconds())))
	c.JSON(200, response.AuthResponse{AccessToken: session.AccessToken, User: response.User(&session.User)})
}
