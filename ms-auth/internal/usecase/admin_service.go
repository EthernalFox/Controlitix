package usecase

import (
	"context"
	cryptorand "crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/EthernalFox/Controlitix/ms-auth/internal/domain"
	"github.com/EthernalFox/Controlitix/ms-auth/internal/infrastructure/password"
	"github.com/EthernalFox/Controlitix/ms-auth/internal/infrastructure/repository"
)

const (
	localIdentitySourceID = 1
	serviceSecretLength   = 32
)

var clientIDPattern = regexp.MustCompile(`^[a-z][a-z0-9-]{2,63}$`)

type adminUserRepository interface {
	FindByID(ctx context.Context, id string) (*domain.User, error)
	FindBySubject(ctx context.Context, subject string) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) (*domain.User, error)
	Update(
		ctx context.Context,
		id string,
		displayName *string,
		email *string,
		passwordHash *string,
		isActive *bool,
	) (*domain.User, error)
	GetRoles(ctx context.Context, userID string) ([]string, error)
	SetRoles(ctx context.Context, userID string, roles []string, grantedBy *string) error
	CountActiveAdminsExcluding(ctx context.Context, excludeUserID string) (int, error)
}

type adminServiceAccountRepository interface {
	FindByID(ctx context.Context, id string) (*domain.ServiceAccount, error)
	Create(ctx context.Context, account *domain.ServiceAccount) (*domain.ServiceAccount, error)
	Update(ctx context.Context, account *domain.ServiceAccount) error
	UpdateSecretHash(ctx context.Context, id string, hash string) error
	Delete(ctx context.Context, id string) error
}

type adminSourceRepository interface {
	FindByID(ctx context.Context, id int) (*domain.IdentitySource, error)
}

type adminRefreshRepository interface {
	RevokeAllForUser(ctx context.Context, userID string, at time.Time, reason string) error
}

type adminAuditRecorder interface {
	Record(ctx context.Context, event domain.AuditEvent) error
}

type CreateUserInput struct {
	Username    string
	DisplayName string
	Email       string
	Password    string
	Roles       []string
}

type UpdateUserInput struct {
	DisplayName *string
	Email       *string
	Password    *string
	IsActive    *bool
}

type CreateServiceAccountInput struct {
	ClientID    string
	DisplayName string
	Scopes      []string
}

type UpdateServiceAccountInput struct {
	DisplayName *string
	Scopes      []string
	IsActive    *bool
}

type AdminService struct {
	users    adminUserRepository
	services adminServiceAccountRepository
	sources  adminSourceRepository
	audit    adminAuditRecorder
	refresh  adminRefreshRepository
	hasher   *password.Hasher
	logger   *slog.Logger
	clock    func() time.Time
	rand     io.Reader
}

func NewAdminService(
	users *repository.UserRepository,
	services *repository.ServiceAccountRepository,
	sources *repository.IdentitySourceRepository,
	audit *AuditService,
	refresh *repository.RefreshTokenRepository,
	hasher *password.Hasher,
	logger *slog.Logger,
) *AdminService {
	if logger == nil {
		logger = slog.Default()
	}

	return &AdminService{
		users:    users,
		services: services,
		sources:  sources,
		audit:    audit,
		refresh:  refresh,
		hasher:   hasher,
		logger:   logger,
		clock:    time.Now,
		rand:     cryptorand.Reader,
	}
}

func (s *AdminService) CreateUser(
	ctx context.Context,
	actorSubject string,
	input CreateUserInput,
) (*domain.User, error) {
	username := strings.TrimSpace(input.Username)
	password := strings.TrimSpace(input.Password)
	if username == "" || password == "" {
		return nil, domain.ErrValidation
	}

	roles, err := validateRoles(input.Roles)
	if err != nil {
		return nil, err
	}

	source, err := s.sources.FindByID(ctx, localIdentitySourceID)
	if err != nil {
		return nil, fmt.Errorf("find local identity source: %w", err)
	}
	if source == nil || !source.IsEnabled {
		return nil, domain.ErrSourceDisabled
	}

	passwordHash, err := s.hasher.Hash(password)
	if err != nil {
		return nil, fmt.Errorf("hash user password: %w", err)
	}

	subjectID, err := s.newUUIDv4String()
	if err != nil {
		return nil, fmt.Errorf("generate user subject id: %w", err)
	}

	createdUser, err := s.users.Create(ctx, &domain.User{
		Subject:      "local:" + subjectID,
		SourceID:     localIdentitySourceID,
		Username:     username,
		DisplayName:  strings.TrimSpace(input.DisplayName),
		Email:        strings.TrimSpace(input.Email),
		PasswordHash: passwordHash,
		IsActive:     true,
	})
	if err != nil {
		return nil, err
	}

	if err := s.users.SetRoles(ctx, createdUser.ID, roles, nil); err != nil {
		return nil, fmt.Errorf("set created user roles: %w", err)
	}

	s.recordAudit(ctx, domain.AuditEvent{
		ActorSubject: actorSubject,
		Action:       domain.AuditAdminUserCreate,
		Target:       createdUser.ID,
		Result:       domain.AuditResultSuccess,
		Metadata: map[string]any{
			"username": createdUser.Username,
			"roles":    roles,
		},
	})

	return createdUser, nil
}

