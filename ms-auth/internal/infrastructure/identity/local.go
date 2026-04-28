package identity

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/EthernalFox/Controlitix/ms-auth/internal/domain"
	"github.com/EthernalFox/Controlitix/ms-auth/internal/infrastructure/password"
	"github.com/EthernalFox/Controlitix/ms-auth/internal/infrastructure/repository"
)

const fallbackDummyHash = "$argon2id$v=19$m=65536,t=3,p=2$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"

type localUserRepository interface {
	FindBySourceAndUsername(
		ctx context.Context,
		sourceID int,
		username string,
	) (*domain.User, error)
}

type LocalProvider struct {
	users     localUserRepository
	hasher    *password.Hasher
	logger    *slog.Logger
	dummyHash string
}

func NewLocalProvider(
	users *repository.UserRepository,
	hasher *password.Hasher,
	logger *slog.Logger,
) *LocalProvider {
	if logger == nil {
		logger = slog.Default()
	}

	dummyHash, err := hasher.Hash("controlitix:invalid-local-provider-password")
	if err != nil {
		logger.Error(
			"failed to generate dummy hash, fallback hash will be used",
			"method",
			"LocalProvider.New",
			"error",
			err,
		)
		dummyHash = fallbackDummyHash
	}

	return &LocalProvider{
		users:     users,
		hasher:    hasher,
		logger:    logger,
		dummyHash: dummyHash,
	}
}

func (p *LocalProvider) Type() string {
	return "local"
}

func (p *LocalProvider) Authenticate(
	ctx context.Context,
	source domain.IdentitySource,
	creds domain.Credentials,
) (*domain.ExternalIdentity, error) {
	user, err := p.users.FindBySourceAndUsername(ctx, source.ID, creds.Username)
	if err != nil {
		return nil, fmt.Errorf("find local user by username: %w", err)
	}

	if user == nil || user.PasswordHash == "" {
		p.verifyDummy(creds.Password)
		return nil, domain.ErrInvalidCredentials
	}

	match, err := p.hasher.Verify(creds.Password, user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("verify local password: %w", err)
	}
	if !match {
		return nil, domain.ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, domain.ErrUserDisabled
	}

	return &domain.ExternalIdentity{
		Subject:     user.Subject,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		Email:       user.Email,
		Groups:      nil,
		Raw:         nil,
	}, nil
}

func (p *LocalProvider) verifyDummy(password string) {
	_, err := p.hasher.Verify(password, p.dummyHash)
	if err != nil {
		p.logger.Error(
			"failed to verify dummy hash",
			"method",
			"LocalProvider.Authenticate",
			"error",
			err,
		)
	}
}
