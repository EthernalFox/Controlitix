package http

import (
	"context"
	"errors"
	"log/slog"
	stdhttp "net/http"
	"strings"

	"github.com/EthernalFox/Controlitix/ms-auth/internal/domain"
	"github.com/EthernalFox/Controlitix/ms-auth/internal/usecase"
	"github.com/go-chi/chi/v5"
)

type ServiceTokenHandler struct {
	service serviceTokenIssuerUsecase
	logger  *slog.Logger
}

type serviceTokenIssuerUsecase interface {
	Issue(
		ctx context.Context,
		input usecase.ServiceTokenInput,
	) (*usecase.ServiceTokenResult, error)
}

type serviceTokenRequest struct {
	GrantType    string `json:"grant_type"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Scope        string `json:"scope,omitempty"`
}

type serviceTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

func NewServiceTokenHandler(
	service serviceTokenIssuerUsecase,
	logger *slog.Logger,
) *ServiceTokenHandler {
	if logger == nil {
		logger = slog.Default()
	}

	return &ServiceTokenHandler{
		service: service,
		logger:  logger,
	}
}

func (h *ServiceTokenHandler) Register(router chi.Router) {
	router.Post("/service-token", h.issueServiceToken)
}

func (h *ServiceTokenHandler) issueServiceToken(
	responseWriter stdhttp.ResponseWriter,
	request *stdhttp.Request,
) {
	var payload serviceTokenRequest
	if err := decodeJSONBody(request.Body, &payload); err != nil {
		WriteProblem(responseWriter, Problem{
			Type:   "/errors/validation",
			Title:  "Validation failed",
			Status: stdhttp.StatusBadRequest,
		})
		return
	}

	result, err := h.service.Issue(request.Context(), usecase.ServiceTokenInput{
		GrantType:    payload.GrantType,
		ClientID:     payload.ClientID,
		ClientSecret: payload.ClientSecret,
		Scope:        payload.Scope,
		UserAgent:    request.UserAgent(),
		IP:           extractClientIP(request),
	})
	if err != nil {
		h.writeIssueError(responseWriter, err)
		return
	}

	writeJSON(responseWriter, stdhttp.StatusOK, serviceTokenResponse{
		AccessToken: result.AccessToken,
		TokenType:   result.TokenType,
		ExpiresIn:   result.ExpiresIn,
		Scope:       strings.TrimSpace(result.Scope),
	})
}

func (h *ServiceTokenHandler) writeIssueError(
	responseWriter stdhttp.ResponseWriter,
	err error,
) {
	switch {
	case errors.Is(err, domain.ErrUnsupportedGrant):
		WriteProblem(responseWriter, Problem{
			Type:   "/errors/auth/unsupported-grant",
			Title:  "Unsupported grant type",
			Status: stdhttp.StatusBadRequest,
		})
	case errors.Is(err, domain.ErrInvalidScope):
		WriteProblem(responseWriter, Problem{
			Type:   "/errors/auth/invalid-scope",
			Title:  "Invalid scope",
			Status: stdhttp.StatusBadRequest,
		})
	case errors.Is(err, domain.ErrInvalidClient):
		WriteProblem(responseWriter, Problem{
			Type:   "/errors/auth/invalid-client",
			Title:  "Invalid client",
			Status: stdhttp.StatusUnauthorized,
		})
	case errors.Is(err, domain.ErrClientDisabled):
		WriteProblem(responseWriter, Problem{
			Type:   "/errors/auth/client-disabled",
			Title:  "Client disabled",
			Status: stdhttp.StatusForbidden,
		})
	case errors.Is(err, domain.ErrValidation):
		WriteProblem(responseWriter, Problem{
			Type:   "/errors/validation",
			Title:  "Validation failed",
			Status: stdhttp.StatusBadRequest,
		})
	default:
		h.logger.Error(
			"failed to issue service token",
			"method",
			"ServiceTokenHandler.issueServiceToken",
			"error",
			err,
		)
		WriteProblem(responseWriter, Problem{
			Type:   "/errors/internal",
			Title:  "Internal error",
			Status: stdhttp.StatusInternalServerError,
		})
	}
}
