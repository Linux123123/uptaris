package models

import "time"

// TwoFactorChallenge authorizes only completion of a first-factor sign-in.
type TwoFactorChallenge struct {
	Model
	UserID    uint      `json:"-"`
	TokenHash string    `json:"-"`
	ExpiresAt time.Time `json:"-"`
	Attempts  int       `json:"-"`
}

func (TwoFactorChallenge) TableName() string { return "two_factor_challenges" }
