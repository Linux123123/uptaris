package handlers

import (
	"strconv"

	"github.com/uptaris/uptaris/backend/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/uptaris/uptaris/backend/internal/auth"
	"github.com/uptaris/uptaris/backend/internal/models"
	"github.com/uptaris/uptaris/backend/internal/request"
	"github.com/uptaris/uptaris/backend/internal/response"
)

// @Summary List servers
// @Tags servers
// @Produce json
// @Security bearerauth
// @Param page query int false "Page number" default(1) minimum(1)
// @Param pageSize query int false "Results per page" default(20) maximum(100) minimum(1)
// @Param status query string false "Status" Enums(up,down,paused)
// @Success 200 {object} response.ServerListResponse
// @Failure 400 {object} response.ErrorResponse "invalid_query: invalid query parameter"
// @Failure 401 {object} response.ErrorResponse "authentication_required, invalid_token, revoked_token, or stale_token"
// @Failure 413 {object} response.ErrorResponse "body_too_large: request body exceeds 64 KiB"
// @Failure 500 {object} response.ErrorResponse "internal_error: unexpected server error"
// @Failure 503 {object} response.ErrorResponse "service_unavailable: database operation failed"
// @Router /servers [get]
func (h *Handlers) Servers(c *gin.Context) {
	currentActor, ok := middleware.Require(c, h.auth, "viewer", "operator", "admin")
	if !ok {
		return
	}
	pageNumber, pageSize, ok := request.Page(c)
	if !ok {
		return
	}
	filters := models.Filters{}
	filters.Status, ok = request.Filter(c, "status", "up", "down", "paused")
	if !ok {
		return
	}
	rows, total, err := models.ListServers(c.Request.Context(), h.db, currentActor.ID, currentActor.Role == "admin", models.Page{Number: pageNumber, Size: pageSize}, filters)
	if !response.ResourceError(c, err, "server") {
		return
	}
	response.List(c, rows, total, pageNumber, pageSize)
}

// @Summary Create server
// @Tags servers
// @Accept json
// @Produce json
// @Security bearerauth
// @Param body body request.ServerInput true "Server fields"
// @Success 201 {object} models.Server
// @Header 201 {string} Location "Created resource URI"
// @Failure 400 {object} response.ErrorResponse "invalid_json: malformed JSON body"
// @Failure 401 {object} response.ErrorResponse "authentication_required, invalid_token, revoked_token, or stale_token"
// @Failure 403 {object} response.ErrorResponse "forbidden or origin_not_allowed: role or origin not permitted"
// @Failure 413 {object} response.ErrorResponse "body_too_large: request body exceeds 64 KiB"
// @Failure 415 {object} response.ErrorResponse "unsupported_media_type: non-JSON request body"
// @Failure 422 {object} response.ErrorResponse "validation_failed: invalid request field"
// @Failure 500 {object} response.ErrorResponse "internal_error: unexpected server error"
// @Failure 503 {object} response.ErrorResponse "service_unavailable: database operation failed"
// @Router /servers [post]
func (h *Handlers) CreateServer(c *gin.Context) {
	currentActor, ok := middleware.Require(c, h.auth, "operator", "admin")
	if !ok {
		return
	}
	var input request.ServerInput
	if !request.JSON(c, &input) {
		return
	}
	server := models.Server{
		OwnerID:         currentActor.ID,
		Name:            input.Name,
		Address:         input.Address,
		OperatingSystem: input.OperatingSystem,
		Description:     input.Description,
		Status:          defaultOf(input.Status, "up"),
	}
	if !request.Valid(c, request.PrepareServer(&server)) {
		return
	}
	if err := server.Create(c.Request.Context(), h.db); err != nil {
		response.Error(c, err)
		return
	}
	c.Header("Location", "/api/v1/servers/"+strconv.FormatUint(uint64(server.ID), 10))
	c.JSON(201, server)
}

func (h *Handlers) oneServer(c *gin.Context) (auth.Identity, *models.Server, bool) {
	currentActor, ok := middleware.Require(c, h.auth, "viewer", "operator", "admin")
	if !ok {
		return currentActor, nil, false
	}
	serverID, ok := request.ID(c, "serverId")
	if !ok {
		return currentActor, nil, false
	}
	server, err := models.GetServer(c.Request.Context(), h.db, currentActor.ID, currentActor.Role == "admin", serverID)
	if !response.ResourceError(c, err, "server") {
		return currentActor, nil, false
	}
	return currentActor, server, true
}