func (s *AdminService) UpdateUser(
	ctx context.Context,
	actorSubject string,
	userID string,
	input UpdateUserInput,
) (*domain.User, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("find user for update: %w", err)
	}
	if user == nil {
		return nil, domain.ErrUserNotFound
	}

	var passwordHash *string
	if input.Password != nil {
		password := strings.TrimSpace(*input.Password)
		if password == "" {
			return nil, domain.ErrValidation
		}
		hashed, hashErr := s.hasher.Hash(password)
		if hashErr != nil {
			return nil, fmt.Errorf("hash updated user password: %w", hashErr)
		}
		passwordHash = &hashed
	}

	if input.IsActive != nil && !*input.IsActive && user.IsActive {
		if err := s.ensureCanDeactivateUser(ctx, user, actorSubject); err != nil {
			return nil, err
		}
	}

	updatedUser, err := s.users.Update(
		ctx,
		userID,
		input.DisplayName,
		input.Email,
		passwordHash,
		input.IsActive,
	)
	if err != nil {
		return nil, err
	}

	if input.IsActive != nil && !*input.IsActive && user.IsActive {
		if err := s.refresh.RevokeAllForUser(
			ctx,
			userID,
			s.clock().UTC(),
			"admin_deactivate",
		); err != nil {
			return nil, fmt.Errorf("revoke sessions on user deactivation: %w", err)
		}
	}

	s.recordAudit(ctx, domain.AuditEvent{
		ActorSubject: actorSubject,
		Action:       domain.AuditAdminUserUpdate,
		Target:       userID,
		Result:       domain.AuditResultSuccess,
		Metadata: map[string]any{
			"display_name_updated": input.DisplayName != nil,
			"email_updated":        input.Email != nil,
			"password_updated":     input.Password != nil,
			"is_active_updated":    input.IsActive != nil,
		},
	})

	return updatedUser, nil
}

func (s *AdminService) DeactivateUser(
	ctx context.Context,
	actorSubject string,
	userID string,
) error {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("find user for deactivation: %w", err)
	}
	if user == nil {
		return domain.ErrUserNotFound
	}
	if !user.IsActive {
		return nil
	}

	if err := s.ensureCanDeactivateUser(ctx, user, actorSubject); err != nil {
		return err
	}

	inactive := false
	if _, err := s.users.Update(ctx, userID, nil, nil, nil, &inactive); err != nil {
		return err
	}

	if err := s.refresh.RevokeAllForUser(
		ctx,
		userID,
		s.clock().UTC(),
		"admin_deactivate",
	); err != nil {
		return fmt.Errorf("revoke user sessions on deactivation: %w", err)
	}

	s.recordAudit(ctx, domain.AuditEvent{
		ActorSubject: actorSubject,
		Action:       domain.AuditAdminUserDeactivate,
		Target:       userID,
		Result:       domain.AuditResultSuccess,
	})

	return nil
}

func (s *AdminService) SetUserRoles(
	ctx context.Context,
	actorSubject string,
	userID string,
	roles []string,
) error {
	validatedRoles, err := validateRoles(roles)
	if err != nil {
		return err
	}

	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("find user for role update: %w", err)
	}
	if user == nil {
		return domain.ErrUserNotFound
	}

	currentRoles, err := s.users.GetRoles(ctx, userID)
	if err != nil {
		return fmt.Errorf("get current user roles: %w", err)
	}

	currentIsAdmin := containsRole(currentRoles, "admin")
	nextIsAdmin := containsRole(validatedRoles, "admin")

	if user.Subject == actorSubject && currentIsAdmin && !nextIsAdmin {
		return domain.ErrForbiddenSelfDemote
	}

	if user.IsActive && currentIsAdmin && !nextIsAdmin {
		count, countErr := s.users.CountActiveAdminsExcluding(ctx, user.ID)
		if countErr != nil {
			return fmt.Errorf("count remaining active admins for role update: %w", countErr)
		}
		if count == 0 {
			return domain.ErrLastAdmin
		}
	}

	if err := s.users.SetRoles(ctx, userID, validatedRoles, nil); err != nil {
		return fmt.Errorf("set user roles: %w", err)
	}

	s.recordAudit(ctx, domain.AuditEvent{
		ActorSubject: actorSubject,
		Action:       domain.AuditAdminUserSetRoles,
		Target:       userID,
		Result:       domain.AuditResultSuccess,
		Metadata: map[string]any{
			"roles": validatedRoles,
		},
	})

	return nil
}

