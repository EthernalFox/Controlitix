package repository

import (
	"context"

	"github.com/EthernalFox/Controlitix/ms-auth/internal/domain"
)

type UserLookupAdapter struct {
	users   *UserRepository
	sources *IdentitySourceRepository
}

func NewUserLookupAdapter(
	users *UserRepository,
	sources *IdentitySourceRepository,
) *UserLookupAdapter {
	return &UserLookupAdapter{
		users:   users,
		sources: sources,
	}
}

func (a *UserLookupAdapter) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	user, err := a.users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domain.ErrUserNotFound
	}

	return user, nil
}

func (a *UserLookupAdapter) GetRoles(ctx context.Context, userID string) ([]string, error) {
	roles, err := a.users.GetRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	return roles, nil
}

func (a *UserLookupAdapter) GetSourceType(ctx context.Context, sourceID int) (string, error) {
	source, err := a.sources.FindByID(ctx, sourceID)
	if err != nil {
		return "", err
	}
	if source == nil {
		return "", domain.ErrSourceNotFound
	}

	return source.Type, nil
}
