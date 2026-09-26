package handlers

import (
	"context"

	"github.com/uptaris/uptaris/backend/internal/auth"
	"github.com/uptaris/uptaris/backend/internal/config"
	"github.com/uptaris/uptaris/backend/internal/users"
	"gorm.io/gorm"
)

type Handlers struct {
	db     *gorm.DB
	auth   *auth.Service
	users  *users.Service
	cfg    config.Config
	health func(context.Context) error
}

func New(db *gorm.DB, authentication *auth.Service, accounts *users.Service, cfg config.Config, health func(context.Context) error) *Handlers {
	return &Handlers{
		db:     db,
		auth:   authentication,
		users:  accounts,
		cfg:    cfg,
		health: health,
	}
}

func defaultOf(v, d string) string {
	if v == "" {
		return d
	}
	return v
}
