package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/uptaris/uptaris/backend/internal/auth"
	"github.com/uptaris/uptaris/backend/internal/inventory"
	"github.com/uptaris/uptaris/backend/internal/middleware"
	"github.com/uptaris/uptaris/backend/internal/request"
	"github.com/uptaris/uptaris/backend/internal/response"
)

// @Summary Get owner-scoped dashboard
// @Tags system
// @Produce json
// @Security bearerauth
// @Success 200 {object} response.DashboardResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 503 {object} response.ErrorResponse
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
// @Failure 503 {object} response.ErrorResponse
// @Router /status [get]
func (h *Handlers) Status(c *gin.Context) {
	h.summary(c, nil)
}

// @Summary List accessible incidents across servers
// @Tags incidents
// @Produce json
// @Security bearerauth
// @Param page query int false "Page" default(1)
// @Param pageSize query int false "Page size" default(20) maximum(100)
// @Param status query string false "Status" Enums(open,acknowledged,resolved)
// @Param severity query string false "Severity" Enums(low,medium,high,critical)
// @Success 200 {object} response.IncidentOverviewResponse
// @Failure 400 {object} response.BadRequestError
// @Failure 401 {object} response.UnauthorizedError
// @Failure 503 {object} response.ErrorResponse
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
	rows, total, err := h.inventory.IncidentsOverview(c.Request.Context(), identity, inventory.Page{Number: page, Size: size}, inventory.Filters{Status: status, Severity: severity})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.List(c, rows, total, page, size)
}

func (h *Handlers) summary(c *gin.Context, identity *auth.Identity) {
	result, err := h.inventory.Summary(c.Request.Context(), identity)
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
