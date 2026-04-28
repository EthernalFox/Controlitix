package http

import (
	"errors"
	"fmt"
	"log/slog"
	stdhttp "net/http"
	"strconv"
	"strings"
	"time"

	"github.com/EthernalFox/Controlitix/ms-auth/internal/domain"
	"github.com/EthernalFox/Controlitix/ms-auth/internal/infrastructure/repository"
	"github.com/EthernalFox/Controlitix/ms-auth/internal/usecase"
	"github.com/go-chi/chi/v5"
)

const (
	defaultPageLimit = 50
	maxPageLimit     = 200
)

type AdminHandler struct {
	service  *usecase.AdminService
	users    *repository.UserRepository
	sources  *repository.IdentitySourceRepository
	services *repository.ServiceAccountRepository
	audit    *repository.AuditRepository
	logger   *slog.Logger
}

type listResponse struct {
	Data   any `json:"data"`
	Total  int `json:"total"`
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

type createUserRequest struct {
	Username    string   `json:"username"`
	DisplayName string   `json:"display_name"`
	Email       string   `json:"email"`
	Password    string   `json:"password"`
	Roles       []string `json:"roles"`
}

type updateUserRequest struct {
	DisplayName *string `json:"display_name,omitempty"`
	Email       *string `json:"email,omitempty"`
	Password    *string `json:"password,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
}

type setUserRolesRequest struct {
	Roles []string `json:"roles"`
}

type createServiceAccountRequest struct {
	ClientID    string   `json:"client_id"`
	DisplayName string   `json:"display_name"`
	Scopes      []string `json:"scopes"`
}

type updateServiceAccountRequest struct {
	DisplayName *string  `json:"display_name,omitempty"`
	Scopes      []string `json:"scopes,omitempty"`
	IsActive    *bool    `json:"is_active,omitempty"`
}

type userResponse struct {
	ID          string     `json:"id"`
	Subject     string     `json:"subject"`
	Username    string     `json:"username"`
	DisplayName string     `json:"display_name"`
	Email       string     `json:"email"`
	Roles       []string   `json:"roles"`
	Source      string     `json:"source"`
	IsActive    bool       `json:"is_active"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type serviceAccountResponse struct {
	ID          string     `json:"id"`
	ClientID    string     `json:"client_id"`
	DisplayName string     `json:"display_name"`
	Scopes      []string   `json:"scopes"`
	IsActive    bool       `json:"is_active"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type createServiceAccountResponse struct {
	ServiceAccount serviceAccountResponse `json:"service_account"`
	ClientSecret   string                 `json:"client_secret"`
}

type rotateSecretResponse struct {
	ClientSecret string `json:"client_secret"`
}

func NewAdminHandler(
	service *usecase.AdminService,
	users *repository.UserRepository,
	sources *repository.IdentitySourceRepository,
	services *repository.ServiceAccountRepository,
	audit *repository.AuditRepository,
	logger *slog.Logger,
) *AdminHandler {
	if logger == nil {
		logger = slog.Default()
	}

	return &AdminHandler{
		service:  service,
		users:    users,
		sources:  sources,
		services: services,
		audit:    audit,
		logger:   logger,
	}
}

func (h *AdminHandler) Register(router chi.Router) {
	router.Get("/admin/users", h.listUsers)
	router.Post("/admin/users", h.createUser)
	router.Get("/admin/users/{id}", h.getUser)
	router.Patch("/admin/users/{id}", h.updateUser)
	router.Delete("/admin/users/{id}", h.deleteUser)
	router.Post("/admin/users/{id}/roles", h.setUserRoles)
	router.Delete("/admin/users/{id}/roles/{role}", h.removeUserRole)
	router.Post("/admin/users/{id}/revoke-sessions", h.revokeUserSessions)

	router.Get("/admin/roles", h.listRoles)

	router.Get("/admin/identity-sources", h.listIdentitySources)
	router.Get("/admin/identity-sources/{id}", h.getIdentitySource)
	router.Post("/admin/identity-sources", h.notImplemented)
	router.Patch("/admin/identity-sources/{id}", h.notImplemented)
	router.Delete("/admin/identity-sources/{id}", h.notImplemented)
	router.Get("/admin/identity-sources/{id}/role-mappings", h.listIdentitySourceRoleMappings)
	router.Post("/admin/identity-sources/{id}/role-mappings", h.notImplemented)
	router.Delete("/admin/identity-sources/{id}/role-mappings/{mid}", h.notImplemented)

	router.Get("/admin/service-accounts", h.listServiceAccounts)
	router.Post("/admin/service-accounts", h.createServiceAccount)
	router.Get("/admin/service-accounts/{id}", h.getServiceAccount)
	router.Patch("/admin/service-accounts/{id}", h.updateServiceAccount)
	router.Post("/admin/service-accounts/{id}/rotate-secret", h.rotateServiceAccountSecret)
	router.Delete("/admin/service-accounts/{id}", h.deleteServiceAccount)

	router.Get("/admin/audit", h.listAudit)
}

func (h *AdminHandler) listUsers(
	responseWriter stdhttp.ResponseWriter,
	request *stdhttp.Request,
) {
	offset, limit, err := parsePagination(request)
	if err != nil {
		h.writeAdminError(responseWriter, domain.ErrValidation, err)
		return
	}

	filter := repository.UserListFilter{
		Offset: offset,
		Limit:  limit,
		Query:  strings.TrimSpace(request.URL.Query().Get("q")),
	}
	if sourceIDRaw := strings.TrimSpace(request.URL.Query().Get("source_id")); sourceIDRaw != "" {
		sourceID, parseErr := strconv.Atoi(sourceIDRaw)
		if parseErr != nil {
			h.writeAdminError(responseWriter, domain.ErrValidation, parseErr)
			return
		}
		filter.SourceID = &sourceID
	}
	if activeRaw := strings.TrimSpace(request.URL.Query().Get("is_active")); activeRaw != "" {
		isActive, parseErr := strconv.ParseBool(activeRaw)
		if parseErr != nil {
			h.writeAdminError(responseWriter, domain.ErrValidation, parseErr)
			return
		}
		filter.IsActive = &isActive
	}

	users, total, err := h.users.List(request.Context(), filter)
	if err != nil {
		h.writeAdminError(responseWriter, err, err)
		return
	}

	data := make([]userResponse, 0, len(users))
	for _, user := range users {
		if user == nil {
			continue
		}

		roles, rolesErr := h.users.GetRoles(request.Context(), user.ID)
		if rolesErr != nil {
			h.writeAdminError(responseWriter, rolesErr, rolesErr)
			return
		}
		source, sourceErr := h.sources.FindByID(request.Context(), user.SourceID)
		if sourceErr != nil {
			h.writeAdminError(responseWriter, sourceErr, sourceErr)
			return
		}

		sourceType := ""
		if source != nil {
			sourceType = source.Type
		}
		data = append(data, mapUserResponse(*user, roles, sourceType))
	}

	writeJSON(responseWriter, stdhttp.StatusOK, listResponse{
		Data:   data,
		Total:  total,
		Offset: offset,
		Limit:  limit,
	})
}

func (h *AdminHandler) createUser(
	responseWriter stdhttp.ResponseWriter,
	request *stdhttp.Request,
) {
	actor, ok := PrincipalFromContext(request.Context())
	if !ok {
		h.writeAdminError(responseWriter, domain.ErrForbidden, domain.ErrForbidden)
		return
	}

	var payload createUserRequest
	if err := decodeJSONBody(request.Body, &payload); err != nil {
		h.writeAdminError(responseWriter, domain.ErrValidation, err)
		return
	}

	createdUser, err := h.service.CreateUser(request.Context(), actor.Subject, usecase.CreateUserInput{
		Username:    payload.Username,
		DisplayName: payload.DisplayName,
		Email:       payload.Email,
		Password:    payload.Password,
		Roles:       payload.Roles,
	})
	if err != nil {
		h.writeAdminError(responseWriter, err, err)
		return
	}

	roles, err := h.users.GetRoles(request.Context(), createdUser.ID)
	if err != nil {
		h.writeAdminError(responseWriter, err, err)
		return
	}

	writeJSON(responseWriter, stdhttp.StatusCreated, mapUserResponse(*createdUser, roles, "local"))
}

func (h *AdminHandler) getUser(
	responseWriter stdhttp.ResponseWriter,
	request *stdhttp.Request,
) {
	userID := chi.URLParam(request, "id")

	user, err := h.users.FindByID(request.Context(), userID)
	if err != nil {
		h.writeAdminError(responseWriter, err, err)
		return
	}
	if user == nil {
		h.writeAdminError(responseWriter, domain.ErrUserNotFound, domain.ErrUserNotFound)
		return
	}

	roles, err := h.users.GetRoles(request.Context(), user.ID)
	if err != nil {
		h.writeAdminError(responseWriter, err, err)
		return
	}
	source, err := h.sources.FindByID(request.Context(), user.SourceID)
	if err != nil {
		h.writeAdminError(responseWriter, err, err)
		return
	}

	sourceType := ""
	if source != nil {
		sourceType = source.Type
	}

	writeJSON(responseWriter, stdhttp.StatusOK, mapUserResponse(*user, roles, sourceType))
}

func (h *AdminHandler) updateUser(
	responseWriter stdhttp.ResponseWriter,
	request *stdhttp.Request,
) {
	actor, ok := PrincipalFromContext(request.Context())
	if !ok {
		h.writeAdminError(responseWriter, domain.ErrForbidden, domain.ErrForbidden)
		return
	}

	userID := chi.URLParam(request, "id")
	var payload updateUserRequest
	if err := decodeJSONBody(request.Body, &payload); err != nil {
		h.writeAdminError(responseWriter, domain.ErrValidation, err)
		return
	}

	updatedUser, err := h.service.UpdateUser(request.Context(), actor.Subject, userID, usecase.UpdateUserInput{
		DisplayName: payload.DisplayName,
		Email:       payload.Email,
		Password:    payload.Password,
		IsActive:    payload.IsActive,
	})
	if err != nil {
		h.writeAdminError(responseWriter, err, err)
		return
	}

	roles, err := h.users.GetRoles(request.Context(), updatedUser.ID)
	if err != nil {
		h.writeAdminError(responseWriter, err, err)
		return
	}
	source, err := h.sources.FindByID(request.Context(), updatedUser.SourceID)
	if err != nil {
		h.writeAdminError(responseWriter, err, err)
		return
	}

	sourceType := ""
	if source != nil {
		sourceType = source.Type
	}

	writeJSON(responseWriter, stdhttp.StatusOK, mapUserResponse(*updatedUser, roles, sourceType))
}

func (h *AdminHandler) deleteUser(
	responseWriter stdhttp.ResponseWriter,
	request *stdhttp.Request,
) {
	actor, ok := PrincipalFromContext(request.Context())
	if !ok {
		h.writeAdminError(responseWriter, domain.ErrForbidden, domain.ErrForbidden)
		return
	}

	userID := chi.URLParam(request, "id")
	if err := h.service.DeactivateUser(request.Context(), actor.Subject, userID); err != nil {
		h.writeAdminError(responseWriter, err, err)
		return
	}

	responseWriter.WriteHeader(stdhttp.StatusNoContent)
}

func (h *AdminHandler) setUserRoles(
	responseWriter stdhttp.ResponseWriter,
	request *stdhttp.Request,
) {
	actor, ok := PrincipalFromContext(request.Context())
	if !ok {
		h.writeAdminError(responseWriter, domain.ErrForbidden, domain.ErrForbidden)
		return
	}

	userID := chi.URLParam(request, "id")
	var payload setUserRolesRequest
	if err := decodeJSONBody(request.Body, &payload); err != nil {
		h.writeAdminError(responseWriter, domain.ErrValidation, err)
		return
	}

	if err := h.service.SetUserRoles(request.Context(), actor.Subject, userID, payload.Roles); err != nil {
		h.writeAdminError(responseWriter, err, err)
		return
	}

	responseWriter.WriteHeader(stdhttp.StatusNoContent)
}

func (h *AdminHandler) removeUserRole(
	responseWriter stdhttp.ResponseWriter,
	request *stdhttp.Request,
) {
	actor, ok := PrincipalFromContext(request.Context())
	if !ok {
		h.writeAdminError(responseWriter, domain.ErrForbidden, domain.ErrForbidden)
		return
	}

	userID := chi.URLParam(request, "id")
	role := strings.TrimSpace(chi.URLParam(request, "role"))
	if role == "" {
		h.writeAdminError(responseWriter, domain.ErrValidation, domain.ErrValidation)
		return
	}

	if err := h.service.RemoveUserRole(request.Context(), actor.Subject, userID, role); err != nil {
		h.writeAdminError(responseWriter, err, err)
		return
	}

	responseWriter.WriteHeader(stdhttp.StatusNoContent)
}

func (h *AdminHandler) revokeUserSessions(
	responseWriter stdhttp.ResponseWriter,
	request *stdhttp.Request,
) {
	actor, ok := PrincipalFromContext(request.Context())
	if !ok {
		h.writeAdminError(responseWriter, domain.ErrForbidden, domain.ErrForbidden)
		return
	}

	userID := chi.URLParam(request, "id")
	if err := h.service.RevokeUserSessions(request.Context(), actor.Subject, userID); err != nil {
		h.writeAdminError(responseWriter, err, err)
		return
	}

	responseWriter.WriteHeader(stdhttp.StatusNoContent)
}

func (h *AdminHandler) listRoles(
	responseWriter stdhttp.ResponseWriter,
	request *stdhttp.Request,
) {
	roles, err := h.users.ListRolesCatalog(request.Context())
	if err != nil {
		h.writeAdminError(responseWriter, err, err)
		return
	}

	writeJSON(responseWriter, stdhttp.StatusOK, roles)
}

func (h *AdminHandler) listIdentitySources(
	responseWriter stdhttp.ResponseWriter,
	request *stdhttp.Request,
) {
	sources, err := h.sources.List(request.Context())
	if err != nil {
		h.writeAdminError(responseWriter, err, err)
		return
	}

	writeJSON(responseWriter, stdhttp.StatusOK, sources)
}

func (h *AdminHandler) getIdentitySource(
	responseWriter stdhttp.ResponseWriter,
	request *stdhttp.Request,
) {
	sourceID, err := strconv.Atoi(chi.URLParam(request, "id"))
	if err != nil {
		h.writeAdminError(responseWriter, domain.ErrValidation, err)
		return
	}

	source, err := h.sources.FindByID(request.Context(), sourceID)
	if err != nil {
		h.writeAdminError(responseWriter, err, err)
		return
	}
	if source == nil {
		h.writeAdminError(responseWriter, domain.ErrSourceNotFound, domain.ErrSourceNotFound)
		return
	}

	writeJSON(responseWriter, stdhttp.StatusOK, source)
}

func (h *AdminHandler) listIdentitySourceRoleMappings(
	responseWriter stdhttp.ResponseWriter,
	_ *stdhttp.Request,
) {
	writeJSON(responseWriter, stdhttp.StatusOK, map[string]any{"data": []any{}})
}

func (h *AdminHandler) listServiceAccounts(
	responseWriter stdhttp.ResponseWriter,
	request *stdhttp.Request,
) {
	offset, limit, err := parsePagination(request)
	if err != nil {
		h.writeAdminError(responseWriter, domain.ErrValidation, err)
		return
	}

	accounts, total, err := h.services.List(request.Context(), offset, limit)
	if err != nil {
		h.writeAdminError(responseWriter, err, err)
		return
	}

	data := make([]serviceAccountResponse, 0, len(accounts))
	for _, account := range accounts {
		if account == nil {
			continue
		}
		data = append(data, mapServiceAccountResponse(*account))
	}

	writeJSON(responseWriter, stdhttp.StatusOK, listResponse{
		Data:   data,
		Total:  total,
		Offset: offset,
		Limit:  limit,
	})
}

func (h *AdminHandler) createServiceAccount(
	responseWriter stdhttp.ResponseWriter,
	request *stdhttp.Request,
) {
	actor, ok := PrincipalFromContext(request.Context())
	if !ok {
		h.writeAdminError(responseWriter, domain.ErrForbidden, domain.ErrForbidden)
		return
	}

	var payload createServiceAccountRequest
	if err := decodeJSONBody(request.Body, &payload); err != nil {
		h.writeAdminError(responseWriter, domain.ErrValidation, err)
		return
	}

	account, secret, err := h.service.CreateServiceAccount(
		request.Context(),
		actor.Subject,
		usecase.CreateServiceAccountInput{
			ClientID:    payload.ClientID,
			DisplayName: payload.DisplayName,
			Scopes:      payload.Scopes,
		},
	)
	if err != nil {
		h.writeAdminError(responseWriter, err, err)
		return
	}

	writeJSON(responseWriter, stdhttp.StatusCreated, createServiceAccountResponse{
		ServiceAccount: mapServiceAccountResponse(*account),
		ClientSecret:   secret,
	})
}

func (h *AdminHandler) getServiceAccount(
	responseWriter stdhttp.ResponseWriter,
	request *stdhttp.Request,
) {
	accountID := chi.URLParam(request, "id")
	account, err := h.services.FindByID(request.Context(), accountID)
	if err != nil {
		h.writeAdminError(responseWriter, err, err)
		return
	}
	if account == nil {
		h.writeAdminError(responseWriter, domain.ErrServiceNotFound, domain.ErrServiceNotFound)
		return
	}

	writeJSON(responseWriter, stdhttp.StatusOK, mapServiceAccountResponse(*account))
}

func (h *AdminHandler) updateServiceAccount(
	responseWriter stdhttp.ResponseWriter,
	request *stdhttp.Request,
) {
	actor, ok := PrincipalFromContext(request.Context())
	if !ok {
		h.writeAdminError(responseWriter, domain.ErrForbidden, domain.ErrForbidden)
		return
	}

	accountID := chi.URLParam(request, "id")
	var payload updateServiceAccountRequest
	if err := decodeJSONBody(request.Body, &payload); err != nil {
		h.writeAdminError(responseWriter, domain.ErrValidation, err)
		return
	}

	if err := h.service.UpdateServiceAccount(
		request.Context(),
		actor.Subject,
		accountID,
		usecase.UpdateServiceAccountInput{
			DisplayName: payload.DisplayName,
			Scopes:      payload.Scopes,
			IsActive:    payload.IsActive,
		},
	); err != nil {
		h.writeAdminError(responseWriter, err, err)
		return
	}

	account, err := h.services.FindByID(request.Context(), accountID)
	if err != nil {
		h.writeAdminError(responseWriter, err, err)
		return
	}
	if account == nil {
		h.writeAdminError(responseWriter, domain.ErrServiceNotFound, domain.ErrServiceNotFound)
		return
	}

	writeJSON(responseWriter, stdhttp.StatusOK, mapServiceAccountResponse(*account))
}

func (h *AdminHandler) rotateServiceAccountSecret(
	responseWriter stdhttp.ResponseWriter,
	request *stdhttp.Request,
) {
	actor, ok := PrincipalFromContext(request.Context())
	if !ok {
		h.writeAdminError(responseWriter, domain.ErrForbidden, domain.ErrForbidden)
		return
	}

	accountID := chi.URLParam(request, "id")
	secret, err := h.service.RotateServiceAccountSecret(request.Context(), actor.Subject, accountID)
	if err != nil {
		h.writeAdminError(responseWriter, err, err)
		return
	}

	writeJSON(responseWriter, stdhttp.StatusOK, rotateSecretResponse{
		ClientSecret: secret,
	})
}

func (h *AdminHandler) deleteServiceAccount(
	responseWriter stdhttp.ResponseWriter,
	request *stdhttp.Request,
) {
	actor, ok := PrincipalFromContext(request.Context())
	if !ok {
		h.writeAdminError(responseWriter, domain.ErrForbidden, domain.ErrForbidden)
		return
	}

	accountID := chi.URLParam(request, "id")
	if err := h.service.DeleteServiceAccount(request.Context(), actor.Subject, accountID); err != nil {
		h.writeAdminError(responseWriter, err, err)
		return
	}

	responseWriter.WriteHeader(stdhttp.StatusNoContent)
}

func (h *AdminHandler) listAudit(
	responseWriter stdhttp.ResponseWriter,
	request *stdhttp.Request,
) {
	offset, limit, err := parsePagination(request)
	if err != nil {
		h.writeAdminError(responseWriter, domain.ErrValidation, err)
		return
	}

	filter := repository.AuditListFilter{
		Offset: offset,
		Limit:  limit,
		Action: strings.TrimSpace(request.URL.Query().Get("action")),
		Actor:  strings.TrimSpace(request.URL.Query().Get("actor")),
	}
	if value := strings.TrimSpace(request.URL.Query().Get("from")); value != "" {
		from, parseErr := time.Parse(time.RFC3339, value)
		if parseErr != nil {
			h.writeAdminError(responseWriter, domain.ErrValidation, parseErr)
			return
		}
		filter.From = &from
	}
	if value := strings.TrimSpace(request.URL.Query().Get("to")); value != "" {
		to, parseErr := time.Parse(time.RFC3339, value)
		if parseErr != nil {
			h.writeAdminError(responseWriter, domain.ErrValidation, parseErr)
			return
		}
		filter.To = &to
	}

	records, total, err := h.audit.List(request.Context(), filter)
	if err != nil {
		h.writeAdminError(responseWriter, err, err)
		return
	}

	writeJSON(responseWriter, stdhttp.StatusOK, listResponse{
		Data:   records,
		Total:  total,
		Offset: offset,
		Limit:  limit,
	})
}

func (h *AdminHandler) notImplemented(
	responseWriter stdhttp.ResponseWriter,
	_ *stdhttp.Request,
) {
	WriteProblem(responseWriter, Problem{
		Type:   "/errors/not-implemented",
		Title:  "Not implemented",
		Status: stdhttp.StatusNotImplemented,
	})
}

func parsePagination(request *stdhttp.Request) (int, int, error) {
	offset := 0
	limit := defaultPageLimit

	if rawOffset := strings.TrimSpace(request.URL.Query().Get("offset")); rawOffset != "" {
		value, err := strconv.Atoi(rawOffset)
		if err != nil || value < 0 {
			return 0, 0, domain.ErrValidation
		}
		offset = value
	}
	if rawLimit := strings.TrimSpace(request.URL.Query().Get("limit")); rawLimit != "" {
		value, err := strconv.Atoi(rawLimit)
		if err != nil || value <= 0 {
			return 0, 0, domain.ErrValidation
		}
		limit = value
	}
	if limit > maxPageLimit {
		limit = maxPageLimit
	}

	return offset, limit, nil
}

func mapUserResponse(
	user domain.User,
	roles []string,
	sourceType string,
) userResponse {
	return userResponse{
		ID:          user.ID,
		Subject:     user.Subject,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		Email:       user.Email,
		Roles:       roles,
		Source:      sourceType,
		IsActive:    user.IsActive,
		LastLoginAt: user.LastLoginAt,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}

func mapServiceAccountResponse(
	account domain.ServiceAccount,
) serviceAccountResponse {
	return serviceAccountResponse{
		ID:          account.ID,
		ClientID:    account.ClientID,
		DisplayName: account.DisplayName,
		Scopes:      account.Scopes,
		IsActive:    account.IsActive,
		LastUsedAt:  account.LastUsedAt,
		CreatedAt:   account.CreatedAt,
		UpdatedAt:   account.UpdatedAt,
	}
}

func (h *AdminHandler) writeAdminError(
	responseWriter stdhttp.ResponseWriter,
	err error,
	logError error,
) {
	switch {
	case errors.Is(err, domain.ErrValidation):
		WriteProblem(responseWriter, Problem{
			Type:   "/errors/validation",
			Title:  "Validation failed",
			Status: stdhttp.StatusUnprocessableEntity,
		})
	case errors.Is(err, domain.ErrForbidden):
		WriteProblem(responseWriter, Problem{
			Type:   "/errors/auth/forbidden",
			Title:  "Forbidden",
			Status: stdhttp.StatusForbidden,
		})
	case errors.Is(err, domain.ErrForbiddenSelfDemote):
		WriteProblem(responseWriter, Problem{
			Type:   "/errors/auth/forbidden-self-demote",
			Title:  "Forbidden self demote",
			Status: stdhttp.StatusConflict,
		})
	case errors.Is(err, domain.ErrLastAdmin):
		WriteProblem(responseWriter, Problem{
			Type:   "/errors/auth/last-admin",
			Title:  "Last admin",
			Status: stdhttp.StatusConflict,
		})
	case errors.Is(err, domain.ErrUserNotFound),
		errors.Is(err, domain.ErrSourceNotFound),
		errors.Is(err, domain.ErrServiceNotFound):
		WriteProblem(responseWriter, Problem{
			Type:   "/errors/not-found",
			Title:  "Not found",
			Status: stdhttp.StatusNotFound,
		})
	case errors.Is(err, domain.ErrClientIDExists):
		WriteProblem(responseWriter, Problem{
			Type:   "/errors/conflict/client-id-exists",
			Title:  "Client id already exists",
			Status: stdhttp.StatusConflict,
		})
	case errors.Is(err, domain.ErrUserExists):
		WriteProblem(responseWriter, Problem{
			Type:   "/errors/conflict/user-exists",
			Title:  "User already exists",
			Status: stdhttp.StatusConflict,
		})
	case errors.Is(err, domain.ErrNotImplemented):
		WriteProblem(responseWriter, Problem{
			Type:   "/errors/not-implemented",
			Title:  "Not implemented",
			Status: stdhttp.StatusNotImplemented,
		})
	default:
		h.logger.Error(
			"admin request failed",
			"method",
			"AdminHandler.writeAdminError",
			"error",
			fmt.Sprintf("%v", logError),
		)
		WriteProblem(responseWriter, Problem{
			Type:   "/errors/internal",
			Title:  "Internal error",
			Status: stdhttp.StatusInternalServerError,
		})
	}
}
