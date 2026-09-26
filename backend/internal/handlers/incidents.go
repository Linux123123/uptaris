package handlers

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uptaris/uptaris/backend/internal/auth"
	"github.com/uptaris/uptaris/backend/internal/inventory"
	"github.com/uptaris/uptaris/backend/internal/models"
	"github.com/uptaris/uptaris/backend/internal/request"
	"github.com/uptaris/uptaris/backend/internal/response"
)

// @Summary List monitor incidents
// @Tags incidents
// @Security bearerauth
// @Param serverId path int true "Server ID"
// @Param monitorId path int true "Monitor ID"
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Results per page" default(20)
// @Param severity query string false "Severity" Enums(low,medium,high,critical)
// @Param status query string false "Status" Enums(open,acknowledged,resolved)
// @Success 200 {object} response.IncidentListResponse
// @Failure 400 {object} response.BadRequestError
// @Failure 401 {object} response.UnauthorizedError
// @Failure 404 {object} response.NotFoundError
// @Produce json
// @Failure 503 {object} response.ErrorResponse "Service temporarily unavailable"
// @Router /servers/{serverId}/monitors/{monitorId}/incidents [get]
func (h *Handlers) Incidents(c *gin.Context) {
	_, _, monitor, ok := h.oneMonitor(c)
	if !ok {
		return
	}
	pageNumber, pageSize, ok := request.Page(c)
	if !ok {
		return
	}
	filters := inventory.Filters{}
	filters.Severity, ok = request.Filter(c, "severity", "low", "medium", "high", "critical")
	if !ok {
		return
	}
	filters.Status, ok = request.Filter(c, "status", "open", "acknowledged", "resolved")
	if !ok {
		return
	}
	rows, total, err := h.inventory.Incidents(c.Request.Context(), monitor.ID, inventory.Page{Number: pageNumber, Size: pageSize}, filters)
	if !response.ResourceError(c, err, "incident") {
		return
	}
	response.List(c, rows, total, pageNumber, pageSize)
}

// @Summary Create monitor incident
// @Tags incidents
// @Accept json
// @Security bearerauth
// @Param serverId path int true "Server ID"
// @Param monitorId path int true "Monitor ID"
// @Param body body request.IncidentInput true "Incident fields"
// @Success 201 {object} models.Incident
// @Failure 401 {object} response.UnauthorizedError
// @Failure 403 {object} response.ForbiddenError
// @Failure 404 {object} response.NotFoundError
// @Failure 422 {object} response.ValidationError
// @Produce json
// @Failure 503 {object} response.ErrorResponse "Service temporarily unavailable"
// @Failure 400 {object} response.ErrorResponse "Malformed JSON"
// @Failure 413 {object} response.ErrorResponse "Body exceeds 64 KiB"
// @Failure 415 {object} response.ErrorResponse "JSON content type required"
// @Router /servers/{serverId}/monitors/{monitorId}/incidents [post]
func (h *Handlers) CreateIncident(c *gin.Context) {
	currentActor, _, monitor, ok := h.oneMonitor(c)
	if !ok {
		return
	}
	if currentActor.Role == "viewer" {
		response.Fail(c, 403, "forbidden", "viewer is read only")
		return
	}
	var input request.IncidentInput
	if !request.JSON(c, &input) {
		return
	}
	started := time.Now()
	if input.StartedAt != nil {
		started = *input.StartedAt
	}
	incident := models.Incident{
		MonitorID:   monitor.ID,
		Title:       input.Title,
		Description: input.Description,
		Severity:    input.Severity,
		Status:      defaultOf(input.Status, "open"),
		StartedAt:   started,
		ResolvedAt:  input.ResolvedAt,
	}
	if err := h.inventory.CreateIncident(c.Request.Context(), &incident); err != nil {
		response.Error(c, err)
		return
	}
	c.Header("Location", c.Request.URL.Path+"/"+strconv.FormatUint(uint64(incident.ID), 10))
	c.JSON(201, incident)
}

