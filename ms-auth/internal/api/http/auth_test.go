package http

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/EthernalFox/Controlitix/ms-auth/internal/domain"
	"github.com/EthernalFox/Controlitix/ms-auth/internal/infrastructure/jwks"
	"github.com/EthernalFox/Controlitix/ms-auth/internal/usecase"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
)

type authServiceMock struct {
	authenticate func(
		ctx context.Context,
		sourceType string,
		creds domain.Credentials,
	) (*domain.AuthenticatedUser, error)
}

func (m *authServiceMock) Authenticate(
	ctx context.Context,
	sourceType string,
	creds domain.Credentials,
) (*domain.AuthenticatedUser, error) {
	return m.authenticate(ctx, sourceType, creds)
}

type tokenServiceMock struct {
	issueForUser func(
		ctx context.Context,
		user *domain.User,
		source string,
		roles []string,
		meta usecase.TokenMeta,
	) (*domain.IssuedTokens, error)
	rotate func(
		ctx context.Context,
		refreshToken string,
		meta usecase.TokenMeta,
	) (*domain.IssuedTokens, *domain.User, error)
	revoke func(ctx context.Context, refreshToken string, reason string) error
}

func (m *tokenServiceMock) IssueForUser(
	ctx context.Context,
	user *domain.User,
	source string,
	roles []string,
	meta usecase.TokenMeta,
) (*domain.IssuedTokens, error) {
	return m.issueForUser(ctx, user, source, roles, meta)
}

func (m *tokenServiceMock) Rotate(
	ctx context.Context,
	refreshToken string,
	meta usecase.TokenMeta,
) (*domain.IssuedTokens, *domain.User, error) {
	return m.rotate(ctx, refreshToken, meta)
}

func (m *tokenServiceMock) Revoke(
	ctx context.Context,
	refreshToken string,
	reason string,
) error {
	return m.revoke(ctx, refreshToken, reason)
}

type userRepositoryMock struct {
	findBySubject func(ctx context.Context, subject string) (*domain.User, error)
	getRoles      func(ctx context.Context, userID string) ([]string, error)
}

func (m *userRepositoryMock) FindBySubject(ctx context.Context, subject string) (*domain.User, error) {
	return m.findBySubject(ctx, subject)
}

func (m *userRepositoryMock) GetRoles(ctx context.Context, userID string) ([]string, error) {
	return m.getRoles(ctx, userID)
}

type sourceRepositoryMock struct {
	findByID func(ctx context.Context, id int) (*domain.IdentitySource, error)
}

func (m *sourceRepositoryMock) FindByID(ctx context.Context, id int) (*domain.IdentitySource, error) {
	return m.findByID(ctx, id)
}

type auditRecorderMock struct {
	mutex  sync.Mutex
	events []domain.AuditEvent
	record func(ctx context.Context, event domain.AuditEvent) error
}

