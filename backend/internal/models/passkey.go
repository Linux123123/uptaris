package models

import (
	"encoding/json"
	"time"
)

type Passkey struct {
	ID           uint            `gorm:"primaryKey" json:"id,string"`
	UserID       uint            `gorm:"index;not null"`
	RPID         string          `gorm:"not null"`
	CredentialID []byte          `gorm:"uniqueIndex;not null"`
	Credential   json.RawMessage `gorm:"type:jsonb;not null"`
	Name         string          `gorm:"not null"`
	LastUsedAt   *time.Time      `json:"lastUsedAt"`
	CreatedAt    time.Time       `json:"createdAt"`
	UpdatedAt    time.Time       `json:"-"`
}
