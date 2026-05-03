package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
	"github.com/EthernalFox/Controlitix/shared/authctx"
)

type auditRecorderStub struct {
	events []domain.AuditEvent
}

func (stub *auditRecorderStub) Record(_ context.Context, event domain.AuditEvent) {
	stub.events = append(stub.events, event)
}

func TestRequireAnyRole(t *testing.T) {
	tests := []struct {
		name       string
		roles      []string
		required   []string
		statusCode int
	}{
		{name: "operator allowed", roles: []string{"operator"}, required: []string{"operator", "admin"}, statusCode: http.StatusOK},
		{name: "admin bypass", roles: []string{"admin"}, required: []string{"operator"}, statusCode: http.StatusOK},
		{name: "engineer forbidden for acknowledge", roles: []string{"engineer"}, required: []string{"operator", "admin"}, statusCode: http.StatusForbidden},
		{name: "missing principal", roles: nil, required: []string{"operator"}, statusCode: http.StatusUnauthorized},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			recorder := &auditRecorderStub{}
			handler := RequireAnyRole(recorder, testCase.required...)(http.HandlerFunc(func(responseWriter http.ResponseWriter, _ *http.Request) {
				responseWriter.WriteHeader(http.StatusOK)
			}))

			request := httptest.NewRequest(http.MethodGet, "/api/alarms", nil)
			if testCase.roles != nil {
				principal := authctx.Principal{Subject: "user-1", Roles: testCase.roles, Username: "user-1"}
				request = request.WithContext(authctx.WithPrincipal(request.Context(), principal))
			}
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)
			if response.Code != testCase.statusCode {
				t.Fatalf("expected status %d, got %d", testCase.statusCode, response.Code)
			}

			if testCase.statusCode == http.StatusForbidden && len(recorder.events) != 1 {
				t.Fatalf("expected forbidden audit event")
			}
		})
	}
}
