package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/uptaris/uptaris/backend/internal/handlers"
)

func registerMonitors(v1 *gin.RouterGroup, api *handlers.Handlers) {
	v1.GET("/servers/:serverId/monitors", api.Monitors)
	v1.POST("/servers/:serverId/monitors", api.CreateMonitor)
	v1.GET("/servers/:serverId/monitors/:monitorId", api.GetMonitor)
	v1.PATCH("/servers/:serverId/monitors/:monitorId", api.UpdateMonitor)
	v1.DELETE("/servers/:serverId/monitors/:monitorId", api.DeleteMonitor)
}
