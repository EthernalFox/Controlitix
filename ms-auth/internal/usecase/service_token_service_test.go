package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/EthernalFox/Controlitix/ms-auth/internal/domain"
	"github.com/EthernalFox/Controlitix/ms-auth/internal/infrastructure/password"
)

type serviceTokenAccountsMock struct {
	findByClientID func(ctx context.Context, clientID string) (*domain.ServiceAccount, error)
	updateLastUsed func(ctx context.Context, id string, at time.Time) error
}

func (m *serviceTokenAccountsMock) FindByClientID(
	ctx context.Context,
	clientID string,
) (*domain.ServiceAccount, error) {
	return m.findByClientID(ctx, clientID)
}

func (m *serviceTokenAccountsMock) UpdateLastUsed(
	ctx context.Context,
	id string,
	at time.Time,
) error {
	return m.updateLastUsed(ctx, id, at)
}

type serviceTokenIssuerMock struct {
	issue func(
		ctx context.Context,
		clientID string,
		scopes []string,
		meta TokenMeta,
	) (*domain.IssuedTokens, error)
}

func (m *serviceTokenIssuerMock) IssueForService(
	ctx context.Context,
	clientID string,
	scopes []string,
	meta TokenMeta,
) (*domain.IssuedTokens, error) {
	return m.issue(ctx, clientID, scopes, meta)
}

type serviceTokenAuditMock struct {
	events []domain.AuditEvent
}

func (m *serviceTokenAuditMock) Record(_ context.Context, event domain.AuditEvent) error {
	m.events = append(m.events, event)
	return nil
}

func TestServiceTokenServiceIssueSuccess(t *testing.T) {
	t.Parallel()

	hasher := testHasher(t)
	secret := "super-secret"
	hash, err := hasher.Hash(secret)
	if err != nil {
		t.Fatalf("hash secret: %v", err)
	}

	updateLastUsedCalled := false
	issuedScopes := []string{}
	audit := &serviceTokenAuditMock{}
	service := &ServiceTokenService{
		accounts: &serviceTokenAccountsMock{
			findByClientID: func(
				ctx context.Context,
				clientID string,
			) (*domain.ServiceAccount, error) {
				return &domain.ServiceAccount{
					ID:               "svc-id",
					ClientID:         clientID,
					ClientSecretHash: hash,
					Scopes:           []string{"config.read", "tags.values.write"},
					IsActive:         true,
				}, nil
			},
			updateLastUsed: func(ctx context.Context, id string, at time.Time) error {
				updateLastUsedCalled = true
				if id != "svc-id" {
					t.Fatalf("unexpected service account id: %s", id)
				}
				return nil
			},
		},
		tokens: &serviceTokenIssuerMock{
			issue: func(
				ctx context.Context,
				clientID string,
				scopes []string,
				meta TokenMeta,
			) (*domain.IssuedTokens, error) {
				issuedScopes = scopes
				return &domain.IssuedTokens{
					AccessToken: "service-access",
					ExpiresIn:   900,
				}, nil
			},
		},
		hasher: hasher,
		audit:  audit,
		clock:  func() time.Time { return time.Date(2026, 4, 25, 15, 0, 0, 0, time.UTC) },
	}

	result, err := service.Issue(context.Background(), ServiceTokenInput{
		GrantType:    "client_credentials",
		ClientID:     "ms-poll",
		ClientSecret: secret,
		Scope:        "config.read",
	})
	if err != nil {
		t.Fatalf("issue service token: %v", err)
	}

	if result.AccessToken != "service-access" {
		t.Fatalf("unexpected access token: %s", result.AccessToken)
	}
	if result.Scope != "config.read" {
		t.Fatalf("unexpected scope: %s", result.Scope)
	}
	if !updateLastUsedCalled {
		t.Fatal("expected update last used call")
	}
	if len(issuedScopes) != 1 || issuedScopes[0] != "config.read" {
		t.Fatalf("unexpected issued scopes: %#v", issuedScopes)
	}
	if len(audit.events) != 1 || audit.events[0].Action != domain.AuditServiceTokenSuccess {
		t.Fatalf("unexpected audit events: %#v", audit.events)
	}
}

