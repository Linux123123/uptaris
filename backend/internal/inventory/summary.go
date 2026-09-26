package inventory

import (
	"context"

	"github.com/uptaris/uptaris/backend/internal/auth"
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
	ServerID    uint   `json:"serverId"`
	MonitorName string `json:"monitorName"`
}

func (s *Service) activeMonitors(ctx context.Context, currentActor *auth.Identity) *gorm.DB {
	query := s.db.WithContext(ctx).Model(&models.Monitor{}).
		Joins("JOIN servers ON servers.id = monitors.server_id AND servers.deleted_at IS NULL")
	if currentActor != nil && currentActor.Role != "admin" {
		query = query.Where("servers.owner_id = ?", currentActor.ID)
	}
	return query
}

func (s *Service) activeIncidents(ctx context.Context, currentActor *auth.Identity) *gorm.DB {
	query := s.db.WithContext(ctx).Model(&models.Incident{}).
		Joins("JOIN monitors ON monitors.id = incidents.monitor_id AND monitors.deleted_at IS NULL").
		Joins("JOIN servers ON servers.id = monitors.server_id AND servers.deleted_at IS NULL")
	if currentActor != nil && currentActor.Role != "admin" {
		query = query.Where("servers.owner_id = ?", currentActor.ID)
	}
	return query
}

func (s *Service) Summary(ctx context.Context, currentActor *auth.Identity) (Summary, error) {
	var result Summary
	query := s.db.WithContext(ctx).Model(&models.Server{})
	if currentActor != nil && currentActor.Role != "admin" {
		query = query.Where("owner_id = ?", currentActor.ID)
	}
	if err := query.Count(&result.Servers).Error; err != nil {
		return result, err
	}
	if err := s.activeMonitors(ctx, currentActor).Count(&result.Monitors).Error; err != nil {
		return result, err
	}
	if err := s.activeIncidents(ctx, currentActor).Where("incidents.status != ?", "resolved").Count(&result.OpenIncidents).Error; err != nil {
		return result, err
	}
	return result, nil
}

func (s *Service) IncidentsOverview(ctx context.Context, identity auth.Identity, page Page, filters Filters) ([]IncidentRow, int, error) {
	query := s.activeIncidents(ctx, &identity)
	if filters.Status != "" {
		query = query.Where("incidents.status = ?", filters.Status)
	}
	if filters.Severity != "" {
		query = query.Where("incidents.severity = ?", filters.Severity)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	rows := []IncidentRow{}
	err := query.Select("incidents.*, monitors.server_id, monitors.name AS monitor_name").Order("incidents.started_at DESC, incidents.id DESC").Offset((page.Number - 1) * page.Size).Limit(page.Size).Scan(&rows).Error
	return rows, int(total), err
}
