package handlers

import (
	"context"

	"github.com/uptaris/uptaris/backend/internal/auth"
	"github.com/uptaris/uptaris/backend/internal/config"
	"github.com/uptaris/uptaris/backend/internal/inventory"
	"github.com/uptaris/uptaris/backend/internal/users"
)

type Handlers struct {
	auth      *auth.Service
	inventory *inventory.Service
	users     *users.Service
	cfg       config.Config
	health    func(context.Context) error
}

func New(authentication *auth.Service, resources *inventory.Service, accounts *users.Service, cfg config.Config, health func(context.Context) error) *Handlers {
	return &Handlers{
		auth:      authentication,
		inventory: resources,
		users:     accounts,
		cfg:       cfg,
		health:    health,
	}
}

func defaultOf(v, d string) string {
	if v == "" {
		return d
	}
	return v
}
