package http

import (
	"context"
	"encoding/json"
	stdhttp "net/http"
	"net/http/httptest"
	"testing"
)

func TestRequireRoleAllowsAdmin(t *testing.T) {
	t.Parallel()

	handler := RequireRole("admin")(stdhttp.HandlerFunc(func(
		responseWriter stdhttp.ResponseWriter,
		_ *stdhttp.Request,
	) {
		responseWriter.WriteHeader(stdhttp.StatusNoContent)
	}))

	request := httptest.NewRequest(stdhttp.MethodGet, "/admin/users", nil)
	request = request.WithContext(context.WithValue(request.Context(), principalContextKey, Principal{
		Subject: "local:admin",
		Roles:   []string{"admin"},
	}))

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != stdhttp.StatusNoContent {
		t.Fatalf("unexpected status: %d", response.Code)
	}
}

func TestRequireRoleBlocksNonAdmin(t *testing.T) {
	t.Parallel()

	handler := RequireRole("admin")(stdhttp.HandlerFunc(func(
		responseWriter stdhttp.ResponseWriter,
		_ *stdhttp.Request,
	) {
		responseWriter.WriteHeader(stdhttp.StatusNoContent)
	}))

	request := httptest.NewRequest(stdhttp.MethodGet, "/admin/users", nil)
	request = request.WithContext(context.WithValue(request.Context(), principalContextKey, Principal{
		Subject: "local:engineer",
		Roles:   []string{"engineer"},
	}))

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != stdhttp.StatusForbidden {
		t.Fatalf("unexpected status: %d", response.Code)
	}

	var problem Problem
	if err := json.Unmarshal(response.Body.Bytes(), &problem); err != nil {
		t.Fatalf("unmarshal problem response: %v", err)
	}
	if problem.Type != "/errors/auth/forbidden" {
		t.Fatalf("unexpected problem type: %s", problem.Type)
	}
}
