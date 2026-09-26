package request

type UserRoleInput struct {
	Role string `json:"role" binding:"required" example:"viewer"`
}
