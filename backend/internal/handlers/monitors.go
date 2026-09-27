package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/uptaris/uptaris/backend/internal/auth"
	"github.com/uptaris/uptaris/backend/internal/models"
	"github.com/uptaris/uptaris/backend/internal/request"
	"github.com/uptaris/uptaris/backend/internal/response"
)

// @Summary List server monitors
// @Tags monitors
// @Security bearerauth
// @Param serverId path int true "Server ID" minimum(1)
// @Param page query int false "Page number" default(1) minimum(1)
// @Param pageSize query int false "Results per page" default(20) minimum(1) maximum(100)
// @Param type query string false "Type" Enums(http,tcp,icmp)
// @Param status query string false "Status" Enums(up,down,paused)
// @Success 200 {object} response.MonitorListResponse
// @Produce json
// @Failure 400 {object} response.ErrorResponse "invalid_id or invalid_query: invalid ID or query parameter"
// @Failure 401 {object} response.ErrorResponse "authentication_required, invalid_token, revoked_token, or stale_token"
// @Failure 404 {object} response.ErrorResponse "not_found: resource not found"
// @Failure 413 {object} response.ErrorResponse "body_too_large: request body exceeds 64 KiB"
// @Failure 500 {object} response.ErrorResponse "internal_error: unexpected server error"
// @Failure 503 {object} response.ErrorResponse "service_unavailable: database operation failed"
// @Router /servers/{serverId}/monitors [get]
func (h *Handlers) Monitors(c *gin.Context) {
	_, server, ok := h.oneServer(c)
	if !ok {
		return
	}

	pageNumber, pageSize, ok := request.Page(c)
	if !ok {
		return
	}

	monitorType, ok := request.Filter(c, "type", "http", "tcp", "icmp")
	if !ok {
		return
	}

	status, ok := request.Filter(c, "status", "up", "down", "paused")
	if !ok {
		return
	}

	rows, total, err := models.ListMonitors(c.Request.Context(), h.db, server.ID, pageNumber, pageSize, monitorType, status)
	if !response.ResourceError(c, err, "monitor") {
		return
	}

	response.List(c, rows, total, pageNumber, pageSize)
}

// @Summary Create server monitor
// @Tags monitors
// @Accept json
// @Security bearerauth
// @Param serverId path int true "Server ID" minimum(1)
// @Param body body request.MonitorInput true "Monitor fields"
// @Success 201 {object} models.Monitor
// @Produce json
// @Header 201 {string} Location "Created resource URI"
// @Failure 400 {object} response.ErrorResponse "invalid_id or invalid_json: invalid ID or JSON body"
// @Failure 401 {object} response.ErrorResponse "authentication_required, invalid_token, revoked_token, or stale_token"
// @Failure 403 {object} response.ErrorResponse "forbidden or origin_not_allowed: role or origin not permitted"
// @Failure 404 {object} response.ErrorResponse "not_found: resource not found"
// @Failure 413 {object} response.ErrorResponse "body_too_large: request body exceeds 64 KiB"
// @Failure 415 {object} response.ErrorResponse "unsupported_media_type: non-JSON request body"
// @Failure 422 {object} response.ErrorResponse "validation_failed: invalid request field"
// @Failure 500 {object} response.ErrorResponse "internal_error: unexpected server error"
// @Failure 503 {object} response.ErrorResponse "service_unavailable: database operation failed"
// @Router /servers/{serverId}/monitors [post]
func (h *Handlers) CreateMonitor(c *gin.Context) {
	currentActor, server, ok := h.oneServer(c)
	if !ok {
		return
	}

	if currentActor.Role == "viewer" {
		response.Fail(c, 403, "forbidden", "viewer is read only")
		return
	}

	var input request.MonitorInput

	if !request.JSON(c, &input) {
		return
	}

	monitor := models.Monitor{
		ServerID:        server.ID,
		Name:            input.Name,
		Type:            input.Type,
		Target:          input.Target,
		IntervalSeconds: input.IntervalSeconds,
		ExpectedHealth:  input.ExpectedHealth,
		Status:          defaultOf(input.Status, "up"),
	}

	if !request.Valid(c, request.PrepareMonitor(&monitor)) {
		return
	}

	if err := monitor.Create(c.Request.Context(), h.db); err != nil {
		response.Error(c, err)
		return
	}

	c.Header("Location", c.Request.URL.Path+"/"+strconv.FormatUint(uint64(monitor.ID), 10))
	c.JSON(201, monitor)
}

