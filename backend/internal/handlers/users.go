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
// @Param page query int false "Page number" default(1) minimum(1)
// @Param pageSize query int false "Results per page" default(20) minimum(1) maximum(100)
// @Success 200 {object} response.UserListResponse
// @Produce json
// @Failure 400 {object} response.ErrorResponse "invalid_query: invalid query parameter"
// @Failure 401 {object} response.ErrorResponse "authentication_required, invalid_token, revoked_token, or stale_token"
// @Failure 403 {object} response.ErrorResponse "forbidden: administrator role required"
// @Failure 413 {object} response.ErrorResponse "body_too_large: request body exceeds 64 KiB"
// @Failure 500 {object} response.ErrorResponse "internal_error: unexpected server error"
// @Failure 503 {object} response.ErrorResponse "service_unavailable: database operation failed"
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
// @Failure 400 {object} response.ErrorResponse "invalid_id or invalid_json: invalid ID or JSON body"
// @Failure 401 {object} response.ErrorResponse "authentication_required, invalid_token, revoked_token, or stale_token"
// @Failure 403 {object} response.ErrorResponse "forbidden or origin_not_allowed: role or origin not permitted"
// @Failure 404 {object} response.ErrorResponse "not_found: resource not found"
// @Failure 409 {object} response.ErrorResponse "self_role_change: cannot change own role"
// @Failure 413 {object} response.ErrorResponse "body_too_large: request body exceeds 64 KiB"
// @Failure 415 {object} response.ErrorResponse "unsupported_media_type: non-JSON request body"
// @Failure 422 {object} response.ErrorResponse "validation_failed: invalid request field"
// @Failure 500 {object} response.ErrorResponse "internal_error: unexpected server error"
// @Failure 503 {object} response.ErrorResponse "service_unavailable: database operation failed"
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
	if !request.Valid(c, request.Validate(&input)) {
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
// @Param userId path int true "User ID" minimum(1)
// @Success 204
// @Produce json
// @Failure 400 {object} response.ErrorResponse "invalid_id: invalid resource ID"
// @Failure 401 {object} response.ErrorResponse "authentication_required, invalid_token, revoked_token, or stale_token"
// @Failure 403 {object} response.ErrorResponse "forbidden or origin_not_allowed: role or origin not permitted"
// @Failure 404 {object} response.ErrorResponse "not_found: resource not found"
// @Failure 409 {object} response.ErrorResponse "self_delete: cannot delete own account"
// @Failure 413 {object} response.ErrorResponse "body_too_large: request body exceeds 64 KiB"
// @Failure 415 {object} response.ErrorResponse "unsupported_media_type: non-JSON request body"
// @Failure 500 {object} response.ErrorResponse "internal_error: unexpected server error"
// @Failure 503 {object} response.ErrorResponse "service_unavailable: database operation failed"
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
