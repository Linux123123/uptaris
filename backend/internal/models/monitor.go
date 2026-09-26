package models

import (
	"time"
)

type Monitor struct {
	Model
	ServerID        uint       `gorm:"index;not null" json:"serverId"`
	Name            string     `json:"name"`
	Type            string     `gorm:"index" json:"type"`
	Target          string     `json:"target"`
	IntervalSeconds int        `json:"intervalSeconds"`
	ExpectedHealth  string     `json:"expectedHealth"`
	Status          string     `gorm:"index" json:"status"`
	LastCheckedAt   *time.Time `json:"lastCheckedAt"`
}