func (h *Handlers) oneMonitor(c *gin.Context) (auth.Identity, *models.Server, *models.Monitor, bool) {
	currentActor, server, ok := h.oneServer(c)
	if !ok {
		return currentActor, nil, nil, false
	}

	monitorID, ok := request.ID(c, "monitorId")
	if !ok {
		return currentActor, nil, nil, false
	}

	monitor, err := models.GetMonitor(c.Request.Context(), h.db, server.ID, monitorID)
	if !response.ResourceError(c, err, "monitor") {
		return currentActor, nil, nil, false
	}

	return currentActor, server, monitor, true
}

// @Summary Get monitor
// @Tags monitors
// @Security bearerauth
// @Param serverId path int true "Server ID" minimum(1)
// @Param monitorId path int true "Monitor ID" minimum(1)
// @Success 200 {object} response.MonitorResponse
// @Produce json
// @Failure 400 {object} response.ErrorResponse "invalid_id: invalid resource ID"
// @Failure 401 {object} response.ErrorResponse "authentication_required, invalid_token, revoked_token, or stale_token"
// @Failure 404 {object} response.ErrorResponse "not_found: resource not found"
// @Failure 413 {object} response.ErrorResponse "body_too_large: request body exceeds 64 KiB"
// @Failure 500 {object} response.ErrorResponse "internal_error: unexpected server error"
// @Failure 503 {object} response.ErrorResponse "service_unavailable: database operation failed"
// @Router /servers/{serverId}/monitors/{monitorId} [get]
func (h *Handlers) GetMonitor(c *gin.Context) {
	_, _, monitor, ok := h.oneMonitor(c)
	if ok {
		c.JSON(200, response.MonitorResponse{Data: *monitor, Links: map[string]string{"self": c.Request.URL.Path, "incidents": c.Request.URL.Path + "/incidents"}})
	}
}

// @Summary Update monitor
// @Tags monitors
// @Accept json
// @Security bearerauth
// @Param serverId path int true "Server ID" minimum(1)
// @Param monitorId path int true "Monitor ID" minimum(1)
// @Param body body request.MonitorPatchInput true "Fields to update"
// @Success 200 {object} models.Monitor
// @Produce json
// @Failure 400 {object} response.ErrorResponse "invalid_id or invalid_json: invalid ID or JSON body"
// @Failure 401 {object} response.ErrorResponse "authentication_required, invalid_token, revoked_token, or stale_token"
// @Failure 403 {object} response.ErrorResponse "forbidden or origin_not_allowed: role or origin not permitted"
// @Failure 404 {object} response.ErrorResponse "not_found: resource not found"
// @Failure 413 {object} response.ErrorResponse "body_too_large: request body exceeds 64 KiB"
// @Failure 415 {object} response.ErrorResponse "unsupported_media_type: non-JSON request body"
// @Failure 422 {object} response.ErrorResponse "validation_failed: invalid request field"
// @Failure 500 {object} response.ErrorResponse "internal_error: unexpected server error"
// @Failure 503 {object} response.ErrorResponse "service_unavailable: database operation failed"
// @Router /servers/{serverId}/monitors/{monitorId} [patch]
func (h *Handlers) UpdateMonitor(c *gin.Context) {
	currentActor, _, monitor, ok := h.oneMonitor(c)
	if !ok {
		return
	}

	if currentActor.Role == "viewer" {
		response.Fail(c, 403, "forbidden", "viewer is read only")
		return
	}

	var input request.MonitorPatchInput

	if !request.JSON(c, &input) {
		return
	}

	err := monitor.Update(c.Request.Context(), h.db, func(value *models.Monitor) error {
		input.Apply(value)

		return request.PrepareMonitor(value)
	})
	if !response.ResourceError(c, err, "monitor") {
		return
	}

	c.JSON(200, monitor)
}

// @Summary Delete monitor and incidents
// @Tags monitors
// @Security bearerauth
// @Param serverId path int true "Server ID" minimum(1)
// @Param monitorId path int true "Monitor ID" minimum(1)
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
// @Router /servers/{serverId}/monitors/{monitorId} [delete]
func (h *Handlers) DeleteMonitor(c *gin.Context) {
	currentActor, _, monitor, ok := h.oneMonitor(c)
	if !ok {
		return
	}

	if currentActor.Role == "viewer" {
		response.Fail(c, 403, "forbidden", "viewer is read only")
		return
	}

	if err := monitor.Delete(c.Request.Context(), h.db); err != nil {
		response.Error(c, err)
		return
	}

	c.Status(204)
}
