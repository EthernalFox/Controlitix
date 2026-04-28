package usecase

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/EthernalFox/Controlitix/ms-auth/internal/domain"
	"github.com/EthernalFox/Controlitix/ms-auth/internal/infrastructure/jwks"
	"github.com/golang-jwt/jwt/v5"
)

type refreshStoreMock struct {
	create          func(ctx context.Context, rec *domain.RefreshRecord) error
	findByHash      func(ctx context.Context, tokenHash string) (*domain.RefreshRecord, error)
	markUsed        func(ctx context.Context, id string, at time.Time) error
	revokeFamily    func(ctx context.Context, familyID string, at time.Time, reason string) error
	revokeAllForUser func(ctx context.Context, userID string, at time.Time, reason string) error
}

func (m *refreshStoreMock) Create(ctx context.Context, rec *domain.RefreshRecord) error {
	return m.create(ctx, rec)
}

func (m *refreshStoreMock) FindByHash(ctx context.Context, tokenHash string) (*domain.RefreshRecord, error) {
	return m.findByHash(ctx, tokenHash)
}

func (m *refreshStoreMock) MarkUsed(ctx context.Context, id string, at time.Time) error {
	return m.markUsed(ctx, id, at)
}

func (m *refreshStoreMock) RevokeFamily(ctx context.Context, familyID string, at time.Time, reason string) error {
	return m.revokeFamily(ctx, familyID, at, reason)
}

func (m *refreshStoreMock) RevokeAllForUser(ctx context.Context, userID string, at time.Time, reason string) error {
	return m.revokeAllForUser(ctx, userID, at, reason)
}

type userLookupMock struct {
	getByID      func(ctx context.Context, userID string) (*domain.User, error)
	getRoles     func(ctx context.Context, userID string) ([]string, error)
	getSourceType func(ctx context.Context, sourceID int) (string, error)
}

func (m *userLookupMock) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	return m.getByID(ctx, userID)
}

func (m *userLookupMock) GetRoles(ctx context.Context, userID string) ([]string, error) {
	return m.getRoles(ctx, userID)
}

func (m *userLookupMock) GetSourceType(ctx context.Context, sourceID int) (string, error) {
	return m.getSourceType(ctx, sourceID)
}

