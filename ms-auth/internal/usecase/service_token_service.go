package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/EthernalFox/Controlitix/ms-auth/internal/domain"
	"github.com/EthernalFox/Controlitix/ms-auth/internal/infrastructure/password"
	"github.com/EthernalFox/Controlitix/ms-auth/internal/infrastructure/repository"
)

type serviceTokenAccountRepository interface {
	FindByClientID(ctx context.Context, clientID string) (*domain.ServiceAccount, error)
	UpdateLastUsed(ctx context.Context, id string, at time.Time) error
}

type serviceTokenIssuer interface {
	IssueForService(
		ctx context.Context,
		clientID string,
		scopes []string,
		meta TokenMeta,
	) (*domain.IssuedTokens, error)
}

type serviceTokenAudit interface {
	Record(ctx context.Context, event domain.AuditEvent) error
}

type ServiceTokenInput struct {
	GrantType    string
	ClientID     string
	ClientSecret string
	Scope        string
	UserAgent    string
	IP           string
}

type ServiceTokenResult struct {
	AccessToken string
	TokenType   string
	ExpiresIn   int
	Scope       string
}

type ServiceTokenService struct {
	accounts serviceTokenAccountRepository
	tokens   serviceTokenIssuer
	hasher   *password.Hasher
	audit    serviceTokenAudit
	logger   *slog.Logger
	clock    func() time.Time
}

func NewServiceTokenService(
	accounts *repository.ServiceAccountRepository,
	tokens *TokenService,
	hasher *password.Hasher,
	audit *AuditService,
	logger *slog.Logger,
) *ServiceTokenService {
	if logger == nil {
		logger = slog.Default()
	}

	return &ServiceTokenService{
		accounts: accounts,
		tokens:   tokens,
		hasher:   hasher,
		audit:    audit,
		logger:   logger,
		clock:    time.Now,
	}
}

func (s *ServiceTokenService) Issue(
	ctx context.Context,
	input ServiceTokenInput,
) (*ServiceTokenResult, error) {
	grantType := strings.TrimSpace(input.GrantType)
	if grantType != "client_credentials" {
		return nil, domain.ErrUnsupportedGrant
	}

	clientID := strings.TrimSpace(input.ClientID)
	clientSecret := strings.TrimSpace(input.ClientSecret)
	if clientID == "" || clientSecret == "" {
		return nil, domain.ErrValidation
	}

	account, err := s.accounts.FindByClientID(ctx, clientID)
	if err != nil {
		return nil, fmt.Errorf("find service account by client id: %w", err)
	}
	if account == nil {
		s.recordAudit(ctx, domain.AuditEvent{
			Action:    domain.AuditServiceTokenFailure,
			Target:    clientID,
			Result:    domain.AuditResultFailure,
			Reason:    "unknown_client",
			IP:        input.IP,
			UserAgent: input.UserAgent,
		})
		return nil, domain.ErrInvalidClient
	}

	verified, err := s.hasher.Verify(clientSecret, account.ClientSecretHash)
	if err != nil {
		return nil, fmt.Errorf("verify service account secret: %w", err)
	}
	if !verified {
		s.recordAudit(ctx, domain.AuditEvent{
			Action:    domain.AuditServiceTokenFailure,
			Target:    clientID,
			Result:    domain.AuditResultFailure,
			Reason:    "invalid_secret",
			IP:        input.IP,
			UserAgent: input.UserAgent,
		})
		return nil, domain.ErrInvalidClient
	}

	if !account.IsActive {
		s.recordAudit(ctx, domain.AuditEvent{
			Action:    domain.AuditServiceTokenFailure,
			Target:    clientID,
			Result:    domain.AuditResultFailure,
			Reason:    "client_disabled",
			IP:        input.IP,
			UserAgent: input.UserAgent,
		})
		return nil, domain.ErrClientDisabled
	}

	grantedScopes, err := resolveGrantedScopes(account.Scopes, input.Scope)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidScope) {
			s.recordAudit(ctx, domain.AuditEvent{
				Action:    domain.AuditServiceTokenFailure,
				Target:    clientID,
				Result:    domain.AuditResultFailure,
				Reason:    "invalid_scope",
				IP:        input.IP,
				UserAgent: input.UserAgent,
			})
		}
		return nil, err
	}

	issuedTokens, err := s.tokens.IssueForService(
		ctx,
		clientID,
		grantedScopes,
		TokenMeta{
			UserAgent: input.UserAgent,
			IP:        input.IP,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("issue service token: %w", err)
	}

	if err := s.accounts.UpdateLastUsed(ctx, account.ID, s.clock().UTC()); err != nil {
		return nil, fmt.Errorf("update service account last used: %w", err)
	}

	s.recordAudit(ctx, domain.AuditEvent{
		ActorSubject: "svc:" + clientID,
		Action:       domain.AuditServiceTokenSuccess,
		Target:       clientID,
		Result:       domain.AuditResultSuccess,
		IP:           input.IP,
		UserAgent:    input.UserAgent,
	})

	return &ServiceTokenResult{
		AccessToken: issuedTokens.AccessToken,
		TokenType:   "Bearer",
		ExpiresIn:   issuedTokens.ExpiresIn,
		Scope:       strings.Join(grantedScopes, " "),
	}, nil
}

func resolveGrantedScopes(
	allowedScopes []string,
	requestedScope string,
) ([]string, error) {
	allowed := make(map[string]struct{}, len(allowedScopes))
	normalizedAllowed := make([]string, 0, len(allowedScopes))
	for _, scope := range allowedScopes {
		scope = strings.TrimSpace(scope)
		if scope == "" {
			continue
		}
		if _, exists := allowed[scope]; exists {
			continue
		}
		allowed[scope] = struct{}{}
		normalizedAllowed = append(normalizedAllowed, scope)
	}

	requested := normalizeRequestedScopes(strings.Fields(requestedScope))
	if len(requested) == 0 {
		return normalizedAllowed, nil
	}

	for _, scope := range requested {
		if _, exists := allowed[scope]; !exists {
			return nil, domain.ErrInvalidScope
		}
	}

	return requested, nil
}

func normalizeRequestedScopes(scopes []string) []string {
	normalized := make([]string, 0, len(scopes))
	seen := make(map[string]struct{}, len(scopes))
	for _, scope := range scopes {
		scope = strings.TrimSpace(scope)
		if scope == "" {
			continue
		}
		if _, exists := seen[scope]; exists {
			continue
		}
		seen[scope] = struct{}{}
		normalized = append(normalized, scope)
	}

	return normalized
}

func (s *ServiceTokenService) recordAudit(ctx context.Context, event domain.AuditEvent) {
	if s.audit == nil {
		return
	}
	if err := s.audit.Record(ctx, event); err != nil {
		s.logger.Warn(
			"failed to record service token audit event",
			"method",
			"ServiceTokenService.recordAudit",
			"action",
			event.Action,
			"result",
			event.Result,
			"error",
			err,
		)
	}
}
