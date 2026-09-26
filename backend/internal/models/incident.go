package models

import (
	"time"
)

type Incident struct {
	Model
	MonitorID   uint       `gorm:"index;not null" json:"monitorId"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Severity    string     `gorm:"index" json:"severity"`
	Status      string     `gorm:"index" json:"status"`
	StartedAt   time.Time  `json:"startedAt"`
	ResolvedAt  *time.Time `json:"resolvedAt"`
}
