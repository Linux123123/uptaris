package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/uptaris/uptaris/backend/internal/handlers"
)

func registerIncidents(v1 *gin.RouterGroup, api *handlers.Handlers) {
	v1.GET("/servers/:serverId/monitors/:monitorId/incidents", api.Incidents)
	v1.POST("/servers/:serverId/monitors/:monitorId/incidents", api.CreateIncident)
	v1.GET("/servers/:serverId/monitors/:monitorId/incidents/:incidentId", api.GetIncident)
	v1.PATCH("/servers/:serverId/monitors/:monitorId/incidents/:incidentId", api.UpdateIncident)
	v1.DELETE("/servers/:serverId/monitors/:monitorId/incidents/:incidentId", api.DeleteIncident)
}
