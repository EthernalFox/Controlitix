package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/EthernalFox/Controlitix/ms-auth/internal/domain"
	"github.com/EthernalFox/Controlitix/ms-auth/internal/infrastructure/jwks"
	"github.com/EthernalFox/Controlitix/ms-auth/internal/infrastructure/repository"
	"github.com/golang-jwt/jwt/v5"
)

type UserLookup interface {
	GetByID(ctx context.Context, userID string) (*domain.User, error)
	GetRoles(ctx context.Context, userID string) ([]string, error)
	GetSourceType(ctx context.Context, sourceID int) (string, error)
}

type refreshTokenStore interface {
	Create(ctx context.Context, rec *domain.RefreshRecord) error
	FindByHash(ctx context.Context, tokenHash string) (*domain.RefreshRecord, error)
	MarkUsed(ctx context.Context, id string, at time.Time) error
	RevokeFamily(ctx context.Context, familyID string, at time.Time, reason string) error
	RevokeAllForUser(ctx context.Context, userID string, at time.Time, reason string) error
}

type TokenService struct {
	keystore   *jwks.Keystore
	refresh    refreshTokenStore
	userLookup UserLookup
	cfg        TokenConfig
	logger     *slog.Logger
	clock      func() time.Time
	rand       io.Reader
}

type TokenConfig struct {
	Issuer          string
	AudienceUser    string
	AudienceService string
	AccessTTL       time.Duration
	RefreshTTL      time.Duration
}

type TokenMeta struct {
	UserAgent string
	IP        string
}

func NewTokenService(
	keystore *jwks.Keystore,
	refresh *repository.RefreshTokenRepository,
	userLookup UserLookup,
	cfg TokenConfig,
	logger *slog.Logger,
) *TokenService {
	if logger == nil {
		logger = slog.Default()
	}

	return &TokenService{
		keystore:   keystore,
		refresh:    refresh,
		userLookup: userLookup,
		cfg:        cfg,
		logger:     logger,
		clock:      time.Now,
		rand:       rand.Reader,
	}
}

func (s *TokenService) IssueForUser(
	ctx context.Context,
	user *domain.User,
	source string,
	roles []string,
	meta TokenMeta,
) (*domain.IssuedTokens, error) {
	if user == nil {
		return nil, errors.New("user is nil")
	}

	now := s.clock().UTC()
	jti, err := newUUIDv4String(s.rand)
	if err != nil {
		return nil, fmt.Errorf("generate user access token jti: %w", err)
	}
	familyID, err := newUUIDv4String(s.rand)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token family id: %w", err)
	}

	accessToken, err := s.signAccessToken(domain.AccessClaims{
		Issuer:      s.cfg.Issuer,
		Audience:    []string{s.cfg.AudienceUser},
		Subject:     user.Subject,
		IssuedAt:    now.Unix(),
		NotBefore:   now.Unix(),
		ExpiresAt:   now.Add(s.cfg.AccessTTL).Unix(),
		JWTID:       jti,
		Source:      source,
		Roles:       normalizeList(roles),
		Scope:       "",
		Username:    user.Username,
		DisplayName: user.DisplayName,
	})
	if err != nil {
		return nil, fmt.Errorf("issue access token for user: %w", err)
	}

	refreshToken, tokenHash, err := s.generateRefreshTokenPair()
	if err != nil {
		return nil, fmt.Errorf("generate refresh token for user: %w", err)
	}

	if err := s.refresh.Create(ctx, &domain.RefreshRecord{
		UserID:    user.ID,
		FamilyID:  familyID,
		TokenHash: tokenHash,
		IssuedAt:  now,
		ExpiresAt: now.Add(s.cfg.RefreshTTL),
		UserAgent: meta.UserAgent,
		IP:        meta.IP,
	}); err != nil {
		return nil, fmt.Errorf("create refresh token for user: %w", err)
	}

	return &domain.IssuedTokens{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		ExpiresIn:        int(s.cfg.AccessTTL.Seconds()),
		RefreshExpiresIn: int(s.cfg.RefreshTTL.Seconds()),
		FamilyID:         familyID,
	}, nil
}