func TestTokenServiceIssueForUser(t *testing.T) {
	t.Parallel()

	keystore := testKeystore(t)
	now := time.Date(2026, 4, 25, 14, 0, 0, 0, time.UTC)
	user := &domain.User{
		ID:          "user-id",
		Subject:     "local:user-id",
		SourceID:    1,
		Username:    "admin",
		DisplayName: "Administrator",
		IsActive:    true,
	}

	var createdRecord *domain.RefreshRecord
	service := &TokenService{
		keystore: keystore,
		refresh: &refreshStoreMock{
			create: func(ctx context.Context, rec *domain.RefreshRecord) error {
				createdRecord = rec
				return nil
			},
			findByHash: func(ctx context.Context, tokenHash string) (*domain.RefreshRecord, error) {
				return nil, nil
			},
			markUsed: func(ctx context.Context, id string, at time.Time) error {
				return nil
			},
			revokeFamily: func(ctx context.Context, familyID string, at time.Time, reason string) error {
				return nil
			},
			revokeAllForUser: func(ctx context.Context, userID string, at time.Time, reason string) error {
				return nil
			},
		},
		userLookup: nil,
		cfg: TokenConfig{
			Issuer:          "controlitix-auth",
			AudienceUser:    "controlitix-api",
			AudienceService: "controlitix-internal",
			AccessTTL:       15 * time.Minute,
			RefreshTTL:      14 * 24 * time.Hour,
		},
		logger: slog.Default(),
		clock: func() time.Time {
			return now
		},
		rand: bytes.NewReader(bytes.Repeat([]byte{0x11}, 4096)),
	}

	tokens, err := service.IssueForUser(
		context.Background(),
		user,
		"local",
		[]string{"admin"},
		TokenMeta{UserAgent: "ua", IP: "127.0.0.1"},
	)
	if err != nil {
		t.Fatalf("issue for user: %v", err)
	}
	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Fatal("expected access and refresh tokens")
	}
	if createdRecord == nil {
		t.Fatal("expected refresh record to be created")
	}

	expectedRefreshHash, err := hashOpaqueRefreshToken(tokens.RefreshToken)
	if err != nil {
		t.Fatalf("hash opaque refresh token: %v", err)
	}
	if createdRecord.TokenHash != expectedRefreshHash {
		t.Fatal("refresh record hash does not match issued refresh token")
	}
	if createdRecord.UserID != user.ID {
		t.Fatalf("unexpected refresh record user id: %s", createdRecord.UserID)
	}

	claims := parseAndValidateJWTClaims(t, tokens.AccessToken, keystore.PrivateKey().Public().(*rsa.PublicKey))
	issuer, err := claims.GetIssuer()
	if err != nil || issuer != "controlitix-auth" {
		t.Fatalf("unexpected issuer: %s (err=%v)", issuer, err)
	}
	subject, err := claims.GetSubject()
	if err != nil || subject != user.Subject {
		t.Fatalf("unexpected subject: %s (err=%v)", subject, err)
	}
	audience, err := claims.GetAudience()
	if err != nil || len(audience) != 1 || audience[0] != "controlitix-api" {
		t.Fatalf("unexpected audience: %#v (err=%v)", audience, err)
	}
	if claims["src"] != "local" {
		t.Fatalf("unexpected src claim: %#v", claims["src"])
	}
	roles, ok := claims["roles"].([]any)
	if !ok || len(roles) != 1 || roles[0] != "admin" {
		t.Fatalf("unexpected roles claim: %#v", claims["roles"])
	}
}

func TestTokenServiceIssueForService(t *testing.T) {
	t.Parallel()

	keystore := testKeystore(t)
	service := &TokenService{
		keystore: keystore,
		refresh: &refreshStoreMock{
			create: func(ctx context.Context, rec *domain.RefreshRecord) error {
				t.Fatal("refresh create should not be called for service token")
				return nil
			},
			findByHash: func(ctx context.Context, tokenHash string) (*domain.RefreshRecord, error) {
				return nil, nil
			},
			markUsed: func(ctx context.Context, id string, at time.Time) error {
				return nil
			},
			revokeFamily: func(ctx context.Context, familyID string, at time.Time, reason string) error {
				return nil
			},
			revokeAllForUser: func(ctx context.Context, userID string, at time.Time, reason string) error {
				return nil
			},
		},
		cfg: TokenConfig{
			Issuer:          "controlitix-auth",
			AudienceUser:    "controlitix-api",
			AudienceService: "controlitix-internal",
			AccessTTL:       15 * time.Minute,
			RefreshTTL:      14 * 24 * time.Hour,
		},
		logger: slog.Default(),
		clock: func() time.Time {
			return time.Date(2026, 4, 25, 14, 0, 0, 0, time.UTC)
		},
		rand: bytes.NewReader(bytes.Repeat([]byte{0x22}, 4096)),
	}

	tokens, err := service.IssueForService(
		context.Background(),
		"ms-poll",
		[]string{"tags.values.write", "config.read"},
		TokenMeta{},
	)
	if err != nil {
		t.Fatalf("issue for service: %v", err)
	}
	if tokens.RefreshToken != "" || tokens.RefreshExpiresIn != 0 || tokens.FamilyID != "" {
		t.Fatalf("unexpected refresh payload for service token: %#v", tokens)
	}

	claims := parseAndValidateJWTClaims(t, tokens.AccessToken, keystore.PrivateKey().Public().(*rsa.PublicKey))
	audience, err := claims.GetAudience()
	if err != nil || len(audience) != 1 || audience[0] != "controlitix-internal" {
		t.Fatalf("unexpected audience: %#v (err=%v)", audience, err)
	}
}

