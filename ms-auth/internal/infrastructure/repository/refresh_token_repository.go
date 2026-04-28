package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/EthernalFox/Controlitix/ms-auth/internal/domain"
)

type RefreshTokenRepository struct {
	db *sql.DB
}

func NewRefreshTokenRepository(db *sql.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) Create(ctx context.Context, rec *domain.RefreshRecord) error {
	if rec == nil {
		return errors.New("refresh record is nil")
	}

	const query = `
INSERT INTO auth.refresh_tokens (
	user_id,
	family_id,
	token_hash,
	parent_id,
	user_agent,
	ip,
	issued_at,
	expires_at
)
VALUES ($1, $2, $3, $4, $5, NULLIF($6, '')::inet, $7, $8)
`

	if _, err := r.db.ExecContext(
		ctx,
		query,
		rec.UserID,
		rec.FamilyID,
		rec.TokenHash,
		rec.ParentID,
		nullIfEmpty(rec.UserAgent),
		rec.IP,
		rec.IssuedAt,
		rec.ExpiresAt,
	); err != nil {
		return fmt.Errorf("create refresh token: %w", err)
	}

	return nil
}

func (r *RefreshTokenRepository) FindByHash(
	ctx context.Context,
	tokenHash string,
) (*domain.RefreshRecord, error) {
	const query = `
SELECT id, user_id, family_id, token_hash, parent_id, issued_at, expires_at, used_at, revoked_at, user_agent, ip::text
FROM auth.refresh_tokens
WHERE token_hash = $1
`

	record, err := scanRefreshRecord(r.db.QueryRowContext(ctx, query, tokenHash))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find refresh token by hash: %w", err)
	}

	return record, nil
}

func (r *RefreshTokenRepository) MarkUsed(
	ctx context.Context,
	id string,
	at time.Time,
) error {
	const query = `
UPDATE auth.refresh_tokens
SET used_at = $2
WHERE id = $1
`

	if _, err := r.db.ExecContext(ctx, query, id, at); err != nil {
		return fmt.Errorf("mark refresh token used: %w", err)
	}

	return nil
}

func (r *RefreshTokenRepository) RevokeFamily(
	ctx context.Context,
	familyID string,
	at time.Time,
	reason string,
) error {
	const query = `
UPDATE auth.refresh_tokens
SET revoked_at = $1, revoke_reason = $2
WHERE family_id = $3 AND revoked_at IS NULL
`

	if _, err := r.db.ExecContext(ctx, query, at, reason, familyID); err != nil {
		return fmt.Errorf("revoke refresh token family: %w", err)
	}

	return nil
}

func (r *RefreshTokenRepository) RevokeAllForUser(
	ctx context.Context,
	userID string,
	at time.Time,
	reason string,
) error {
	const query = `
UPDATE auth.refresh_tokens
SET revoked_at = $1, revoke_reason = $2
WHERE user_id = $3 AND revoked_at IS NULL
`

	if _, err := r.db.ExecContext(ctx, query, at, reason, userID); err != nil {
		return fmt.Errorf("revoke refresh tokens for user: %w", err)
	}

	return nil
}

func (r *RefreshTokenRepository) DeleteExpired(ctx context.Context, now time.Time) (int64, error) {
	const query = `
DELETE FROM auth.refresh_tokens
WHERE expires_at < $1
`

	result, err := r.db.ExecContext(ctx, query, now)
	if err != nil {
		return 0, fmt.Errorf("delete expired refresh tokens: %w", err)
	}

	affectedRows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("read deleted refresh tokens count: %w", err)
	}

	return affectedRows, nil
}

func scanRefreshRecord(scanner interface{ Scan(dest ...any) error }) (*domain.RefreshRecord, error) {
	var (
		record      domain.RefreshRecord
		parentID    sql.NullString
		usedAt      sql.NullTime
		revokedAt   sql.NullTime
		userAgent   sql.NullString
		ip          sql.NullString
	)

	if err := scanner.Scan(
		&record.ID,
		&record.UserID,
		&record.FamilyID,
		&record.TokenHash,
		&parentID,
		&record.IssuedAt,
		&record.ExpiresAt,
		&usedAt,
		&revokedAt,
		&userAgent,
		&ip,
	); err != nil {
		return nil, err
	}

	if parentID.Valid {
		record.ParentID = &parentID.String
	}
	if usedAt.Valid {
		used := usedAt.Time
		record.UsedAt = &used
	}
	if revokedAt.Valid {
		revoked := revokedAt.Time
		record.RevokedAt = &revoked
	}
	if userAgent.Valid {
		record.UserAgent = userAgent.String
	}
	if ip.Valid {
		record.IP = ip.String
	}

	return &record, nil
}

func nullIfEmpty(value string) any {
	if value == "" {
		return nil
	}

	return value
}
