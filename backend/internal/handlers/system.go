package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/uptaris/uptaris/backend/internal/auth"
	"github.com/uptaris/uptaris/backend/internal/middleware"
	"github.com/uptaris/uptaris/backend/internal/models"
	"github.com/uptaris/uptaris/backend/internal/request"
	"github.com/uptaris/uptaris/backend/internal/response"
)

// @Summary Get owner-scoped dashboard
// @Tags system
// @Produce json
// @Security bearerauth
// @Success 200 {object} response.DashboardResponse
// @Failure 401 {object} response.ErrorResponse "authentication_required, invalid_token, revoked_token, or stale_token"
// @Failure 413 {object} response.ErrorResponse "body_too_large: request body exceeds 64 KiB"
// @Failure 500 {object} response.ErrorResponse "internal_error: unexpected server error"
// @Failure 503 {object} response.ErrorResponse "service_unavailable: database operation failed"
// @Router /dashboard [get]
func (h *Handlers) Dashboard(c *gin.Context) {
	identity, ok := middleware.Require(c, h.auth, "viewer", "operator", "admin")
	if !ok {
		return
	}
	h.summary(c, &identity)
}

// @Summary Get public aggregate status
// @Tags system
// @Produce json
// @Success 200 {object} response.StatusResponse
// @Failure 413 {object} response.ErrorResponse "body_too_large: request body exceeds 64 KiB"
// @Failure 500 {object} response.ErrorResponse "internal_error: unexpected server error"
// @Failure 503 {object} response.ErrorResponse "service_unavailable: database operation failed"
// @Router /status [get]
func (h *Handlers) Status(c *gin.Context) {
	h.summary(c, nil)
}

// @Summary List accessible incidents across servers
// @Tags incidents
// @Produce json
// @Security bearerauth
// @Param page query int false "Page" default(1) minimum(1)
// @Param pageSize query int false "Page size" default(20) maximum(100) minimum(1)
// @Param status query string false "Status" Enums(open,acknowledged,resolved)
// @Param severity query string false "Severity" Enums(low,medium,high,critical)
// @Success 200 {object} response.IncidentOverviewResponse
// @Failure 400 {object} response.ErrorResponse "invalid_query: invalid query parameter"
// @Failure 401 {object} response.ErrorResponse "authentication_required, invalid_token, revoked_token, or stale_token"
// @Failure 413 {object} response.ErrorResponse "body_too_large: request body exceeds 64 KiB"
// @Failure 500 {object} response.ErrorResponse "internal_error: unexpected server error"
// @Failure 503 {object} response.ErrorResponse "service_unavailable: database operation failed"
// @Router /incidents [get]
func (h *Handlers) AllIncidents(c *gin.Context) {
	identity, ok := middleware.Require(c, h.auth, "viewer", "operator", "admin")
	if !ok {
		return
	}
	page, size, ok := request.Page(c)
	if !ok {
		return
	}
	status, ok := request.Filter(c, "status", "open", "acknowledged", "resolved")
	if !ok {
		return
	}
	severity, ok := request.Filter(c, "severity", "low", "medium", "high", "critical")
	if !ok {
		return
	}
	rows, total, err := models.ListIncidentOverview(c.Request.Context(), h.db, ownerScope(&identity), models.Page{Number: page, Size: size}, models.Filters{Status: status, Severity: severity})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.List(c, rows, total, page, size)
}

func (h *Handlers) summary(c *gin.Context, identity *auth.Identity) {
	result, err := models.AggregateSummary(c.Request.Context(), h.db, ownerScope(identity))
	if err != nil {
		response.Error(c, err)
		return
	}
	c.JSON(200, result)
}

func (h *Handlers) Health(c *gin.Context) {
	if err := h.health(c.Request.Context()); err != nil {
		c.JSON(503, gin.H{"status": "unavailable"})
		return
	}
	c.JSON(200, gin.H{"status": "ok"})
}

func NotFound(c *gin.Context) { response.Fail(c, 404, "not_found", "route not found") }

func MethodNotAllowed(c *gin.Context) {
	response.Fail(c, 405, "method_not_allowed", "method not allowed")
}

func ownerScope(identity *auth.Identity) *uint {
	if identity == nil || identity.Role == "admin" {
		return nil
	}
	return &identity.ID
}
