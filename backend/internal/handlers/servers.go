package handlers

import (
	"strconv"

	"github.com/uptaris/uptaris/backend/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/uptaris/uptaris/backend/internal/auth"
	"github.com/uptaris/uptaris/backend/internal/inventory"
	"github.com/uptaris/uptaris/backend/internal/models"
	"github.com/uptaris/uptaris/backend/internal/request"
	"github.com/uptaris/uptaris/backend/internal/response"
)

// @Summary List servers
// @Tags servers
// @Produce json
// @Security bearerauth
// @Param page query int false "Page number" default(1) minimum(1)
// @Param pageSize query int false "Results per page" default(20) maximum(100)
// @Param status query string false "Status" Enums(up,down,paused)
// @Success 200 {object} response.ServerListResponse
// @Failure 400 {object} response.BadRequestError "Malformed ID or query parameter"
// @Failure 401 {object} response.UnauthorizedError "Missing, invalid, expired, or revoked token"
// @Failure 503 {object} response.ErrorResponse "Service temporarily unavailable"
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
	filters := inventory.Filters{}
	filters.Status, ok = request.Filter(c, "status", "up", "down", "paused")
	if !ok {
		return
	}
	rows, total, err := h.inventory.Servers(c.Request.Context(), currentActor, inventory.Page{Number: pageNumber, Size: pageSize}, filters)
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
// @Failure 401 {object} response.UnauthorizedError
// @Failure 403 {object} response.ForbiddenError "Viewer role cannot create resources"
// @Failure 422 {object} response.ValidationError
// @Failure 503 {object} response.ErrorResponse "Service temporarily unavailable"
// @Failure 400 {object} response.ErrorResponse "Malformed JSON"
// @Failure 413 {object} response.ErrorResponse "Body exceeds 64 KiB"
// @Failure 415 {object} response.ErrorResponse "JSON content type required"
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
	if err := h.inventory.CreateServer(c.Request.Context(), &server); err != nil {
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
	server, err := h.inventory.Server(c.Request.Context(), currentActor, serverID)
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
// @Failure 400 {object} response.BadRequestError
// @Failure 401 {object} response.UnauthorizedError
// @Failure 404 {object} response.NotFoundError
// @Failure 503 {object} response.ErrorResponse "Service temporarily unavailable"
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
// @Failure 400 {object} response.BadRequestError
// @Failure 401 {object} response.UnauthorizedError
// @Failure 403 {object} response.ForbiddenError
// @Failure 404 {object} response.NotFoundError
// @Failure 422 {object} response.ValidationError
// @Failure 503 {object} response.ErrorResponse "Service temporarily unavailable"
// @Failure 413 {object} response.ErrorResponse "Body exceeds 64 KiB"
// @Failure 415 {object} response.ErrorResponse "JSON content type required"
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

	err := h.inventory.UpdateServer(c.Request.Context(), server, input)
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
// @Failure 400 {object} response.BadRequestError
// @Failure 401 {object} response.UnauthorizedError
// @Failure 403 {object} response.ForbiddenError
// @Failure 404 {object} response.NotFoundError
// @Produce json
// @Failure 503 {object} response.ErrorResponse "Service temporarily unavailable"
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
	if err := h.inventory.DeleteServer(c.Request.Context(), server.ID); err != nil {
		response.Error(c, err)
		return
	}
	c.Status(204)
}