func (h *Handlers) oneIncident(c *gin.Context) (auth.Identity, *models.Incident, bool) {
	currentActor, _, monitor, ok := h.oneMonitor(c)
	if !ok {
		return currentActor, nil, false
	}
	iid, ok := request.ID(c, "incidentId")
	if !ok {
		return currentActor, nil, false
	}
	incident, err := h.inventory.Incident(c.Request.Context(), monitor.ID, iid)
	if !response.ResourceError(c, err, "incident") {
		return currentActor, nil, false
	}
	return currentActor, incident, true
}

// @Summary Get incident
// @Tags incidents
// @Security bearerauth
// @Param serverId path int true "Server ID"
// @Param monitorId path int true "Monitor ID"
// @Param incidentId path int true "Incident ID"
// @Success 200 {object} response.IncidentResponse
// @Failure 400 {object} response.BadRequestError
// @Failure 401 {object} response.UnauthorizedError
// @Failure 404 {object} response.NotFoundError
// @Produce json
// @Failure 503 {object} response.ErrorResponse "Service temporarily unavailable"
// @Router /servers/{serverId}/monitors/{monitorId}/incidents/{incidentId} [get]
func (h *Handlers) GetIncident(c *gin.Context) {
	_, incident, ok := h.oneIncident(c)
	if ok {
		c.JSON(200, response.IncidentResponse{Data: *incident, Links: map[string]string{"self": c.Request.URL.Path}})
	}
}

// @Summary Update incident
// @Tags incidents
// @Accept json
// @Security bearerauth
// @Param serverId path int true "Server ID"
// @Param monitorId path int true "Monitor ID"
// @Param incidentId path int true "Incident ID"
// @Param body body request.IncidentPatchInput true "Fields to update"
// @Success 200 {object} models.Incident
// @Failure 400 {object} response.BadRequestError
// @Failure 401 {object} response.UnauthorizedError
// @Failure 403 {object} response.ForbiddenError
// @Failure 404 {object} response.NotFoundError
// @Failure 422 {object} response.ValidationError
// @Produce json
// @Failure 503 {object} response.ErrorResponse "Service temporarily unavailable"
// @Failure 413 {object} response.ErrorResponse "Body exceeds 64 KiB"
// @Failure 415 {object} response.ErrorResponse "JSON content type required"
// @Router /servers/{serverId}/monitors/{monitorId}/incidents/{incidentId} [patch]
func (h *Handlers) UpdateIncident(c *gin.Context) {
	currentActor, incident, ok := h.oneIncident(c)
	if !ok {
		return
	}
	if currentActor.Role == "viewer" {
		response.Fail(c, 403, "forbidden", "viewer is read only")
		return
	}
	var input request.IncidentPatchInput
	if !request.JSON(c, &input) {
		return
	}

	err := h.inventory.UpdateIncident(c.Request.Context(), incident, input)
	if !response.ResourceError(c, err, "incident") {
		return
	}
	c.JSON(200, incident)
}

// @Summary Delete incident
// @Tags incidents
// @Security bearerauth
// @Param serverId path int true "Server ID"
// @Param monitorId path int true "Monitor ID"
// @Param incidentId path int true "Incident ID"
// @Success 204
// @Failure 400 {object} response.BadRequestError
// @Failure 401 {object} response.UnauthorizedError
// @Failure 403 {object} response.ForbiddenError
// @Failure 404 {object} response.NotFoundError
// @Produce json
// @Failure 503 {object} response.ErrorResponse "Service temporarily unavailable"
// @Router /servers/{serverId}/monitors/{monitorId}/incidents/{incidentId} [delete]
func (h *Handlers) DeleteIncident(c *gin.Context) {
	currentActor, incident, ok := h.oneIncident(c)
	if !ok {
		return
	}
	if currentActor.Role == "viewer" {
		response.Fail(c, 403, "forbidden", "viewer is read only")
		return
	}
	if err := h.inventory.DeleteIncident(c.Request.Context(), incident.ID); err != nil {
		response.Error(c, err)
		return
	}
	c.Status(204)
}
