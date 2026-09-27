package models

import (
	"encoding/json"
	"time"
)

type PasskeyCeremony struct {
	ID          uint   `gorm:"primaryKey"`
	TokenHash   string `gorm:"uniqueIndex;not null"`
	Kind        string `gorm:"not null"`
	UserID      *uint
	SessionID   *uint
	SessionData json.RawMessage `gorm:"type:jsonb;not null"`
	ExpiresAt   time.Time
	ConsumedAt  *time.Time
}
