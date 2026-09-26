package auth

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/uptaris/uptaris/backend/internal/models"
	"gorm.io/gorm"
)

func (s *Service) Register(ctx context.Context, email, password string) (*models.User, error) {
	hash, err := HashPassword(password)
	if err != nil {
		return nil, failure("internal_error", "password processing failed")
	}
	user := models.User{Email: strings.ToLower(email), PasswordHash: hash, Role: "viewer"}
	err = s.db.WithContext(ctx).Create(&user).Error
	var conflict *pgconn.PgError
	if errors.As(err, &conflict) && conflict.Code == "23505" {
		return nil, failure("email_exists", "email already registered")
	}
	return &user, err
}

func (s *Service) Login(ctx context.Context, email, password string) (*Session, error) {
	var user models.User
	err := s.db.WithContext(ctx).Where("email = ?", strings.ToLower(email)).First(&user).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if err != nil || CheckPassword(user.PasswordHash, password) != nil {
		return nil, failure("invalid_credentials", "email or password invalid")
	}
	return s.issue(ctx, user)
}

func (s *Service) User(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	err := s.db.WithContext(ctx).First(&user, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, failure("invalid_token", "user unavailable")
	}
	return &user, err
}