func (s *TokenService) IssueForService(
	ctx context.Context,
	clientID string,
	scopes []string,
	meta TokenMeta,
) (*domain.IssuedTokens, error) {
	_ = ctx
	_ = meta

	now := s.clock().UTC()
	jti, err := newUUIDv4String(s.rand)
	if err != nil {
		return nil, fmt.Errorf("generate service access token jti: %w", err)
	}

	accessToken, err := s.signAccessToken(domain.AccessClaims{
		Issuer:    s.cfg.Issuer,
		Audience:  []string{s.cfg.AudienceService},
		Subject:   "svc:" + clientID,
		IssuedAt:  now.Unix(),
		NotBefore: now.Unix(),
		ExpiresAt: now.Add(s.cfg.AccessTTL).Unix(),
		JWTID:     jti,
		Source:    "local",
		Roles:     []string{},
		Scope:     strings.Join(normalizeList(scopes), " "),
		ClientID:  clientID,
	})
	if err != nil {
		return nil, fmt.Errorf("issue access token for service: %w", err)
	}

	return &domain.IssuedTokens{
		AccessToken:      accessToken,
		RefreshToken:     "",
		ExpiresIn:        int(s.cfg.AccessTTL.Seconds()),
		RefreshExpiresIn: 0,
		FamilyID:         "",
	}, nil
}

