package inventory

import (
	"context"

	"github.com/uptaris/uptaris/backend/internal/auth"
	"github.com/uptaris/uptaris/backend/internal/database"
	"github.com/uptaris/uptaris/backend/internal/models"
	"gorm.io/gorm"
)

func (s *Service) Server(ctx context.Context, actor auth.Identity, id uint) (*models.Server, error) {
	var value models.Server
	err := serverScope(s.db.WithContext(ctx), actor).First(&value, id).Error
	return &value, err
}

func (s *Service) Servers(ctx context.Context, actor auth.Identity, page Page, filters Filters) ([]models.Server, int, error) {
	query := serverScope(s.db.WithContext(ctx), actor)
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	return listRows[models.Server](query, page, "created_at desc, id desc")
}

func (s *Service) CreateServer(ctx context.Context, value *models.Server) error {
	if !ValidateServer(value) {
		return ErrValidation
	}
	return s.db.WithContext(ctx).Create(value).Error
}

func (s *Service) DeleteServer(ctx context.Context, id uint) error {
	return database.Transaction(s.db.WithContext(ctx), func(tx *gorm.DB) error {
		monitors := tx.Model(&models.Monitor{}).Select("id").Where("server_id = ?", id)
		if err := tx.Where("monitor_id IN (?)", monitors).Delete(&models.Incident{}).Error; err != nil {
			return err
		}
		if err := tx.Where("server_id = ?", id).Delete(&models.Monitor{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.Server{}, id).Error
	})
}

type ServerPatch struct {
	Name            *string `json:"name" example:"Public API"`
	Address         *string `json:"address" example:"api.example.com"`
	OperatingSystem *string `json:"operatingSystem" example:"Ubuntu 24.04"`
	Description     *string `json:"description" example:"Primary customer API"`
	Status          *string `json:"status"`
}

func (s *Service) UpdateServer(ctx context.Context, server *models.Server, input ServerPatch) error {
	return update(ctx, s, server, func() bool {
		if input.Name != nil {
			server.Name = *input.Name
		}
		if input.Address != nil {
			server.Address = *input.Address
		}
		if input.OperatingSystem != nil {
			server.OperatingSystem = *input.OperatingSystem
		}
		if input.Description != nil {
			server.Description = *input.Description
		}
		if input.Status != nil {
			server.Status = *input.Status
		}
		return ValidateServer(server)
	})
}
