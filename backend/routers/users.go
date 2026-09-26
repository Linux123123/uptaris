package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/uptaris/uptaris/backend/internal/handlers"
)

func registerUsers(v1 *gin.RouterGroup, api *handlers.Handlers) {
	v1.GET("/admin/users", api.Users)
	v1.PATCH("/admin/users/:userId", api.UpdateUserRole)
	v1.DELETE("/admin/users/:userId", api.DeleteUser)
}
