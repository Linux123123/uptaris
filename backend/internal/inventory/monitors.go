package inventory

import (
	"context"

	"github.com/uptaris/uptaris/backend/internal/models"
	"gorm.io/gorm"
)

func (s *Service) Monitor(ctx context.Context, serverID uint, id uint) (*models.Monitor, error) {
	var value models.Monitor
	err := s.db.WithContext(ctx).Where("server_id = ?", serverID).First(&value, id).Error
	return &value, err
}

func (s *Service) Monitors(ctx context.Context, serverID uint, page Page, filters Filters) ([]models.Monitor, int, error) {
	query := s.db.WithContext(ctx).Where("server_id = ?", serverID)
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.Type != "" {
		query = query.Where("type = ?", filters.Type)
	}
	return listRows[models.Monitor](query, page, "created_at desc, id desc")
}

func (s *Service) CreateMonitor(ctx context.Context, value *models.Monitor) error {
	if !ValidateMonitor(value) {
		return ErrValidation
	}
	return s.db.WithContext(ctx).Create(value).Error
}

func (s *Service) DeleteMonitor(ctx context.Context, id uint) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("monitor_id = ?", id).Delete(&models.Incident{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.Monitor{}, id).Error
	})
}

type MonitorPatch struct {
	Name            *string `json:"name" example:"Public API"`
	Type            *string `json:"type" example:"http"`
	Target          *string `json:"target" example:"https://api.example.com/health"`
	IntervalSeconds *int    `json:"intervalSeconds" example:"60"`
	ExpectedHealth  *string `json:"expectedHealth" example:"200"`
	Status          *string `json:"status"`
}

func (s *Service) UpdateMonitor(ctx context.Context, monitor *models.Monitor, input MonitorPatch) error {
	return update(ctx, s, monitor, func() bool {
		if input.Name != nil {
			monitor.Name = *input.Name
		}
		if input.Type != nil {
			monitor.Type = *input.Type
		}
		if input.Target != nil {
			monitor.Target = *input.Target
		}
		if input.IntervalSeconds != nil {
			monitor.IntervalSeconds = *input.IntervalSeconds
		}
		if input.ExpectedHealth != nil {
			monitor.ExpectedHealth = *input.ExpectedHealth
		}
		if input.Status != nil {
			monitor.Status = *input.Status
		}
		return ValidateMonitor(monitor)
	})
}
