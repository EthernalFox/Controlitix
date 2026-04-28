package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/EthernalFox/Controlitix/ms-auth/internal/domain"
	"github.com/EthernalFox/Controlitix/ms-auth/internal/infrastructure/identity"
	"github.com/EthernalFox/Controlitix/ms-auth/internal/infrastructure/repository"
)

type authSourceRepository interface {
	FindByType(ctx context.Context, sourceType string) ([]*domain.IdentitySource, error)
	ResolveRoles(ctx context.Context, sourceID int, externalGroups []string) ([]string, error)
}

type authUserRepository interface {
	FindBySubject(ctx context.Context, subject string) (*domain.User, error)
	Create(ctx context.Context, u *domain.User) (*domain.User, error)
	UpdateLastLogin(ctx context.Context, id string, at time.Time) error
	GetRoles(ctx context.Context, userID string) ([]string, error)
}

type AuthService struct {
	sources  authSourceRepository
	users    authUserRepository
	registry *identity.Registry
	logger   *slog.Logger
	clock    func() time.Time
}

func NewAuthService(
	sources *repository.IdentitySourceRepository,
	users *repository.UserRepository,
	registry *identity.Registry,
	logger *slog.Logger,
) *AuthService {
	if logger == nil {
		logger = slog.Default()
	}

	return &AuthService{
		sources:  sources,
		users:    users,
		registry: registry,
		logger:   logger,
		clock:    time.Now,
	}
}

func (s *AuthService) Authenticate(
	ctx context.Context,
	sourceType string,
	creds domain.Credentials,
) (*domain.AuthenticatedUser, error) {
	source, err := s.resolveSource(ctx, sourceType)
	if err != nil {
		return nil, err
	}

	provider, ok := s.registry.Get(source.Type)
	if !ok {
		return nil, domain.ErrSourceNotFound
	}

	externalIdentity, err := provider.Authenticate(ctx, *source, creds)
	if err != nil {
		return nil, err
	}
	if externalIdentity == nil {
		return nil, domain.ErrInvalidCredentials
	}

	user, err := s.users.FindBySubject(ctx, externalIdentity.Subject)
	if err != nil {
		return nil, fmt.Errorf("find user by subject: %w", err)
	}

	if user == nil {
		user, err = s.users.Create(ctx, &domain.User{
			Subject:      externalIdentity.Subject,
			SourceID:     source.ID,
			Username:     externalIdentity.Username,
			DisplayName:  externalIdentity.DisplayName,
			Email:        externalIdentity.Email,
			PasswordHash: "",
			IsActive:     true,
		})
		if err != nil {
			return nil, fmt.Errorf("jit create user: %w", err)
		}
	}

	if !user.IsActive {
		return nil, domain.ErrUserDisabled
	}

	localRoles, err := s.users.GetRoles(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("get user roles: %w", err)
	}

	mappedRoles, err := s.sources.ResolveRoles(ctx, source.ID, externalIdentity.Groups)
	if err != nil {
		return nil, fmt.Errorf("resolve mapped roles: %w", err)
	}

	allRoles := mergeRoles(localRoles, mappedRoles)
	loginAt := s.clock()
	if err := s.users.UpdateLastLogin(ctx, user.ID, loginAt); err != nil {
		return nil, fmt.Errorf("update last login: %w", err)
	}
	user.LastLoginAt = &loginAt

	return &domain.AuthenticatedUser{
		User:   *user,
		Source: *source,
		Roles:  allRoles,
	}, nil
}

func (s *AuthService) resolveSource(
	ctx context.Context,
	sourceType string,
) (*domain.IdentitySource, error) {
	sourceType = strings.ToLower(strings.TrimSpace(sourceType))
	sources, err := s.sources.FindByType(ctx, sourceType)
	if err != nil {
		return nil, fmt.Errorf("find identity source by type: %w", err)
	}
	if len(sources) == 0 {
		return nil, domain.ErrSourceNotFound
	}

	for _, source := range sources {
		if source == nil {
			continue
		}
		if source.IsEnabled {
			return source, nil
		}
	}

	return nil, domain.ErrSourceDisabled
}

func mergeRoles(roleSets ...[]string) []string {
	unique := make(map[string]struct{})
	roles := make([]string, 0)

	for _, roleSet := range roleSets {
		for _, role := range roleSet {
			role = strings.TrimSpace(role)
			if role == "" {
				continue
			}
			if _, exists := unique[role]; exists {
				continue
			}
			unique[role] = struct{}{}
			roles = append(roles, role)
		}
	}

	if roles == nil {
		return []string{}
	}

	return roles
}
