package request

import (
	"strings"

	"github.com/uptaris/uptaris/backend/internal/models"
)

type ServerInput struct {
	Name            string `json:"name" binding:"required,max=200" example:"Public API"`
	Address         string `json:"address" binding:"required,max=2048" example:"api.example.com"`
	OperatingSystem string `json:"operatingSystem" binding:"required,max=200" example:"Ubuntu 24.04"`
	Description     string `json:"description" binding:"max=10000" example:"Primary customer API"`
	Status          string `json:"status" binding:"omitempty,oneof=up down paused" example:"up" enums:"up,down,paused"`
}

type ServerPatchInput struct {
	Name            *string `json:"name" binding:"omitempty,max=200" example:"Public API"`
	Address         *string `json:"address" binding:"omitempty,max=2048" example:"api.example.com"`
	OperatingSystem *string `json:"operatingSystem" binding:"omitempty,max=200" example:"Ubuntu 24.04"`
	Description     *string `json:"description" binding:"omitempty,max=10000" example:"Primary customer API"`
	Status          *string `json:"status" binding:"omitempty,oneof=up down paused" example:"up" enums:"up,down,paused"`
}

func (input ServerPatchInput) Apply(value *models.Server) {
	if input.Name != nil {
		value.Name = *input.Name
	}

	if input.Address != nil {
		value.Address = *input.Address
	}

	if input.OperatingSystem != nil {
		value.OperatingSystem = *input.OperatingSystem
	}

	if input.Description != nil {
		value.Description = *input.Description
	}

	if input.Status != nil {
		value.Status = *input.Status
	}
}

func PrepareServer(value *models.Server) error {
	value.Name = strings.TrimSpace(value.Name)
	value.Address = strings.TrimSpace(value.Address)
	value.OperatingSystem = strings.TrimSpace(value.OperatingSystem)

	return Validate(value)
}
