package identity

import (
	"context"
	"errors"
	"log/slog"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/EthernalFox/Controlitix/ms-auth/internal/domain"
	"github.com/EthernalFox/Controlitix/ms-auth/internal/infrastructure/password"
)

type localUserRepositoryMock struct {
	findBySourceAndUsername func(
		ctx context.Context,
		sourceID int,
		username string,
	) (*domain.User, error)
}

func (m *localUserRepositoryMock) FindBySourceAndUsername(
	ctx context.Context,
	sourceID int,
	username string,
) (*domain.User, error) {
	return m.findBySourceAndUsername(ctx, sourceID, username)
}

func TestLocalProviderAuthenticate(t *testing.T) {
	t.Parallel()

	pepper := []byte(strings.Repeat("p", 32))
	hasher, err := password.NewHasher(pepper)
	if err != nil {
		t.Fatalf("new hasher: %v", err)
	}

	validHash, err := hasher.Hash("correct-password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	testSource := domain.IdentitySource{
		ID:        1,
		Type:      "local",
		IsEnabled: true,
	}

	tests := []struct {
		name      string
		user      *domain.User
		password  string
		wantError error
	}{
		{
			name: "valid credentials",
			user: &domain.User{
				Subject:      "local:123",
				Username:     "admin",
				DisplayName:  "Admin",
				Email:        "admin@example.local",
				PasswordHash: validHash,
				IsActive:     true,
			},
			password:  "correct-password",
			wantError: nil,
		},
		{
			name: "invalid password",
			user: &domain.User{
				Subject:      "local:123",
				Username:     "admin",
				PasswordHash: validHash,
				IsActive:     true,
			},
			password:  "wrong-password",
			wantError: domain.ErrInvalidCredentials,
		},
		{
			name:      "missing user",
			user:      nil,
			password:  "any-password",
			wantError: domain.ErrInvalidCredentials,
		},
		{
			name: "empty password hash",
			user: &domain.User{
				Subject:      "local:123",
				Username:     "admin",
				PasswordHash: "",
				IsActive:     true,
			},
			password:  "any-password",
			wantError: domain.ErrInvalidCredentials,
		},
		{
			name: "disabled user",
			user: &domain.User{
				Subject:      "local:123",
				Username:     "admin",
				PasswordHash: validHash,
				IsActive:     false,
			},
			password:  "correct-password",
			wantError: domain.ErrUserDisabled,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			userRepo := &localUserRepositoryMock{
				findBySourceAndUsername: func(
					ctx context.Context,
					sourceID int,
					username string,
				) (*domain.User, error) {
					return tc.user, nil
				},
			}

			provider := &LocalProvider{
				users:     userRepo,
				hasher:    hasher,
				logger:    slog.Default(),
				dummyHash: fallbackDummyHash,
			}

			externalIdentity, authErr := provider.Authenticate(
				context.Background(),
				testSource,
				domain.Credentials{
					Username: "admin",
					Password: tc.password,
				},
			)

			if tc.wantError == nil {
				if authErr != nil {
					t.Fatalf("authenticate: %v", authErr)
				}
				if externalIdentity == nil {
					t.Fatal("expected external identity")
				}
				if externalIdentity.Subject != tc.user.Subject {
					t.Fatalf("unexpected subject: %s", externalIdentity.Subject)
				}
				return
			}

			if !errors.Is(authErr, tc.wantError) {
				t.Fatalf("expected error %v, got %v", tc.wantError, authErr)
			}
		})
	}
}

func TestLocalProviderAuthenticateConstantTimeApproximation(t *testing.T) {
	pepper := []byte(strings.Repeat("p", 32))
	hasher, err := password.NewHasher(pepper)
	if err != nil {
		t.Fatalf("new hasher: %v", err)
	}

	validHash, err := hasher.Hash("correct-password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	testSource := domain.IdentitySource{
		ID:        1,
		Type:      "local",
		IsEnabled: true,
	}

	missingUserProvider := &LocalProvider{
		users: &localUserRepositoryMock{
			findBySourceAndUsername: func(
				ctx context.Context,
				sourceID int,
				username string,
			) (*domain.User, error) {
				return nil, nil
			},
		},
		hasher:    hasher,
		logger:    slog.Default(),
		dummyHash: fallbackDummyHash,
	}

	wrongPasswordProvider := &LocalProvider{
		users: &localUserRepositoryMock{
			findBySourceAndUsername: func(
				ctx context.Context,
				sourceID int,
				username string,
			) (*domain.User, error) {
				return &domain.User{
					Subject:      "local:123",
					Username:     "admin",
					PasswordHash: validHash,
					IsActive:     true,
				}, nil
			},
		},
		hasher:    hasher,
		logger:    slog.Default(),
		dummyHash: fallbackDummyHash,
	}

	startMissing := time.Now()
	_, err = missingUserProvider.Authenticate(
		context.Background(),
		testSource,
		domain.Credentials{Username: "missing", Password: "wrong-password"},
	)
	missingDuration := time.Since(startMissing)
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("missing user error: %v", err)
	}

	startWrong := time.Now()
	_, err = wrongPasswordProvider.Authenticate(
		context.Background(),
		testSource,
		domain.Credentials{Username: "admin", Password: "wrong-password"},
	)
	wrongPasswordDuration := time.Since(startWrong)
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("wrong password error: %v", err)
	}

	denominator := float64(maxDuration(missingDuration, wrongPasswordDuration))
	if denominator == 0 {
		t.Fatal("unexpected zero duration")
	}

	difference := math.Abs(float64(missingDuration - wrongPasswordDuration))
	ratio := difference / denominator
	if ratio > 0.20 {
		t.Fatalf(
			"duration difference is outside 20%% tolerance: missing=%s wrong=%s ratio=%.2f",
			missingDuration,
			wrongPasswordDuration,
			ratio,
		)
	}
}

func maxDuration(left, right time.Duration) time.Duration {
	if left > right {
		return left
	}

	return right
}
