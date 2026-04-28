package http

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	stdhttp "net/http"
	"strings"
	"time"

	"github.com/EthernalFox/Controlitix/ms-auth/internal/infrastructure/jwks"
	"github.com/golang-jwt/jwt/v5"
)

const (
	requestIDHeaderName = "X-Request-ID"
	defaultJWTIssuer    = "controlitix-auth"
)

type contextKey string

const (
	requestIDContextKey contextKey = "request_id"
	principalContextKey contextKey = "principal"
)

type Principal struct {
	Subject     string
	Roles       []string
	Scopes      []string
	Source      string
	Username    string
	DisplayName string
	ClientID    string
	TokenID     string
}

func ChainMiddleware(
	handler stdhttp.Handler,
	middlewares ...func(stdhttp.Handler) stdhttp.Handler,
) stdhttp.Handler {
	if len(middlewares) == 0 {
		return handler
	}

	wrapped := handler
	for index := len(middlewares) - 1; index >= 0; index-- {
		wrapped = middlewares[index](wrapped)
	}

	return wrapped
}

func RequestID(next stdhttp.Handler) stdhttp.Handler {
	return stdhttp.HandlerFunc(func(responseWriter stdhttp.ResponseWriter, request *stdhttp.Request) {
		requestID := strings.TrimSpace(request.Header.Get(requestIDHeaderName))
		if requestID == "" {
			requestID = generateRequestID()
		}

		responseWriter.Header().Set(requestIDHeaderName, requestID)
		requestContext := context.WithValue(request.Context(), requestIDContextKey, requestID)
		next.ServeHTTP(responseWriter, request.WithContext(requestContext))
	})
}

func AccessLog(logger *slog.Logger) func(stdhttp.Handler) stdhttp.Handler {
	if logger == nil {
		logger = slog.Default()
	}

	return func(next stdhttp.Handler) stdhttp.Handler {
		return stdhttp.HandlerFunc(func(responseWriter stdhttp.ResponseWriter, request *stdhttp.Request) {
			startedAt := time.Now()
			wrappedWriter := &statusWriter{
				ResponseWriter: responseWriter,
				statusCode:     stdhttp.StatusOK,
			}

			next.ServeHTTP(wrappedWriter, request)

			logger.Info(
				"http request handled",
				"method",
				request.Method,
				"path",
				request.URL.Path,
				"status",
				wrappedWriter.statusCode,
				"duration_ms",
				time.Since(startedAt).Milliseconds(),
				"request_id",
				RequestIDFromContext(request.Context()),
			)
		})
	}
}

func RequireAuth(
	keystore *jwks.Keystore,
	audiences ...string,
) func(stdhttp.Handler) stdhttp.Handler {
	return RequireAuthWithIssuer(keystore, defaultJWTIssuer, audiences...)
}

func RequireAuthWithIssuer(
	keystore *jwks.Keystore,
	issuer string,
	audiences ...string,
) func(stdhttp.Handler) stdhttp.Handler {
	issuer = strings.TrimSpace(issuer)
	if issuer == "" {
		issuer = defaultJWTIssuer
	}

	normalizedAudiences := make([]string, 0, len(audiences))
	for _, audience := range audiences {
		audience = strings.TrimSpace(audience)
		if audience == "" {
			continue
		}
		normalizedAudiences = append(normalizedAudiences, audience)
	}

	return func(next stdhttp.Handler) stdhttp.Handler {
		return stdhttp.HandlerFunc(func(responseWriter stdhttp.ResponseWriter, request *stdhttp.Request) {
			authorizationHeader := strings.TrimSpace(request.Header.Get("Authorization"))
			if !strings.HasPrefix(strings.ToLower(authorizationHeader), "bearer ") {
				WriteProblem(responseWriter, Problem{
					Type:   "/errors/auth/invalid-token",
					Title:  "Invalid token",
					Status: stdhttp.StatusUnauthorized,
				})
				return
			}

			tokenString := strings.TrimSpace(authorizationHeader[len("Bearer "):])
			principal, parseError := parsePrincipal(
				tokenString,
				keystore,
				issuer,
				normalizedAudiences,
			)
			if parseError != nil {
				switch {
				case errors.Is(parseError, jwt.ErrTokenExpired):
					WriteProblem(responseWriter, Problem{
						Type:   "/errors/auth/token-expired",
						Title:  "Token expired",
						Status: stdhttp.StatusUnauthorized,
					})
				default:
					WriteProblem(responseWriter, Problem{
						Type:   "/errors/auth/invalid-token",
						Title:  "Invalid token",
						Status: stdhttp.StatusUnauthorized,
					})
				}
				return
			}

			requestContext := context.WithValue(request.Context(), principalContextKey, principal)
			next.ServeHTTP(responseWriter, request.WithContext(requestContext))
		})
	}
}