func TestTokenServiceRotateValidToken(t *testing.T) {
	t.Parallel()

	keystore := testKeystore(t)
	now := time.Date(2026, 4, 25, 15, 0, 0, 0, time.UTC)
	rawRefresh := bytes.Repeat([]byte{0x33}, 32)
	refreshToken := base64.RawURLEncoding.EncodeToString(rawRefresh)
	refreshTokenHash, err := hashOpaqueRefreshToken(refreshToken)
	if err != nil {
		t.Fatalf("hash refresh token: %v", err)
	}

	record := &domain.RefreshRecord{
		ID:        "refresh-id",
		UserID:    "user-id",
		FamilyID:  "family-id",
		TokenHash: refreshTokenHash,
		IssuedAt:  now.Add(-time.Hour),
		ExpiresAt: now.Add(time.Hour),
	}

	markUsedCalled := false
	var createdRecord *domain.RefreshRecord
	service := &TokenService{
		keystore: keystore,
		refresh: &refreshStoreMock{
			create: func(ctx context.Context, rec *domain.RefreshRecord) error {
				createdRecord = rec
				return nil
			},
			findByHash: func(ctx context.Context, tokenHash string) (*domain.RefreshRecord, error) {
				if tokenHash != refreshTokenHash {
					t.Fatalf("unexpected token hash: %s", tokenHash)
				}
				return record, nil
			},
			markUsed: func(ctx context.Context, id string, at time.Time) error {
				markUsedCalled = true
				if id != record.ID {
					t.Fatalf("unexpected refresh id: %s", id)
				}
				if !at.Equal(now) {
					t.Fatalf("unexpected used_at: %s", at)
				}
				return nil
			},
			revokeFamily: func(ctx context.Context, familyID string, at time.Time, reason string) error {
				return nil
			},
			revokeAllForUser: func(ctx context.Context, userID string, at time.Time, reason string) error {
				return nil
			},
		},
		userLookup: &userLookupMock{
			getByID: func(ctx context.Context, userID string) (*domain.User, error) {
				return &domain.User{
					ID:          "user-id",
					Subject:     "local:user-id",
					SourceID:    1,
					Username:    "admin",
					DisplayName: "Administrator",
					IsActive:    true,
				}, nil
			},
			getRoles: func(ctx context.Context, userID string) ([]string, error) {
				return []string{"admin"}, nil
			},
			getSourceType: func(ctx context.Context, sourceID int) (string, error) {
				return "local", nil
			},
		},
		cfg: TokenConfig{
			Issuer:          "controlitix-auth",
			AudienceUser:    "controlitix-api",
			AudienceService: "controlitix-internal",
			AccessTTL:       15 * time.Minute,
			RefreshTTL:      14 * 24 * time.Hour,
		},
		logger: slog.Default(),
		clock: func() time.Time {
			return now
		},
		rand: bytes.NewReader(bytes.Repeat([]byte{0x44}, 4096)),
	}

	tokens, user, err := service.Rotate(
		context.Background(),
		refreshToken,
		TokenMeta{UserAgent: "ua", IP: "127.0.0.1"},
	)
	if err != nil {
		t.Fatalf("rotate: %v", err)
	}
	if !markUsedCalled {
		t.Fatal("expected mark used to be called")
	}
	if user == nil || user.ID != "user-id" {
		t.Fatalf("unexpected user returned from rotate: %#v", user)
	}
	if tokens == nil || tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Fatalf("unexpected issued tokens from rotate: %#v", tokens)
	}
	if createdRecord == nil {
		t.Fatal("expected new refresh record to be created")
	}
	if createdRecord.FamilyID != record.FamilyID {
		t.Fatalf("expected family to remain the same, got %s", createdRecord.FamilyID)
	}
	if createdRecord.ParentID == nil || *createdRecord.ParentID != record.ID {
		t.Fatalf("unexpected parent id: %#v", createdRecord.ParentID)
	}
}