func (m *auditRecorderMock) Record(ctx context.Context, event domain.AuditEvent) error {
	if m.record != nil {
		return m.record(ctx, event)
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.events = append(m.events, event)
	return nil
}

func (m *auditRecorderMock) Events() []domain.AuditEvent {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	result := make([]domain.AuditEvent, len(m.events))
	copy(result, m.events)
	return result
}

type limiterMock struct {
	allow     func(key string) (bool, time.Duration)
	onFailure func(key string)
	onSuccess func(key string)
}

func (m *limiterMock) Allow(key string) (bool, time.Duration) {
	return m.allow(key)
}

func (m *limiterMock) OnFailure(key string) {
	if m.onFailure != nil {
		m.onFailure(key)
	}
}

func (m *limiterMock) OnSuccess(key string) {
	if m.onSuccess != nil {
		m.onSuccess(key)
	}
}

func TestAuthHandlerLoginSuccess(t *testing.T) {
	t.Parallel()

	env := newAuthTestEnv(t)
	env.auth.authenticate = func(
		ctx context.Context,
		sourceType string,
		creds domain.Credentials,
	) (*domain.AuthenticatedUser, error) {
		return &domain.AuthenticatedUser{
			User: domain.User{
				ID:          "user-id",
				Subject:     "local:user-id",
				Username:    "admin",
				DisplayName: "Administrator",
				Email:       "admin@example.local",
			},
			Source: domain.IdentitySource{
				ID:   1,
				Type: "local",
			},
			Roles: []string{"admin"},
		}, nil
	}
	env.tokens.issueForUser = func(
		ctx context.Context,
		user *domain.User,
		source string,
		roles []string,
		meta usecase.TokenMeta,
	) (*domain.IssuedTokens, error) {
		return &domain.IssuedTokens{
			AccessToken:      "access",
			RefreshToken:     "refresh",
			ExpiresIn:        900,
			RefreshExpiresIn: 1209600,
		}, nil
	}
	env.limiter.allow = func(key string) (bool, time.Duration) { return true, 0 }

	response := performJSONRequest(
		t,
		env.router,
		http.MethodPost,
		"/api/auth/login",
		map[string]any{
			"username": "admin",
			"password": "secret",
		},
		nil,
	)

	if response.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", response.Code)
	}
	if !strings.Contains(response.Header().Get("Set-Cookie"), "refresh_token=refresh") {
		t.Fatalf("refresh cookie not set: %s", response.Header().Get("Set-Cookie"))
	}
	if !strings.Contains(response.Header().Get("Set-Cookie"), "HttpOnly") {
		t.Fatalf("refresh cookie is not HttpOnly: %s", response.Header().Get("Set-Cookie"))
	}
	if !strings.Contains(response.Header().Get("Set-Cookie"), "SameSite=Strict") {
		t.Fatalf("refresh cookie must be SameSite=Strict: %s", response.Header().Get("Set-Cookie"))
	}

	var payload loginResponse
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response body: %v", err)
	}
	if payload.AccessToken != "access" {
		t.Fatalf("unexpected access token: %#v", payload.AccessToken)
	}
	if payload.TokenType != "Bearer" {
		t.Fatalf("unexpected token type: %s", payload.TokenType)
	}
	if payload.ExpiresIn != 900 {
		t.Fatalf("unexpected expires_in: %d", payload.ExpiresIn)
	}
	if payload.RefreshToken != "refresh" {
		t.Fatalf("unexpected refresh token in body: %s", payload.RefreshToken)
	}
	if payload.User.ID != "user-id" {
		t.Fatalf("unexpected user.id: %s", payload.User.ID)
	}
	if payload.User.Source != "local" {
		t.Fatalf("unexpected user.source: %s", payload.User.Source)
	}
	if len(payload.User.Roles) != 1 || payload.User.Roles[0] != "admin" {
		t.Fatalf("unexpected user.roles: %#v", payload.User.Roles)
	}

	events := env.audit.Events()
	if len(events) != 1 {
		t.Fatalf("expected one audit event, got %d", len(events))
	}
	if events[0].Action != domain.AuditLoginSuccess {
		t.Fatalf("unexpected audit action: %s", events[0].Action)
	}
	if events[0].Result != domain.AuditResultSuccess {
		t.Fatalf("unexpected audit result: %s", events[0].Result)
	}
	if events[0].ActorSubject != "local:user-id" {
		t.Fatalf("unexpected audit actor: %s", events[0].ActorSubject)
	}
}

func TestAuthHandlerLoginValidation(t *testing.T) {
	t.Parallel()

	env := newAuthTestEnv(t)
	env.limiter.allow = func(key string) (bool, time.Duration) { return true, 0 }
	response := performRequest(
		env.router,
		http.MethodPost,
		"/api/auth/login",
		strings.NewReader(""),
		nil,
	)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d", response.Code)
	}
	assertProblemType(t, response, "/errors/validation")
}

func TestAuthHandlerLoginInvalidCredentials(t *testing.T) {
	t.Parallel()

	env := newAuthTestEnv(t)
	var failedKey string
	env.limiter.allow = func(key string) (bool, time.Duration) { return true, 0 }
	env.limiter.onFailure = func(key string) {
		failedKey = key
	}
	env.auth.authenticate = func(
		ctx context.Context,
		sourceType string,
		creds domain.Credentials,
	) (*domain.AuthenticatedUser, error) {
		return nil, domain.ErrInvalidCredentials
	}

	response := performJSONRequest(
		t,
		env.router,
		http.MethodPost,
		"/api/auth/login",
		map[string]any{
			"username": "admin",
			"password": "wrong",
		},
		map[string]string{
			"X-Forwarded-For": "203.0.113.9",
		},
	)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected status: %d", response.Code)
	}
	assertProblemType(t, response, "/errors/auth/invalid-credentials")
	if failedKey != "203.0.113.9|admin" {
		t.Fatalf("unexpected limiter key: %s", failedKey)
	}

	events := env.audit.Events()
	if len(events) != 1 {
		t.Fatalf("expected one audit event, got %d", len(events))
	}
	if events[0].Action != domain.AuditLoginFailure {
		t.Fatalf("unexpected audit action: %s", events[0].Action)
	}
	if events[0].Reason != "invalid_credentials" {
		t.Fatalf("unexpected audit reason: %s", events[0].Reason)
	}
	if events[0].Result != domain.AuditResultFailure {
		t.Fatalf("unexpected audit result: %s", events[0].Result)
	}
}

