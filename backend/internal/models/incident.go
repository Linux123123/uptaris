package models

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type Incident struct {
	Model
	MonitorID   uint       `gorm:"index;not null" json:"monitorId,string"`
	Title       string     `json:"title" binding:"required,max=200"`
	Description string     `json:"description" binding:"max=10000"`
	Severity    string     `gorm:"index" json:"severity" binding:"required,oneof=low medium high critical"`
	Status      string     `gorm:"index" json:"status" binding:"required,oneof=open acknowledged resolved"`
	StartedAt   time.Time  `json:"startedAt" binding:"required"`
	ResolvedAt  *time.Time `json:"resolvedAt"`
}

func (value *Incident) Create(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).Create(value).Error
}

func (value *Incident) Update(ctx context.Context, db *gorm.DB, change func(*Incident) error) error {
	return update(ctx, db, value, change)
}

func GetIncident(ctx context.Context, db *gorm.DB, monitorID, id uint) (*Incident, error) {
	value := new(Incident)
	err := db.WithContext(ctx).Where("monitor_id = ?", monitorID).First(value, id).Error

	return value, err
}

func ListIncidents(ctx context.Context, db *gorm.DB, monitorID uint, page, pageSize int, status, severity string) ([]Incident, int, error) {
	query := db.WithContext(ctx).Where("monitor_id = ?", monitorID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if severity != "" {
		query = query.Where("severity = ?", severity)
	}

	return listRows[Incident](query, page, pageSize, "started_at desc, id desc")
}

func (value *Incident) Delete(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).Delete(value).Error
}