// @Summary Get server
// @Tags servers
// @Produce json
// @Security bearerauth
// @Param serverId path int true "Server ID" minimum(1)
// @Success 200 {object} response.ServerResponse
// @Failure 400 {object} response.ErrorResponse "invalid_id: invalid resource ID"
// @Failure 401 {object} response.ErrorResponse "authentication_required, invalid_token, revoked_token, or stale_token"
// @Failure 404 {object} response.ErrorResponse "not_found: resource not found"
// @Failure 413 {object} response.ErrorResponse "body_too_large: request body exceeds 64 KiB"
// @Failure 500 {object} response.ErrorResponse "internal_error: unexpected server error"
// @Failure 503 {object} response.ErrorResponse "service_unavailable: database operation failed"
// @Router /servers/{serverId} [get]
func (h *Handlers) GetServer(c *gin.Context) {
	_, server, ok := h.oneServer(c)
	if ok {
		c.JSON(200, response.ServerResponse{Data: *server, Links: map[string]string{"self": c.Request.URL.Path, "monitors": c.Request.URL.Path + "/monitors"}})
	}
}

// @Summary Update server
// @Tags servers
// @Accept json
// @Produce json
// @Security bearerauth
// @Param serverId path int true "Server ID" minimum(1)
// @Param body body request.ServerPatchInput true "Fields to update"
// @Success 200 {object} models.Server
// @Failure 400 {object} response.ErrorResponse "invalid_id or invalid_json: invalid ID or JSON body"
// @Failure 401 {object} response.ErrorResponse "authentication_required, invalid_token, revoked_token, or stale_token"
// @Failure 403 {object} response.ErrorResponse "forbidden or origin_not_allowed: role or origin not permitted"
// @Failure 404 {object} response.ErrorResponse "not_found: resource not found"
// @Failure 413 {object} response.ErrorResponse "body_too_large: request body exceeds 64 KiB"
// @Failure 415 {object} response.ErrorResponse "unsupported_media_type: non-JSON request body"
// @Failure 422 {object} response.ErrorResponse "validation_failed: invalid request field"
// @Failure 500 {object} response.ErrorResponse "internal_error: unexpected server error"
// @Failure 503 {object} response.ErrorResponse "service_unavailable: database operation failed"
// @Router /servers/{serverId} [patch]
func (h *Handlers) UpdateServer(c *gin.Context) {
	currentActor, server, ok := h.oneServer(c)
	if !ok {
		return
	}
	if currentActor.Role == "viewer" {
		response.Fail(c, 403, "forbidden", "viewer is read only")
		return
	}
	var input request.ServerPatchInput
	if !request.JSON(c, &input) {
		return
	}

	err := server.Update(c.Request.Context(), h.db, func(value *models.Server) error {
		input.Apply(value)
		return request.PrepareServer(value)
	})
	if !response.ResourceError(c, err, "server") {
		return
	}
	c.JSON(200, server)
}

// @Summary Delete server and descendants
// @Tags servers
// @Security bearerauth
// @Param serverId path int true "Server ID" minimum(1)
// @Success 204
// @Produce json
// @Failure 400 {object} response.ErrorResponse "invalid_id: invalid resource ID"
// @Failure 401 {object} response.ErrorResponse "authentication_required, invalid_token, revoked_token, or stale_token"
// @Failure 403 {object} response.ErrorResponse "forbidden or origin_not_allowed: role or origin not permitted"
// @Failure 404 {object} response.ErrorResponse "not_found: resource not found"
// @Failure 413 {object} response.ErrorResponse "body_too_large: request body exceeds 64 KiB"
// @Failure 415 {object} response.ErrorResponse "unsupported_media_type: non-JSON request body"
// @Failure 500 {object} response.ErrorResponse "internal_error: unexpected server error"
// @Failure 503 {object} response.ErrorResponse "service_unavailable: database operation failed"
// @Router /servers/{serverId} [delete]
func (h *Handlers) DeleteServer(c *gin.Context) {
	currentActor, server, ok := h.oneServer(c)
	if !ok {
		return
	}
	if currentActor.Role == "viewer" {
		response.Fail(c, 403, "forbidden", "viewer is read only")
		return
	}
	if err := server.Delete(c.Request.Context(), h.db); err != nil {
		response.Error(c, err)
		return
	}
	c.Status(204)
}
