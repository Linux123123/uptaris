package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/uptaris/uptaris/backend/internal/handlers"
)

func registerServers(v1 *gin.RouterGroup, api *handlers.Handlers) {
	v1.GET("/servers", api.Servers)
	v1.POST("/servers", api.CreateServer)
	v1.GET("/servers/:serverId", api.GetServer)
	v1.PATCH("/servers/:serverId", api.UpdateServer)
	v1.DELETE("/servers/:serverId", api.DeleteServer)
}