func (s *AdminService) RemoveUserRole(
	ctx context.Context,
	actorSubject string,
	userID string,
	role string,
) error {
	currentRoles, err := s.users.GetRoles(ctx, userID)
	if err != nil {
		return fmt.Errorf("get current user roles for removal: %w", err)
	}

	nextRoles := make([]string, 0, len(currentRoles))
	for _, currentRole := range currentRoles {
		if currentRole == role {
			continue
		}
		nextRoles = append(nextRoles, currentRole)
	}

	return s.SetUserRoles(ctx, actorSubject, userID, nextRoles)
}

func (s *AdminService) RevokeUserSessions(
	ctx context.Context,
	actorSubject string,
	userID string,
) error {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("find user for session revoke: %w", err)
	}
	if user == nil {
		return domain.ErrUserNotFound
	}

	if err := s.refresh.RevokeAllForUser(
		ctx,
		userID,
		s.clock().UTC(),
		"admin_revoke_sessions",
	); err != nil {
		return fmt.Errorf("revoke user sessions: %w", err)
	}

	s.recordAudit(ctx, domain.AuditEvent{
		ActorSubject: actorSubject,
		Action:       domain.AuditAdminUserRevokeSessions,
		Target:       userID,
		Result:       domain.AuditResultSuccess,
	})

	return nil
}

func (s *AdminService) CreateServiceAccount(
	ctx context.Context,
	actorSubject string,
	input CreateServiceAccountInput,
) (*domain.ServiceAccount, string, error) {
	clientID := strings.TrimSpace(input.ClientID)
	displayName := strings.TrimSpace(input.DisplayName)
	scopes, err := validateScopes(input.Scopes)
	if err != nil {
		return nil, "", err
	}

	if !clientIDPattern.MatchString(clientID) || displayName == "" {
		return nil, "", domain.ErrValidation
	}

	secret, hash, err := s.generateServiceSecret()
	if err != nil {
		return nil, "", err
	}

	createdAccount, err := s.services.Create(ctx, &domain.ServiceAccount{
		ClientID:         clientID,
		ClientSecretHash: hash,
		DisplayName:      displayName,
		Scopes:           scopes,
		IsActive:         true,
	})
	if err != nil {
		return nil, "", err
	}

	s.recordAudit(ctx, domain.AuditEvent{
		ActorSubject: actorSubject,
		Action:       domain.AuditAdminServiceCreate,
		Target:       clientID,
		Result:       domain.AuditResultSuccess,
		Metadata: map[string]any{
			"scopes": scopes,
		},
	})

	return createdAccount, secret, nil
}

func (s *AdminService) RotateServiceAccountSecret(
	ctx context.Context,
	actorSubject string,
	id string,
) (string, error) {
	account, err := s.services.FindByID(ctx, id)
	if err != nil {
		return "", fmt.Errorf("find service account for secret rotation: %w", err)
	}
	if account == nil {
		return "", domain.ErrServiceNotFound
	}

	secret, hash, err := s.generateServiceSecret()
	if err != nil {
		return "", err
	}

	if err := s.services.UpdateSecretHash(ctx, id, hash); err != nil {
		return "", err
	}

	s.recordAudit(ctx, domain.AuditEvent{
		ActorSubject: actorSubject,
		Action:       domain.AuditAdminServiceRotateSecret,
		Target:       account.ClientID,
		Result:       domain.AuditResultSuccess,
	})

	return secret, nil
}

