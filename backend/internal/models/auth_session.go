package models

import "time"

type AuthSession struct {
	Model
	UserID           uint   `gorm:"index;not null"`
	RefreshTokenHash string `gorm:"uniqueIndex;not null"`
	ExpiresAt        time.Time
	RevokedAt        *time.Time `gorm:"index"`
}
