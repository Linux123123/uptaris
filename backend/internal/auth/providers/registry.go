package providers

import (
	"fmt"

	"github.com/uptaris/uptaris/backend/internal/auth"
	"github.com/uptaris/uptaris/backend/internal/auth/providers/github"
	"github.com/uptaris/uptaris/backend/internal/auth/providers/google"
	"github.com/uptaris/uptaris/backend/internal/config"
)

type factory func(config.OAuthProviderConfig) auth.OAuthProvider

var factories = map[string]factory{
	"google": func(configuration config.OAuthProviderConfig) auth.OAuthProvider {
		return google.New(configuration)
	},
	"github": func(configuration config.OAuthProviderConfig) auth.OAuthProvider {
		return github.New(configuration)
	},
}

func Build(configured map[string]config.OAuthProviderConfig) ([]auth.OAuthProvider, error) {
	providers := make([]auth.OAuthProvider, 0, len(configured))
	for id, settings := range configured {
		factory, supported := factories[id]
		if !supported {
			return nil, fmt.Errorf("OAuth provider %q has no registered adapter", id)
		}

		providers = append(providers, factory(settings))
	}

	return providers, nil
}
