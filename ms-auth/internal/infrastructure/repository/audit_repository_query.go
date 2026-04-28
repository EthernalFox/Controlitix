package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/EthernalFox/Controlitix/ms-auth/internal/domain"
)

type AuditListFilter struct {
	Offset int
	Limit  int
	Action string
	Actor  string
	From   *time.Time
	To     *time.Time
}

type AuditRecord struct {
	ID string `json:"id"`
	domain.AuditEvent
}

func (r *AuditRepository) List(
	ctx context.Context,
	filter AuditListFilter,
) ([]AuditRecord, int, error) {
	whereParts := make([]string, 0, 4)
	args := make([]any, 0, 8)
	countArgs := make([]any, 0, 6)

	if value := strings.TrimSpace(filter.Action); value != "" {
		whereParts = append(whereParts, "action = $"+fmt.Sprintf("%d", len(args)+1))
		args = append(args, value)
		countArgs = append(countArgs, value)
	}
	if value := strings.TrimSpace(filter.Actor); value != "" {
		whereParts = append(whereParts, "actor_subject = $"+fmt.Sprintf("%d", len(args)+1))
		args = append(args, value)
		countArgs = append(countArgs, value)
	}
	if filter.From != nil {
		whereParts = append(whereParts, "occurred_at >= $"+fmt.Sprintf("%d", len(args)+1))
		args = append(args, *filter.From)
		countArgs = append(countArgs, *filter.From)
	}
	if filter.To != nil {
		whereParts = append(whereParts, "occurred_at <= $"+fmt.Sprintf("%d", len(args)+1))
		args = append(args, *filter.To)
		countArgs = append(countArgs, *filter.To)
	}

	whereClause := ""
	if len(whereParts) > 0 {
		whereClause = " WHERE " + strings.Join(whereParts, " AND ")
	}

	countQuery := "SELECT count(*) FROM auth.audit_log" + whereClause
	total := 0
	if err := r.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count audit log records: %w", err)
	}

	listQuery := `
SELECT id, occurred_at, actor_subject, action, target, result, reason, ip::text, user_agent, metadata
FROM auth.audit_log` + whereClause + `
ORDER BY occurred_at DESC
OFFSET $` + fmt.Sprintf("%d", len(args)+1) + ` LIMIT $` + fmt.Sprintf("%d", len(args)+2)
	args = append(args, filter.Offset, filter.Limit)

	rows, err := r.db.QueryContext(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list audit log records: %w", err)
	}
	defer rows.Close()

	records := make([]AuditRecord, 0)
	for rows.Next() {
		record, scanErr := scanAuditRecord(rows)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("scan audit log record: %w", scanErr)
		}
		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate audit log records: %w", err)
	}

	return records, total, nil
}

func scanAuditRecord(scanner interface{ Scan(dest ...any) error }) (AuditRecord, error) {
	var (
		record       AuditRecord
		actor        sql.NullString
		target       sql.NullString
		reason       sql.NullString
		ip           sql.NullString
		userAgent    sql.NullString
		metadataRaw  []byte
	)

	if err := scanner.Scan(
		&record.ID,
		&record.OccurredAt,
		&actor,
		&record.Action,
		&target,
		&record.Result,
		&reason,
		&ip,
		&userAgent,
		&metadataRaw,
	); err != nil {
		return AuditRecord{}, err
	}

	if actor.Valid {
		record.ActorSubject = actor.String
	}
	if target.Valid {
		record.Target = target.String
	}
	if reason.Valid {
		record.Reason = reason.String
	}
	if ip.Valid {
		record.IP = ip.String
	}
	if userAgent.Valid {
		record.UserAgent = userAgent.String
	}

	record.Metadata = map[string]any{}
	if len(metadataRaw) > 0 {
		if err := json.Unmarshal(metadataRaw, &record.Metadata); err != nil {
			return AuditRecord{}, fmt.Errorf("unmarshal audit metadata: %w", err)
		}
	}

	return record, nil
}
