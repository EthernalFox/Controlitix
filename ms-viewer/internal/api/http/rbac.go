package http

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
	"github.com/EthernalFox/Controlitix/shared/authctx"
)

type AuditRecorder interface {
	Record(ctx context.Context, event domain.AuditEvent)
}

func RequireRole(auditRecorder AuditRecorder, role string) func(http.Handler) http.Handler {
	return RequireAnyRole(auditRecorder, role)
}

func RequireAnyRole(auditRecorder AuditRecorder, roles ...string) func(http.Handler) http.Handler {
	normalizedRoles := normalizeRoles(roles)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			principal, ok := authctx.FromContext(request.Context())
			if !ok {
				writeProblem(responseWriter, Problem{
					Type:   "/errors/auth/invalid-token",
					Title:  "Invalid token",
					Status: http.StatusUnauthorized,
				})
				return
			}

			if hasAnyRole(principal.Roles, normalizedRoles) || hasRole(principal.Roles, "admin") {
				next.ServeHTTP(responseWriter, request)
				return
			}

			if auditRecorder != nil {
				auditRecorder.Record(request.Context(), domain.AuditEvent{
					Action: "auth.forbidden",
					Target: domain.AuditTarget{
						Type: "http_endpoint",
						ID:   strings.TrimSpace(request.URL.Path),
					},
					Details: map[string]any{
						"method":         request.Method,
						"required_roles": normalizedRoles,
					},
					Result: domain.AuditResultFailure,
				})
			}

			writeProblem(responseWriter, Problem{
				Type:   "/errors/auth/forbidden",
				Title:  "Forbidden",
				Status: http.StatusForbidden,
				Detail: fmt.Sprintf("required any of: %v", normalizedRoles),
			})
		})
	}
}

func normalizeRoles(input []string) []string {
	seen := make(map[string]struct{}, len(input))
	roles := make([]string, 0, len(input))
	for _, role := range input {
		normalized := strings.ToLower(strings.TrimSpace(role))
		if normalized == "" {
			continue
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		roles = append(roles, normalized)
	}
	return roles
}

func hasAnyRole(actual []string, expected []string) bool {
	if len(expected) == 0 {
		return true
	}
	for _, candidate := range expected {
		if hasRole(actual, candidate) {
			return true
		}
	}
	return false
}

func hasRole(actual []string, expected string) bool {
	normalizedExpected := strings.ToLower(strings.TrimSpace(expected))
	if normalizedExpected == "" {
		return false
	}
	for _, role := range actual {
		if strings.ToLower(strings.TrimSpace(role)) == normalizedExpected {
			return true
		}
	}
	return false
}
