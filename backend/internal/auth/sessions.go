package auth

import (
	"context"
	"errors"
	"time"

	"github.com/uptaris/uptaris/backend/internal/database"
	"github.com/uptaris/uptaris/backend/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Session struct {
	User           models.User
	AccessToken    string
	RefreshToken   string
	ExpiresAt      time.Time
	ChallengeToken string
}

// Caller holds the user lock while verifying the first factor and issuing its result.
func (s *Service) issueLocked(tx *gorm.DB, user models.User) (*Session, error) {
	var factor models.TwoFactor

	err := tx.Where("user_id = ? AND enabled_at IS NOT NULL", user.ID).First(&factor).Error
	if err == nil {
		token, err := s.newChallenge(tx, user.ID)
		if err != nil {
			return nil, err
		}

		return &Session{ChallengeToken: token}, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return s.createSession(tx, user)
}

// Caller holds the user lock; verification and session creation share one commit.
func (s *Service) createSession(tx *gorm.DB, user models.User) (*Session, error) {
	refresh, err := NewRefresh()
	if err != nil {
		return nil, err
	}

	row := models.AuthSession{
		UserID:           user.ID,
		RefreshTokenHash: HashRefresh(refresh, s.cfg.RefreshTokenPepper),
		ExpiresAt:        time.Now().Add(s.cfg.RefreshTokenTTL),
	}

	if err := tx.Create(&row).Error; err != nil {
		return nil, err
	}

	access, _, err := Issue(uint64(user.ID), row.ID, user.Role, s.cfg.JWTAccessSecret, s.cfg.AccessTokenTTL)
	if err != nil {
		return nil, err
	}

	return &Session{User: user, AccessToken: access, RefreshToken: refresh, ExpiresAt: row.ExpiresAt}, nil
}

func (s *Service) Refresh(ctx context.Context, refresh string) (*Session, error) {
	hash := HashRefresh(refresh, s.cfg.RefreshTokenPepper)
	next, err := NewRefresh()
	if err != nil {
		return nil, err
	}

	var session models.AuthSession
	var user models.User
	var access string

	err = database.Transaction(s.db.WithContext(ctx), func(tx *gorm.DB) error {
		if err := tx.
			Where("refresh_token_hash = ? AND revoked_at IS NULL AND expires_at > ?", hash, time.Now()).
			First(&session).Error; err != nil {
			return err
		}

		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, session.UserID).Error; err != nil {
			return err
		}

		result := tx.Model(&models.AuthSession{}).
			Where("id = ? AND refresh_token_hash = ? AND revoked_at IS NULL AND expires_at > ?", session.ID, hash, time.Now()).
			Update("refresh_token_hash", HashRefresh(next, s.cfg.RefreshTokenPepper))
		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}

		var err error

		access, _, err = Issue(uint64(user.ID), session.ID, user.Role, s.cfg.JWTAccessSecret, s.cfg.AccessTokenTTL)

		return err
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, failure("invalid_refresh", "refresh token invalid")
	}

	return &Session{User: user, AccessToken: access, RefreshToken: next, ExpiresAt: session.ExpiresAt}, err
}

func (s *Service) Logout(ctx context.Context, identity Identity) error {
	return s.db.WithContext(ctx).Model(&models.AuthSession{}).
		Where("id = ? AND user_id = ?", identity.SessionID, identity.ID).
		Update("revoked_at", time.Now()).Error
}
