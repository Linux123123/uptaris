package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/uptaris/uptaris/backend/internal/middleware"
	"github.com/uptaris/uptaris/backend/internal/request"
	"github.com/uptaris/uptaris/backend/internal/response"
)

// @Summary List users
// @Tags admin
// @Security bearerauth
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Results per page" default(20)
// @Success 200 {object} response.UserListResponse
// @Failure 401 {object} response.UnauthorizedError
// @Failure 403 {object} response.ForbiddenError
// @Produce json
// @Failure 503 {object} response.ErrorResponse "Service temporarily unavailable"
// @Failure 400 {object} response.BadRequestError
// @Router /admin/users [get]
func (h *Handlers) Users(c *gin.Context) {
	_, ok := middleware.Require(c, h.auth, "admin")
	if !ok {
		return
	}
	page, size, ok := request.Page(c)
	if !ok {
		return
	}
	rows, total, err := h.users.List(c.Request.Context(), page, size)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.List(c, rows, total, page, size)
}

// @Summary Update user role
// @Description Changes user authorization role and revokes every active refresh session. Existing access tokens are rejected because their role claim no longer matches stored role.
// @Tags admin
// @Accept json
// @Produce json
// @Security bearerauth
// @Param userId path int true "User ID" minimum(1)
// @Param body body request.UserRoleInput true "New user role"
// @Success 200 {object} response.UserRecord
// @Failure 400 {object} response.BadRequestError "Malformed user ID or JSON"
// @Failure 401 {object} response.UnauthorizedError
// @Failure 403 {object} response.ForbiddenError
// @Failure 404 {object} response.NotFoundError "User not found"
// @Failure 409 {object} response.ConflictError "Administrator cannot change own role"
// @Failure 422 {object} response.ValidationError "Role must be viewer, operator, or admin"
// @Failure 503 {object} response.ErrorResponse "Service temporarily unavailable"
// @Failure 413 {object} response.ErrorResponse "Body exceeds 64 KiB"
// @Failure 415 {object} response.ErrorResponse "JSON content type required"
// @Router /admin/users/{userId} [patch]
func (h *Handlers) UpdateUserRole(c *gin.Context) {
	identity, ok := middleware.Require(c, h.auth, "admin")
	if !ok {
		return
	}
	id, ok := request.ID(c, "userId")
	if !ok {
		return
	}
	var input request.UserRoleInput
	if !request.JSON(c, &input) {
		return
	}
	user, err := h.users.ChangeRole(c.Request.Context(), identity, id, input.Role)
	if !response.ResourceError(c, err, "user") {
		return
	}
	c.JSON(200, user)
}

// @Summary Delete user and revoke sessions
// @Tags admin
// @Security bearerauth
// @Param userId path int true "User ID"
// @Success 204
// @Failure 400 {object} response.BadRequestError
// @Failure 401 {object} response.UnauthorizedError
// @Failure 403 {object} response.ForbiddenError
// @Failure 409 {object} response.ConflictError
// @Produce json
// @Failure 503 {object} response.ErrorResponse "Service temporarily unavailable"
// @Failure 404 {object} response.NotFoundError
// @Router /admin/users/{userId} [delete]
func (h *Handlers) DeleteUser(c *gin.Context) {
	identity, ok := middleware.Require(c, h.auth, "admin")
	if !ok {
		return
	}
	id, ok := request.ID(c, "userId")
	if !ok {
		return
	}
	if !response.ResourceError(c, h.users.Delete(c.Request.Context(), identity, id), "user") {
		return
	}
	c.Status(204)
}
