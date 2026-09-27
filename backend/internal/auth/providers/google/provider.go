package google

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/uptaris/uptaris/backend/internal/auth"
	"github.com/uptaris/uptaris/backend/internal/auth/providers/internal/oauth"
	"github.com/uptaris/uptaris/backend/internal/config"
	"golang.org/x/oauth2"
)

type Provider struct {
	config oauth2.Config
	client *http.Client
}

func New(configuration config.OAuthProviderConfig) *Provider {
	return &Provider{
		config: oauth2.Config{
			ClientID:     configuration.ClientID,
			ClientSecret: configuration.ClientSecret,
			RedirectURL:  configuration.RedirectURL,
			Scopes:       []string{"openid", "email"},
			Endpoint: oauth2.Endpoint{
				AuthURL:   "https://accounts.google.com/o/oauth2/v2/auth",
				TokenURL:  "https://oauth2.googleapis.com/token",
				AuthStyle: oauth2.AuthStyleInParams,
			},
		},
		client: &http.Client{Timeout: 8 * time.Second},
	}
}

func (p *Provider) ID() string { return "google" }

func (p *Provider) DisplayName() string { return "Google" }

func (p *Provider) AuthorizationURL(state string) string {
	// Request refresh access without forcing consent or account selection on every sign-in.
	return p.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (p *Provider) Exchange(ctx context.Context, code string) (auth.OAuthIdentity, error) {
	ctx = context.WithValue(ctx, oauth2.HTTPClient, p.client)
	token, err := p.config.Exchange(ctx, code)
	if err != nil || token.AccessToken == "" {
		return auth.OAuthIdentity{}, fmt.Errorf("google token request failed")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://openidconnect.googleapis.com/v1/userinfo", nil)
	if err != nil {
		return auth.OAuthIdentity{}, err
	}

	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	response, err := p.client.Do(req)
	if err != nil {
		return auth.OAuthIdentity{}, fmt.Errorf("google identity lookup failed")
	}

	defer response.Body.Close()

	var user struct {
		Subject       string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
	}

	if response.StatusCode != http.StatusOK || json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&user) != nil || user.Subject == "" {
		return auth.OAuthIdentity{}, fmt.Errorf("google identity lookup failed")
	}

	return auth.OAuthIdentity{
		Subject:       user.Subject,
		Email:         user.Email,
		EmailVerified: user.EmailVerified,
		Tokens:        oauth.Tokens(token),
	}, nil
}
