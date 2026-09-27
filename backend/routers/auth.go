package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/uptaris/uptaris/backend/internal/handlers"
	"github.com/uptaris/uptaris/backend/internal/middleware"
)

func registerAuth(v1 *gin.RouterGroup, api *handlers.Handlers) {
	limit := middleware.AuthLimit()
	v1.POST("/auth/register", limit, api.Register)
	v1.POST("/auth/login", limit, api.Login)
	v1.POST("/auth/passkeys/login/options", limit, api.BeginPasskeyLogin)
	v1.POST("/auth/passkeys/login/verify", limit, api.FinishPasskeyLogin)
	v1.POST("/auth/two-factor/verify", limit, api.VerifyTwoFactor)
	v1.POST("/auth/two-factor/cancel", limit, api.CancelTwoFactor)
	v1.POST("/auth/refresh", api.Refresh)
	v1.POST("/auth/logout", api.Logout)
	v1.GET("/auth/me", api.Me)
	v1.GET("/auth/oauth/:provider", limit, api.OAuthLogin)
	v1.GET("/auth/oauth/:provider/callback", limit, api.OAuthCallback)
	v1.POST("/auth/oauth/:provider/link", limit, api.OAuthLink)
	v1.DELETE("/auth/oauth/:provider/link", api.OAuthUnlink)
	v1.GET("/auth/providers", api.OAuthProviders)
}
