package domain

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserDisabled       = errors.New("user disabled")
	ErrSourceDisabled     = errors.New("identity source disabled")
	ErrUserNotFound       = errors.New("user not found")
	ErrSourceNotFound     = errors.New("identity source not found")
	ErrUnsupportedGrant   = errors.New("unsupported grant type")
	ErrInvalidClient      = errors.New("invalid client")
	ErrClientDisabled     = errors.New("client disabled")
	ErrInvalidScope       = errors.New("invalid scope")
	ErrInvalidRefresh     = errors.New("invalid refresh token")
	ErrRefreshReused      = errors.New("refresh token reused")
	ErrValidation         = errors.New("validation failed")
	ErrForbidden          = errors.New("forbidden")
	ErrForbiddenSelfDemote = errors.New("forbidden self demote")
	ErrLastAdmin          = errors.New("last admin")
	ErrClientIDExists     = errors.New("client id already exists")
	ErrUserExists         = errors.New("user already exists")
	ErrServiceNotFound    = errors.New("service account not found")
	ErrNotImplemented     = errors.New("not implemented")
)
