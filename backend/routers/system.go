package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/uptaris/uptaris/backend/internal/handlers"
)

func registerSystem(v1 *gin.RouterGroup, api *handlers.Handlers) {
	v1.GET("/status", api.Status)
	v1.GET("/dashboard", api.Dashboard)
	v1.GET("/incidents", api.AllIncidents)
}
