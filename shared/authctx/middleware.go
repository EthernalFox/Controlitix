package authctx

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
)

type Problem struct {
	Type     string         `json:"type"`
	Title    string         `json:"title"`
	Status   int            `json:"status"`
	Detail   string         `json:"detail,omitempty"`
	Instance string         `json:"instance,omitempty"`
	Extra    map[string]any `json:"-"`
}

func RequireAuth(validator *Validator) func(http.Handler) http.Handler {
	if validator != nil && validator.disabled {
		logger := validator.logger
		if logger == nil {
			logger = slog.Default()
		}
		logger.Warn("authentication disabled, allowing all requests")
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(
			responseWriter http.ResponseWriter,
			request *http.Request,
		) {
			if validator == nil {
				writeAuthError(responseWriter, ErrInvalidSignature)
				return
			}

			if validator.disabled {
				principal, err := validator.Parse(request.Context(), "")
				if err != nil || principal == nil {
					writeAuthError(responseWriter, ErrInvalidSignature)
					return
				}
				requestContext := WithPrincipal(request.Context(), *principal)
				next.ServeHTTP(responseWriter, request.WithContext(requestContext))
				return
			}

			authorizationHeader := strings.TrimSpace(request.Header.Get("Authorization"))
			parts := strings.Fields(authorizationHeader)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				writeAuthError(responseWriter, ErrMissingAuth)
				return
			}

			principal, err := validator.Parse(request.Context(), strings.TrimSpace(parts[1]))
			if err != nil {
				writeAuthError(responseWriter, err)
				return
			}

			requestContext := WithPrincipal(request.Context(), *principal)
			next.ServeHTTP(responseWriter, request.WithContext(requestContext))
		})
	}
}

func RequireRole(roles ...string) func(http.Handler) http.Handler {
	requiredRoles := normalizeStringList(roles)
	if len(requiredRoles) == 0 {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			principal, ok := FromContext(request.Context())
			if !ok {
				writeForbidden(responseWriter)
				return
			}

			for _, role := range requiredRoles {
				if principal.HasRole(role) {
					next.ServeHTTP(responseWriter, request)
					return
				}
			}

			writeForbidden(responseWriter)
		})
	}
}

func RequireScope(scopes ...string) func(http.Handler) http.Handler {
	requiredScopes := normalizeStringList(scopes)
	if len(requiredScopes) == 0 {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			principal, ok := FromContext(request.Context())
			if !ok {
				writeForbidden(responseWriter)
				return
			}

			for _, scope := range requiredScopes {
				if !principal.HasScope(scope) {
					writeForbidden(responseWriter)
					return
				}
			}

			next.ServeHTTP(responseWriter, request)
		})
	}
}

func RequireAudience(audiences ...string) func(http.Handler) http.Handler {
	requiredAudiences := normalizeStringList(audiences)
	if len(requiredAudiences) == 0 {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			principal, ok := FromContext(request.Context())
			if !ok {
				writeForbidden(responseWriter)
				return
			}

			if !hasAny(normalizeStringList(principal.Audience), requiredAudiences) {
				writeForbidden(responseWriter)
				return
			}

			next.ServeHTTP(responseWriter, request)
		})
	}
}

func WriteProblem(responseWriter http.ResponseWriter, problem Problem) {
	if problem.Status == 0 {
		problem.Status = http.StatusInternalServerError
	}
	if problem.Type == "" {
		problem.Type = "/errors/internal"
	}
	if problem.Title == "" {
		problem.Title = "Internal error"
	}

	payload := map[string]any{
		"type":   problem.Type,
		"title":  problem.Title,
		"status": problem.Status,
	}
	if problem.Detail != "" {
		payload["detail"] = problem.Detail
	}
	if problem.Instance != "" {
		payload["instance"] = problem.Instance
	}
	for key, value := range problem.Extra {
		payload[key] = value
	}

	responseWriter.Header().Set("Content-Type", "application/problem+json")
	responseWriter.WriteHeader(problem.Status)

	encoder := json.NewEncoder(responseWriter)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(payload)
}

func writeForbidden(responseWriter http.ResponseWriter) {
	WriteProblem(responseWriter, Problem{
		Type:   "/errors/auth/forbidden",
		Title:  "Forbidden",
		Status: http.StatusForbidden,
	})
}

func writeAuthError(responseWriter http.ResponseWriter, authError error) {
	switch {
	case errors.Is(authError, ErrTokenExpired):
		WriteProblem(responseWriter, Problem{
			Type:   "/errors/auth/token-expired",
			Title:  "Token expired",
			Status: http.StatusUnauthorized,
		})
	case errors.Is(authError, ErrTokenNotYetValid):
		WriteProblem(responseWriter, Problem{
			Type:   "/errors/auth/token-not-yet-valid",
			Title:  "Token not yet valid",
			Status: http.StatusUnauthorized,
		})
	case errors.Is(authError, ErrInvalidIssuer):
		WriteProblem(responseWriter, Problem{
			Type:   "/errors/auth/invalid-issuer",
			Title:  "Invalid issuer",
			Status: http.StatusUnauthorized,
		})
	case errors.Is(authError, ErrInvalidAudience):
		WriteProblem(responseWriter, Problem{
			Type:   "/errors/auth/invalid-audience",
			Title:  "Invalid audience",
			Status: http.StatusUnauthorized,
		})
	case errors.Is(authError, ErrUnknownKey):
		WriteProblem(responseWriter, Problem{
			Type:   "/errors/auth/unknown-key",
			Title:  "Unknown key",
			Status: http.StatusUnauthorized,
		})
	default:
		WriteProblem(responseWriter, Problem{
			Type:   "/errors/auth/invalid-token",
			Title:  "Invalid token",
			Status: http.StatusUnauthorized,
		})
	}
}