func (s *TokenService) Rotate(
	ctx context.Context,
	refreshToken string,
	meta TokenMeta,
) (*domain.IssuedTokens, *domain.User, error) {
	now := s.clock().UTC()
	tokenHash, err := hashOpaqueRefreshToken(refreshToken)
	if err != nil {
		return nil, nil, domain.ErrInvalidRefresh
	}

	record, err := s.refresh.FindByHash(ctx, tokenHash)
	if err != nil {
		return nil, nil, fmt.Errorf("find refresh token by hash: %w", err)
	}
	if record == nil || record.RevokedAt != nil || !now.Before(record.ExpiresAt) {
		return nil, nil, domain.ErrInvalidRefresh
	}

	if record.UsedAt != nil {
		if revokeError := s.refresh.RevokeFamily(ctx, record.FamilyID, now, "reuse_detected"); revokeError != nil {
			return nil, nil, fmt.Errorf("revoke refresh family on reuse detection: %w", revokeError)
		}
		user, getUserError := s.userLookup.GetByID(ctx, record.UserID)
		if getUserError != nil && !errors.Is(getUserError, domain.ErrUserNotFound) {
			return nil, nil, fmt.Errorf("load user on refresh reuse: %w", getUserError)
		}
		return nil, user, domain.ErrRefreshReused
	}

	if err := s.refresh.MarkUsed(ctx, record.ID, now); err != nil {
		return nil, nil, fmt.Errorf("mark refresh token as used: %w", err)
	}

	user, err := s.userLookup.GetByID(ctx, record.UserID)
	if err != nil {
		return nil, nil, fmt.Errorf("load user by refresh token: %w", err)
	}
	roles, err := s.userLookup.GetRoles(ctx, user.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("load user roles by refresh token: %w", err)
	}
	sourceType, err := s.userLookup.GetSourceType(ctx, user.SourceID)
	if err != nil {
		return nil, nil, fmt.Errorf("load source type by refresh token: %w", err)
	}

	if !user.IsActive {
		if revokeError := s.refresh.RevokeFamily(ctx, record.FamilyID, now, "user_disabled"); revokeError != nil {
			return nil, nil, fmt.Errorf("revoke refresh family for disabled user: %w", revokeError)
		}
		return nil, user, domain.ErrUserDisabled
	}

	jti, err := newUUIDv4String(s.rand)
	if err != nil {
		return nil, nil, fmt.Errorf("generate rotated access token jti: %w", err)
	}

	accessToken, err := s.signAccessToken(domain.AccessClaims{
		Issuer:      s.cfg.Issuer,
		Audience:    []string{s.cfg.AudienceUser},
		Subject:     user.Subject,
		IssuedAt:    now.Unix(),
		NotBefore:   now.Unix(),
		ExpiresAt:   now.Add(s.cfg.AccessTTL).Unix(),
		JWTID:       jti,
		Source:      sourceType,
		Roles:       normalizeList(roles),
		Scope:       "",
		Username:    user.Username,
		DisplayName: user.DisplayName,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("issue rotated access token: %w", err)
	}

	newRefreshToken, newTokenHash, err := s.generateRefreshTokenPair()
	if err != nil {
		return nil, nil, fmt.Errorf("generate rotated refresh token: %w", err)
	}
	parentID := record.ID
	if err := s.refresh.Create(ctx, &domain.RefreshRecord{
		UserID:    user.ID,
		FamilyID:  record.FamilyID,
		TokenHash: newTokenHash,
		ParentID:  &parentID,
		IssuedAt:  now,
		ExpiresAt: now.Add(s.cfg.RefreshTTL),
		UserAgent: meta.UserAgent,
		IP:        meta.IP,
	}); err != nil {
		return nil, nil, fmt.Errorf("create rotated refresh token: %w", err)
	}

	return &domain.IssuedTokens{
		AccessToken:      accessToken,
		RefreshToken:     newRefreshToken,
		ExpiresIn:        int(s.cfg.AccessTTL.Seconds()),
		RefreshExpiresIn: int(s.cfg.RefreshTTL.Seconds()),
		FamilyID:         record.FamilyID,
	}, user, nil
}

func (s *TokenService) Revoke(
	ctx context.Context,
	refreshToken string,
	reason string,
) error {
	now := s.clock().UTC()
	tokenHash, err := hashOpaqueRefreshToken(refreshToken)
	if err != nil {
		return nil
	}

	record, err := s.refresh.FindByHash(ctx, tokenHash)
	if err != nil {
		return fmt.Errorf("find refresh token by hash for revoke: %w", err)
	}
	if record == nil {
		return nil
	}

	reason = strings.TrimSpace(reason)
	if reason == "" {
		reason = "logout"
	}

	if err := s.refresh.RevokeFamily(ctx, record.FamilyID, now, reason); err != nil {
		return fmt.Errorf("revoke refresh token family: %w", err)
	}

	return nil
}

func (s *TokenService) signAccessToken(claims domain.AccessClaims) (string, error) {
	jwtClaims := jwt.MapClaims{
		"iss":   claims.Issuer,
		"aud":   claims.Audience,
		"sub":   claims.Subject,
		"iat":   claims.IssuedAt,
		"exp":   claims.ExpiresAt,
		"nbf":   claims.NotBefore,
		"jti":   claims.JWTID,
		"src":   claims.Source,
		"roles": claims.Roles,
		"scope": claims.Scope,
	}
	if claims.Username != "" {
		jwtClaims["username"] = claims.Username
	}
	if claims.DisplayName != "" {
		jwtClaims["display_name"] = claims.DisplayName
	}
	if claims.ClientID != "" {
		jwtClaims["client_id"] = claims.ClientID
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwtClaims)
	token.Header["kid"] = s.keystore.KeyID()

	signedToken, err := token.SignedString(s.keystore.PrivateKey())
	if err != nil {
		return "", fmt.Errorf("sign jwt: %w", err)
	}

	return signedToken, nil
}

func (s *TokenService) generateRefreshTokenPair() (string, string, error) {
	rawToken := make([]byte, 32)
	if _, err := io.ReadFull(s.rand, rawToken); err != nil {
		return "", "", fmt.Errorf("generate random refresh token bytes: %w", err)
	}

	token := base64.RawURLEncoding.EncodeToString(rawToken)
	tokenHash := sha256.Sum256(rawToken)

	return token, hex.EncodeToString(tokenHash[:]), nil
}

func hashOpaqueRefreshToken(refreshToken string) (string, error) {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return "", errors.New("refresh token is empty")
	}

	rawToken, err := base64.RawURLEncoding.DecodeString(refreshToken)
	if err != nil {
		return "", fmt.Errorf("decode refresh token: %w", err)
	}
	if len(rawToken) == 0 {
		return "", errors.New("refresh token is empty")
	}

	tokenHash := sha256.Sum256(rawToken)
	return hex.EncodeToString(tokenHash[:]), nil
}

func newUUIDv4String(randomReader io.Reader) (string, error) {
	randomBytes := make([]byte, 16)
	if _, err := io.ReadFull(randomReader, randomBytes); err != nil {
		return "", fmt.Errorf("read random bytes for uuid: %w", err)
	}

	randomBytes[6] = (randomBytes[6] & 0x0f) | 0x40
	randomBytes[8] = (randomBytes[8] & 0x3f) | 0x80

	hexValue := hex.EncodeToString(randomBytes)
	return fmt.Sprintf(
		"%s-%s-%s-%s-%s",
		hexValue[0:8],
		hexValue[8:12],
		hexValue[12:16],
		hexValue[16:20],
		hexValue[20:32],
	), nil
}

func normalizeList(input []string) []string {
	seen := make(map[string]struct{}, len(input))
	result := make([]string, 0, len(input))
	for _, value := range input {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