func TestTokenServiceRotateReusedToken(t *testing.T) {
	t.Parallel()

	keystore := testKeystore(t)
	now := time.Date(2026, 4, 25, 16, 0, 0, 0, time.UTC)
	rawRefresh := bytes.Repeat([]byte{0x55}, 32)
	refreshToken := base64.RawURLEncoding.EncodeToString(rawRefresh)
	refreshTokenHash, err := hashOpaqueRefreshToken(refreshToken)
	if err != nil {
		t.Fatalf("hash refresh token: %v", err)
	}

	revokeCalled := false
	service := &TokenService{
		keystore: keystore,
		refresh: &refreshStoreMock{
			create: func(ctx context.Context, rec *domain.RefreshRecord) error {
				return nil
			},
			findByHash: func(ctx context.Context, tokenHash string) (*domain.RefreshRecord, error) {
				usedAt := now.Add(-time.Minute)
				return &domain.RefreshRecord{
					ID:        "refresh-id",
					UserID:    "user-id",
					FamilyID:  "family-id",
					TokenHash: refreshTokenHash,
					IssuedAt:  now.Add(-time.Hour),
					ExpiresAt: now.Add(time.Hour),
					UsedAt:    &usedAt,
				}, nil
			},
			markUsed: func(ctx context.Context, id string, at time.Time) error {
				t.Fatal("mark used should not be called for reused token")
				return nil
			},
			revokeFamily: func(ctx context.Context, familyID string, at time.Time, reason string) error {
				revokeCalled = true
				if familyID != "family-id" || reason != "reuse_detected" {
					t.Fatalf("unexpected revoke family payload: family=%s reason=%s", familyID, reason)
				}
				return nil
			},
			revokeAllForUser: func(ctx context.Context, userID string, at time.Time, reason string) error {
				return nil
			},
		},
		userLookup: &userLookupMock{
			getByID: func(ctx context.Context, userID string) (*domain.User, error) {
				return &domain.User{ID: userID}, nil
			},
			getRoles: func(ctx context.Context, userID string) ([]string, error) {
				return []string{}, nil
			},
			getSourceType: func(ctx context.Context, sourceID int) (string, error) {
				return "local", nil
			},
		},
		cfg: TokenConfig{
			Issuer:          "controlitix-auth",
			AudienceUser:    "controlitix-api",
			AudienceService: "controlitix-internal",
			AccessTTL:       15 * time.Minute,
			RefreshTTL:      14 * 24 * time.Hour,
		},
		logger: slog.Default(),
		clock: func() time.Time {
			return now
		},
		rand: bytes.NewReader(bytes.Repeat([]byte{0x66}, 4096)),
	}

	_, user, err := service.Rotate(context.Background(), refreshToken, TokenMeta{})
	if !errors.Is(err, domain.ErrRefreshReused) {
		t.Fatalf("expected ErrRefreshReused, got %v", err)
	}
	if !revokeCalled {
		t.Fatal("expected revoke family to be called for reused token")
	}
	if user == nil || user.ID != "user-id" {
		t.Fatalf("unexpected user returned for reused token: %#v", user)
	}
}

