package repository

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
)

type NotificationRepository struct {
	pool *pgxpool.Pool
}

func NewNotificationRepository(pool *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{pool: pool}
}

func (repository *NotificationRepository) CreateNotification(
	ctx context.Context,
	alarmEventID uuid.UUID,
	chatID int64,
) (bool, error) {
	const query = `
INSERT INTO alarms.notifications (alarm_event_id, chat_id, status, attempt)
VALUES ($1, $2, $3, 0)
ON CONFLICT (alarm_event_id, chat_id) DO NOTHING
RETURNING id
`

	var notificationID uuid.UUID
	scanError := repository.pool.QueryRow(
		ctx,
		query,
		alarmEventID,
		chatID,
		int16(domain.NotificationStatusPending),
	).Scan(&notificationID)
	if scanError != nil {
		if errors.Is(scanError, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("insert notification: %w", scanError)
	}

	return true, nil
}

func (repository *NotificationRepository) MarkNotificationSent(
	ctx context.Context,
	alarmEventID uuid.UUID,
	chatID int64,
) error {
	const query = `
UPDATE alarms.notifications
SET status = $3,
    attempt = attempt + 1,
    sent_at = now(),
    updated_at = now(),
    last_error = NULL
WHERE alarm_event_id = $1
  AND chat_id = $2
`

	if _, execError := repository.pool.Exec(
		ctx,
		query,
		alarmEventID,
		chatID,
		int16(domain.NotificationStatusSent),
	); execError != nil {
		return fmt.Errorf("mark notification sent: %w", execError)
	}

	return nil
}

func (repository *NotificationRepository) MarkNotificationPendingRetry(
	ctx context.Context,
	alarmEventID uuid.UUID,
	chatID int64,
	lastError string,
	retryAfter *time.Duration,
) error {
	updatedAt := time.Now().UTC()
	if retryAfter != nil && *retryAfter > 0 {
		updatedAt = updatedAt.Add(*retryAfter)
	}

	const query = `
UPDATE alarms.notifications
SET status = $3,
    attempt = attempt + 1,
    last_error = $4,
    sent_at = NULL,
    updated_at = $5
WHERE alarm_event_id = $1
  AND chat_id = $2
`

	if _, execError := repository.pool.Exec(
		ctx,
		query,
		alarmEventID,
		chatID,
		int16(domain.NotificationStatusPending),
		lastError,
		updatedAt,
	); execError != nil {
		return fmt.Errorf("mark notification pending retry: %w", execError)
	}

	return nil
}

func (repository *NotificationRepository) MarkNotificationFailed(
	ctx context.Context,
	alarmEventID uuid.UUID,
	chatID int64,
	lastError string,
) error {
	const query = `
UPDATE alarms.notifications
SET status = $3,
    attempt = attempt + 1,
    last_error = $4,
    updated_at = now()
WHERE alarm_event_id = $1
  AND chat_id = $2
`

	if _, execError := repository.pool.Exec(
		ctx,
		query,
		alarmEventID,
		chatID,
		int16(domain.NotificationStatusFailed),
		lastError,
	); execError != nil {
		return fmt.Errorf("mark notification failed: %w", execError)
	}

	return nil
}

func (repository *NotificationRepository) MarkNotificationSkipped(
	ctx context.Context,
	alarmEventID uuid.UUID,
	chatID int64,
	reason string,
) error {
	const query = `
UPDATE alarms.notifications
SET status = $3,
    last_error = $4,
    sent_at = NULL,
    updated_at = now()
WHERE alarm_event_id = $1
  AND chat_id = $2
`

	if _, execError := repository.pool.Exec(
		ctx,
		query,
		alarmEventID,
		chatID,
		int16(domain.NotificationStatusSkipped),
		reason,
	); execError != nil {
		return fmt.Errorf("mark notification skipped: %w", execError)
	}

	return nil
}

func (repository *NotificationRepository) ListPendingNotifications(
	ctx context.Context,
	maxAttempts int,
	backoffBase time.Duration,
	limit int,
) ([]domain.PendingNotification, error) {
	if backoffBase <= 0 {
		backoffBase = 60 * time.Second
	}
	if limit <= 0 {
		limit = 100
	}

	backoffSeconds := int64(math.Max(1, backoffBase.Seconds()))

	const query = `
SELECT
    n.id,
    n.alarm_event_id,
    n.chat_id,
    n.attempt,
    n.updated_at,
    n.last_error,
    e.id,
    e.tag_id,
    e.event_type,
    e.state_from,
    e.state_to,
    e.value,
    e.quality,
    e.ts,
    e.actor_id,
    e.note,
    e.created_at
FROM alarms.notifications n
JOIN alarms.events e ON e.id = n.alarm_event_id
WHERE n.status = $1
  AND n.attempt < $2
  AND n.updated_at <= now() - (
    ($3::bigint * (1::bigint << GREATEST(n.attempt - 1, 0)::int)) * interval '1 second'
  )
ORDER BY n.updated_at ASC, n.id ASC
LIMIT $4
`

	rows, queryError := repository.pool.Query(
		ctx,
		query,
		int16(domain.NotificationStatusPending),
		maxAttempts,
		backoffSeconds,
		limit,
	)
	if queryError != nil {
		return nil, fmt.Errorf("query pending notifications: %w", queryError)
	}
	defer rows.Close()

	items := make([]domain.PendingNotification, 0)
	for rows.Next() {
		var (
			item        domain.PendingNotification
			eventType   int16
			stateFrom   int16
			stateTo     int16
			qualityCode int16
		)

		if scanError := rows.Scan(
			&item.NotificationID,
			&item.AlarmEventID,
			&item.ChatID,
			&item.Attempt,
			&item.UpdatedAt,
			&item.LastError,
			&item.Event.ID,
			&item.Event.TagID,
			&eventType,
			&stateFrom,
			&stateTo,
			&item.Event.Value,
			&qualityCode,
			&item.Event.TS,
			&item.Event.ActorID,
			&item.Event.Note,
			&item.Event.CreatedAt,
		); scanError != nil {
			return nil, fmt.Errorf("scan pending notification: %w", scanError)
		}

		item.Event.EventType = domain.AlarmEventType(eventType)
		item.Event.StateFrom = domain.AlarmState(stateFrom)
		item.Event.StateTo = domain.AlarmState(stateTo)
		item.Event.Quality = domain.QualityFromCode(qualityCode)
		items = append(items, item)
	}
	if rowsError := rows.Err(); rowsError != nil {
		return nil, fmt.Errorf("iterate pending notifications: %w", rowsError)
	}

	return items, nil
}

func (repository *NotificationRepository) CreateEscalation(
	ctx context.Context,
	alarmEventID uuid.UUID,
	tagID uuid.UUID,
	state domain.AlarmState,
	fireAt time.Time,
) error {
	const query = `
INSERT INTO alarms.escalations (alarm_event_id, tag_id, state, fire_at)
SELECT $1, $2, $3, $4
WHERE NOT EXISTS (
    SELECT 1
    FROM alarms.escalations
    WHERE alarm_event_id = $1
)
`

	if _, execError := repository.pool.Exec(
		ctx,
		query,
		alarmEventID,
		tagID,
		int16(state),
		fireAt.UTC(),
	); execError != nil {
		return fmt.Errorf("insert escalation: %w", execError)
	}

	return nil
}

func (repository *NotificationRepository) ListDueEscalations(
	ctx context.Context,
	limit int,
) ([]domain.EscalationRecord, error) {
	if limit <= 0 {
		limit = 100
	}

	const query = `
SELECT id, alarm_event_id, tag_id, state, fire_at, processed, skipped, created_at
FROM alarms.escalations
WHERE processed = false
  AND fire_at <= now()
ORDER BY fire_at ASC, id ASC
LIMIT $1
`

	rows, queryError := repository.pool.Query(ctx, query, limit)
	if queryError != nil {
		return nil, fmt.Errorf("query due escalations: %w", queryError)
	}
	defer rows.Close()

	items := make([]domain.EscalationRecord, 0)
	for rows.Next() {
		var (
			record    domain.EscalationRecord
			stateCode int16
		)
		if scanError := rows.Scan(
			&record.ID,
			&record.AlarmEventID,
			&record.TagID,
			&stateCode,
			&record.FireAt,
			&record.Processed,
			&record.Skipped,
			&record.CreatedAt,
		); scanError != nil {
			return nil, fmt.Errorf("scan escalation: %w", scanError)
		}
		record.State = domain.AlarmState(stateCode)
		items = append(items, record)
	}
	if rowsError := rows.Err(); rowsError != nil {
		return nil, fmt.Errorf("iterate escalations: %w", rowsError)
	}

	return items, nil
}

func (repository *NotificationRepository) MarkEscalationProcessed(
	ctx context.Context,
	escalationID uuid.UUID,
	skipped bool,
) error {
	const query = `
UPDATE alarms.escalations
SET processed = true,
    skipped = $2
WHERE id = $1
`

	if _, execError := repository.pool.Exec(ctx, query, escalationID, skipped); execError != nil {
		return fmt.Errorf("mark escalation processed: %w", execError)
	}

	return nil
}

func (repository *NotificationRepository) IsTagAcknowledgedForState(
	ctx context.Context,
	tagID uuid.UUID,
	state domain.AlarmState,
) (bool, error) {
	const query = `
SELECT 1
FROM alarms.acks
WHERE tag_id = $1
  AND state >= $2
`

	var marker int
	scanError := repository.pool.QueryRow(ctx, query, tagID, int16(state)).Scan(&marker)
	if scanError != nil {
		if errors.Is(scanError, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("check ack for state: %w", scanError)
	}

	return true, nil
}

func (repository *NotificationRepository) GetAlarmEventByID(
	ctx context.Context,
	alarmEventID uuid.UUID,
) (domain.AlarmEvent, error) {
	const query = `
SELECT id, tag_id, event_type, state_from, state_to, value, quality, ts, actor_id, note, created_at
FROM alarms.events
WHERE id = $1
`

	var (
		event       domain.AlarmEvent
		eventType   int16
		stateFrom   int16
		stateTo     int16
		qualityCode int16
	)

	scanError := repository.pool.QueryRow(ctx, query, alarmEventID).Scan(
		&event.ID,
		&event.TagID,
		&eventType,
		&stateFrom,
		&stateTo,
		&event.Value,
		&qualityCode,
		&event.TS,
		&event.ActorID,
		&event.Note,
		&event.CreatedAt,
	)
	if scanError != nil {
		if errors.Is(scanError, pgx.ErrNoRows) {
			return domain.AlarmEvent{}, fmt.Errorf("alarm event %s not found: %w", alarmEventID.String(), domain.ErrNotFound)
		}
		return domain.AlarmEvent{}, fmt.Errorf("query alarm event: %w", scanError)
	}

	event.EventType = domain.AlarmEventType(eventType)
	event.StateFrom = domain.AlarmState(stateFrom)
	event.StateTo = domain.AlarmState(stateTo)
	event.Quality = domain.QualityFromCode(qualityCode)
	return event, nil
}

func (repository *NotificationRepository) WithSchedulerLock(
	ctx context.Context,
	lockKey int64,
	callback func(context.Context) error,
) (bool, error) {
	if callback == nil {
		return false, nil
	}

	var acquired bool
	scanError := repository.pool.QueryRow(
		ctx,
		`SELECT pg_try_advisory_lock($1)`,
		lockKey,
	).Scan(&acquired)
	if scanError != nil {
		return false, fmt.Errorf("try advisory lock: %w", scanError)
	}
	if !acquired {
		return false, nil
	}

	defer func() {
		_, _ = repository.pool.Exec(context.Background(), `SELECT pg_advisory_unlock($1)`, lockKey)
	}()

	if callbackError := callback(ctx); callbackError != nil {
		return true, callbackError
	}

	return true, nil
}
