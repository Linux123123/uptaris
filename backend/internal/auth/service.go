package auth

import (
	"strings"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/uptaris/uptaris/backend/internal/config"
	"gorm.io/gorm"
)

type Service struct {
	db        *gorm.DB
	cfg       config.Config
	providers map[string]OAuthProvider
	passkeys  *webauthn.WebAuthn
}

type Error struct {
	Code    string
	Message string
}

func New(db *gorm.DB, cfg config.Config, providers ...OAuthProvider) *Service {
	passkeys, err := webauthn.New(&webauthn.Config{
		RPID:                  cfg.EffectiveWebAuthnRPID(),
		RPDisplayName:         "Uptaris",
		RPOrigins:             []string{cfg.FrontendURL},
		AttestationPreference: protocol.PreferNoAttestation,
	})
	if err != nil {
		panic(err)
	}

	registered := make(map[string]OAuthProvider, len(providers))
	for _, provider := range providers {
		if provider != nil && strings.TrimSpace(provider.ID()) != "" {
			registered[provider.ID()] = provider
		}
	}

	return &Service{
		db:        db,
		cfg:       cfg,
		providers: registered,
		passkeys:  passkeys,
	}
}

func failure(code, message string) error { return &Error{Code: code, Message: message} }

func (e *Error) Error() string { return e.Message }
