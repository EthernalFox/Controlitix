package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/EthernalFox/Controlitix/ms-auth/internal/domain"
)

type IdentitySourceRepository struct {
	db *sql.DB
}

func NewIdentitySourceRepository(db *sql.DB) *IdentitySourceRepository {
	return &IdentitySourceRepository{db: db}
}

func (r *IdentitySourceRepository) FindByID(ctx context.Context, id int) (*domain.IdentitySource, error) {
	const query = `
SELECT id, type, name, config, is_enabled
FROM auth.identity_sources
WHERE id = $1
`

	source, err := scanIdentitySource(r.db.QueryRowContext(ctx, query, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find identity source by id: %w", err)
	}

	return source, nil
}

func (r *IdentitySourceRepository) FindByType(ctx context.Context, sourceType string) ([]*domain.IdentitySource, error) {
	const query = `
SELECT id, type, name, config, is_enabled
FROM auth.identity_sources
WHERE type = $1
ORDER BY id
`

	rows, err := r.db.QueryContext(ctx, query, sourceType)
	if err != nil {
		return nil, fmt.Errorf("find identity sources by type: %w", err)
	}
	defer rows.Close()

	sources := make([]*domain.IdentitySource, 0)
	for rows.Next() {
		source, scanErr := scanIdentitySource(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scan identity source by type: %w", scanErr)
		}
		sources = append(sources, source)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate identity sources by type: %w", err)
	}

	return sources, nil
}

func (r *IdentitySourceRepository) ListEnabled(ctx context.Context) ([]*domain.IdentitySource, error) {
	const query = `
SELECT id, type, name, config, is_enabled
FROM auth.identity_sources
WHERE is_enabled = true
ORDER BY id
`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list enabled identity sources: %w", err)
	}
	defer rows.Close()

	sources := make([]*domain.IdentitySource, 0)
	for rows.Next() {
		source, scanErr := scanIdentitySource(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scan enabled identity source: %w", scanErr)
		}
		sources = append(sources, source)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate enabled identity sources: %w", err)
	}

	return sources, nil
}

func (r *IdentitySourceRepository) List(ctx context.Context) ([]*domain.IdentitySource, error) {
	const query = `
SELECT id, type, name, config, is_enabled
FROM auth.identity_sources
ORDER BY id
`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list identity sources: %w", err)
	}
	defer rows.Close()

	sources := make([]*domain.IdentitySource, 0)
	for rows.Next() {
		source, scanErr := scanIdentitySource(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scan identity source from list: %w", scanErr)
		}
		sources = append(sources, source)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate identity sources list: %w", err)
	}

	return sources, nil
}

func (r *IdentitySourceRepository) ResolveRoles(
	ctx context.Context,
	sourceID int,
	externalGroups []string,
) ([]string, error) {
	canonicalGroups := make([]string, 0, len(externalGroups))
	for _, group := range externalGroups {
		group = strings.ToLower(strings.TrimSpace(group))
		if group == "" {
			continue
		}
		canonicalGroups = append(canonicalGroups, group)
	}
	if len(canonicalGroups) == 0 {
		return []string{}, nil
	}

	const query = `
SELECT DISTINCT role_code
FROM auth.role_mappings
WHERE source_id = $1
  AND lower(external_group) = ANY($2::text[])
ORDER BY role_code
`

	rows, err := r.db.QueryContext(ctx, query, sourceID, canonicalGroups)
	if err != nil {
		return nil, fmt.Errorf("resolve role mappings: %w", err)
	}
	defer rows.Close()

	roles := make([]string, 0)
	for rows.Next() {
		var role string
		if scanErr := rows.Scan(&role); scanErr != nil {
			return nil, fmt.Errorf("scan resolved role: %w", scanErr)
		}
		roles = append(roles, role)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate resolved roles: %w", err)
	}

	return roles, nil
}

func scanIdentitySource(scanner interface{ Scan(dest ...any) error }) (*domain.IdentitySource, error) {
	var (
		source    domain.IdentitySource
		configRaw []byte
	)

	if err := scanner.Scan(
		&source.ID,
		&source.Type,
		&source.Name,
		&configRaw,
		&source.IsEnabled,
	); err != nil {
		return nil, err
	}

	if len(configRaw) == 0 {
		source.Config = map[string]any{}
		return &source, nil
	}

	source.Config = map[string]any{}
	if err := json.Unmarshal(configRaw, &source.Config); err != nil {
		return nil, fmt.Errorf("decode source config: %w", err)
	}

	return &source, nil
}