func RequireRole(requiredRoles ...string) func(stdhttp.Handler) stdhttp.Handler {
	normalizedRequiredRoles := normalizeStringSlice(requiredRoles)

	return func(next stdhttp.Handler) stdhttp.Handler {
		return stdhttp.HandlerFunc(func(responseWriter stdhttp.ResponseWriter, request *stdhttp.Request) {
			principal, ok := PrincipalFromContext(request.Context())
			if !ok {
				WriteProblem(responseWriter, Problem{
					Type:   "/errors/auth/forbidden",
					Title:  "Forbidden",
					Status: stdhttp.StatusForbidden,
				})
				return
			}

			if len(normalizedRequiredRoles) > 0 && !hasAnyAudience(principal.Roles, normalizedRequiredRoles) {
				WriteProblem(responseWriter, Problem{
					Type:   "/errors/auth/forbidden",
					Title:  "Forbidden",
					Status: stdhttp.StatusForbidden,
				})
				return
			}

			next.ServeHTTP(responseWriter, request)
		})
	}
}

func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	principal, ok := ctx.Value(principalContextKey).(Principal)
	return principal, ok
}

func RequestIDFromContext(ctx context.Context) string {
	requestID, _ := ctx.Value(requestIDContextKey).(string)
	return requestID
}

func ParsePrincipalToken(
	tokenString string,
	keystore *jwks.Keystore,
	issuer string,
	audiences ...string,
) (Principal, error) {
	return parsePrincipal(tokenString, keystore, issuer, audiences)
}

type statusWriter struct {
	stdhttp.ResponseWriter
	statusCode int
}

func (writer *statusWriter) WriteHeader(statusCode int) {
	writer.statusCode = statusCode
	writer.ResponseWriter.WriteHeader(statusCode)
}

func parsePrincipal(
	tokenString string,
	keystore *jwks.Keystore,
	issuer string,
	audiences []string,
) (Principal, error) {
	if keystore == nil || keystore.PrivateKey() == nil {
		return Principal{}, errors.New("jwt keystore is not initialized")
	}
	if strings.TrimSpace(tokenString) == "" {
		return Principal{}, errors.New("empty bearer token")
	}

	token, err := jwt.Parse(tokenString, func(parsedToken *jwt.Token) (any, error) {
		if parsedToken.Method != jwt.SigningMethodRS256 {
			return nil, fmt.Errorf("unexpected jwt signing method: %s", parsedToken.Method.Alg())
		}

		return &keystore.PrivateKey().PublicKey, nil
	}, jwt.WithIssuer(issuer))
	if err != nil {
		return Principal{}, err
	}
	if !token.Valid {
		return Principal{}, errors.New("jwt token is invalid")
	}

	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return Principal{}, errors.New("unsupported jwt claims payload")
	}

	claimAudiences, err := mapClaims.GetAudience()
	if err != nil {
		return Principal{}, fmt.Errorf("read jwt audience claim: %w", err)
	}
	if len(audiences) > 0 && !hasAnyAudience(claimAudiences, audiences) {
		return Principal{}, errors.New("jwt audience mismatch")
	}

	subject, err := mapClaims.GetSubject()
	if err != nil {
		return Principal{}, fmt.Errorf("read jwt subject claim: %w", err)
	}

	roles, _ := stringSliceClaim(mapClaims["roles"])
	scope := stringClaim(mapClaims["scope"])
	scopes := []string{}
	if strings.TrimSpace(scope) != "" {
		scopes = strings.Fields(scope)
	}

	return Principal{
		Subject:     subject,
		Roles:       roles,
		Scopes:      scopes,
		Source:      stringClaim(mapClaims["src"]),
		Username:    stringClaim(mapClaims["username"]),
		DisplayName: stringClaim(mapClaims["display_name"]),
		ClientID:    stringClaim(mapClaims["client_id"]),
		TokenID:     stringClaim(mapClaims["jti"]),
	}, nil
}

func hasAnyAudience(actual []string, expected []string) bool {
	expectedSet := make(map[string]struct{}, len(expected))
	for _, value := range expected {
		expectedSet[value] = struct{}{}
	}

	for _, audience := range actual {
		if _, ok := expectedSet[audience]; ok {
			return true
		}
	}

	return false
}

func stringClaim(raw any) string {
	value, _ := raw.(string)
	return strings.TrimSpace(value)
}

func stringSliceClaim(raw any) ([]string, bool) {
	switch typed := raw.(type) {
	case []string:
		return normalizeStringSlice(typed), true
	case []any:
		values := make([]string, 0, len(typed))
		for _, item := range typed {
			text, ok := item.(string)
			if !ok {
				continue
			}
			values = append(values, text)
		}
		return normalizeStringSlice(values), true
	default:
		return nil, false
	}
}

func normalizeStringSlice(input []string) []string {
	normalized := make([]string, 0, len(input))
	seen := make(map[string]struct{}, len(input))
	for _, value := range input {
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

func generateRequestID() string {
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return fmt.Sprintf("fallback-%d", time.Now().UnixNano())
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
	)
}