func TestAuthHandlerLoginUserDisabled(t *testing.T) {
	t.Parallel()

	env := newAuthTestEnv(t)
	env.limiter.allow = func(key string) (bool, time.Duration) { return true, 0 }
	env.auth.authenticate = func(
		ctx context.Context,
		sourceType string,
		creds domain.Credentials,
	) (*domain.AuthenticatedUser, error) {
		return nil, domain.ErrUserDisabled
	}

	response := performJSONRequest(
		t,
		env.router,
		http.MethodPost,
		"/api/auth/login",
		map[string]any{
			"username": "admin",
			"password": "secret",
		},
		nil,
	)
	if response.Code != http.StatusForbidden {
		t.Fatalf("unexpected status: %d", response.Code)
	}
	assertProblemType(t, response, "/errors/auth/user-disabled")
}

func TestAuthHandlerLoginRateLimited(t *testing.T) {
	t.Parallel()

	env := newAuthTestEnv(t)
	env.limiter.allow = func(key string) (bool, time.Duration) {
		return false, 5 * time.Second
	}

	response := performJSONRequest(
		t,
		env.router,
		http.MethodPost,
		"/api/auth/login",
		map[string]any{
			"username": "admin",
			"password": "secret",
		},
		nil,
	)
	if response.Code != http.StatusTooManyRequests {
		t.Fatalf("unexpected status: %d", response.Code)
	}
	if response.Header().Get("Retry-After") == "" {
		t.Fatal("expected Retry-After header")
	}
	assertProblemType(t, response, "/errors/auth/rate-limited")

	events := env.audit.Events()
	if len(events) != 1 {
		t.Fatalf("expected one audit event, got %d", len(events))
	}
	if events[0].Action != domain.AuditRateLimited {
		t.Fatalf("unexpected audit action: %s", events[0].Action)
	}
	if events[0].Reason != "too_many_attempts" {
		t.Fatalf("unexpected audit reason: %s", events[0].Reason)
	}
}

func TestAuthHandlerRefreshSuccess(t *testing.T) {
	t.Parallel()

	env := newAuthTestEnv(t)
	env.tokens.rotate = func(
		ctx context.Context,
		refreshToken string,
		meta usecase.TokenMeta,
	) (*domain.IssuedTokens, *domain.User, error) {
		return &domain.IssuedTokens{
				AccessToken:      "new-access",
				RefreshToken:     "new-refresh",
				ExpiresIn:        900,
				RefreshExpiresIn: 1209600,
			}, &domain.User{
				ID:      "user-id",
				Subject: "local:user-id",
			}, nil
	}

	response := performJSONRequest(
		t,
		env.router,
		http.MethodPost,
		"/api/auth/refresh",
		map[string]any{
			"refresh_token": "token",
		},
		nil,
	)
	if response.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", response.Code)
	}
	if !strings.Contains(response.Header().Get("Set-Cookie"), "refresh_token=new-refresh") {
		t.Fatalf("refresh cookie not updated: %s", response.Header().Get("Set-Cookie"))
	}

	var payload refreshResponse
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response body: %v", err)
	}
	if payload.AccessToken != "new-access" {
		t.Fatalf("unexpected access token: %s", payload.AccessToken)
	}
	if payload.TokenType != "Bearer" {
		t.Fatalf("unexpected token type: %s", payload.TokenType)
	}
	if payload.RefreshToken != "new-refresh" {
		t.Fatalf("unexpected refresh token: %s", payload.RefreshToken)
	}
	events := env.audit.Events()
	if len(events) != 1 {
		t.Fatalf("expected one audit event, got %d", len(events))
	}
	if events[0].Action != domain.AuditTokenRefresh {
		t.Fatalf("unexpected audit action: %s", events[0].Action)
	}
	if events[0].Result != domain.AuditResultSuccess {
		t.Fatalf("unexpected audit result: %s", events[0].Result)
	}
}

