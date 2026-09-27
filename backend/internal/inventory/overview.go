package inventory

import (
	"context"

	"github.com/uptaris/uptaris/backend/internal/models"
	"gorm.io/gorm"
)

type Summary struct {
	Servers       int64 `json:"servers"`
	Monitors      int64 `json:"monitors"`
	OpenIncidents int64 `json:"openIncidents"`
}

type IncidentRow struct {
	models.Incident
	ServerID    uint   `json:"serverId,string"`
	MonitorName string `json:"monitorName"`
}

func activeMonitors(ctx context.Context, db *gorm.DB, ownerID *uint) *gorm.DB {
	query := db.WithContext(ctx).Model(&models.Monitor{}).
		Joins("JOIN servers ON servers.id = monitors.server_id AND servers.deleted_at IS NULL")
	if ownerID != nil {
		query = query.Where("servers.owner_id = ?", *ownerID)
	}

	return query
}

func activeIncidents(ctx context.Context, db *gorm.DB, ownerID *uint) *gorm.DB {
	query := db.WithContext(ctx).Model(&models.Incident{}).
		Joins("JOIN monitors ON monitors.id = incidents.monitor_id AND monitors.deleted_at IS NULL").
		Joins("JOIN servers ON servers.id = monitors.server_id AND servers.deleted_at IS NULL")
	if ownerID != nil {
		query = query.Where("servers.owner_id = ?", *ownerID)
	}

	return query
}

func AggregateSummary(ctx context.Context, db *gorm.DB, ownerID *uint) (Summary, error) {
	var result Summary

	query := db.WithContext(ctx).Model(&models.Server{})
	if ownerID != nil {
		query = query.Where("owner_id = ?", *ownerID)
	}

	if err := query.Count(&result.Servers).Error; err != nil {
		return result, err
	}

	if err := activeMonitors(ctx, db, ownerID).Count(&result.Monitors).Error; err != nil {
		return result, err
	}

	if err := activeIncidents(ctx, db, ownerID).Where("incidents.status != ?", "resolved").Count(&result.OpenIncidents).Error; err != nil {
		return result, err
	}

	return result, nil
}

func ListIncidentOverview(ctx context.Context, db *gorm.DB, ownerID *uint, page, pageSize int, status, severity string) ([]IncidentRow, int, error) {
	query := activeIncidents(ctx, db, ownerID)
	if status != "" {
		query = query.Where("incidents.status = ?", status)
	}

	if severity != "" {
		query = query.Where("incidents.severity = ?", severity)
	}

	var total int64

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	rows := []IncidentRow{}
	err := query.Select("incidents.*, monitors.server_id, monitors.name AS monitor_name").
		Order("incidents.started_at DESC, incidents.id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&rows).Error

	return rows, int(total), err
}