func (s *AdminService) UpdateServiceAccount(
	ctx context.Context,
	actorSubject string,
	id string,
	input UpdateServiceAccountInput,
) error {
	account, err := s.services.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("find service account for update: %w", err)
	}
	if account == nil {
		return domain.ErrServiceNotFound
	}

	if input.DisplayName != nil {
		value := strings.TrimSpace(*input.DisplayName)
		if value == "" {
			return domain.ErrValidation
		}
		account.DisplayName = value
	}
	if input.IsActive != nil {
		account.IsActive = *input.IsActive
	}
	if input.Scopes != nil {
		scopes, validationErr := validateScopes(input.Scopes)
		if validationErr != nil {
			return validationErr
		}
		account.Scopes = scopes
	}

	if err := s.services.Update(ctx, account); err != nil {
		return err
	}

	s.recordAudit(ctx, domain.AuditEvent{
		ActorSubject: actorSubject,
		Action:       domain.AuditAdminServiceUpdate,
		Target:       account.ClientID,
		Result:       domain.AuditResultSuccess,
	})

	return nil
}

func (s *AdminService) DeleteServiceAccount(
	ctx context.Context,
	actorSubject string,
	id string,
) error {
	account, err := s.services.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("find service account for delete: %w", err)
	}
	if account == nil {
		return domain.ErrServiceNotFound
	}

	if err := s.services.Delete(ctx, id); err != nil {
		return err
	}

	s.recordAudit(ctx, domain.AuditEvent{
		ActorSubject: actorSubject,
		Action:       domain.AuditAdminServiceDelete,
		Target:       account.ClientID,
		Result:       domain.AuditResultSuccess,
	})

	return nil
}

func (s *AdminService) ensureCanDeactivateUser(
	ctx context.Context,
	user *domain.User,
	actorSubject string,
) error {
	roles, err := s.users.GetRoles(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("get roles for deactivation checks: %w", err)
	}

	if user.Subject == actorSubject && containsRole(roles, "admin") {
		return domain.ErrForbiddenSelfDemote
	}

	if user.IsActive && containsRole(roles, "admin") {
		count, countErr := s.users.CountActiveAdminsExcluding(ctx, user.ID)
		if countErr != nil {
			return fmt.Errorf("count active admins for deactivation checks: %w", countErr)
		}
		if count == 0 {
			return domain.ErrLastAdmin
		}
	}

	return nil
}

func (s *AdminService) generateServiceSecret() (string, string, error) {
	rawSecret := make([]byte, serviceSecretLength)
	if _, err := io.ReadFull(s.rand, rawSecret); err != nil {
		return "", "", fmt.Errorf("generate service account secret: %w", err)
	}

	secret := base64.RawURLEncoding.EncodeToString(rawSecret)
	hash, err := s.hasher.Hash(secret)
	if err != nil {
		return "", "", fmt.Errorf("hash service account secret: %w", err)
	}

	return secret, hash, nil
}

func (s *AdminService) newUUIDv4String() (string, error) {
	randomBytes := make([]byte, 16)
	if _, err := io.ReadFull(s.rand, randomBytes); err != nil {
		return "", fmt.Errorf("read random bytes for uuid: %w", err)
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
	), nil
}

func validateRoles(input []string) ([]string, error) {
	roles := make([]string, 0, len(input))
	seen := make(map[string]struct{}, len(input))
	for _, role := range input {
		role = strings.TrimSpace(role)
		if role == "" {
			continue
		}
		if _, exists := seen[role]; exists {
			continue
		}
		if !domain.IsAllowedRole(role) {
			return nil, domain.ErrValidation
		}
		seen[role] = struct{}{}
		roles = append(roles, role)
	}

	return roles, nil
}

func validateScopes(input []string) ([]string, error) {
	scopes := make([]string, 0, len(input))
	seen := make(map[string]struct{}, len(input))
	for _, scope := range input {
		scope = strings.TrimSpace(scope)
		if scope == "" {
			continue
		}
		if strings.Contains(scope, " ") {
			return nil, domain.ErrValidation
		}
		if _, exists := seen[scope]; exists {
			continue
		}
		seen[scope] = struct{}{}
		scopes = append(scopes, scope)
	}

	return scopes, nil
}

func containsRole(roles []string, role string) bool {
	for _, value := range roles {
		if value == role {
			return true
		}
	}

	return false
}

func (s *AdminService) recordAudit(ctx context.Context, event domain.AuditEvent) {
	if s.audit == nil {
		return
	}
	if err := s.audit.Record(ctx, event); err != nil {
		s.logger.Warn(
			"failed to record admin audit event",
			"method",
			"AdminService.recordAudit",
			"action",
			event.Action,
			"result",
			event.Result,
			"error",
			err,
		)
	}
}
