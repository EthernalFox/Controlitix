package authctx

import "errors"

var (
	ErrMissingAuth      = errors.New("missing authorization header")
	ErrTokenMalformed   = errors.New("token malformed")
	ErrInvalidSignature = errors.New("invalid signature")
	ErrInvalidIssuer    = errors.New("invalid issuer")
	ErrInvalidAudience  = errors.New("invalid audience")
	ErrTokenExpired     = errors.New("token expired")
	ErrTokenNotYetValid = errors.New("token not yet valid")
	ErrUnknownKey       = errors.New("unknown signing key")
	ErrForbidden        = errors.New("forbidden")
)
