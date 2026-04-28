package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	stdhttp "net/http"
	"net/http/httptest"
	"testing"

	"github.com/EthernalFox/Controlitix/ms-auth/internal/domain"
	"github.com/EthernalFox/Controlitix/ms-auth/internal/usecase"
	"github.com/go-chi/chi/v5"
)

type serviceTokenUsecaseMock struct {
	issue func(ctx context.Context, input usecase.ServiceTokenInput) (*usecase.ServiceTokenResult, error)
}

func (m *serviceTokenUsecaseMock) Issue(
	ctx context.Context,
	input usecase.ServiceTokenInput,
) (*usecase.ServiceTokenResult, error) {
	return m.issue(ctx, input)
}

func TestServiceTokenHandlerSuccess(t *testing.T) {
	t.Parallel()

	mock := &serviceTokenUsecaseMock{
		issue: func(ctx context.Context, input usecase.ServiceTokenInput) (*usecase.ServiceTokenResult, error) {
			return &usecase.ServiceTokenResult{
				AccessToken: "svc-token",
				TokenType:   "Bearer",
				ExpiresIn:   900,
				Scope:       "config.read",
			}, nil
		},
	}
	handler := NewServiceTokenHandler(mock, nil)

	router := chi.NewRouter()
	router.Route("/api/auth", handler.Register)

	response := performServiceTokenRequest(
		t,
		router,
		map[string]any{
			"grant_type":    "client_credentials",
			"client_id":     "ms-poll",
			"client_secret": "secret",
			"scope":         "config.read",
		},
	)
	if response.Code != stdhttp.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", response.Code, response.Body.String())
	}
}

func TestServiceTokenHandlerErrorMapping(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		err         error
		status      int
		problemType string
	}{
		{
			name:        "unsupported grant",
			err:         domain.ErrUnsupportedGrant,
			status:      stdhttp.StatusBadRequest,
			problemType: "/errors/auth/unsupported-grant",
		},
		{
			name:        "invalid scope",
			err:         domain.ErrInvalidScope,
			status:      stdhttp.StatusBadRequest,
			problemType: "/errors/auth/invalid-scope",
		},
		{
			name:        "invalid client",
			err:         domain.ErrInvalidClient,
			status:      stdhttp.StatusUnauthorized,
			problemType: "/errors/auth/invalid-client",
		},
		{
			name:        "client disabled",
			err:         domain.ErrClientDisabled,
			status:      stdhttp.StatusForbidden,
			problemType: "/errors/auth/client-disabled",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mock := &serviceTokenUsecaseMock{
				issue: func(ctx context.Context, input usecase.ServiceTokenInput) (*usecase.ServiceTokenResult, error) {
					return nil, tc.err
				},
			}
			handler := NewServiceTokenHandler(mock, nil)

			router := chi.NewRouter()
			router.Route("/api/auth", handler.Register)

			response := performServiceTokenRequest(
				t,
				router,
				map[string]any{
					"grant_type":    "client_credentials",
					"client_id":     "ms-poll",
					"client_secret": "secret",
					"scope":         "config.read",
				},
			)
			if response.Code != tc.status {
				t.Fatalf("unexpected status: %d body=%s", response.Code, response.Body.String())
			}

			var problem Problem
			if err := json.Unmarshal(response.Body.Bytes(), &problem); err != nil {
				t.Fatalf("unmarshal problem body: %v", err)
			}
			if problem.Type != tc.problemType {
				t.Fatalf("unexpected problem type: got %s want %s", problem.Type, tc.problemType)
			}
		})
	}
}

func TestServiceTokenHandlerValidation(t *testing.T) {
	t.Parallel()

	mock := &serviceTokenUsecaseMock{
		issue: func(ctx context.Context, input usecase.ServiceTokenInput) (*usecase.ServiceTokenResult, error) {
			return nil, errors.New("unexpected call")
		},
	}
	handler := NewServiceTokenHandler(mock, nil)

	router := chi.NewRouter()
	router.Route("/api/auth", handler.Register)

	request := httptest.NewRequest(stdhttp.MethodPost, "/api/auth/service-token", bytes.NewReader([]byte("{")))
	request.Header.Set("Content-Type", "application/json")

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != stdhttp.StatusBadRequest {
		t.Fatalf("unexpected status: %d body=%s", response.Code, response.Body.String())
	}
}

func performServiceTokenRequest(
	t *testing.T,
	handler stdhttp.Handler,
	payload map[string]any,
) *httptest.ResponseRecorder {
	t.Helper()

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal request payload: %v", err)
	}

	request := httptest.NewRequest(
		stdhttp.MethodPost,
		"/api/auth/service-token",
		bytes.NewReader(bodyBytes),
	)
	request.Header.Set("Content-Type", "application/json")

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}
