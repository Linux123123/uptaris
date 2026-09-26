package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/uptaris/uptaris/backend/internal/auth"
	"github.com/uptaris/uptaris/backend/internal/inventory"
	"github.com/uptaris/uptaris/backend/internal/models"
	"github.com/uptaris/uptaris/backend/internal/request"
	"github.com/uptaris/uptaris/backend/internal/response"
)

// @Summary List server monitors
// @Tags monitors
// @Security bearerauth
// @Param serverId path int true "Server ID"
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Results per page" default(20)
// @Param type query string false "Type" Enums(http,tcp,icmp)
// @Param status query string false "Status" Enums(up,down,paused)
// @Success 200 {object} response.MonitorListResponse
// @Failure 400 {object} response.BadRequestError
// @Failure 401 {object} response.UnauthorizedError
// @Failure 404 {object} response.NotFoundError
// @Produce json
// @Failure 503 {object} response.ErrorResponse "Service temporarily unavailable"
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
	filters := inventory.Filters{}
	filters.Type, ok = request.Filter(c, "type", "http", "tcp", "icmp")
	if !ok {
		return
	}
	filters.Status, ok = request.Filter(c, "status", "up", "down", "paused")
	if !ok {
		return
	}
	rows, total, err := h.inventory.Monitors(c.Request.Context(), server.ID, inventory.Page{Number: pageNumber, Size: pageSize}, filters)
	if !response.ResourceError(c, err, "monitor") {
		return
	}
	response.List(c, rows, total, pageNumber, pageSize)
}

// @Summary Create server monitor
// @Tags monitors
// @Accept json
// @Security bearerauth
// @Param serverId path int true "Server ID"
// @Param body body request.MonitorInput true "Monitor fields"
// @Success 201 {object} models.Monitor
// @Failure 401 {object} response.UnauthorizedError
// @Failure 403 {object} response.ForbiddenError
// @Failure 404 {object} response.NotFoundError
// @Failure 422 {object} response.ValidationError
// @Produce json
// @Failure 503 {object} response.ErrorResponse "Service temporarily unavailable"
// @Failure 400 {object} response.ErrorResponse "Malformed JSON"
// @Failure 413 {object} response.ErrorResponse "Body exceeds 64 KiB"
// @Failure 415 {object} response.ErrorResponse "JSON content type required"
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
	if err := h.inventory.CreateMonitor(c.Request.Context(), &monitor); err != nil {
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
	monitor, err := h.inventory.Monitor(c.Request.Context(), server.ID, monitorID)
	if !response.ResourceError(c, err, "monitor") {
		return currentActor, nil, nil, false
	}
	return currentActor, server, monitor, true
}

// @Summary Get monitor
// @Tags monitors
// @Security bearerauth
// @Param serverId path int true "Server ID"
// @Param monitorId path int true "Monitor ID"
// @Success 200 {object} response.MonitorResponse
// @Failure 400 {object} response.BadRequestError
// @Failure 401 {object} response.UnauthorizedError
// @Failure 404 {object} response.NotFoundError
// @Produce json
// @Failure 503 {object} response.ErrorResponse "Service temporarily unavailable"
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
// @Param serverId path int true "Server ID"
// @Param monitorId path int true "Monitor ID"
// @Param body body request.MonitorPatchInput true "Fields to update"
// @Success 200 {object} models.Monitor
// @Failure 400 {object} response.BadRequestError
// @Failure 401 {object} response.UnauthorizedError
// @Failure 403 {object} response.ForbiddenError
// @Failure 404 {object} response.NotFoundError
// @Failure 422 {object} response.ValidationError
// @Produce json
// @Failure 503 {object} response.ErrorResponse "Service temporarily unavailable"
// @Failure 413 {object} response.ErrorResponse "Body exceeds 64 KiB"
// @Failure 415 {object} response.ErrorResponse "JSON content type required"
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

	err := h.inventory.UpdateMonitor(c.Request.Context(), monitor, input)
	if !response.ResourceError(c, err, "monitor") {
		return
	}
	c.JSON(200, monitor)
}

// @Summary Delete monitor and incidents
// @Tags monitors
// @Security bearerauth
// @Param serverId path int true "Server ID"
// @Param monitorId path int true "Monitor ID"
// @Success 204
// @Failure 400 {object} response.BadRequestError
// @Failure 401 {object} response.UnauthorizedError
// @Failure 403 {object} response.ForbiddenError
// @Failure 404 {object} response.NotFoundError
// @Produce json
// @Failure 503 {object} response.ErrorResponse "Service temporarily unavailable"
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
	if err := h.inventory.DeleteMonitor(c.Request.Context(), monitor.ID); err != nil {
		response.Error(c, err)
		return
	}
	c.Status(204)
}
