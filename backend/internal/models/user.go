package models

type User struct {
	Model
	Email          string  `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash   *string `json:"-"`
	WebAuthnHandle []byte  `gorm:"column:webauthn_handle" json:"-"`
	Role           string  `gorm:"not null;default:viewer" json:"role"`
}
