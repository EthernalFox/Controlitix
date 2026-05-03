package authctx

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const defaultLeeway = 30 * time.Second

type Validator struct {
	cache        *JWKSCache
	expectedIss  string
	expectedAuds []string
	clock        func() time.Time
	leeway       time.Duration
	logger       *slog.Logger
	disabled     bool
}

type ValidatorOptions struct {
	JWKSCache         *JWKSCache
	ExpectedIssuer    string
	ExpectedAudiences []string
	Leeway            time.Duration
	Clock             func() time.Time
	Logger            *slog.Logger
	Disabled          bool
}

func NewValidator(options ValidatorOptions) *Validator {
	clock := options.Clock
	if clock == nil {
		clock = time.Now
	}

	leeway := options.Leeway
	if leeway <= 0 {
		leeway = defaultLeeway
	}

	logger := options.Logger
	if logger == nil {
		logger = slog.Default()
	}

	return &Validator{
		cache:        options.JWKSCache,
		expectedIss:  strings.TrimSpace(options.ExpectedIssuer),
		expectedAuds: normalizeStringList(options.ExpectedAudiences),
		clock:        clock,
		leeway:       leeway,
		logger:       logger,
		disabled:     options.Disabled,
	}
}

func (validator *Validator) Parse(
	ctx context.Context,
	tokenString string,
) (*Principal, error) {
	if validator.disabled {
		return &Principal{
			Subject:   "dev:anonymous",
			Roles:     []string{"admin"},
			Source:    "dev",
			ExpiresAt: validator.clock().Add(24 * time.Hour),
		}, nil
	}

	if strings.TrimSpace(tokenString) == "" {
		return nil, ErrTokenMalformed
	}
	if validator.cache == nil {
		return nil, ErrInvalidSignature
	}

	token, err := jwt.Parse(tokenString, func(parsedToken *jwt.Token) (any, error) {
		if parsedToken.Method == nil || parsedToken.Method.Alg() != jwt.SigningMethodRS256.Alg() {
			return nil, ErrInvalidSignature
		}

		kid, _ := parsedToken.Header["kid"].(string)
		kid = strings.TrimSpace(kid)
		if kid == "" {
			return nil, ErrTokenMalformed
		}

		publicKey, keyError := validator.cache.GetKey(ctx, kid)
		if keyError != nil {
			return nil, keyError
		}

		return publicKey, nil
	}, jwt.WithoutClaimsValidation())
	if err != nil {
		return nil, mapJWTParseError(err)
	}
	if token == nil || !token.Valid {
		return nil, ErrInvalidSignature
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrTokenMalformed
	}

	expiresAt, err := claims.GetExpirationTime()
	if err != nil || expiresAt == nil {
		return nil, ErrTokenMalformed
	}
	if validator.clock().After(expiresAt.Time.Add(validator.leeway)) {
		return nil, ErrTokenExpired
	}

	notBefore, err := claims.GetNotBefore()
	if err != nil {
		return nil, ErrTokenMalformed
	}
	if notBefore != nil && validator.clock().Add(validator.leeway).Before(notBefore.Time) {
		return nil, ErrTokenNotYetValid
	}

	issuer, err := claims.GetIssuer()
	if err != nil {
		return nil, ErrTokenMalformed
	}
	issuer = strings.TrimSpace(issuer)
	if validator.expectedIss != "" && issuer != validator.expectedIss {
		return nil, ErrInvalidIssuer
	}

	audiences, err := claims.GetAudience()
	if err != nil {
		return nil, ErrTokenMalformed
	}
	audiences = normalizeStringList(audiences)
	if len(validator.expectedAuds) > 0 && !hasAny(audiences, validator.expectedAuds) {
		return nil, ErrInvalidAudience
	}

	subject, err := claims.GetSubject()
	if err != nil || strings.TrimSpace(subject) == "" {
		return nil, ErrTokenMalformed
	}

	tokenID := extractStringClaim(claims["jti"])
	roles, _ := extractStringSliceClaim(claims["roles"])
	scope := strings.TrimSpace(extractStringClaim(claims["scope"]))
	scopes := []string{}
	if scope != "" {
		scopes = strings.Fields(scope)
	}

	return &Principal{
		Subject:     strings.TrimSpace(subject),
		Roles:       roles,
		Scopes:      scopes,
		Audience:    audiences,
		Issuer:      issuer,
		Source:      extractStringClaim(claims["src"]),
		Username:    extractStringClaim(claims["username"]),
		DisplayName: extractStringClaim(claims["display_name"]),
		ClientID:    extractStringClaim(claims["client_id"]),
		TokenID:     strings.TrimSpace(tokenID),
		ExpiresAt:   expiresAt.Time,
	}, nil
}

func mapJWTParseError(err error) error {
	switch {
	case errors.Is(err, ErrTokenMalformed):
		return ErrTokenMalformed
	case errors.Is(err, ErrUnknownKey):
		return ErrUnknownKey
	case errors.Is(err, ErrInvalidSignature):
		return ErrInvalidSignature
	case errors.Is(err, jwt.ErrTokenMalformed):
		return ErrTokenMalformed
	case errors.Is(err, jwt.ErrTokenSignatureInvalid):
		return ErrInvalidSignature
	default:
		return ErrInvalidSignature
	}
}

func extractStringClaim(rawValue any) string {
	value, _ := rawValue.(string)
	return strings.TrimSpace(value)
}

func extractStringSliceClaim(rawValue any) ([]string, bool) {
	switch value := rawValue.(type) {
	case []string:
		return normalizeStringList(value), true
	case []any:
		result := make([]string, 0, len(value))
		for _, item := range value {
			text, ok := item.(string)
			if !ok {
				continue
			}

			result = append(result, text)
		}
		return normalizeStringList(result), true
	default:
		return nil, false
	}
}

func normalizeStringList(values []string) []string {
	normalized := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}

		seen[value] = struct{}{}
		normalized = append(normalized, value)
	}

	return normalized
}

func hasAny(actual []string, expected []string) bool {
	if len(actual) == 0 || len(expected) == 0 {
		return false
	}

	expectedSet := make(map[string]struct{}, len(expected))
	for _, value := range expected {
		expectedSet[value] = struct{}{}
	}

	for _, value := range actual {
		if _, ok := expectedSet[value]; ok {
			return true
		}
	}

	return false
}
