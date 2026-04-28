package authctx

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRequireAuthWithoutHeader(t *testing.T) {
	t.Parallel()

	validator := NewValidator(ValidatorOptions{
		Logger: testLogger(),
	})

	handler := RequireAuth(validator)(http.HandlerFunc(func(
		responseWriter http.ResponseWriter,
		request *http.Request,
	) {
		responseWriter.WriteHeader(http.StatusNoContent)
	}))

	request := httptest.NewRequest(http.MethodGet, "/objects", nil)
	responseRecorder := httptest.NewRecorder()
	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", responseRecorder.Code)
	}
	if responseRecorder.Header().Get("Content-Type") != "application/problem+json" {
		t.Fatalf("unexpected content type: %s", responseRecorder.Header().Get("Content-Type"))
	}

	var responseBody map[string]any
	if err := json.Unmarshal(responseRecorder.Body.Bytes(), &responseBody); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if responseBody["type"] != "/errors/auth/invalid-token" {
		t.Fatalf("unexpected problem type: %v", responseBody["type"])
	}
}

func TestRequireAuthValidBearerAddsPrincipalToContext(t *testing.T) {
	t.Parallel()

	validator := NewValidator(ValidatorOptions{
		Disabled: true,
		Logger:   testLogger(),
		Clock: func() time.Time {
			return time.Unix(1_700_000_000, 0)
		},
	})

	handler := RequireAuth(validator)(http.HandlerFunc(func(
		responseWriter http.ResponseWriter,
		request *http.Request,
	) {
		principal, ok := FromContext(request.Context())
		if !ok {
			t.Fatal("principal is not set in context")
		}
		if principal.Subject != "dev:anonymous" {
			t.Fatalf("unexpected subject: %s", principal.Subject)
		}

		responseWriter.WriteHeader(http.StatusNoContent)
	}))

	request := httptest.NewRequest(http.MethodGet, "/objects", nil)
	request.Header.Set("Authorization", "Bearer any-token")
	responseRecorder := httptest.NewRecorder()
	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", responseRecorder.Code)
	}
}

func TestRequireRoleRejectsMissingRole(t *testing.T) {
	t.Parallel()

	handler := withPrincipal(
		Principal{Roles: []string{"engineer"}},
		RequireRole("admin")(http.HandlerFunc(func(
			responseWriter http.ResponseWriter,
			request *http.Request,
		) {
			responseWriter.WriteHeader(http.StatusNoContent)
		})),
	)

	request := httptest.NewRequest(http.MethodGet, "/objects", nil)
	responseRecorder := httptest.NewRecorder()
	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", responseRecorder.Code)
	}
}

func TestRequireRoleAllowsMatchingRole(t *testing.T) {
	t.Parallel()

	handler := withPrincipal(
		Principal{Roles: []string{"admin"}},
		RequireRole("admin")(http.HandlerFunc(func(
			responseWriter http.ResponseWriter,
			request *http.Request,
		) {
			responseWriter.WriteHeader(http.StatusNoContent)
		})),
	)

	request := httptest.NewRequest(http.MethodGet, "/objects", nil)
	responseRecorder := httptest.NewRecorder()
	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", responseRecorder.Code)
	}
}

func TestRequireScopeRequiresAllScopes(t *testing.T) {
	t.Parallel()

	handler := withPrincipal(
		Principal{Scopes: []string{"x"}},
		RequireScope("x", "y")(http.HandlerFunc(func(
			responseWriter http.ResponseWriter,
			request *http.Request,
		) {
			responseWriter.WriteHeader(http.StatusNoContent)
		})),
	)

	request := httptest.NewRequest(http.MethodGet, "/objects", nil)
	responseRecorder := httptest.NewRecorder()
	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", responseRecorder.Code)
	}
}

func TestRequireScopeAllowsWhenAllScopesPresent(t *testing.T) {
	t.Parallel()

	handler := withPrincipal(
		Principal{Scopes: []string{"x", "y", "z"}},
		RequireScope("x", "y")(http.HandlerFunc(func(
			responseWriter http.ResponseWriter,
			request *http.Request,
		) {
			responseWriter.WriteHeader(http.StatusNoContent)
		})),
	)

	request := httptest.NewRequest(http.MethodGet, "/objects", nil)
	responseRecorder := httptest.NewRecorder()
	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", responseRecorder.Code)
	}
}

func TestRequireAudienceRejectsMismatchedAudience(t *testing.T) {
	t.Parallel()

	handler := withPrincipal(
		Principal{Audience: []string{"controlitix-internal"}},
		RequireAudience("controlitix-api")(http.HandlerFunc(func(
			responseWriter http.ResponseWriter,
			request *http.Request,
		) {
			responseWriter.WriteHeader(http.StatusNoContent)
		})),
	)

	request := httptest.NewRequest(http.MethodGet, "/objects", nil)
	responseRecorder := httptest.NewRecorder()
	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", responseRecorder.Code)
	}
}

func TestRequireAudienceAllowsMatchedAudience(t *testing.T) {
	t.Parallel()

	handler := withPrincipal(
		Principal{Audience: []string{"controlitix-api"}},
		RequireAudience("controlitix-api")(http.HandlerFunc(func(
			responseWriter http.ResponseWriter,
			request *http.Request,
		) {
			responseWriter.WriteHeader(http.StatusNoContent)
		})),
	)

	request := httptest.NewRequest(http.MethodGet, "/objects", nil)
	responseRecorder := httptest.NewRecorder()
	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", responseRecorder.Code)
	}
}

func withPrincipal(principal Principal, next http.Handler) http.Handler {
	return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		request = request.WithContext(WithPrincipal(request.Context(), principal))
		next.ServeHTTP(responseWriter, request)
	})
}
