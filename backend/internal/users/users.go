package users

import (
	"context"
	"errors"
	"slices"
	"time"

	"github.com/uptaris/uptaris/backend/internal/auth"
	"github.com/uptaris/uptaris/backend/internal/database"
	"github.com/uptaris/uptaris/backend/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Service struct{ db *gorm.DB }

func New(db *gorm.DB) *Service { return &Service{db: db} }

var ErrOwnRole = errors.New("administrator cannot change own role")
var ErrOwnDelete = errors.New("administrator cannot delete own account")
var ErrRole = errors.New("role must be viewer, operator, or admin")

func (s *Service) List(ctx context.Context, page, size int) ([]models.User, int, error) {
	db := s.db.WithContext(ctx)
	var total int64
	if err := db.Model(&models.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	rows := []models.User{}
	err := db.Select("id,email,role,created_at,updated_at").Order("id").Offset((page - 1) * size).Limit(size).Find(&rows).Error
	return rows, int(total), err
}

func (s *Service) ChangeRole(ctx context.Context, actor auth.Identity, userID uint, role string) (*models.User, error) {
	if !slices.Contains([]string{"viewer", "operator", "admin"}, role) {
		return nil, ErrRole
	}
	if userID == actor.ID {
		return nil, ErrOwnRole
	}
	var user models.User
	err := database.Transaction(s.db.WithContext(ctx), func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, userID).Error; err != nil {
			return err
		}
		if err := tx.Model(&user).Update("role", role).Error; err != nil {
			return err
		}
		return tx.Model(&models.AuthSession{}).Where("user_id = ? AND revoked_at IS NULL", userID).Update("revoked_at", time.Now()).Error
	})
	user.Role = role
	return &user, err
}

func (s *Service) Delete(ctx context.Context, actor auth.Identity, userID uint) error {
	if userID == actor.ID {
		return ErrOwnDelete
	}
	err := database.Transaction(s.db.WithContext(ctx), func(tx *gorm.DB) error {
		var user models.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, userID).Error; err != nil {
			return err
		}
		if err := tx.Delete(&user).Error; err != nil {
			return err
		}
		return tx.Model(&models.AuthSession{}).Where("user_id = ?", userID).Update("revoked_at", time.Now()).Error
	})
	return err
}
