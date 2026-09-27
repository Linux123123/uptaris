package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/uptaris/uptaris/backend/internal/handlers"
	"github.com/uptaris/uptaris/backend/internal/middleware"
)

func registerAccounts(v1 *gin.RouterGroup, api *handlers.Handlers) {
	limit := middleware.AuthLimit()
	v1.GET("/account", api.AccountSettings)
	v1.GET("/account/passkeys", api.ListPasskeys)
	v1.POST("/account/passkeys/options", limit, api.BeginPasskeyRegistration)
	v1.POST("/account/passkeys/verify", limit, api.FinishPasskeyRegistration)
	v1.DELETE("/account/passkeys/:id", limit, api.DeletePasskey)
	v1.PUT("/account/password", limit, api.SetPassword)
	v1.POST("/account/two-factor/setup", limit, api.SetupTwoFactor)
	v1.POST("/account/two-factor/confirm", limit, api.ConfirmTwoFactor)
	v1.POST("/account/two-factor/disable", limit, api.DisableTwoFactor)
	v1.POST("/account/two-factor/backup-codes", limit, api.RegenerateBackupCodes)
}
