package handlers

import (
	"crypto/subtle"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uptaris/uptaris/backend/internal/auth"
	"github.com/uptaris/uptaris/backend/internal/middleware"
	"github.com/uptaris/uptaris/backend/internal/response"
)

// @Summary List configured OAuth providers
// @Tags auth
// @Success 200 {array} auth.OAuthProviderInfo
// @Produce json
// @Router /auth/providers [get]
func (h *Handlers) OAuthProviders(c *gin.Context) {
	c.JSON(http.StatusOK, h.auth.OAuthProviders())
}

// @Summary Start OAuth sign-in
// @Tags auth
// @Param provider path string true "OAuth provider ID"
// @Success 302 "Redirect to provider authorization"
// @Failure 503 {object} response.ErrorResponse
// @Router /auth/oauth/{provider} [get]
func (h *Handlers) OAuthLogin(c *gin.Context) {
	provider := c.Param("provider")
	authorizeURL, state, err := h.auth.StartOAuth(c.Request.Context(), provider, "login", 0)
	if err != nil {
		response.Error(c, err)
		return
	}

	h.oauthCookie(c, state+"|login", 600)
	c.Redirect(http.StatusFound, authorizeURL)
}

// @Summary Start authenticated OAuth connection
// @Tags auth
// @Security bearerauth
// @Param provider path string true "OAuth provider ID"
// @Success 200 {object} map[string]string
// @Failure 401 {object} response.ErrorResponse
// @Failure 503 {object} response.ErrorResponse
// @Router /auth/oauth/{provider}/link [post]
func (h *Handlers) OAuthLink(c *gin.Context) {
	identity, ok := middleware.Require(c, h.auth, "viewer", "operator", "admin")
	if !ok {
		return
	}

	provider := c.Param("provider")
	authorizeURL, state, err := h.auth.StartOAuth(c.Request.Context(), provider, "link", identity.ID)
	if err != nil {
		response.Error(c, err)
		return
	}

	h.oauthCookie(c, state+"|link", 600)
	c.JSON(http.StatusOK, gin.H{"authorizeUrl": authorizeURL})
}

// @Summary Complete OAuth sign-in or connection
// @Tags auth
// @Param provider path string true "OAuth provider ID"
// @Param code query string true "Provider authorization code"
// @Param state query string true "One-time state value"
// @Success 302 "Redirect to app"
// @Router /auth/oauth/{provider}/callback [get]
func (h *Handlers) OAuthCallback(c *gin.Context) {
	h.oauthCookie(c, "", -1)
	cookieValue, stateErr := c.Cookie("uptaris_oauth_state")
	state, intent, _ := strings.Cut(cookieValue, "|")
	provider := c.Param("provider")
	queryState := c.Query("state")
	if stateErr != nil || queryState == "" || len(state) != len(queryState) || subtle.ConstantTimeCompare([]byte(state), []byte(queryState)) != 1 {
		h.oauthRedirect(c, intent, "invalid_state")
		return
	}

	if c.Query("error") != "" {
		if err := h.auth.CancelOAuth(c.Request.Context(), provider, queryState); err != nil {
			h.oauthRedirect(c, intent, "invalid_state")
			return
		}

		h.oauthRedirect(c, intent, "cancelled")
		return
	}

	session, err := h.auth.CompleteOAuth(c.Request.Context(), provider, queryState, c.Query("code"))
	if err != nil {
		failureCode := "failed"
		if authErr, ok := err.(*auth.Error); ok {
			failureCode = authErr.Code
		}

		h.oauthRedirect(c, intent, failureCode)
		return
	}

	if session == nil {
		c.Redirect(http.StatusFound, h.cfg.FrontendURL+"/app/settings")
		return
	}

	if session.ChallengeToken != "" {
		h.challengeCookie(c, session.ChallengeToken, 300)
		h.cookie(c, "", -1)
		c.Redirect(http.StatusFound, h.cfg.FrontendURL+"/two-factor")
		return
	}

	h.challengeCookie(c, "", -1)
	h.cookie(c, session.RefreshToken, max(1, int(time.Until(session.ExpiresAt).Seconds())))
	c.Redirect(http.StatusFound, h.cfg.FrontendURL+"/oauth/callback")
}

// @Summary Disconnect OAuth provider
// @Tags auth
// @Security bearerauth
// @Param provider path string true "OAuth provider ID"
// @Success 204
// @Failure 401 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse "last_signin_method"
// @Router /auth/oauth/{provider}/link [delete]
func (h *Handlers) OAuthUnlink(c *gin.Context) {
	identity, ok := middleware.Require(c, h.auth, "viewer", "operator", "admin")
	if !ok {
		return
	}

	if err := h.auth.UnlinkOAuth(c.Request.Context(), c.Param("provider"), identity.ID); err != nil {
		response.Error(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handlers) oauthRedirect(c *gin.Context, intent, code string) {
	path := "/login"
	if intent == "link" {
		path = "/app/settings"
	}

	c.Redirect(http.StatusFound, h.cfg.FrontendURL+path+"?oauth_error="+url.QueryEscape(code))
}

func (h *Handlers) oauthCookie(c *gin.Context, value string, maxAge int) {
	// Cross-site linking needs None; otherwise Lax permits the provider callback.
	sameSite := http.SameSiteLaxMode
	if h.cfg.CookieSameSite == "none" {
		sameSite = http.SameSiteNoneMode
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "uptaris_oauth_state",
		Value:    value,
		Path:     "/api/v1/auth/oauth",
		HttpOnly: true,
		Secure:   h.cfg.CookieSecure,
		SameSite: sameSite,
		MaxAge:   maxAge,
	})
}
