package models

type User struct {
	Model
	Email        string `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string `json:"-"`
	Role         string `gorm:"not null;default:viewer" json:"role"`
}
