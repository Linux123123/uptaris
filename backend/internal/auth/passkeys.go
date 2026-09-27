package auth

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/uptaris/uptaris/backend/internal/database"
	"github.com/uptaris/uptaris/backend/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const passkeyCeremonyTTL = 5 * time.Minute

type PasskeyOptions struct {
	CeremonyToken string `json:"ceremonyToken"`
	OptionsJSON   any    `json:"optionsJSON"`
}

type PasskeyInfo struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	CreatedAt  time.Time  `json:"createdAt"`
	LastUsedAt *time.Time `json:"lastUsedAt"`
}

type passkeyUser struct {
	user        models.User
	credentials []webauthn.Credential
}

func (u passkeyUser) WebAuthnID() []byte { return u.user.WebAuthnHandle }

func (u passkeyUser) WebAuthnName() string { return u.user.Email }

func (u passkeyUser) WebAuthnDisplayName() string { return u.user.Email }

func (u passkeyUser) WebAuthnCredentials() []webauthn.Credential { return u.credentials }

func (s *Service) passkeyCredentials(tx *gorm.DB, userID uint) ([]webauthn.Credential, error) {
	var rows []models.Passkey

	if err := tx.Where("user_id = ? AND rp_id = ?", userID, s.cfg.EffectiveWebAuthnRPID()).Find(&rows).Error; err != nil {
		return nil, err
	}

	credentials := make([]webauthn.Credential, 0, len(rows))
	for _, row := range rows {
		var credential webauthn.Credential

		if err := json.Unmarshal(row.Credential, &credential); err != nil {
			return nil, failure("internal_error", "stored passkey is invalid")
		}

		credentials = append(credentials, credential)
	}

	return credentials, nil
}

func (s *Service) newPasskeyCeremony(tx *gorm.DB, kind string, userID, sessionID *uint, data *webauthn.SessionData) (string, error) {
	if err := tx.Where("expires_at <= ?", time.Now()).Delete(&models.PasskeyCeremony{}).Error; err != nil {
		return "", err
	}

	token, err := NewRefresh()
	if err != nil {
		return "", err
	}

	encoded, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	row := models.PasskeyCeremony{
		TokenHash:   HashRefresh(token, s.cfg.RefreshTokenPepper),
		Kind:        kind,
		UserID:      userID,
		SessionID:   sessionID,
		SessionData: encoded,
		ExpiresAt:   time.Now().Add(passkeyCeremonyTTL),
	}

	return token, tx.Create(&row).Error
}