func TestAuthHandlerRefreshReuseDetected(t *testing.T) {
	t.Parallel()

	env := newAuthTestEnv(t)
	env.tokens.rotate = func(
		ctx context.Context,
		refreshToken string,
		meta usecase.TokenMeta,
	) (*domain.IssuedTokens, *domain.User, error) {
		return nil, &domain.User{Subject: "local:user-id"}, domain.ErrRefreshReused
	}

	response := performJSONRequest(
		t,
		env.router,
		http.MethodPost,
		"/api/auth/refresh",
		map[string]any{
			"refresh_token": "token",
		},
		nil,
	)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected status: %d", response.Code)
	}
	assertProblemType(t, response, "/errors/auth/invalid-refresh")

	events := env.audit.Events()
	if len(events) != 1 {
		t.Fatalf("expected one audit event, got %d", len(events))
	}
	if events[0].Action != domain.AuditTokenReuseDetected {
		t.Fatalf("unexpected audit action: %s", events[0].Action)
	}
	if events[0].Reason != "reuse_detected" {
		t.Fatalf("unexpected audit reason: %s", events[0].Reason)
	}
	if events[0].ActorSubject != "local:user-id" {
		t.Fatalf("unexpected audit actor: %s", events[0].ActorSubject)
	}
}

func TestAuthHandlerLogoutUnknownToken(t *testing.T) {
	t.Parallel()

	env := newAuthTestEnv(t)
	env.tokens.revoke = func(ctx context.Context, refreshToken string, reason string) error {
		return nil
	}

	response := performJSONRequest(
		t,
		env.router,
		http.MethodPost,
		"/api/auth/logout",
		map[string]any{
			"refresh_token": "unknown",
		},
		nil,
	)
	if response.Code != http.StatusNoContent {
		t.Fatalf("unexpected status: %d", response.Code)
	}
	if !strings.Contains(response.Header().Get("Set-Cookie"), "Max-Age=0") {
		t.Fatalf("logout must clear refresh cookie: %s", response.Header().Get("Set-Cookie"))
	}
	events := env.audit.Events()
	if len(events) != 1 {
		t.Fatalf("expected one audit event, got %d", len(events))
	}
	if events[0].Action != domain.AuditLogout {
		t.Fatalf("unexpected audit action: %s", events[0].Action)
	}
	if events[0].Result != domain.AuditResultSuccess {
		t.Fatalf("unexpected audit result: %s", events[0].Result)
	}
}

func TestAuthHandlerUserinfoWithoutAuthorization(t *testing.T) {
	t.Parallel()

	env := newAuthTestEnv(t)
	response := performRequest(
		env.router,
		http.MethodGet,
		"/api/auth/userinfo",
		nil,
		nil,
	)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected status: %d", response.Code)
	}
	assertProblemType(t, response, "/errors/auth/invalid-token")
}

func TestAuthHandlerUserinfoSuccess(t *testing.T) {
	t.Parallel()

	env := newAuthTestEnv(t)
	env.users.findBySubject = func(ctx context.Context, subject string) (*domain.User, error) {
		return &domain.User{
			ID:          "user-id",
			Subject:     subject,
			Username:    "admin",
			DisplayName: "Administrator",
			Email:       "admin@example.local",
			SourceID:    1,
			LastLoginAt: timePointer(time.Now().UTC()),
		}, nil
	}
	env.users.getRoles = func(ctx context.Context, userID string) ([]string, error) {
		return []string{"admin"}, nil
	}
	env.sources.findByID = func(ctx context.Context, id int) (*domain.IdentitySource, error) {
		return &domain.IdentitySource{ID: id, Type: "local"}, nil
	}

	accessToken := createUserJWT(t, env.keystore, "controlitix-auth", "controlitix-api", "local:user-id")
	response := performRequest(
		env.router,
		http.MethodGet,
		"/api/auth/userinfo",
		nil,
		map[string]string{
			"Authorization": "Bearer " + accessToken,
		},
	)
	if response.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", response.Code, response.Body.String())
	}

	var payload userinfoResponse
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal userinfo response: %v", err)
	}
	if payload.Source != "local" {
		t.Fatalf("unexpected source: %#v", payload.Source)
	}
	if len(payload.Roles) != 1 || payload.Roles[0] != "admin" {
		t.Fatalf("unexpected roles: %#v", payload.Roles)
	}
}

type authTestEnv struct {
	router   http.Handler
	auth     *authServiceMock
	tokens   *tokenServiceMock
	users    *userRepositoryMock
	sources  *sourceRepositoryMock
	audit    *auditRecorderMock
	limiter  *limiterMock
	keystore *jwks.Keystore
}

