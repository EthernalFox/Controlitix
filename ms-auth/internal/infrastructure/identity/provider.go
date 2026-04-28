package identity

import (
	"context"
	"strings"

	"github.com/EthernalFox/Controlitix/ms-auth/internal/domain"
)

type Provider interface {
	Type() string
	Authenticate(
		ctx context.Context,
		source domain.IdentitySource,
		creds domain.Credentials,
	) (*domain.ExternalIdentity, error)
}

type Registry struct {
	byType map[string]Provider
}

func NewRegistry(providers ...Provider) *Registry {
	byType := make(map[string]Provider, len(providers))
	for _, provider := range providers {
		if provider == nil {
			continue
		}
		providerType := strings.ToLower(strings.TrimSpace(provider.Type()))
		if providerType == "" {
			continue
		}
		byType[providerType] = provider
	}

	return &Registry{
		byType: byType,
	}
}

func (r *Registry) Get(sourceType string) (Provider, bool) {
	if r == nil {
		return nil, false
	}

	provider, ok := r.byType[strings.ToLower(strings.TrimSpace(sourceType))]
	return provider, ok
}
