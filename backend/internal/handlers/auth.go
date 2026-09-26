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
// @Failure 400 {object} response.BadRequestError
// @Failure 409 {object} response.ConflictError
// @Failure 422 {object} response.ValidationError
// @Produce json
// @Failure 503 {object} response.ErrorResponse "Service temporarily unavailable"
// @Failure 413 {object} response.ErrorResponse "Body exceeds 64 KiB"
// @Failure 415 {object} response.ErrorResponse "JSON content type required"
// @Failure 429 {object} response.ErrorResponse "Too many authentication attempts"
// @Router /auth/register [post]
func (h *Handlers) Register(c *gin.Context) {
	var input request.Credentials
	if !request.JSON(c, &input) {
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
// @Failure 400 {object} response.BadRequestError
// @Failure 401 {object} response.UnauthorizedError
// @Produce json
// @Failure 503 {object} response.ErrorResponse "Service temporarily unavailable"
// @Failure 413 {object} response.ErrorResponse "Body exceeds 64 KiB"
// @Failure 415 {object} response.ErrorResponse "JSON content type required"
// @Failure 422 {object} response.ValidationError
// @Failure 429 {object} response.ErrorResponse "Too many authentication attempts"
// @Router /auth/login [post]
func (h *Handlers) Login(c *gin.Context) {
	var input request.Credentials
	if !request.JSON(c, &input) {
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
// @Description Rotates the HttpOnly refresh cookie atomically within the existing session.
// @Tags auth
// @Success 200 {object} response.AuthResponse
// @Failure 401 {object} response.ErrorResponse
// @Produce json
// @Failure 503 {object} response.ErrorResponse "Service temporarily unavailable"
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
// @Failure 401 {object} response.ErrorResponse
// @Produce json
// @Failure 503 {object} response.ErrorResponse "Service temporarily unavailable"
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
// @Failure 401 {object} response.ErrorResponse
// @Produce json
// @Failure 503 {object} response.ErrorResponse "Service temporarily unavailable"
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