func newAuthTestEnv(t *testing.T) *authTestEnv {
	t.Helper()

	keystore := testKeystore(t)
	auth := &authServiceMock{
		authenticate: func(
			ctx context.Context,
			sourceType string,
			creds domain.Credentials,
		) (*domain.AuthenticatedUser, error) {
			return nil, errors.New("unexpected auth authenticate call")
		},
	}
	tokens := &tokenServiceMock{
		issueForUser: func(
			ctx context.Context,
			user *domain.User,
			source string,
			roles []string,
			meta usecase.TokenMeta,
		) (*domain.IssuedTokens, error) {
			return nil, errors.New("unexpected token issue call")
		},
		rotate: func(
			ctx context.Context,
			refreshToken string,
			meta usecase.TokenMeta,
		) (*domain.IssuedTokens, *domain.User, error) {
			return nil, nil, errors.New("unexpected token rotate call")
		},
		revoke: func(ctx context.Context, refreshToken string, reason string) error {
			return errors.New("unexpected token revoke call")
		},
	}
	users := &userRepositoryMock{
		findBySubject: func(ctx context.Context, subject string) (*domain.User, error) {
			return nil, errors.New("unexpected user lookup call")
		},
		getRoles: func(ctx context.Context, userID string) ([]string, error) {
			return nil, errors.New("unexpected user roles call")
		},
	}
	sources := &sourceRepositoryMock{
		findByID: func(ctx context.Context, id int) (*domain.IdentitySource, error) {
			return nil, errors.New("unexpected source lookup call")
		},
	}
	audit := &auditRecorderMock{}
	limiter := &limiterMock{
		allow: func(key string) (bool, time.Duration) { return true, 0 },
	}

	handler := NewAuthHandler(
		auth,
		tokens,
		users,
		sources,
		audit,
		limiter,
		keystore,
		AuthHandlerConfig{
			RefreshCookieName:   "refresh_token",
			RefreshCookiePath:   "/api/auth",
			RefreshCookieSecure: false,
			RefreshCookieTTL:    time.Hour,
			DevMode:             true,
			JWTIssuer:           "controlitix-auth",
			JWTAudienceUser:     "controlitix-api",
		},
		nil,
	)

	rootRouter := chi.NewRouter()
	rootRouter.Route("/api/auth", func(authRouter chi.Router) {
		handler.Register(authRouter)
	})

	return &authTestEnv{
		router:   rootRouter,
		auth:     auth,
		tokens:   tokens,
		users:    users,
		sources:  sources,
		audit:    audit,
		limiter:  limiter,
		keystore: keystore,
	}
}

func performJSONRequest(
	t *testing.T,
	handler http.Handler,
	method string,
	path string,
	payload map[string]any,
	headers map[string]string,
) *httptest.ResponseRecorder {
	t.Helper()

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal request payload: %v", err)
	}

	request := httptest.NewRequest(method, path, bytes.NewReader(bodyBytes))
	request.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		request.Header.Set(key, value)
	}

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func performRequest(
	handler http.Handler,
	method string,
	path string,
	body io.Reader,
	headers map[string]string,
) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, body)
	for key, value := range headers {
		request.Header.Set(key, value)
	}

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
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

func createUserJWT(
	t *testing.T,
	keystore *jwks.Keystore,
	issuer string,
	audience string,
	subject string,
) string {
	t.Helper()

	now := time.Now().UTC()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss":          issuer,
		"aud":          []string{audience},
		"sub":          subject,
		"iat":          now.Unix(),
		"nbf":          now.Unix(),
		"exp":          now.Add(time.Hour).Unix(),
		"jti":          "jti",
		"src":          "local",
		"roles":        []string{"admin"},
		"scope":        "",
		"username":     "admin",
		"display_name": "Administrator",
	})
	token.Header["kid"] = keystore.KeyID()

	signedToken, err := token.SignedString(keystore.PrivateKey())
	if err != nil {
		t.Fatalf("sign test jwt: %v", err)
	}

	return signedToken
}

func timePointer(value time.Time) *time.Time {
	return &value
}

func assertProblemType(
	t *testing.T,
	response *httptest.ResponseRecorder,
	expectedType string,
) {
	t.Helper()

	var problem Problem
	if err := json.Unmarshal(response.Body.Bytes(), &problem); err != nil {
		t.Fatalf("unmarshal problem response: %v", err)
	}
	if problem.Type != expectedType {
		t.Fatalf("unexpected problem type: got %s want %s", problem.Type, expectedType)
	}
}
