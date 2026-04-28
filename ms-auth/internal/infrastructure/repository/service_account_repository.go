package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/EthernalFox/Controlitix/ms-auth/internal/domain"
	"github.com/jackc/pgx/v5/pgconn"
)

type ServiceAccountRepository struct {
	db *sql.DB
}

func NewServiceAccountRepository(db *sql.DB) *ServiceAccountRepository {
	return &ServiceAccountRepository{db: db}
}

func (r *ServiceAccountRepository) FindByClientID(
	ctx context.Context,
	clientID string,
) (*domain.ServiceAccount, error) {
	const query = `
SELECT id, client_id, client_secret_hash, display_name, scopes, is_active, created_at, updated_at, last_used_at
FROM auth.service_accounts
WHERE client_id = $1
`

	account, err := scanServiceAccount(r.db.QueryRowContext(ctx, query, clientID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find service account by client id: %w", err)
	}

	return account, nil
}

func (r *ServiceAccountRepository) FindByID(
	ctx context.Context,
	id string,
) (*domain.ServiceAccount, error) {
	const query = `
SELECT id, client_id, client_secret_hash, display_name, scopes, is_active, created_at, updated_at, last_used_at
FROM auth.service_accounts
WHERE id = $1
`

	account, err := scanServiceAccount(r.db.QueryRowContext(ctx, query, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find service account by id: %w", err)
	}

	return account, nil
}

func (r *ServiceAccountRepository) List(
	ctx context.Context,
	offset int,
	limit int,
) ([]*domain.ServiceAccount, int, error) {
	const countQuery = `SELECT count(*) FROM auth.service_accounts`

	total := 0
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count service accounts: %w", err)
	}

	const listQuery = `
SELECT id, client_id, client_secret_hash, display_name, scopes, is_active, created_at, updated_at, last_used_at
FROM auth.service_accounts
ORDER BY created_at DESC
OFFSET $1 LIMIT $2
`

	rows, err := r.db.QueryContext(ctx, listQuery, offset, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("list service accounts: %w", err)
	}
	defer rows.Close()

	accounts := make([]*domain.ServiceAccount, 0)
	for rows.Next() {
		account, scanErr := scanServiceAccount(rows)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("scan service account: %w", scanErr)
		}
		accounts = append(accounts, account)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate service accounts: %w", err)
	}

	return accounts, total, nil
}

func (r *ServiceAccountRepository) Create(
	ctx context.Context,
	account *domain.ServiceAccount,
) (*domain.ServiceAccount, error) {
	if account == nil {
		return nil, errors.New("service account is nil")
	}

	const query = `
INSERT INTO auth.service_accounts (client_id, client_secret_hash, display_name, scopes, is_active)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, created_at, updated_at
`

	created := *account
	created.Scopes = normalizeScopes(account.Scopes)
	if err := r.db.QueryRowContext(
		ctx,
		query,
		created.ClientID,
		created.ClientSecretHash,
		created.DisplayName,
		created.Scopes,
		created.IsActive,
	).Scan(&created.ID, &created.CreatedAt, &created.UpdatedAt); err != nil {
		if isUniqueViolation(err) {
			return nil, domain.ErrClientIDExists
		}
		return nil, fmt.Errorf("create service account: %w", err)
	}

	return &created, nil
}

func (r *ServiceAccountRepository) Update(
	ctx context.Context,
	account *domain.ServiceAccount,
) error {
	if account == nil {
		return errors.New("service account is nil")
	}

	const query = `
UPDATE auth.service_accounts
SET display_name = $2,
    scopes = $3,
    is_active = $4,
    updated_at = now()
WHERE id = $1
`

	result, err := r.db.ExecContext(
		ctx,
		query,
		account.ID,
		account.DisplayName,
		normalizeScopes(account.Scopes),
		account.IsActive,
	)
	if err != nil {
		return fmt.Errorf("update service account: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read affected rows for service account update: %w", err)
	}
	if rowsAffected == 0 {
		return domain.ErrServiceNotFound
	}

	return nil
}

func (r *ServiceAccountRepository) UpdateSecretHash(
	ctx context.Context,
	id string,
	hash string,
) error {
	const query = `
UPDATE auth.service_accounts
SET client_secret_hash = $2,
    updated_at = now()
WHERE id = $1
`

	result, err := r.db.ExecContext(ctx, query, id, hash)
	if err != nil {
		return fmt.Errorf("update service account secret hash: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read affected rows for service account secret update: %w", err)
	}
	if rowsAffected == 0 {
		return domain.ErrServiceNotFound
	}

	return nil
}

func (r *ServiceAccountRepository) Delete(ctx context.Context, id string) error {
	const query = `DELETE FROM auth.service_accounts WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete service account: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read affected rows for service account delete: %w", err)
	}
	if rowsAffected == 0 {
		return domain.ErrServiceNotFound
	}

	return nil
}

func (r *ServiceAccountRepository) UpdateLastUsed(
	ctx context.Context,
	id string,
	at time.Time,
) error {
	const query = `
UPDATE auth.service_accounts
SET last_used_at = $2,
    updated_at = now()
WHERE id = $1
`

	if _, err := r.db.ExecContext(ctx, query, id, at); err != nil {
		return fmt.Errorf("update service account last used: %w", err)
	}

	return nil
}

func scanServiceAccount(scanner interface{ Scan(dest ...any) error }) (*domain.ServiceAccount, error) {
	var (
		account    domain.ServiceAccount
		lastUsedAt sql.NullTime
		scopes     []string
	)

	if err := scanner.Scan(
		&account.ID,
		&account.ClientID,
		&account.ClientSecretHash,
		&account.DisplayName,
		&scopes,
		&account.IsActive,
		&account.CreatedAt,
		&account.UpdatedAt,
		&lastUsedAt,
	); err != nil {
		return nil, err
	}

	account.Scopes = normalizeScopes(scopes)
	if lastUsedAt.Valid {
		usedAt := lastUsedAt.Time
		account.LastUsedAt = &usedAt
	}

	return &account, nil
}

func normalizeScopes(scopes []string) []string {
	normalized := make([]string, 0, len(scopes))
	seen := make(map[string]struct{}, len(scopes))
	for _, scope := range scopes {
		scope = strings.TrimSpace(scope)
		if scope == "" {
			continue
		}
		if strings.Contains(scope, " ") {
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

func isUniqueViolation(err error) bool {
	var pgError *pgconn.PgError
	return errors.As(err, &pgError) && pgError.Code == "23505"
}
