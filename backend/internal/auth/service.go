package auth

import (
	"github.com/uptaris/uptaris/backend/internal/config"
	"gorm.io/gorm"
)

type Service struct {
	db  *gorm.DB
	cfg config.Config
}

type Error struct {
	Code    string
	Message string
}

func New(db *gorm.DB, cfg config.Config) *Service { return &Service{db: db, cfg: cfg} }

func failure(code, message string) error { return &Error{Code: code, Message: message} }
func (e *Error) Error() string           { return e.Message }
