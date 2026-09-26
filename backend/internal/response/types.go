package response

import (
	"github.com/uptaris/uptaris/backend/internal/models"
)

// ErrorResponse is returned for every non-success API response.
type ErrorResponse struct {
	Error struct {
		Code    string `json:"code" binding:"required"`
		Message string `json:"message" binding:"required"`
	} `json:"error" binding:"required"`
}

type Pagination struct {
	Page       int `json:"page" example:"1"`
	PageSize   int `json:"pageSize" example:"20"`
	Total      int `json:"total" example:"5"`
	TotalPages int `json:"totalPages" example:"1"`
}

type Links struct {
	Self     string  `json:"self"`
	Next     *string `json:"next"`
	Previous *string `json:"previous"`
}

type ServerListResponse struct {
	Data       []models.Server `json:"data"`
	Pagination Pagination      `json:"pagination"`
	Links      Links           `json:"links"`
}

type MonitorListResponse struct {
	Data       []models.Monitor `json:"data"`
	Pagination Pagination       `json:"pagination"`
	Links      Links            `json:"links"`
}

type IncidentListResponse struct {
	Data       []models.Incident `json:"data"`
	Pagination Pagination        `json:"pagination"`
	Links      Links             `json:"links"`
}

type UserListResponse struct {
	Data       []models.User `json:"data"`
	Pagination Pagination    `json:"pagination"`
	Links      Links         `json:"links"`
}

type AuthResponse struct {
	AccessToken string       `json:"accessToken"`
	User        UserResponse `json:"user"`
}

type DashboardResponse struct {
	Servers       int64 `json:"servers"`
	Monitors      int64 `json:"monitors"`
	OpenIncidents int64 `json:"openIncidents"`
}

type StatusResponse struct {
	Servers       int64 `json:"servers"`
	Monitors      int64 `json:"monitors"`
	OpenIncidents int64 `json:"openIncidents"`
}

type CreateUserResponse struct {
	ID    uint   `json:"id,string"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type UserResponse struct {
	ID    uint   `json:"id,string"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type ServerResponse struct {
	Data  models.Server     `json:"data"`
	Links map[string]string `json:"_links"`
}

type MonitorResponse struct {
	Data  models.Monitor    `json:"data"`
	Links map[string]string `json:"_links"`
}

type IncidentResponse struct {
	Data  models.Incident   `json:"data"`
	Links map[string]string `json:"_links"`
}

func User(user *models.User) UserResponse {
	return UserResponse{ID: user.ID, Email: user.Email, Role: user.Role}
}

type IncidentOverviewResponse struct {
	Data       []models.IncidentRow `json:"data"`
	Pagination Pagination           `json:"pagination"`
	Links      Links                `json:"links"`
}

// UserRecord is the public user model returned by administration operations.
type UserRecord = models.User