func TestServiceTokenServiceIssueInvalidSecret(t *testing.T) {
	t.Parallel()

	hasher := testHasher(t)
	hash, err := hasher.Hash("actual-secret")
	if err != nil {
		t.Fatalf("hash secret: %v", err)
	}

	audit := &serviceTokenAuditMock{}
	service := &ServiceTokenService{
		accounts: &serviceTokenAccountsMock{
			findByClientID: func(ctx context.Context, clientID string) (*domain.ServiceAccount, error) {
				return &domain.ServiceAccount{
					ID:               "svc-id",
					ClientID:         clientID,
					ClientSecretHash: hash,
					Scopes:           []string{"config.read"},
					IsActive:         true,
				}, nil
			},
			updateLastUsed: func(ctx context.Context, id string, at time.Time) error {
				return errors.New("unexpected update last used call")
			},
		},
		tokens: &serviceTokenIssuerMock{
			issue: func(ctx context.Context, clientID string, scopes []string, meta TokenMeta) (*domain.IssuedTokens, error) {
				return nil, errors.New("unexpected issue call")
			},
		},
		hasher: hasher,
		audit:  audit,
	}

	_, err = service.Issue(context.Background(), ServiceTokenInput{
		GrantType:    "client_credentials",
		ClientID:     "ms-poll",
		ClientSecret: "wrong-secret",
	})
	if !errors.Is(err, domain.ErrInvalidClient) {
		t.Fatalf("expected ErrInvalidClient, got %v", err)
	}
	if len(audit.events) != 1 || audit.events[0].Reason != "invalid_secret" {
		t.Fatalf("unexpected audit events: %#v", audit.events)
	}
}

func TestServiceTokenServiceIssueErrors(t *testing.T) {
	t.Parallel()

	hasher := testHasher(t)
	hash, err := hasher.Hash("svc-secret")
	if err != nil {
		t.Fatalf("hash secret: %v", err)
	}

	baseService := func(account *domain.ServiceAccount) *ServiceTokenService {
		return &ServiceTokenService{
			accounts: &serviceTokenAccountsMock{
				findByClientID: func(ctx context.Context, clientID string) (*domain.ServiceAccount, error) {
					return account, nil
				},
				updateLastUsed: func(ctx context.Context, id string, at time.Time) error {
					return nil
				},
			},
			tokens: &serviceTokenIssuerMock{
				issue: func(ctx context.Context, clientID string, scopes []string, meta TokenMeta) (*domain.IssuedTokens, error) {
					return &domain.IssuedTokens{AccessToken: "x", ExpiresIn: 1}, nil
				},
			},
			hasher: hasher,
			audit:  &serviceTokenAuditMock{},
		}
	}

	_, err = baseService(nil).Issue(context.Background(), ServiceTokenInput{
		GrantType:    "client_credentials",
		ClientID:     "missing",
		ClientSecret: "any",
	})
	if !errors.Is(err, domain.ErrInvalidClient) {
		t.Fatalf("expected ErrInvalidClient, got %v", err)
	}

	_, err = baseService(&domain.ServiceAccount{
		ID:               "id",
		ClientID:         "ms-poll",
		ClientSecretHash: hash,
		Scopes:           []string{"config.read"},
		IsActive:         false,
	}).Issue(context.Background(), ServiceTokenInput{
		GrantType:    "client_credentials",
		ClientID:     "ms-poll",
		ClientSecret: "svc-secret",
	})
	if !errors.Is(err, domain.ErrClientDisabled) {
		t.Fatalf("expected ErrClientDisabled, got %v", err)
	}

	_, err = baseService(&domain.ServiceAccount{
		ID:               "id",
		ClientID:         "ms-poll",
		ClientSecretHash: hash,
		Scopes:           []string{"config.read"},
		IsActive:         true,
	}).Issue(context.Background(), ServiceTokenInput{
		GrantType:    "client_credentials",
		ClientID:     "ms-poll",
		ClientSecret: "svc-secret",
		Scope:        "tags.values.write",
	})
	if !errors.Is(err, domain.ErrInvalidScope) {
		t.Fatalf("expected ErrInvalidScope, got %v", err)
	}

	_, err = baseService(&domain.ServiceAccount{
		ID:               "id",
		ClientID:         "ms-poll",
		ClientSecretHash: hash,
		Scopes:           []string{"config.read"},
		IsActive:         true,
	}).Issue(context.Background(), ServiceTokenInput{
		GrantType:    "password",
		ClientID:     "ms-poll",
		ClientSecret: "svc-secret",
	})
	if !errors.Is(err, domain.ErrUnsupportedGrant) {
		t.Fatalf("expected ErrUnsupportedGrant, got %v", err)
	}
}

func testHasher(t *testing.T) *password.Hasher {
	t.Helper()

	hasher, err := password.NewHasher([]byte("01234567890123456789012345678901"))
	if err != nil {
		t.Fatalf("new hasher: %v", err)
	}

	return hasher
}
