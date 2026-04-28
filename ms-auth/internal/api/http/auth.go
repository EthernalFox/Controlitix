package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	stdhttp "net/http"
	"net/netip"
	"strconv"
	"strings"
	"time"

	"github.com/EthernalFox/Controlitix/ms-auth/internal/domain"
	"github.com/EthernalFox/Controlitix/ms-auth/internal/infrastructure/jwks"
	"github.com/EthernalFox/Controlitix/ms-auth/internal/usecase"
	"github.com/go-chi/chi/v5"
)

const (
	defaultRefreshCookieName   = "refresh_token"
	defaultRefreshCookiePath   = "/api/auth"
)

type authService interface {
	Authenticate(
		ctx context.Context,
		sourceType string,
		creds domain.Credentials,
	) (*domain.AuthenticatedUser, error)
}

type tokenService interface {
	IssueForUser(
		ctx context.Context,
		user *domain.User,
		source string,
		roles []string,
		meta usecase.TokenMeta,
	) (*domain.IssuedTokens, error)
	Rotate(
		ctx context.Context,
		refreshToken string,
		meta usecase.TokenMeta,
	) (*domain.IssuedTokens, *domain.User, error)
	Revoke(ctx context.Context, refreshToken string, reason string) error
}

type userQueryRepository interface {
	FindBySubject(ctx context.Context, subject string) (*domain.User, error)
	GetRoles(ctx context.Context, userID string) ([]string, error)
}

type sourceQueryRepository interface {
	FindByID(ctx context.Context, id int) (*domain.IdentitySource, error)
}

type auditRecorder interface {
	Record(ctx context.Context, event domain.AuditEvent) error
}

type rateLimiter interface {
	Allow(key string) (bool, time.Duration)
	OnFailure(key string)
	OnSuccess(key string)
}

type AuthHandler struct {
	auth    authService
	tokens  tokenService
	users   userQueryRepository
	sources sourceQueryRepository
	audit   auditRecorder
	limiter rateLimiter
	ks      *jwks.Keystore
	cfg     AuthHandlerConfig
	logger  *slog.Logger
}

