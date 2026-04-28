package authctx

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestValidatorParseSuccess(t *testing.T) {
	t.Parallel()

	now := time.Unix(1_700_000_000, 0)
	privateKey, validator := newValidatorFixture(t, now)

	token := signRS256Token(t, privateKey, "kid-1", defaultClaims(now))
	principal, err := validator.Parse(context.Background(), token)
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}

	if principal.Subject != "local:user-1" {
		t.Fatalf("unexpected subject: %s", principal.Subject)
	}
	if !principal.HasRole("engineer") {
		t.Fatalf("expected role engineer, got %#v", principal.Roles)
	}
	if !principal.HasScope("config.read") || !principal.HasScope("tags.values.write") {
		t.Fatalf("unexpected scopes: %#v", principal.Scopes)
	}
	if len(principal.Audience) != 1 || principal.Audience[0] != "controlitix-api" {
		t.Fatalf("unexpected audience: %#v", principal.Audience)
	}
	if principal.Issuer != "controlitix-auth" {
		t.Fatalf("unexpected issuer: %s", principal.Issuer)
	}
	if principal.Source != "local" {
		t.Fatalf("unexpected source: %s", principal.Source)
	}
	if principal.Username != "engineer" {
		t.Fatalf("unexpected username: %s", principal.Username)
	}
	if principal.DisplayName != "Engineer User" {
		t.Fatalf("unexpected display name: %s", principal.DisplayName)
	}
	if principal.ClientID != "editor-ui" {
		t.Fatalf("unexpected client id: %s", principal.ClientID)
	}
	if principal.TokenID != "token-id-1" {
		t.Fatalf("unexpected token id: %s", principal.TokenID)
	}
}

func TestValidatorRejectsHS256AlgorithmConfusion(t *testing.T) {
	t.Parallel()

	now := time.Unix(1_700_000_000, 0)
	privateKey, validator := newValidatorFixture(t, now)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, defaultClaims(now))
	token.Header["kid"] = "kid-1"
	secret := x509.MarshalPKCS1PublicKey(&privateKey.PublicKey)
	signedToken, err := token.SignedString(secret)
	if err != nil {
		t.Fatalf("sign hs256 token: %v", err)
	}

	if _, err := validator.Parse(context.Background(), signedToken); err != ErrInvalidSignature {
		t.Fatalf("expected ErrInvalidSignature, got %v", err)
	}
}

func TestValidatorRejectsNoneAlgorithm(t *testing.T) {
	t.Parallel()

	now := time.Unix(1_700_000_000, 0)
	_, validator := newValidatorFixture(t, now)

	token := jwt.NewWithClaims(jwt.SigningMethodNone, defaultClaims(now))
	token.Header["kid"] = "kid-1"
	signedToken, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("sign none token: %v", err)
	}

	if _, err := validator.Parse(context.Background(), signedToken); err != ErrInvalidSignature {
		t.Fatalf("expected ErrInvalidSignature, got %v", err)
	}
}

func TestValidatorExpiredToken(t *testing.T) {
	t.Parallel()

	now := time.Unix(1_700_000_000, 0)
	privateKey, validator := newValidatorFixture(t, now)

	claims := defaultClaims(now)
	claims["exp"] = now.Add(-2 * time.Minute).Unix()
	token := signRS256Token(t, privateKey, "kid-1", claims)

	if _, err := validator.Parse(context.Background(), token); err != ErrTokenExpired {
		t.Fatalf("expected ErrTokenExpired, got %v", err)
	}
}

func TestValidatorNotYetValidToken(t *testing.T) {
	t.Parallel()

	now := time.Unix(1_700_000_000, 0)
	privateKey, validator := newValidatorFixture(t, now)

	claims := defaultClaims(now)
	claims["nbf"] = now.Add(2 * time.Minute).Unix()
	token := signRS256Token(t, privateKey, "kid-1", claims)

	if _, err := validator.Parse(context.Background(), token); err != ErrTokenNotYetValid {
		t.Fatalf("expected ErrTokenNotYetValid, got %v", err)
	}
}

func TestValidatorInvalidIssuer(t *testing.T) {
	t.Parallel()

	now := time.Unix(1_700_000_000, 0)
	privateKey, validator := newValidatorFixture(t, now)

	claims := defaultClaims(now)
	claims["iss"] = "unexpected-issuer"
	token := signRS256Token(t, privateKey, "kid-1", claims)

	if _, err := validator.Parse(context.Background(), token); err != ErrInvalidIssuer {
		t.Fatalf("expected ErrInvalidIssuer, got %v", err)
	}
}

