package request

type UserRoleInput struct {
	Role string `json:"role" binding:"required,oneof=viewer operator admin" example:"viewer" enums:"viewer,operator,admin"`
}
