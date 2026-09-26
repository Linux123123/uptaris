package handlers

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uptaris/uptaris/backend/internal/auth"
	"github.com/uptaris/uptaris/backend/internal/models"
	"github.com/uptaris/uptaris/backend/internal/request"
	"github.com/uptaris/uptaris/backend/internal/response"
)

// @Summary List monitor incidents
// @Tags incidents
// @Security bearerauth
// @Param serverId path int true "Server ID" minimum(1)
// @Param monitorId path int true "Monitor ID" minimum(1)
// @Param page query int false "Page number" default(1) minimum(1)
// @Param pageSize query int false "Results per page" default(20) minimum(1) maximum(100)
// @Param severity query string false "Severity" Enums(low,medium,high,critical)
// @Param status query string false "Status" Enums(open,acknowledged,resolved)
// @Success 200 {object} response.IncidentListResponse
// @Produce json
// @Failure 400 {object} response.ErrorResponse "invalid_id or invalid_query: invalid ID or query parameter"
// @Failure 401 {object} response.ErrorResponse "authentication_required, invalid_token, revoked_token, or stale_token"
// @Failure 404 {object} response.ErrorResponse "not_found: resource not found"
// @Failure 413 {object} response.ErrorResponse "body_too_large: request body exceeds 64 KiB"
// @Failure 500 {object} response.ErrorResponse "internal_error: unexpected server error"
// @Failure 503 {object} response.ErrorResponse "service_unavailable: database operation failed"
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
	filters := models.Filters{}
	filters.Severity, ok = request.Filter(c, "severity", "low", "medium", "high", "critical")
	if !ok {
		return
	}
	filters.Status, ok = request.Filter(c, "status", "open", "acknowledged", "resolved")
	if !ok {
		return
	}
	rows, total, err := models.ListIncidents(c.Request.Context(), h.db, monitor.ID, models.Page{Number: pageNumber, Size: pageSize}, filters)
	if !response.ResourceError(c, err, "incident") {
		return
	}
	response.List(c, rows, total, pageNumber, pageSize)
}

// @Summary Create monitor incident
// @Tags incidents
// @Accept json
// @Security bearerauth
// @Param serverId path int true "Server ID" minimum(1)
// @Param monitorId path int true "Monitor ID" minimum(1)
// @Param body body request.IncidentInput true "Incident fields"
// @Success 201 {object} models.Incident
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
	if !request.Valid(c, request.PrepareIncident(&incident)) {
		return
	}
	if err := incident.Create(c.Request.Context(), h.db); err != nil {
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
	incident, err := models.GetIncident(c.Request.Context(), h.db, monitor.ID, iid)
	if !response.ResourceError(c, err, "incident") {
		return currentActor, nil, false
	}
	return currentActor, incident, true
}

// @Summary Get incident
// @Tags incidents
// @Security bearerauth
// @Param serverId path int true "Server ID" minimum(1)
// @Param monitorId path int true "Monitor ID" minimum(1)
// @Param incidentId path int true "Incident ID" minimum(1)
// @Success 200 {object} response.IncidentResponse
// @Produce json
// @Failure 400 {object} response.ErrorResponse "invalid_id: invalid resource ID"
// @Failure 401 {object} response.ErrorResponse "authentication_required, invalid_token, revoked_token, or stale_token"
// @Failure 404 {object} response.ErrorResponse "not_found: resource not found"
// @Failure 413 {object} response.ErrorResponse "body_too_large: request body exceeds 64 KiB"
// @Failure 500 {object} response.ErrorResponse "internal_error: unexpected server error"
// @Failure 503 {object} response.ErrorResponse "service_unavailable: database operation failed"
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
// @Param serverId path int true "Server ID" minimum(1)
// @Param monitorId path int true "Monitor ID" minimum(1)
// @Param incidentId path int true "Incident ID" minimum(1)
// @Param body body request.IncidentPatchInput true "Fields to update"
// @Success 200 {object} models.Incident
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

	err := incident.Update(c.Request.Context(), h.db, func(value *models.Incident) error {
		input.Apply(value)
		return request.PrepareIncident(value)
	})
	if !response.ResourceError(c, err, "incident") {
		return
	}
	c.JSON(200, incident)
}

// @Summary Delete incident
// @Tags incidents
// @Security bearerauth
// @Param serverId path int true "Server ID" minimum(1)
// @Param monitorId path int true "Monitor ID" minimum(1)
// @Param incidentId path int true "Incident ID" minimum(1)
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
	if err := incident.Delete(c.Request.Context(), h.db); err != nil {
		response.Error(c, err)
		return
	}
	c.Status(204)
}
