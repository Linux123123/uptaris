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
	v1.POST("/auth/refresh", api.Refresh)
	v1.POST("/auth/logout", api.Logout)
	v1.GET("/auth/me", api.Me)
}
