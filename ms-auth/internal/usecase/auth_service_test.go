package usecase

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/EthernalFox/Controlitix/ms-auth/internal/domain"
	"github.com/EthernalFox/Controlitix/ms-auth/internal/infrastructure/identity"
)

type sourceRepositoryMock struct {
	findByType   func(ctx context.Context, sourceType string) ([]*domain.IdentitySource, error)
	resolveRoles func(ctx context.Context, sourceID int, externalGroups []string) ([]string, error)
}

func (m *sourceRepositoryMock) FindByType(
	ctx context.Context,
	sourceType string,
) ([]*domain.IdentitySource, error) {
	return m.findByType(ctx, sourceType)
}

func (m *sourceRepositoryMock) ResolveRoles(
	ctx context.Context,
	sourceID int,
	externalGroups []string,
) ([]string, error) {
	return m.resolveRoles(ctx, sourceID, externalGroups)
}

type userRepositoryMock struct {
	findBySubject   func(ctx context.Context, subject string) (*domain.User, error)
	create          func(ctx context.Context, user *domain.User) (*domain.User, error)
	updateLastLogin func(ctx context.Context, id string, at time.Time) error
	getRoles        func(ctx context.Context, userID string) ([]string, error)
}

func (m *userRepositoryMock) FindBySubject(
	ctx context.Context,
	subject string,
) (*domain.User, error) {
	return m.findBySubject(ctx, subject)
}

func (m *userRepositoryMock) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	return m.create(ctx, user)
}

func (m *userRepositoryMock) UpdateLastLogin(
	ctx context.Context,
	id string,
	at time.Time,
) error {
	return m.updateLastLogin(ctx, id, at)
}

func (m *userRepositoryMock) GetRoles(ctx context.Context, userID string) ([]string, error) {
	return m.getRoles(ctx, userID)
}

type providerMock struct {
	sourceType   string
	authenticate func(
		ctx context.Context,
		source domain.IdentitySource,
		creds domain.Credentials,
	) (*domain.ExternalIdentity, error)
}

func (m *providerMock) Type() string {
	return m.sourceType
}

func (m *providerMock) Authenticate(
	ctx context.Context,
	source domain.IdentitySource,
	creds domain.Credentials,
) (*domain.ExternalIdentity, error) {
	return m.authenticate(ctx, source, creds)
}

func TestAuthServiceAuthenticateSuccess(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)
	source := &domain.IdentitySource{
		ID:        1,
		Type:      "local",
		Name:      "Local",
		IsEnabled: true,
	}
	user := &domain.User{
		ID:       "user-1",
		Subject:  "local:user-1",
		SourceID: 1,
		Username: "admin",
		IsActive: true,
	}

	updateLastLoginCalled := false
	mockSources := &sourceRepositoryMock{
		findByType: func(ctx context.Context, sourceType string) ([]*domain.IdentitySource, error) {
			if sourceType != "local" {
				t.Fatalf("unexpected source type: %s", sourceType)
			}
			return []*domain.IdentitySource{source}, nil
		},
		resolveRoles: func(ctx context.Context, sourceID int, externalGroups []string) ([]string, error) {
			return []string{"engineer"}, nil
		},
	}
	mockUsers := &userRepositoryMock{
		findBySubject: func(ctx context.Context, subject string) (*domain.User, error) {
			return user, nil
		},
		create: func(ctx context.Context, user *domain.User) (*domain.User, error) {
			t.Fatal("unexpected create call")
			return nil, nil
		},
		updateLastLogin: func(ctx context.Context, id string, at time.Time) error {
			updateLastLoginCalled = true
			if id != "user-1" {
				t.Fatalf("unexpected user id: %s", id)
			}
			if !at.Equal(now) {
				t.Fatalf("unexpected login time: %s", at)
			}
			return nil
		},
		getRoles: func(ctx context.Context, userID string) ([]string, error) {
			if userID != "user-1" {
				t.Fatalf("unexpected user id in get roles: %s", userID)
			}
			return []string{"admin"}, nil
		},
	}
	mockProvider := &providerMock{
		sourceType: "local",
		authenticate: func(
			ctx context.Context,
			source domain.IdentitySource,
			creds domain.Credentials,
		) (*domain.ExternalIdentity, error) {
			return &domain.ExternalIdentity{
				Subject:  "local:user-1",
				Username: "admin",
				Groups:   []string{"any-group"},
			}, nil
		},
	}

	service := &AuthService{
		sources:  mockSources,
		users:    mockUsers,
		registry: identity.NewRegistry(mockProvider),
		logger:   slog.Default(),
		clock: func() time.Time {
			return now
		},
	}

	authenticated, err := service.Authenticate(
		context.Background(),
		"local",
		domain.Credentials{Username: "admin", Password: "secret"},
	)
	if err != nil {
		t.Fatalf("authenticate: %v", err)
	}

	if authenticated == nil {
		t.Fatal("expected authenticated user")
	}
	if !updateLastLoginCalled {
		t.Fatal("expected update last login to be called")
	}
	if len(authenticated.Roles) != 2 || authenticated.Roles[0] != "admin" || authenticated.Roles[1] != "engineer" {
		t.Fatalf("unexpected roles: %#v", authenticated.Roles)
	}
}

func TestAuthServiceAuthenticateSourceValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		sources     []*domain.IdentitySource
		wantError   error
		sourceType  string
		resolveType string
	}{
		{
			name:       "unknown source type",
			sources:    []*domain.IdentitySource{},
			wantError:  domain.ErrSourceNotFound,
			sourceType: "unknown",
		},
		{
			name: "inactive source",
			sources: []*domain.IdentitySource{
				{
					ID:        1,
					Type:      "local",
					IsEnabled: false,
				},
			},
			wantError:  domain.ErrSourceDisabled,
			sourceType: "local",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			service := &AuthService{
				sources: &sourceRepositoryMock{
					findByType: func(ctx context.Context, sourceType string) ([]*domain.IdentitySource, error) {
						return tc.sources, nil
					},
					resolveRoles: func(ctx context.Context, sourceID int, externalGroups []string) ([]string, error) {
						return []string{}, nil
					},
				},
				users: &userRepositoryMock{
					findBySubject: func(ctx context.Context, subject string) (*domain.User, error) {
						return nil, nil
					},
					create: func(ctx context.Context, user *domain.User) (*domain.User, error) {
						return user, nil
					},
					updateLastLogin: func(ctx context.Context, id string, at time.Time) error {
						return nil
					},
					getRoles: func(ctx context.Context, userID string) ([]string, error) {
						return []string{}, nil
					},
				},
				registry: identity.NewRegistry(&providerMock{
					sourceType: "local",
					authenticate: func(
						ctx context.Context,
						source domain.IdentitySource,
						creds domain.Credentials,
					) (*domain.ExternalIdentity, error) {
						return &domain.ExternalIdentity{Subject: "local:1"}, nil
					},
				}),
				logger: slog.Default(),
				clock:  time.Now,
			}

			_, err := service.Authenticate(
				context.Background(),
				tc.sourceType,
				domain.Credentials{Username: "admin", Password: "secret"},
			)
			if !errors.Is(err, tc.wantError) {
				t.Fatalf("expected error %v, got %v", tc.wantError, err)
			}
		})
	}
}

func TestAuthServiceAuthenticateJITCreate(t *testing.T) {
	t.Parallel()

	source := &domain.IdentitySource{
		ID:        2,
		Type:      "local",
		Name:      "Local",
		IsEnabled: true,
	}

	createCalled := false
	service := &AuthService{
		sources: &sourceRepositoryMock{
			findByType: func(ctx context.Context, sourceType string) ([]*domain.IdentitySource, error) {
				return []*domain.IdentitySource{source}, nil
			},
			resolveRoles: func(ctx context.Context, sourceID int, externalGroups []string) ([]string, error) {
				return []string{}, nil
			},
		},
		users: &userRepositoryMock{
			findBySubject: func(ctx context.Context, subject string) (*domain.User, error) {
				return nil, nil
			},
			create: func(ctx context.Context, user *domain.User) (*domain.User, error) {
				createCalled = true
				if user.Subject != "ldap:external-user" {
					t.Fatalf("unexpected subject: %s", user.Subject)
				}
				user.ID = "new-user-id"
				user.IsActive = true
				return user, nil
			},
			updateLastLogin: func(ctx context.Context, id string, at time.Time) error {
				return nil
			},
			getRoles: func(ctx context.Context, userID string) ([]string, error) {
				return []string{}, nil
			},
		},
		registry: identity.NewRegistry(&providerMock{
			sourceType: "local",
			authenticate: func(
				ctx context.Context,
				source domain.IdentitySource,
				creds domain.Credentials,
			) (*domain.ExternalIdentity, error) {
				return &domain.ExternalIdentity{
					Subject:     "ldap:external-user",
					Username:    "external.user",
					DisplayName: "External User",
					Email:       "ext@example.local",
				}, nil
			},
		}),
		logger: slog.Default(),
		clock:  time.Now,
	}

	authenticated, err := service.Authenticate(
		context.Background(),
		"local",
		domain.Credentials{Username: "external.user", Password: "secret"},
	)
	if err != nil {
		t.Fatalf("authenticate: %v", err)
	}

	if !createCalled {
		t.Fatal("expected jit create to be called")
	}
	if authenticated == nil {
		t.Fatal("expected authenticated user")
	}
	if authenticated.User.ID != "new-user-id" {
		t.Fatalf("unexpected created user id: %s", authenticated.User.ID)
	}
}
