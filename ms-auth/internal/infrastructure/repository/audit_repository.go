package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/EthernalFox/Controlitix/ms-auth/internal/domain"
)

type AuditRepository struct {
	db *sql.DB
}

func NewAuditRepository(db *sql.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) Insert(ctx context.Context, event domain.AuditEvent) error {
	metadata := event.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}

	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("marshal audit metadata: %w", err)
	}

	const query = `
INSERT INTO auth.audit_log (
	occurred_at,
	actor_subject,
	action,
	target,
	result,
	reason,
	ip,
	user_agent,
	metadata
)
VALUES (
	$1,
	NULLIF($2, ''),
	$3,
	NULLIF($4, ''),
	$5,
	NULLIF($6, ''),
	NULLIF($7, '')::inet,
	NULLIF($8, ''),
	$9::jsonb
)
`

	if _, err := r.db.ExecContext(
		ctx,
		query,
		event.OccurredAt,
		event.ActorSubject,
		event.Action,
		event.Target,
		event.Result,
		event.Reason,
		event.IP,
		event.UserAgent,
		metadataBytes,
	); err != nil {
		return fmt.Errorf("insert audit event: %w", err)
	}

	return nil
}