type AuthHandlerConfig struct {
	RefreshCookieName   string
	RefreshCookiePath   string
	RefreshCookieSecure bool
	RefreshCookieTTL    time.Duration
	DevMode             bool
	JWTIssuer           string
	JWTAudienceUser     string
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type loginResponse struct {
	AccessToken  string           `json:"access_token"`
	TokenType    string           `json:"token_type"`
	ExpiresIn    int              `json:"expires_in"`
	RefreshToken string           `json:"refresh_token"`
	User         loginUserPayload `json:"user"`
}

type refreshResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

type loginUserPayload struct {
	ID          string   `json:"id"`
	Subject     string   `json:"subject"`
	Username    string   `json:"username"`
	DisplayName string   `json:"display_name"`
	Email       string   `json:"email"`
	Roles       []string `json:"roles"`
	Source      string   `json:"source"`
}

type userinfoResponse struct {
	ID          string     `json:"id"`
	Subject     string     `json:"subject"`
	Username    string     `json:"username"`
	DisplayName string     `json:"display_name"`
	Email       string     `json:"email"`
	Roles       []string   `json:"roles"`
	Source      string     `json:"source"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
}

func NewAuthHandler(
	auth authService,
	tokens tokenService,
	users userQueryRepository,
	sources sourceQueryRepository,
	audit auditRecorder,
	limiter rateLimiter,
	keystore *jwks.Keystore,
	cfg AuthHandlerConfig,
	logger *slog.Logger,
) *AuthHandler {
	if logger == nil {
		logger = slog.Default()
	}
	if strings.TrimSpace(cfg.RefreshCookieName) == "" {
		cfg.RefreshCookieName = defaultRefreshCookieName
	}
	if strings.TrimSpace(cfg.RefreshCookiePath) == "" {
		cfg.RefreshCookiePath = defaultRefreshCookiePath
	}
	if cfg.RefreshCookieTTL <= 0 {
		cfg.RefreshCookieTTL = 14 * 24 * time.Hour
	}
	if !cfg.DevMode {
		cfg.RefreshCookieSecure = true
	}
	if strings.TrimSpace(cfg.JWTIssuer) == "" {
		cfg.JWTIssuer = defaultJWTIssuer
	}
	if strings.TrimSpace(cfg.JWTAudienceUser) == "" {
		cfg.JWTAudienceUser = "controlitix-api"
	}

	return &AuthHandler{
		auth:    auth,
		tokens:  tokens,
		users:   users,
		sources: sources,
		audit:   audit,
		limiter: limiter,
		ks:      keystore,
		cfg:     cfg,
		logger:  logger,
	}
}

func (handler *AuthHandler) Register(router chi.Router) {
	router.Post("/login", handler.login)
	router.Post("/refresh", handler.refresh)
	router.Post("/logout", handler.logout)
	router.With(
		RequireAuthWithIssuer(
			handler.ks,
			handler.cfg.JWTIssuer,
			handler.cfg.JWTAudienceUser,
		),
	).Get("/userinfo", handler.userinfo)
}

func (handler *AuthHandler) login(
	responseWriter stdhttp.ResponseWriter,
	request *stdhttp.Request,
) {
	var payload loginRequest
	if err := decodeJSONBody(request.Body, &payload); err != nil {
		WriteProblem(responseWriter, Problem{
			Type:   "/errors/validation",
			Title:  "Validation failed",
			Status: stdhttp.StatusBadRequest,
		})
		return
	}

	payload.Username = strings.TrimSpace(payload.Username)
	if payload.Username == "" || strings.TrimSpace(payload.Password) == "" {
		WriteProblem(responseWriter, Problem{
			Type:   "/errors/validation",
			Title:  "Validation failed",
			Status: stdhttp.StatusBadRequest,
		})
		return
	}

	clientIP := extractClientIP(request)
	rateLimitKey := fmt.Sprintf("%s|%s", clientIP, strings.ToLower(payload.Username))
	allowed, retryAfter := handler.limiter.Allow(rateLimitKey)
	if !allowed {
		retryAfterSeconds := int(retryAfter.Seconds())
		if retryAfterSeconds <= 0 {
			retryAfterSeconds = 1
		}

		responseWriter.Header().Set("Retry-After", strconv.Itoa(retryAfterSeconds))
		WriteProblem(responseWriter, Problem{
			Type:   "/errors/auth/rate-limited",
			Title:  "Rate limited",
			Status: stdhttp.StatusTooManyRequests,
		})

		handler.recordAudit(request.Context(), domain.AuditEvent{
			Action:    domain.AuditRateLimited,
			Target:    payload.Username,
			Result:    domain.AuditResultFailure,
			Reason:    "too_many_attempts",
			IP:        clientIP,
			UserAgent: request.UserAgent(),
		})
		return
	}

	authenticatedUser, authError := handler.auth.Authenticate(
		request.Context(),
		"local",
		domain.Credentials{
			Username: payload.Username,
			Password: payload.Password,
		},
	)
	if authError != nil {
		handler.limiter.OnFailure(rateLimitKey)

		problem, reason := loginErrorToProblem(authError)
		WriteProblem(responseWriter, problem)
		handler.recordAudit(request.Context(), domain.AuditEvent{
			Action:    domain.AuditLoginFailure,
			Target:    payload.Username,
			Result:    domain.AuditResultFailure,
			Reason:    reason,
			IP:        clientIP,
			UserAgent: request.UserAgent(),
		})
		return
	}

	issuedTokens, issueError := handler.tokens.IssueForUser(
		request.Context(),
		&authenticatedUser.User,
		authenticatedUser.Source.Type,
		authenticatedUser.Roles,
		usecase.TokenMeta{
			UserAgent: request.UserAgent(),
			IP:        clientIP,
		},
	)
	if issueError != nil {
		WriteProblem(responseWriter, Problem{
			Type:   "/errors/internal",
			Title:  "Internal error",
			Status: stdhttp.StatusInternalServerError,
		})
		return
	}

	handler.limiter.OnSuccess(rateLimitKey)
	handler.setRefreshCookie(responseWriter, issuedTokens.RefreshToken)

	writeJSON(responseWriter, stdhttp.StatusOK, loginResponse{
		AccessToken:  issuedTokens.AccessToken,
		TokenType:    "Bearer",
		ExpiresIn:    issuedTokens.ExpiresIn,
		RefreshToken: issuedTokens.RefreshToken,
		User: loginUserPayload{
			ID:          authenticatedUser.User.ID,
			Subject:     authenticatedUser.User.Subject,
			Username:    authenticatedUser.User.Username,
			DisplayName: authenticatedUser.User.DisplayName,
			Email:       authenticatedUser.User.Email,
			Roles:       authenticatedUser.Roles,
			Source:      authenticatedUser.Source.Type,
		},
	})

	handler.recordAudit(request.Context(), domain.AuditEvent{
		ActorSubject: authenticatedUser.User.Subject,
		Action:       domain.AuditLoginSuccess,
		Target:       authenticatedUser.User.ID,
		Result:       domain.AuditResultSuccess,
		IP:           clientIP,
		UserAgent:    request.UserAgent(),
	})
}

func (handler *AuthHandler) refresh(
	responseWriter stdhttp.ResponseWriter,
	request *stdhttp.Request,
) {
	refreshToken, ok := handler.readRefreshToken(request)
	if !ok {
		WriteProblem(responseWriter, Problem{
			Type:   "/errors/validation",
			Title:  "Validation failed",
			Status: stdhttp.StatusBadRequest,
		})
		return
	}

	clientIP := extractClientIP(request)
	issuedTokens, user, rotateError := handler.tokens.Rotate(
		request.Context(),
		refreshToken,
		usecase.TokenMeta{
			UserAgent: request.UserAgent(),
			IP:        clientIP,
		},
	)
	if rotateError != nil {
		switch {
		case errors.Is(rotateError, domain.ErrRefreshReused):
			WriteProblem(responseWriter, Problem{
				Type:   "/errors/auth/invalid-refresh",
				Title:  "Invalid refresh token",
				Status: stdhttp.StatusUnauthorized,
			})
			handler.recordAudit(request.Context(), domain.AuditEvent{
				ActorSubject: safeUserSubject(user),
				Action:       domain.AuditTokenReuseDetected,
				Result:       domain.AuditResultFailure,
				Reason:       "reuse_detected",
				IP:           clientIP,
				UserAgent:    request.UserAgent(),
			})
			return
		case errors.Is(rotateError, domain.ErrInvalidRefresh):
			WriteProblem(responseWriter, Problem{
				Type:   "/errors/auth/invalid-refresh",
				Title:  "Invalid refresh token",
				Status: stdhttp.StatusUnauthorized,
			})
			handler.recordAudit(request.Context(), domain.AuditEvent{
				Action:    domain.AuditTokenRefresh,
				Result:    domain.AuditResultFailure,
				Reason:    "invalid_refresh",
				IP:        clientIP,
				UserAgent: request.UserAgent(),
			})
			return
		case errors.Is(rotateError, domain.ErrUserDisabled):
			WriteProblem(responseWriter, Problem{
				Type:   "/errors/auth/user-disabled",
				Title:  "User disabled",
				Status: stdhttp.StatusForbidden,
			})
			handler.recordAudit(request.Context(), domain.AuditEvent{
				ActorSubject: safeUserSubject(user),
				Action:       domain.AuditTokenRefresh,
				Result:       domain.AuditResultFailure,
				Reason:       "user_disabled",
				IP:           clientIP,
				UserAgent:    request.UserAgent(),
			})
			return
		default:
			WriteProblem(responseWriter, Problem{
				Type:   "/errors/internal",
				Title:  "Internal error",
				Status: stdhttp.StatusInternalServerError,
			})
			return
		}
	}

	handler.setRefreshCookie(responseWriter, issuedTokens.RefreshToken)
	writeJSON(responseWriter, stdhttp.StatusOK, refreshResponse{
		AccessToken:  issuedTokens.AccessToken,
		TokenType:    "Bearer",
		ExpiresIn:    issuedTokens.ExpiresIn,
		RefreshToken: issuedTokens.RefreshToken,
	})

	handler.recordAudit(request.Context(), domain.AuditEvent{
		ActorSubject: safeUserSubject(user),
		Action:       domain.AuditTokenRefresh,
		Result:       domain.AuditResultSuccess,
		IP:           clientIP,
		UserAgent:    request.UserAgent(),
	})
}

func (handler *AuthHandler) logout(
	responseWriter stdhttp.ResponseWriter,
	request *stdhttp.Request,
) {
	refreshToken, hasRefreshToken := handler.readRefreshToken(request)
	if hasRefreshToken {
		if err := handler.tokens.Revoke(request.Context(), refreshToken, "logout"); err != nil {
			handler.logger.Warn(
				"failed to revoke refresh token",
				"method",
				"AuthHandler.logout",
				"error",
				err,
			)
		}
	}

	handler.clearRefreshCookie(responseWriter)

	actor := ""
	authorization := strings.TrimSpace(request.Header.Get("Authorization"))
	if strings.HasPrefix(strings.ToLower(authorization), "bearer ") {
		if principal, err := ParsePrincipalToken(
			strings.TrimSpace(authorization[len("Bearer "):]),
			handler.ks,
			handler.cfg.JWTIssuer,
			handler.cfg.JWTAudienceUser,
		); err == nil {
			actor = principal.Subject
		}
	}

	handler.recordAudit(request.Context(), domain.AuditEvent{
		ActorSubject: actor,
		Action:       domain.AuditLogout,
		Result:       domain.AuditResultSuccess,
		Reason:       "logout",
		IP:           extractClientIP(request),
		UserAgent:    request.UserAgent(),
	})

	responseWriter.WriteHeader(stdhttp.StatusNoContent)
}

func (handler *AuthHandler) userinfo(
	responseWriter stdhttp.ResponseWriter,
	request *stdhttp.Request,
) {
	principal, ok := PrincipalFromContext(request.Context())
	if !ok {
		WriteProblem(responseWriter, Problem{
			Type:   "/errors/auth/invalid-token",
			Title:  "Invalid token",
			Status: stdhttp.StatusUnauthorized,
		})
		return
	}

	user, err := handler.users.FindBySubject(request.Context(), principal.Subject)
	if err != nil {
		WriteProblem(responseWriter, Problem{
			Type:   "/errors/internal",
			Title:  "Internal error",
			Status: stdhttp.StatusInternalServerError,
		})
		return
	}
	if user == nil {
		WriteProblem(responseWriter, Problem{
			Type:   "/errors/auth/invalid-token",
			Title:  "Invalid token",
			Status: stdhttp.StatusUnauthorized,
		})
		return
	}

	roles, err := handler.users.GetRoles(request.Context(), user.ID)
	if err != nil {
		WriteProblem(responseWriter, Problem{
			Type:   "/errors/internal",
			Title:  "Internal error",
			Status: stdhttp.StatusInternalServerError,
		})
		return
	}

	source, err := handler.sources.FindByID(request.Context(), user.SourceID)
	if err != nil {
		WriteProblem(responseWriter, Problem{
			Type:   "/errors/internal",
			Title:  "Internal error",
			Status: stdhttp.StatusInternalServerError,
		})
		return
	}
	if source == nil {
		WriteProblem(responseWriter, Problem{
			Type:   "/errors/auth/invalid-token",
			Title:  "Invalid token",
			Status: stdhttp.StatusUnauthorized,
		})
		return
	}

	writeJSON(responseWriter, stdhttp.StatusOK, userinfoResponse{
		ID:          user.ID,
		Subject:     user.Subject,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		Email:       user.Email,
		Roles:       roles,
		Source:      source.Type,
		LastLoginAt: user.LastLoginAt,
	})
}

func decodeJSONBody(body io.ReadCloser, target any) error {
	defer body.Close()

	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}

	return nil
}

func loginErrorToProblem(err error) (Problem, string) {
	switch {
	case errors.Is(err, domain.ErrInvalidCredentials):
		return Problem{
			Type:   "/errors/auth/invalid-credentials",
			Title:  "Invalid credentials",
			Status: stdhttp.StatusUnauthorized,
		}, "invalid_credentials"
	case errors.Is(err, domain.ErrUserDisabled):
		return Problem{
			Type:   "/errors/auth/user-disabled",
			Title:  "User disabled",
			Status: stdhttp.StatusForbidden,
		}, "user_disabled"
	case errors.Is(err, domain.ErrSourceDisabled):
		return Problem{
			Type:   "/errors/auth/source-disabled",
			Title:  "Identity source disabled",
			Status: stdhttp.StatusServiceUnavailable,
		}, "source_disabled"
	default:
		return Problem{
			Type:   "/errors/internal",
			Title:  "Internal error",
			Status: stdhttp.StatusInternalServerError,
		}, "internal_error"
	}
}

func (handler *AuthHandler) readRefreshToken(request *stdhttp.Request) (string, bool) {
	var payload refreshRequest
	if err := decodeJSONBody(request.Body, &payload); err == nil && strings.TrimSpace(payload.RefreshToken) != "" {
		return strings.TrimSpace(payload.RefreshToken), true
	}

	cookie, err := request.Cookie(handler.cfg.RefreshCookieName)
	if err == nil && strings.TrimSpace(cookie.Value) != "" {
		return strings.TrimSpace(cookie.Value), true
	}

	return "", false
}

func (handler *AuthHandler) setRefreshCookie(
	responseWriter stdhttp.ResponseWriter,
	token string,
) {
	stdhttp.SetCookie(responseWriter, &stdhttp.Cookie{
		Name:     handler.cfg.RefreshCookieName,
		Value:    token,
		Path:     handler.cfg.RefreshCookiePath,
		MaxAge:   int(handler.cfg.RefreshCookieTTL.Seconds()),
		HttpOnly: true,
		Secure:   handler.cfg.RefreshCookieSecure,
		SameSite: stdhttp.SameSiteStrictMode,
	})
}

func (handler *AuthHandler) clearRefreshCookie(responseWriter stdhttp.ResponseWriter) {
	stdhttp.SetCookie(responseWriter, &stdhttp.Cookie{
		Name:     handler.cfg.RefreshCookieName,
		Value:    "",
		Path:     handler.cfg.RefreshCookiePath,
		MaxAge:   0,
		HttpOnly: true,
		Secure:   handler.cfg.RefreshCookieSecure,
		SameSite: stdhttp.SameSiteStrictMode,
	})
}

func extractClientIP(request *stdhttp.Request) string {
	forwardedFor := strings.TrimSpace(request.Header.Get("X-Forwarded-For"))
	if forwardedFor != "" {
		parts := strings.Split(forwardedFor, ",")
		if len(parts) > 0 {
			candidate := strings.TrimSpace(parts[0])
			if parsed, err := netip.ParseAddr(candidate); err == nil {
				return parsed.String()
			}
			return candidate
		}
	}

	hostPort := strings.TrimSpace(request.RemoteAddr)
	if hostPort == "" {
		return ""
	}
	if parsed, err := netip.ParseAddrPort(hostPort); err == nil {
		return parsed.Addr().String()
	}
	if parsed, err := netip.ParseAddr(hostPort); err == nil {
		return parsed.String()
	}

	return hostPort
}

func safeUserSubject(user *domain.User) string {
	if user == nil {
		return ""
	}

	return user.Subject
}

func (handler *AuthHandler) recordAudit(ctx context.Context, event domain.AuditEvent) {
	if handler.audit == nil {
		return
	}
	if err := handler.audit.Record(ctx, event); err != nil {
		handler.logger.Warn(
			"failed to record audit event",
			"method",
			"AuthHandler.recordAudit",
			"action",
			event.Action,
			"result",
			event.Result,
			"error",
			err,
		)
	}
}
