package inventory

import (
	"context"
	"time"

	"github.com/uptaris/uptaris/backend/internal/models"
)

func (s *Service) Incident(ctx context.Context, monitorID uint, id uint) (*models.Incident, error) {
	var value models.Incident
	err := s.db.WithContext(ctx).Where("monitor_id = ?", monitorID).First(&value, id).Error
	return &value, err
}

func (s *Service) Incidents(ctx context.Context, monitorID uint, page Page, filters Filters) ([]models.Incident, int, error) {
	query := s.db.WithContext(ctx).Where("monitor_id = ?", monitorID)
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.Severity != "" {
		query = query.Where("severity = ?", filters.Severity)
	}
	return listRows[models.Incident](query, page, "started_at desc, id desc")
}

func (s *Service) CreateIncident(ctx context.Context, value *models.Incident) error {
	if !ValidateIncident(value) {
		return ErrValidation
	}
	return s.db.WithContext(ctx).Create(value).Error
}

func (s *Service) DeleteIncident(ctx context.Context, id uint) error {
	return s.db.WithContext(ctx).Delete(&models.Incident{}, id).Error
}

type IncidentPatch struct {
	Title       *string    `json:"title" example:"API maintenance"`
	Description *string    `json:"description" example:"Primary customer API"`
	Severity    *string    `json:"severity" example:"low"`
	Status      *string    `json:"status"`
	StartedAt   *time.Time `json:"startedAt" example:"2026-09-26T10:00:00Z"`
	ResolvedAt  *time.Time `json:"resolvedAt" example:"2026-09-26T11:00:00Z"`
}

func (s *Service) UpdateIncident(ctx context.Context, incident *models.Incident, input IncidentPatch) error {
	return update(ctx, s, incident, func() bool {
		if input.Title != nil {
			incident.Title = *input.Title
		}
		if input.Description != nil {
			incident.Description = *input.Description
		}
		if input.Severity != nil {
			incident.Severity = *input.Severity
		}
		if input.Status != nil {
			incident.Status = *input.Status
		}
		if input.StartedAt != nil {
			incident.StartedAt = *input.StartedAt
		}
		if input.ResolvedAt != nil {
			incident.ResolvedAt = input.ResolvedAt
		}
		return ValidateIncident(incident)
	})
}
