package models

import (
	"context"
	"time"

	"github.com/uptaris/uptaris/backend/internal/database"
	"gorm.io/gorm"
)

type Monitor struct {
	Model
	ServerID        uint       `gorm:"index;not null" json:"serverId,string"`
	Name            string     `json:"name" binding:"required,max=200"`
	Type            string     `gorm:"index" json:"type" binding:"required,oneof=http tcp icmp"`
	Target          string     `json:"target" binding:"required,max=2048"`
	IntervalSeconds int        `json:"intervalSeconds" binding:"min=1,max=86400"`
	ExpectedHealth  string     `json:"expectedHealth" binding:"required,max=200"`
	Status          string     `gorm:"index" json:"status" binding:"required,oneof=up down paused"`
	LastCheckedAt   *time.Time `json:"lastCheckedAt"`
}

func (value *Monitor) Create(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).Create(value).Error
}

func (value *Monitor) Update(ctx context.Context, db *gorm.DB, change func(*Monitor) error) error {
	return update(ctx, db, value, change)
}

func GetMonitor(ctx context.Context, db *gorm.DB, serverID, id uint) (*Monitor, error) {
	value := new(Monitor)
	err := db.WithContext(ctx).Where("server_id = ?", serverID).First(value, id).Error
	return value, err
}

func ListMonitors(ctx context.Context, db *gorm.DB, serverID uint, page Page, filters Filters) ([]Monitor, int, error) {
	query := db.WithContext(ctx).Where("server_id = ?", serverID)
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.Type != "" {
		query = query.Where("type = ?", filters.Type)
	}
	return listRows[Monitor](query, page, "created_at desc, id desc")
}

func (value *Monitor) Delete(ctx context.Context, db *gorm.DB) error {
	return database.Transaction(db.WithContext(ctx), func(tx *gorm.DB) error {
		if err := tx.Where("monitor_id = ?", value.ID).Delete(&Incident{}).Error; err != nil {
			return err
		}
		return tx.Delete(value).Error
	})
}