func TestTokenServiceRotateInvalidStates(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 25, 16, 0, 0, 0, time.UTC)
	tests := []struct {
		name   string
		record *domain.RefreshRecord
	}{
		{
			name:   "unknown token",
			record: nil,
		},
		{
			name: "revoked token",
			record: &domain.RefreshRecord{
				ID:        "id",
				UserID:    "user-id",
				FamilyID:  "family-id",
				TokenHash: "hash",
				IssuedAt:  now.Add(-time.Hour),
				ExpiresAt: now.Add(time.Hour),
				RevokedAt: timePointer(now.Add(-time.Minute)),
			},
		},
		{
			name: "expired token",
			record: &domain.RefreshRecord{
				ID:        "id",
				UserID:    "user-id",
				FamilyID:  "family-id",
				TokenHash: "hash",
				IssuedAt:  now.Add(-2 * time.Hour),
				ExpiresAt: now.Add(-time.Minute),
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rawRefresh := bytes.Repeat([]byte{0x77}, 32)
			refreshToken := base64.RawURLEncoding.EncodeToString(rawRefresh)
			refreshTokenHash, err := hashOpaqueRefreshToken(refreshToken)
			if err != nil {
				t.Fatalf("hash refresh token: %v", err)
			}

			service := &TokenService{
				keystore: testKeystore(t),
				refresh: &refreshStoreMock{
					create: func(ctx context.Context, rec *domain.RefreshRecord) error { return nil },
					findByHash: func(ctx context.Context, tokenHash string) (*domain.RefreshRecord, error) {
						if tokenHash != refreshTokenHash {
							t.Fatalf("unexpected token hash: %s", tokenHash)
						}
						return tc.record, nil
					},
					markUsed: func(ctx context.Context, id string, at time.Time) error { return nil },
					revokeFamily: func(ctx context.Context, familyID string, at time.Time, reason string) error {
						return nil
					},
					revokeAllForUser: func(ctx context.Context, userID string, at time.Time, reason string) error {
						return nil
					},
				},
				userLookup: &userLookupMock{
					getByID: func(ctx context.Context, userID string) (*domain.User, error) { return nil, nil },
					getRoles: func(ctx context.Context, userID string) ([]string, error) { return []string{}, nil },
					getSourceType: func(ctx context.Context, sourceID int) (string, error) { return "local", nil },
				},
				cfg: TokenConfig{
					Issuer:          "controlitix-auth",
					AudienceUser:    "controlitix-api",
					AudienceService: "controlitix-internal",
					AccessTTL:       15 * time.Minute,
					RefreshTTL:      14 * 24 * time.Hour,
				},
				logger: slog.Default(),
				clock: func() time.Time {
					return now
				},
				rand: bytes.NewReader(bytes.Repeat([]byte{0x78}, 4096)),
			}

			_, _, err = service.Rotate(context.Background(), refreshToken, TokenMeta{})
			if !errors.Is(err, domain.ErrInvalidRefresh) {
				t.Fatalf("expected ErrInvalidRefresh, got %v", err)
			}
		})
	}
}

func TestTokenServiceRevokeUnknownToken(t *testing.T) {
	t.Parallel()

	service := &TokenService{
		keystore: testKeystore(t),
		refresh: &refreshStoreMock{
			create: func(ctx context.Context, rec *domain.RefreshRecord) error { return nil },
			findByHash: func(ctx context.Context, tokenHash string) (*domain.RefreshRecord, error) {
				return nil, nil
			},
			markUsed: func(ctx context.Context, id string, at time.Time) error { return nil },
			revokeFamily: func(ctx context.Context, familyID string, at time.Time, reason string) error {
				t.Fatal("revoke family should not be called for unknown token")
				return nil
			},
			revokeAllForUser: func(ctx context.Context, userID string, at time.Time, reason string) error {
				return nil
			},
		},
		cfg: TokenConfig{
			Issuer:          "controlitix-auth",
			AudienceUser:    "controlitix-api",
			AudienceService: "controlitix-internal",
			AccessTTL:       15 * time.Minute,
			RefreshTTL:      14 * 24 * time.Hour,
		},
		logger: slog.Default(),
		clock:  time.Now,
		rand:   bytes.NewReader(bytes.Repeat([]byte{0x79}, 4096)),
	}

	if err := service.Revoke(context.Background(), base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{0x80}, 32)), "logout"); err != nil {
		t.Fatalf("revoke unknown token: %v", err)
	}
}