func TestValidatorInvalidAudience(t *testing.T) {
	t.Parallel()

	now := time.Unix(1_700_000_000, 0)
	privateKey, validator := newValidatorFixture(t, now)

	claims := defaultClaims(now)
	claims["aud"] = []string{"unexpected-audience"}
	token := signRS256Token(t, privateKey, "kid-1", claims)

	if _, err := validator.Parse(context.Background(), token); err != ErrInvalidAudience {
		t.Fatalf("expected ErrInvalidAudience, got %v", err)
	}
}

func TestValidatorUnknownKey(t *testing.T) {
	t.Parallel()

	now := time.Unix(1_700_000_000, 0)
	privateKey, validator := newValidatorFixture(t, now)

	token := signRS256Token(t, privateKey, "kid-2", defaultClaims(now))
	if _, err := validator.Parse(context.Background(), token); err != ErrUnknownKey {
		t.Fatalf("expected ErrUnknownKey, got %v", err)
	}
}

func TestValidatorTokenWithoutKid(t *testing.T) {
	t.Parallel()

	now := time.Unix(1_700_000_000, 0)
	privateKey, validator := newValidatorFixture(t, now)

	token := signRS256Token(t, privateKey, "", defaultClaims(now))
	if _, err := validator.Parse(context.Background(), token); err != ErrTokenMalformed {
		t.Fatalf("expected ErrTokenMalformed, got %v", err)
	}
}

func TestValidatorMalformedToken(t *testing.T) {
	t.Parallel()

	now := time.Unix(1_700_000_000, 0)
	_, validator := newValidatorFixture(t, now)

	if _, err := validator.Parse(context.Background(), "%%%"); err != ErrTokenMalformed {
		t.Fatalf("expected ErrTokenMalformed, got %v", err)
	}
}

func TestValidatorDisabledBypass(t *testing.T) {
	t.Parallel()

	now := time.Unix(1_700_000_000, 0)
	validator := NewValidator(ValidatorOptions{
		Disabled: true,
		Clock: func() time.Time {
			return now
		},
		Logger: testLogger(),
	})

	principal, err := validator.Parse(context.Background(), "")
	if err != nil {
		t.Fatalf("parse token in disabled mode: %v", err)
	}
	if principal.Subject != "dev:anonymous" {
		t.Fatalf("unexpected subject in disabled mode: %s", principal.Subject)
	}
	if !principal.HasRole("admin") {
		t.Fatalf("expected admin role in disabled mode, got %#v", principal.Roles)
	}
}

func newValidatorFixture(t *testing.T, now time.Time) (*rsa.PrivateKey, *Validator) {
	t.Helper()

	privateKey := generateTestRSAKey(t)
	cache := &JWKSCache{
		ttl:       time.Hour,
		logger:    testLogger(),
		keys:      map[string]*rsa.PublicKey{"kid-1": &privateKey.PublicKey},
		fetchedAt: now,
	}

	validator := NewValidator(ValidatorOptions{
		JWKSCache:         cache,
		ExpectedIssuer:    "controlitix-auth",
		ExpectedAudiences: []string{"controlitix-api"},
		Clock: func() time.Time {
			return now
		},
		Leeway: 30 * time.Second,
		Logger: testLogger(),
	})

	return privateKey, validator
}

func signRS256Token(
	t *testing.T,
	privateKey *rsa.PrivateKey,
	kid string,
	claims jwt.MapClaims,
) string {
	t.Helper()

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	if kid != "" {
		token.Header["kid"] = kid
	}

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("sign rs256 token: %v", err)
	}

	return tokenString
}

func defaultClaims(now time.Time) jwt.MapClaims {
	return jwt.MapClaims{
		"iss":          "controlitix-auth",
		"aud":          []string{"controlitix-api"},
		"sub":          "local:user-1",
		"exp":          now.Add(time.Hour).Unix(),
		"nbf":          now.Add(-time.Minute).Unix(),
		"jti":          "token-id-1",
		"src":          "local",
		"roles":        []string{"engineer"},
		"scope":        "config.read tags.values.write",
		"username":     "engineer",
		"display_name": "Engineer User",
		"client_id":    "editor-ui",
	}
}
