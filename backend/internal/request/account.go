package request

type PasswordChange struct {
	CurrentPassword string `json:"currentPassword" binding:"omitempty,max=128"`
	NewPassword     string `json:"newPassword" binding:"required,min=8,max=128"`
}
