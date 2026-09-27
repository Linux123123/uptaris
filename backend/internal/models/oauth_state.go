package models

import "time"

type OAuthState struct {
	Model
	StateHash              string `gorm:"uniqueIndex;not null"`
	Provider               string `gorm:"not null"`
	Intent                 string `gorm:"not null"`
	UserID                 *uint
	PKCEVerifierCiphertext []byte    `json:"-"`
	ExpiresAt              time.Time `gorm:"index;not null"`
	ConsumedAt             *time.Time
}

func (OAuthState) TableName() string { return "oauth_states" }