func (s *Service) consumePasskeyCeremony(ctx context.Context, kind, token string, userID, sessionID uint) (webauthn.SessionData, error) {
	var row models.PasskeyCeremony

	err := database.Transaction(s.db.WithContext(ctx), func(tx *gorm.DB) error {
		query := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("token_hash = ? AND kind = ? AND consumed_at IS NULL AND expires_at > ?", HashRefresh(token, s.cfg.RefreshTokenPepper), kind, time.Now())
		if kind == "register" {
			query = query.Where("user_id = ? AND session_id = ?", userID, sessionID)
		}

		if err := query.First(&row).Error; err != nil {
			return err
		}

		return tx.Model(&row).Update("consumed_at", time.Now()).Error
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return webauthn.SessionData{}, failure("invalid_passkey_ceremony", "passkey request expired or already used")
	}

	if err != nil {
		return webauthn.SessionData{}, err
	}

	var data webauthn.SessionData

	if err := json.Unmarshal(row.SessionData, &data); err != nil {
		return webauthn.SessionData{}, failure("internal_error", "stored passkey request is invalid")
	}

	return data, nil
}

func (s *Service) BeginPasskeyRegistration(ctx context.Context, userID, sessionID uint) (PasskeyOptions, error) {
	var result PasskeyOptions

	err := database.Transaction(s.db.WithContext(ctx), func(tx *gorm.DB) error {
		user, err := lockUser(tx, userID)
		if err != nil {
			return err
		}

		if len(user.WebAuthnHandle) == 0 {
			handle := make([]byte, 32)
			if _, err := rand.Read(handle); err != nil {
				return err
			}

			if err := tx.Model(&user).Update("webauthn_handle", handle).Error; err != nil {
				return err
			}

			user.WebAuthnHandle = handle
		}

		credentials, err := s.passkeyCredentials(tx, userID)
		if err != nil {
			return err
		}

		creation, data, err := s.passkeys.BeginRegistration(passkeyUser{user, credentials},
			webauthn.WithAuthenticatorSelection(protocol.AuthenticatorSelection{
				UserVerification: protocol.VerificationRequired,
			}),
			webauthn.WithResidentKeyRequirement(protocol.ResidentKeyRequirementRequired),
			webauthn.WithExclusions(webauthn.Credentials(credentials).CredentialDescriptors()),
		)
		if err != nil {
			return err
		}

		token, err := s.newPasskeyCeremony(tx, "register", &userID, &sessionID, data)
		result = PasskeyOptions{CeremonyToken: token, OptionsJSON: creation.Response}

		return err
	})

	return result, err
}

func (s *Service) FinishPasskeyRegistration(ctx context.Context, userID, sessionID uint, token, name string, response []byte) (PasskeyInfo, error) {
	data, err := s.consumePasskeyCeremony(ctx, "register", token, userID, sessionID)
	if err != nil {
		return PasskeyInfo{}, err
	}

	parsed, err := protocol.ParseCredentialCreationResponseBytes(response)
	if err != nil {
		return PasskeyInfo{}, failure("invalid_passkey_response", "passkey registration failed")
	}

	var result PasskeyInfo

	err = database.Transaction(s.db.WithContext(ctx), func(tx *gorm.DB) error {
		user, err := lockUser(tx, userID)
		if err != nil {
			return err
		}

		credentials, err := s.passkeyCredentials(tx, userID)
		if err != nil {
			return err
		}

		credential, err := s.passkeys.CreateCredential(passkeyUser{user, credentials}, data, parsed)
		if err != nil {
			return failure("invalid_passkey_response", "passkey registration failed")
		}

		encoded, err := json.Marshal(credential)
		if err != nil {
			return err
		}

		if name == "" {
			name = "Passkey"
		}

		row := models.Passkey{
			UserID:       userID,
			RPID:         s.cfg.EffectiveWebAuthnRPID(),
			CredentialID: credential.ID,
			Credential:   encoded,
			Name:         name,
		}

		if err := tx.Create(&row).Error; err != nil {
			return err
		}

		result = PasskeyInfo{
			ID:        strconv.FormatUint(uint64(row.ID), 10),
			Name:      row.Name,
			CreatedAt: row.CreatedAt,
		}

		return nil
	})

	return result, err
}

func (s *Service) BeginPasskeyLogin(ctx context.Context) (PasskeyOptions, error) {
	assertion, data, err := s.passkeys.BeginDiscoverableLogin(webauthn.WithUserVerification(protocol.VerificationRequired))
	if err != nil {
		return PasskeyOptions{}, err
	}

	var token string

	err = database.Transaction(s.db.WithContext(ctx), func(tx *gorm.DB) error {
		var err error
		token, err = s.newPasskeyCeremony(tx, "login", nil, nil, data)

		return err
	})

	return PasskeyOptions{CeremonyToken: token, OptionsJSON: assertion.Response}, err
}

func (s *Service) FinishPasskeyLogin(ctx context.Context, token string, response []byte) (*Session, error) {
	data, err := s.consumePasskeyCeremony(ctx, "login", token, 0, 0)
	if err != nil {
		return nil, err
	}

	parsed, err := protocol.ParseCredentialRequestResponseBytes(response)
	if err != nil {
		return nil, failure("invalid_passkey_response", "passkey sign-in failed")
	}

	var session *Session

	err = database.Transaction(s.db.WithContext(ctx), func(tx *gorm.DB) error {
		var matched models.Passkey
		var owner models.User

		handler := func(rawID, userHandle []byte) (webauthn.User, error) {
			if err := tx.
				Where("credential_id = ? AND rp_id = ?", rawID, s.cfg.EffectiveWebAuthnRPID()).
				First(&matched).Error; err != nil {
				return nil, err
			}

			var err error
			owner, err = lockUser(tx, matched.UserID)
			if err != nil {
				return nil, err
			}

			if !bytes.Equal(owner.WebAuthnHandle, userHandle) {
				return nil, gorm.ErrRecordNotFound
			}

			if err := tx.Where("id = ? AND credential_id = ?", matched.ID, rawID).First(&matched).Error; err != nil {
				return nil, err
			}

			credentials, err := s.passkeyCredentials(tx, owner.ID)
			if err != nil {
				return nil, err
			}

			return passkeyUser{owner, credentials}, nil
		}

		_, credential, err := s.passkeys.ValidatePasskeyLogin(handler, data, parsed)
		if err != nil {
			return failure("invalid_passkey_response", "passkey sign-in failed")
		}

		encoded, err := json.Marshal(credential)
		if err != nil {
			return err
		}

		if err := tx.Model(&matched).Updates(map[string]any{
			"credential":   encoded,
			"last_used_at": time.Now(),
		}).Error; err != nil {
			return err
		}

		// Verified passkey with required UV is a complete sign-in, including for TOTP accounts.
		session, err = s.createSession(tx, owner)

		return err
	})

	return session, err
}

func (s *Service) ListPasskeys(ctx context.Context, userID uint) ([]PasskeyInfo, error) {
	var rows []models.Passkey

	if err := s.db.WithContext(ctx).
		Where("user_id = ? AND rp_id = ?", userID, s.cfg.EffectiveWebAuthnRPID()).
		Order("created_at DESC").
		Find(&rows).Error; err != nil {
		return nil, err
	}

	result := make([]PasskeyInfo, 0, len(rows))
	for _, row := range rows {
		result = append(result, PasskeyInfo{
			ID:         strconv.FormatUint(uint64(row.ID), 10),
			Name:       row.Name,
			CreatedAt:  row.CreatedAt,
			LastUsedAt: row.LastUsedAt,
		})
	}

	return result, nil
}

func (s *Service) DeletePasskey(ctx context.Context, userID, passkeyID uint) error {
	return database.Transaction(s.db.WithContext(ctx), func(tx *gorm.DB) error {
		user, err := lockUser(tx, userID)
		if err != nil {
			return err
		}

		result := tx.Where("id = ? AND user_id = ? AND rp_id = ?", passkeyID, userID, s.cfg.EffectiveWebAuthnRPID()).
			Delete(&models.Passkey{})
		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		canSignIn, err := s.hasSignInMethod(tx, user)
		if err != nil {
			return err
		}

		if !canSignIn {
			return failure("last_signin_method", "Add another sign-in method before removing this passkey")
		}

		return nil
	})
}
