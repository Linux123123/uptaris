package models

import "time"

// TwoFactor stores active or pending authenticator enrollment.
type TwoFactor struct {
	Model
	UserID           uint       `json:"-"`
	SecretCiphertext []byte     `json:"-"`
	EnabledAt        *time.Time `json:"-"`
	ExpiresAt        time.Time  `json:"-"`
	FailedAttempts   int        `json:"-"`
	LockedUntil      *time.Time `json:"-"`
	LastUsedStep     int64      `json:"-"`
}

func (TwoFactor) TableName() string { return "two_factors" }
