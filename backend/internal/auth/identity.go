package auth

import (
	"context"
	"errors"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/uptaris/uptaris/backend/internal/models"
	"gorm.io/gorm"
)

type Identity struct {
	ID        uint
	Role      string
	SessionID uint
}

func (s *Service) Authenticate(ctx context.Context, header string, roles ...string) (Identity, error) {
	raw := strings.TrimPrefix(header, "Bearer ")
	if raw == header || raw == "" {
		return Identity{}, failure("authentication_required", "bearer token required")
	}
	claims, err := Parse(raw, s.cfg.JWTAccessSecret)
	if err != nil {
		return Identity{}, failure("invalid_token", "access token invalid")
	}
	id, err := strconv.ParseUint(claims.Subject, 10, 64)
	if err != nil {
		return Identity{}, failure("invalid_token", "access token invalid")
	}
	var session models.AuthSession
	err = s.db.WithContext(ctx).Where("id = ? AND user_id = ? AND revoked_at IS NULL AND expires_at > ?", claims.SessionID, id, time.Now()).First(&session).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Identity{}, failure("revoked_token", "user session revoked")
	}
	if err != nil {
		return Identity{}, err
	}
	var user models.User
	err = s.db.WithContext(ctx).First(&user, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Identity{}, failure("revoked_token", "user session revoked")
	}
	if err != nil {
		return Identity{}, err
	}
	if claims.Role != user.Role {
		return Identity{}, failure("stale_token", "user role changed; authenticate again")
	}
	if len(roles) > 0 && !slices.Contains(roles, claims.Role) {
		return Identity{}, failure("forbidden", "role not permitted")
	}
	return Identity{ID: uint(id), Role: claims.Role, SessionID: claims.SessionID}, nil
}
