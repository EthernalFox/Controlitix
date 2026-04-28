package authctx

import (
	"context"
	"strings"
	"time"
)

type Principal struct {
	Subject     string
	Roles       []string
	Scopes      []string
	Audience    []string
	Issuer      string
	Source      string
	Username    string
	DisplayName string
	ClientID    string
	TokenID     string
	ExpiresAt   time.Time
}

func (principal Principal) IsService() bool {
	return strings.HasPrefix(strings.TrimSpace(principal.Subject), "svc:")
}

func (principal Principal) HasRole(role string) bool {
	role = strings.TrimSpace(role)
	if role == "" {
		return false
	}

	for _, existingRole := range principal.Roles {
		if strings.TrimSpace(existingRole) == role {
			return true
		}
	}

	return false
}

func (principal Principal) HasScope(scope string) bool {
	scope = strings.TrimSpace(scope)
	if scope == "" {
		return false
	}

	for _, existingScope := range principal.Scopes {
		if strings.TrimSpace(existingScope) == scope {
			return true
		}
	}

	return false
}

type principalContextKey struct{}

func WithPrincipal(ctx context.Context, principal Principal) context.Context {
	return context.WithValue(ctx, principalContextKey{}, principal)
}

func FromContext(ctx context.Context) (Principal, bool) {
	principal, ok := ctx.Value(principalContextKey{}).(Principal)
	return principal, ok
}

func MustFromContext(ctx context.Context) Principal {
	principal, ok := FromContext(ctx)
	if !ok {
		panic("authctx: principal is not available in context")
	}

	return principal
}
