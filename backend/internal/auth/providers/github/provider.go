package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/uptaris/uptaris/backend/internal/auth"
	"github.com/uptaris/uptaris/backend/internal/auth/providers/internal/oauth"
	"github.com/uptaris/uptaris/backend/internal/config"
	"golang.org/x/oauth2"
	oauthgithub "golang.org/x/oauth2/github"
)

const providerID = "github"

type Provider struct {
	config oauth2.Config
	client *http.Client
}

func New(configuration config.OAuthProviderConfig) *Provider {
	endpoint := oauthgithub.Endpoint
	endpoint.AuthStyle = oauth2.AuthStyleInParams

	return &Provider{
		config: oauth2.Config{
			ClientID:     configuration.ClientID,
			ClientSecret: configuration.ClientSecret,
			RedirectURL:  configuration.RedirectURL,
			Scopes:       []string{"read:user", "user:email"},
			Endpoint:     endpoint,
		},
		client: &http.Client{Timeout: 8 * time.Second},
	}
}

func (p *Provider) ID() string { return providerID }

func (p *Provider) DisplayName() string { return "GitHub" }

func (p *Provider) AuthorizationURL(state, verifier string) string {
	return p.config.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier))
}

func (p *Provider) Exchange(ctx context.Context, code, verifier string) (auth.OAuthIdentity, error) {
	ctx = context.WithValue(ctx, oauth2.HTTPClient, p.client)

	token, err := p.config.Exchange(ctx, code, oauth2.VerifierOption(verifier))
	if err != nil || token.AccessToken == "" {
		return auth.OAuthIdentity{}, fmt.Errorf("GitHub token request failed")
	}

	var user struct {
		ID int64 `json:"id"`
	}

	if err := p.get(ctx, token.AccessToken, "https://api.github.com/user", &user); err != nil || user.ID <= 0 {
		return auth.OAuthIdentity{}, fmt.Errorf("GitHub identity lookup failed")
	}

	identity := auth.OAuthIdentity{
		Subject: strconv.FormatInt(user.ID, 10),
		Tokens:  oauth.Tokens(token),
	}

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}

	if err := p.get(ctx, token.AccessToken, "https://api.github.com/user/emails", &emails); err == nil {
		for _, email := range emails {
			if email.Primary && email.Verified && strings.TrimSpace(email.Email) != "" {
				identity.Email = email.Email
				identity.EmailVerified = true
				break
			}
		}
	}

	return identity, nil
}

func (p *Provider) get(ctx context.Context, token, endpoint string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2026-03-10")
	response, err := p.client.Do(req)
	if err != nil {
		return err
	}

	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub API returned %d", response.StatusCode)
	}

	return json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(out)
}