func TestTokenServiceRotateDisabledUserRevokesFamily(t *testing.T) {
	t.Parallel()

	keystore := testKeystore(t)
	now := time.Date(2026, 4, 25, 17, 0, 0, 0, time.UTC)
	rawRefresh := bytes.Repeat([]byte{0x81}, 32)
	refreshToken := base64.RawURLEncoding.EncodeToString(rawRefresh)
	refreshTokenHash, err := hashOpaqueRefreshToken(refreshToken)
	if err != nil {
		t.Fatalf("hash refresh token: %v", err)
	}

	record := &domain.RefreshRecord{
		ID:        "refresh-id",
		UserID:    "user-id",
		FamilyID:  "family-id",
		TokenHash: refreshTokenHash,
		IssuedAt:  now.Add(-time.Hour),
		ExpiresAt: now.Add(time.Hour),
	}

	revokeCalled := false
	service := &TokenService{
		keystore: keystore,
		refresh: &refreshStoreMock{
			create: func(ctx context.Context, rec *domain.RefreshRecord) error { return nil },
			findByHash: func(ctx context.Context, tokenHash string) (*domain.RefreshRecord, error) {
				return record, nil
			},
			markUsed: func(ctx context.Context, id string, at time.Time) error {
				return nil
			},
			revokeFamily: func(ctx context.Context, familyID string, at time.Time, reason string) error {
				revokeCalled = true
				if familyID != "family-id" || reason != "user_disabled" {
					t.Fatalf("unexpected revoke payload: family=%s reason=%s", familyID, reason)
				}
				return nil
			},
			revokeAllForUser: func(ctx context.Context, userID string, at time.Time, reason string) error {
				return nil
			},
		},
		userLookup: &userLookupMock{
			getByID: func(ctx context.Context, userID string) (*domain.User, error) {
				return &domain.User{
					ID:       "user-id",
					Subject:  "local:user-id",
					SourceID: 1,
					IsActive: false,
				}, nil
			},
			getRoles: func(ctx context.Context, userID string) ([]string, error) {
				return []string{"admin"}, nil
			},
			getSourceType: func(ctx context.Context, sourceID int) (string, error) {
				return "local", nil
			},
		},
		cfg: TokenConfig{
			Issuer:          "controlitix-auth",
			AudienceUser:    "controlitix-api",
			AudienceService: "controlitix-internal",
			AccessTTL:       15 * time.Minute,
			RefreshTTL:      14 * 24 * time.Hour,
		},
		logger: slog.Default(),
		clock: func() time.Time {
			return now
		},
		rand: bytes.NewReader(bytes.Repeat([]byte{0x82}, 4096)),
	}

	_, user, err := service.Rotate(context.Background(), refreshToken, TokenMeta{})
	if !errors.Is(err, domain.ErrUserDisabled) {
		t.Fatalf("expected ErrUserDisabled, got %v", err)
	}
	if user == nil || user.ID != "user-id" {
		t.Fatalf("unexpected user for disabled rotate: %#v", user)
	}
	if !revokeCalled {
		t.Fatal("expected revoke family for disabled user")
	}
}

func testKeystore(t *testing.T) *jwks.Keystore {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	keystore, err := jwks.NewKeystore(privateKeyPEM, "test-kid")
	if err != nil {
		t.Fatalf("new keystore: %v", err)
	}

	return keystore
}

func parseAndValidateJWTClaims(t *testing.T, tokenString string, publicKey *rsa.PublicKey) jwt.MapClaims {
	t.Helper()

	parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return publicKey, nil
	}, jwt.WithValidMethods([]string{"RS256"}))
	if err != nil {
		t.Fatalf("parse jwt: %v", err)
	}
	if !parsedToken.Valid {
		t.Fatal("expected jwt to be valid")
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatalf("unexpected claims type: %T", parsedToken.Claims)
	}

	return claims
}

func timePointer(value time.Time) *time.Time {
	return &value
}
