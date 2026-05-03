package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
)

type AlarmRepository struct {
	pool *pgxpool.Pool
}

func NewAlarmRepository(pool *pgxpool.Pool) *AlarmRepository {
	return &AlarmRepository{pool: pool}
}

func (repository *AlarmRepository) LoadSetpoints(ctx context.Context) (map[uuid.UUID]domain.Setpoints, error) {
	const query = `
SELECT tp.tag_id, sp.lolo, sp.lo, sp.hi, sp.hihi
FROM tags.tag_params tp
LEFT JOIN tags.tag_setpoints sp ON sp.param_id = tp.id
JOIN tags.tags t ON t.id = tp.tag_id
WHERE t.deleted_at IS NULL
`

	rows, queryError := repository.pool.Query(ctx, query)
	if queryError != nil {
		return nil, fmt.Errorf("query setpoints: %w", queryError)
	}
	defer rows.Close()

	setpointsByTag := make(map[uuid.UUID]domain.Setpoints)
	for rows.Next() {
		var (
			tagID     uuid.UUID
			setpoints domain.Setpoints
		)
		if scanError := rows.Scan(
			&tagID,
			&setpoints.LoLo,
			&setpoints.Lo,
			&setpoints.Hi,
			&setpoints.HiHi,
		); scanError != nil {
			return nil, fmt.Errorf("scan setpoints: %w", scanError)
		}

		setpointsByTag[tagID] = setpoints
	}
	if rowsError := rows.Err(); rowsError != nil {
		return nil, fmt.Errorf("iterate setpoints: %w", rowsError)
	}

	return setpointsByTag, nil
}

func (repository *AlarmRepository) LoadSetpoint(
	ctx context.Context,
	tagID uuid.UUID,
) (domain.Setpoints, bool, error) {
	const query = `
SELECT sp.lolo, sp.lo, sp.hi, sp.hihi
FROM tags.tag_params tp
LEFT JOIN tags.tag_setpoints sp ON sp.param_id = tp.id
JOIN tags.tags t ON t.id = tp.tag_id
WHERE tp.tag_id = $1
  AND t.deleted_at IS NULL
`

	var setpoints domain.Setpoints
	scanError := repository.pool.QueryRow(ctx, query, tagID).Scan(
		&setpoints.LoLo,
		&setpoints.Lo,
		&setpoints.Hi,
		&setpoints.HiHi,
	)
	if scanError != nil {
		if errors.Is(scanError, pgx.ErrNoRows) {
			return domain.Setpoints{}, false, nil
		}
		return domain.Setpoints{}, false, fmt.Errorf("query setpoint by tag: %w", scanError)
	}

	return setpoints, true, nil
}

func (repository *AlarmRepository) GetState(
	ctx context.Context,
	tagID uuid.UUID,
) (domain.AlarmStateRecord, bool, error) {
	const query = `
SELECT
    s.tag_id,
    s.state,
    s.last_value,
    s.last_quality,
    s.entered_at,
    s.last_seen_at,
    s.suppressed,
    s.updated_at
FROM alarms.states s
WHERE s.tag_id = $1
`

	record := domain.AlarmStateRecord{}
	var stateCode int16
	var qualityCode int16
	scanError := repository.pool.QueryRow(ctx, query, tagID).Scan(
		&record.TagID,
		&stateCode,
		&record.LastValue,
		&qualityCode,
		&record.EnteredAt,
		&record.LastSeenAt,
		&record.Suppressed,
		&record.UpdatedAt,
	)
	if scanError != nil {
		if errors.Is(scanError, pgx.ErrNoRows) {
			return domain.AlarmStateRecord{}, false, nil
		}
		return domain.AlarmStateRecord{}, false, fmt.Errorf("query alarm state: %w", scanError)
	}

	record.State = domain.AlarmState(stateCode)
	record.LastQuality = domain.QualityFromCode(qualityCode)
	return record, true, nil
}

func (repository *AlarmRepository) UpsertState(ctx context.Context, record domain.AlarmStateRecord) error {
	const query = `
INSERT INTO alarms.states (
    tag_id, state, last_value, last_quality, entered_at, last_seen_at, suppressed, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, now())
ON CONFLICT (tag_id) DO UPDATE SET
    state = EXCLUDED.state,
    last_value = EXCLUDED.last_value,
    last_quality = EXCLUDED.last_quality,
    entered_at = EXCLUDED.entered_at,
    last_seen_at = EXCLUDED.last_seen_at,
    suppressed = EXCLUDED.suppressed,
    updated_at = now()
`

	_, execError := repository.pool.Exec(
		ctx,
		query,
		record.TagID,
		int16(record.State),
		record.LastValue,
		record.LastQuality.ToCode(),
		record.EnteredAt.UTC(),
		record.LastSeenAt.UTC(),
		record.Suppressed,
	)
	if execError != nil {
		return fmt.Errorf("upsert alarm state: %w", execError)
	}

	return nil
}

func (repository *AlarmRepository) ApplyTransition(
	ctx context.Context,
	input domain.AlarmTransitionInput,
) (domain.AlarmEvent, error) {
	transaction, beginError := repository.pool.Begin(ctx)
	if beginError != nil {
		return domain.AlarmEvent{}, fmt.Errorf("begin alarm transition tx: %w", beginError)
	}
	defer func() {
		_ = transaction.Rollback(ctx)
	}()

	const upsertStateQuery = `
INSERT INTO alarms.states (
    tag_id, state, last_value, last_quality, entered_at, last_seen_at, suppressed, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, false, now())
ON CONFLICT (tag_id) DO UPDATE SET
    state = EXCLUDED.state,
    last_value = EXCLUDED.last_value,
    last_quality = EXCLUDED.last_quality,
    entered_at = EXCLUDED.entered_at,
    last_seen_at = EXCLUDED.last_seen_at,
    suppressed = EXCLUDED.suppressed,
    updated_at = now()
`

	_, upsertError := transaction.Exec(
		ctx,
		upsertStateQuery,
		input.TagID,
		int16(input.StateTo),
		input.Value,
		input.Quality.ToCode(),
		input.EnteredAt.UTC(),
		input.LastSeenAt.UTC(),
	)
	if upsertError != nil {
		return domain.AlarmEvent{}, fmt.Errorf("upsert transition state: %w", upsertError)
	}

	if input.StateTo == domain.AlarmStateOK {
		if _, deleteAckError := transaction.Exec(ctx, `DELETE FROM alarms.acks WHERE tag_id = $1`, input.TagID); deleteAckError != nil {
			return domain.AlarmEvent{}, fmt.Errorf("delete alarm ack on clear: %w", deleteAckError)
		}
	}

	eventType := domain.AlarmEventRaised
	if input.StateTo == domain.AlarmStateOK {
		eventType = domain.AlarmEventCleared
	}

	const insertEventQuery = `
INSERT INTO alarms.events (
    tag_id, event_type, state_from, state_to, value, quality, ts, actor_id, note
)
VALUES ($1, $2, $3, $4, $5, $6, $7, NULL, NULL)
RETURNING id, created_at
`

	event := domain.AlarmEvent{
		TagID:     input.TagID,
		EventType: eventType,
		StateFrom: input.StateFrom,
		StateTo:   input.StateTo,
		Value:     input.Value,
		Quality:   input.Quality,
		TS:        input.TS.UTC(),
	}

	if scanError := transaction.QueryRow(
		ctx,
		insertEventQuery,
		event.TagID,
		int16(event.EventType),
		int16(event.StateFrom),
		int16(event.StateTo),
		event.Value,
		event.Quality.ToCode(),
		event.TS,
	).Scan(&event.ID, &event.CreatedAt); scanError != nil {
		return domain.AlarmEvent{}, fmt.Errorf("insert alarm transition event: %w", scanError)
	}

	if commitError := transaction.Commit(ctx); commitError != nil {
		return domain.AlarmEvent{}, fmt.Errorf("commit alarm transition tx: %w", commitError)
	}

	return event, nil
}

func (repository *AlarmRepository) Acknowledge(
	ctx context.Context,
	tagID uuid.UUID,
	actorID string,
	note *string,
	ts time.Time,
) (domain.AlarmAcknowledgeResult, error) {
	transaction, beginError := repository.pool.Begin(ctx)
	if beginError != nil {
		return domain.AlarmAcknowledgeResult{}, fmt.Errorf("begin alarm acknowledge tx: %w", beginError)
	}
	defer func() {
		_ = transaction.Rollback(ctx)
	}()

	const lockStateQuery = `
SELECT s.state, a.state
FROM alarms.states s
LEFT JOIN alarms.acks a ON a.tag_id = s.tag_id
WHERE s.tag_id = $1
FOR UPDATE
`

	var (
		stateCode    int16
		ackStateCode *int16
	)
	scanError := transaction.QueryRow(ctx, lockStateQuery, tagID).Scan(&stateCode, &ackStateCode)
	if scanError != nil {
		if errors.Is(scanError, pgx.ErrNoRows) {
			return domain.AlarmAcknowledgeResult{}, fmt.Errorf("alarm %s not found: %w", tagID.String(), domain.ErrNotFound)
		}
		return domain.AlarmAcknowledgeResult{}, fmt.Errorf("lock alarm state: %w", scanError)
	}

	currentState := domain.AlarmState(stateCode)
	if currentState == domain.AlarmStateOK {
		return domain.AlarmAcknowledgeResult{}, domain.ErrAlarmNotActive
	}
	if ackStateCode != nil && domain.AlarmState(*ackStateCode) == currentState {
		return domain.AlarmAcknowledgeResult{}, domain.ErrAlarmAlreadyAcked
	}

	const upsertAckQuery = `
INSERT INTO alarms.acks (tag_id, state, actor_id, note, acked_at)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (tag_id) DO UPDATE SET
    state = EXCLUDED.state,
    actor_id = EXCLUDED.actor_id,
    note = EXCLUDED.note,
    acked_at = EXCLUDED.acked_at
`

	_, upsertAckError := transaction.Exec(
		ctx,
		upsertAckQuery,
		tagID,
		int16(currentState),
		strings.TrimSpace(actorID),
		note,
		ts.UTC(),
	)
	if upsertAckError != nil {
		return domain.AlarmAcknowledgeResult{}, fmt.Errorf("upsert alarm ack: %w", upsertAckError)
	}

	const insertEventQuery = `
INSERT INTO alarms.events (
    tag_id, event_type, state_from, state_to, value, quality, ts, actor_id, note
)
VALUES (
    $1, $2, $3, $4,
    (SELECT last_value FROM alarms.states WHERE tag_id = $1),
    (SELECT last_quality FROM alarms.states WHERE tag_id = $1),
    $5, $6, $7
)
RETURNING id, value, quality, created_at
`

	event := domain.AlarmEvent{
		TagID:     tagID,
		EventType: domain.AlarmEventAcked,
		StateFrom: currentState,
		StateTo:   currentState,
		TS:        ts.UTC(),
	}

	var (
		qualityCode int16
		actorIDCopy = strings.TrimSpace(actorID)
	)
	if scanEventError := transaction.QueryRow(
		ctx,
		insertEventQuery,
		tagID,
		int16(event.EventType),
		int16(currentState),
		int16(currentState),
		event.TS,
		actorIDCopy,
		note,
	).Scan(&event.ID, &event.Value, &qualityCode, &event.CreatedAt); scanEventError != nil {
		return domain.AlarmAcknowledgeResult{}, fmt.Errorf("insert alarm ack event: %w", scanEventError)
	}
	event.Quality = domain.QualityFromCode(qualityCode)
	event.ActorID = &actorIDCopy
	event.Note = note

	if commitError := transaction.Commit(ctx); commitError != nil {
		return domain.AlarmAcknowledgeResult{}, fmt.Errorf("commit alarm acknowledge tx: %w", commitError)
	}

	return domain.AlarmAcknowledgeResult{
		TagID: tagID,
		State: currentState,
		Ack: domain.AlarmAck{
			State:   currentState,
			ActorID: actorIDCopy,
			Note:    note,
			AckedAt: event.TS,
		},
		Event: event,
	}, nil
}

func (repository *AlarmRepository) AcknowledgeBulk(
	ctx context.Context,
	tagIDs []uuid.UUID,
	actorID string,
	note *string,
	ts time.Time,
) (domain.AlarmBulkAcknowledgeResult, error) {
	normalizedActorID := strings.TrimSpace(actorID)
	transaction, beginError := repository.pool.Begin(ctx)
	if beginError != nil {
		return domain.AlarmBulkAcknowledgeResult{}, fmt.Errorf("begin alarm bulk acknowledge tx: %w", beginError)
	}
	defer func() {
		_ = transaction.Rollback(ctx)
	}()

	type stateRow struct {
		state    domain.AlarmState
		ackState *domain.AlarmState
	}

	const lockStatesQuery = `
SELECT s.tag_id, s.state, a.state
FROM alarms.states s
LEFT JOIN alarms.acks a ON a.tag_id = s.tag_id
WHERE s.tag_id = ANY($1)
FOR UPDATE
`

	rows, queryError := transaction.Query(ctx, lockStatesQuery, tagIDs)
	if queryError != nil {
		return domain.AlarmBulkAcknowledgeResult{}, fmt.Errorf("lock bulk alarm states: %w", queryError)
	}
	defer rows.Close()

	stateByTagID := make(map[uuid.UUID]stateRow, len(tagIDs))
	for rows.Next() {
		var (
			tagIDValue    uuid.UUID
			stateCode     int16
			ackStateCode  *int16
			decodedAckState *domain.AlarmState
		)
		if scanError := rows.Scan(&tagIDValue, &stateCode, &ackStateCode); scanError != nil {
			return domain.AlarmBulkAcknowledgeResult{}, fmt.Errorf("scan bulk alarm state: %w", scanError)
		}
		if ackStateCode != nil {
			value := domain.AlarmState(*ackStateCode)
			decodedAckState = &value
		}
		stateByTagID[tagIDValue] = stateRow{
			state:    domain.AlarmState(stateCode),
			ackState: decodedAckState,
		}
	}
	if rowsError := rows.Err(); rowsError != nil {
		return domain.AlarmBulkAcknowledgeResult{}, fmt.Errorf("iterate bulk alarm states: %w", rowsError)
	}

	const upsertAckQuery = `
INSERT INTO alarms.acks (tag_id, state, actor_id, note, acked_at)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (tag_id) DO UPDATE SET
    state = EXCLUDED.state,
    actor_id = EXCLUDED.actor_id,
    note = EXCLUDED.note,
    acked_at = EXCLUDED.acked_at
`

	const insertEventQuery = `
INSERT INTO alarms.events (
    tag_id, event_type, state_from, state_to, value, quality, ts, actor_id, note
)
VALUES (
    $1, $2, $3, $4,
    (SELECT last_value FROM alarms.states WHERE tag_id = $1),
    (SELECT last_quality FROM alarms.states WHERE tag_id = $1),
    $5, $6, $7
)
RETURNING id, value, quality, created_at
`

	ackedAt := ts.UTC()
	results := make([]domain.AlarmBulkAcknowledgeItemResult, 0, len(tagIDs))
	successCount := 0
	failedCount := 0

	for _, requestedTagID := range tagIDs {
		state, exists := stateByTagID[requestedTagID]
		if !exists {
			failedCount++
			results = append(results, domain.AlarmBulkAcknowledgeItemResult{
				TagID:  requestedTagID,
				Status: domain.AlarmBulkAckStatusNotFound,
			})
			continue
		}

		currentState := state.state
		if currentState == domain.AlarmStateOK {
			failedCount++
			results = append(results, domain.AlarmBulkAcknowledgeItemResult{
				TagID:  requestedTagID,
				Status: domain.AlarmBulkAckStatusNotActive,
			})
			continue
		}
		if state.ackState != nil && *state.ackState == currentState {
			stateCopy := currentState
			failedCount++
			results = append(results, domain.AlarmBulkAcknowledgeItemResult{
				TagID:  requestedTagID,
				Status: domain.AlarmBulkAckStatusAlreadyAcked,
				State:  &stateCopy,
			})
			continue
		}

		if _, upsertAckError := transaction.Exec(
			ctx,
			upsertAckQuery,
			requestedTagID,
			int16(currentState),
			normalizedActorID,
			note,
			ackedAt,
		); upsertAckError != nil {
			return domain.AlarmBulkAcknowledgeResult{}, fmt.Errorf("upsert alarm ack (bulk): %w", upsertAckError)
		}

		event := domain.AlarmEvent{
			TagID:     requestedTagID,
			EventType: domain.AlarmEventAcked,
			StateFrom: currentState,
			StateTo:   currentState,
			TS:        ackedAt,
		}
		qualityCode := int16(0)
		if scanEventError := transaction.QueryRow(
			ctx,
			insertEventQuery,
			requestedTagID,
			int16(event.EventType),
			int16(currentState),
			int16(currentState),
			ackedAt,
			normalizedActorID,
			note,
		).Scan(&event.ID, &event.Value, &qualityCode, &event.CreatedAt); scanEventError != nil {
			return domain.AlarmBulkAcknowledgeResult{}, fmt.Errorf("insert alarm ack event (bulk): %w", scanEventError)
		}
		event.Quality = domain.QualityFromCode(qualityCode)
		event.ActorID = &normalizedActorID
		event.Note = note

		stateCopy := currentState
		successCount++
		results = append(results, domain.AlarmBulkAcknowledgeItemResult{
			TagID:  requestedTagID,
			Status: domain.AlarmBulkAckStatusAcked,
			State:  &stateCopy,
			Event:  &event,
		})
	}

	if commitError := transaction.Commit(ctx); commitError != nil {
		return domain.AlarmBulkAcknowledgeResult{}, fmt.Errorf("commit alarm bulk acknowledge tx: %w", commitError)
	}

	return domain.AlarmBulkAcknowledgeResult{
		Items:    results,
		AckedAt:  ackedAt,
		ActorID:  normalizedActorID,
		SuccessN: successCount,
		FailedN:  failedCount,
	}, nil
}

func (repository *AlarmRepository) ListAlarms(
	ctx context.Context,
	query domain.AlarmListQuery,
) (domain.AlarmListResult, error) {
	baseWhere := buildAlarmStateWhere(query)
	if query.Status == domain.AlarmListStatusCleared {
		return repository.listClearedAlarms(ctx, query)
	}

	countQuery := `
SELECT count(*)
FROM alarms.states s
LEFT JOIN alarms.acks a ON a.tag_id = s.tag_id
LEFT JOIN tags.tags t ON t.id = s.tag_id
LEFT JOIN devices.devices d ON d.id = t.device_id
` + baseWhere

	var total int
	if countError := repository.pool.QueryRow(ctx, countQuery, buildAlarmStateArgs(query)...).Scan(&total); countError != nil {
		return domain.AlarmListResult{}, fmt.Errorf("count alarm states: %w", countError)
	}

	itemsQuery := `
SELECT
    s.tag_id, s.state, s.last_value, s.last_quality, s.entered_at, s.last_seen_at, s.suppressed, s.updated_at,
    COALESCE(t.name, '') AS tag_name,
    COALESCE(d.id, '00000000-0000-0000-0000-000000000000')::uuid AS device_id,
    COALESCE(d.name, '') AS device_name,
    COALESCE(o.id, '00000000-0000-0000-0000-000000000000')::uuid AS object_id,
    COALESCE(o.name, '') AS object_name,
    a.state, a.actor_id, a.note, a.acked_at
FROM alarms.states s
LEFT JOIN alarms.acks a ON a.tag_id = s.tag_id
LEFT JOIN tags.tags t ON t.id = s.tag_id
LEFT JOIN devices.devices d ON d.id = t.device_id
LEFT JOIN public.objects o ON o.id = d.object_id
` + baseWhere + `
ORDER BY s.entered_at DESC, s.tag_id ASC
LIMIT $` + fmt.Sprintf("%d", len(buildAlarmStateArgs(query))+1) + `
OFFSET $` + fmt.Sprintf("%d", len(buildAlarmStateArgs(query))+2)

	args := append(buildAlarmStateArgs(query), query.Limit, query.Offset)
	rows, queryError := repository.pool.Query(ctx, itemsQuery, args...)
	if queryError != nil {
		return domain.AlarmListResult{}, fmt.Errorf("query alarm states: %w", queryError)
	}
	defer rows.Close()

	items := make([]domain.AlarmStateRecord, 0)
	for rows.Next() {
		record, scanError := scanAlarmStateRow(rows)
		if scanError != nil {
			return domain.AlarmListResult{}, scanError
		}
		items = append(items, record)
	}
	if rowsError := rows.Err(); rowsError != nil {
		return domain.AlarmListResult{}, fmt.Errorf("iterate alarm states: %w", rowsError)
	}

	return domain.AlarmListResult{
		Items:  items,
		Total:  total,
		Limit:  query.Limit,
		Offset: query.Offset,
	}, nil
}

func (repository *AlarmRepository) listClearedAlarms(
	ctx context.Context,
	query domain.AlarmListQuery,
) (domain.AlarmListResult, error) {
	where, args := buildClearedAlarmsWhere(query)

	countQuery := `
SELECT count(*)
FROM alarms.events e
LEFT JOIN tags.tags t ON t.id = e.tag_id
LEFT JOIN devices.devices d ON d.id = t.device_id
` + where
	var total int
	if countError := repository.pool.QueryRow(ctx, countQuery, args...).Scan(&total); countError != nil {
		return domain.AlarmListResult{}, fmt.Errorf("count cleared alarms: %w", countError)
	}

	itemsQuery := `
SELECT
    e.tag_id,
    e.state_from,
    e.value,
    e.quality,
    e.ts,
    e.ts,
    false,
    e.created_at,
    COALESCE(t.name, '') AS tag_name,
    COALESCE(d.id, '00000000-0000-0000-0000-000000000000')::uuid AS device_id,
    COALESCE(d.name, '') AS device_name,
    COALESCE(o.id, '00000000-0000-0000-0000-000000000000')::uuid AS object_id,
    COALESCE(o.name, '') AS object_name,
    NULL::smallint, NULL::text, NULL::text, NULL::timestamptz
FROM alarms.events e
LEFT JOIN tags.tags t ON t.id = e.tag_id
LEFT JOIN devices.devices d ON d.id = t.device_id
LEFT JOIN public.objects o ON o.id = d.object_id
` + where + `
ORDER BY e.ts DESC, e.tag_id ASC
LIMIT $` + fmt.Sprintf("%d", len(args)+1) + `
OFFSET $` + fmt.Sprintf("%d", len(args)+2)

	rows, queryError := repository.pool.Query(ctx, itemsQuery, append(args, query.Limit, query.Offset)...)
	if queryError != nil {
		return domain.AlarmListResult{}, fmt.Errorf("query cleared alarms: %w", queryError)
	}
	defer rows.Close()

	items := make([]domain.AlarmStateRecord, 0)
	for rows.Next() {
		record, scanError := scanAlarmStateRow(rows)
		if scanError != nil {
			return domain.AlarmListResult{}, scanError
		}
		items = append(items, record)
	}
	if rowsError := rows.Err(); rowsError != nil {
		return domain.AlarmListResult{}, fmt.Errorf("iterate cleared alarms: %w", rowsError)
	}

	return domain.AlarmListResult{
		Items:  items,
		Total:  total,
		Limit:  query.Limit,
		Offset: query.Offset,
	}, nil
}

func (repository *AlarmRepository) GetAlarm(ctx context.Context, tagID uuid.UUID) (domain.AlarmDetail, error) {
	const currentQuery = `
SELECT
    s.tag_id, s.state, s.last_value, s.last_quality, s.entered_at, s.last_seen_at, s.suppressed, s.updated_at,
    COALESCE(t.name, '') AS tag_name,
    COALESCE(d.id, '00000000-0000-0000-0000-000000000000')::uuid AS device_id,
    COALESCE(d.name, '') AS device_name,
    COALESCE(o.id, '00000000-0000-0000-0000-000000000000')::uuid AS object_id,
    COALESCE(o.name, '') AS object_name,
    a.state, a.actor_id, a.note, a.acked_at
FROM alarms.states s
LEFT JOIN alarms.acks a ON a.tag_id = s.tag_id
LEFT JOIN tags.tags t ON t.id = s.tag_id
LEFT JOIN devices.devices d ON d.id = t.device_id
LEFT JOIN public.objects o ON o.id = d.object_id
WHERE s.tag_id = $1
`
	rows, queryError := repository.pool.Query(ctx, currentQuery, tagID)
	if queryError != nil {
		return domain.AlarmDetail{}, fmt.Errorf("query alarm detail state: %w", queryError)
	}
	defer rows.Close()

	var current *domain.AlarmStateRecord
	if rows.Next() {
		record, scanError := scanAlarmStateRow(rows)
		if scanError != nil {
			return domain.AlarmDetail{}, scanError
		}
		current = &record
	}
	if current == nil {
		return domain.AlarmDetail{}, fmt.Errorf("alarm %s not found: %w", tagID.String(), domain.ErrNotFound)
	}

	const eventsQuery = `
SELECT id, tag_id, event_type, state_from, state_to, value, quality, ts, actor_id, note, created_at
FROM alarms.events
WHERE tag_id = $1
ORDER BY ts DESC
LIMIT 50
`

	rows, queryError = repository.pool.Query(ctx, eventsQuery, tagID)
	if queryError != nil {
		return domain.AlarmDetail{}, fmt.Errorf("query alarm events: %w", queryError)
	}
	defer rows.Close()

	events := make([]domain.AlarmEvent, 0, 50)
	for rows.Next() {
		var (
			event       domain.AlarmEvent
			eventType   int16
			stateFrom   int16
			stateTo     int16
			qualityCode int16
		)
		if scanError := rows.Scan(
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
		); scanError != nil {
			return domain.AlarmDetail{}, fmt.Errorf("scan alarm event: %w", scanError)
		}
		event.EventType = domain.AlarmEventType(eventType)
		event.StateFrom = domain.AlarmState(stateFrom)
		event.StateTo = domain.AlarmState(stateTo)
		event.Quality = domain.QualityFromCode(qualityCode)
		events = append(events, event)
	}
	if rowsError := rows.Err(); rowsError != nil {
		return domain.AlarmDetail{}, fmt.Errorf("iterate alarm events: %w", rowsError)
	}

	return domain.AlarmDetail{
		CurrentState: current,
		Events:       events,
	}, nil
}

func (repository *AlarmRepository) ListActiveUnacked(
	ctx context.Context,
	limit int,
) ([]domain.AlarmStateRecord, error) {
	const baseQuery = `
SELECT
    s.tag_id, s.state, s.last_value, s.last_quality, s.entered_at, s.last_seen_at, s.suppressed, s.updated_at,
    COALESCE(t.name, '') AS tag_name,
    COALESCE(d.id, '00000000-0000-0000-0000-000000000000')::uuid AS device_id,
    COALESCE(d.name, '') AS device_name,
    COALESCE(o.id, '00000000-0000-0000-0000-000000000000')::uuid AS object_id,
    COALESCE(o.name, '') AS object_name,
    a.state, a.actor_id, a.note, a.acked_at
FROM alarms.states s
LEFT JOIN alarms.acks a ON a.tag_id = s.tag_id
LEFT JOIN tags.tags t ON t.id = s.tag_id
LEFT JOIN devices.devices d ON d.id = t.device_id
LEFT JOIN public.objects o ON o.id = d.object_id
WHERE s.state > 0
  AND a.tag_id IS NULL
ORDER BY s.entered_at DESC, s.tag_id ASC
`

	query := baseQuery
	args := make([]any, 0, 1)
	if limit > 0 {
		query += "\nLIMIT $1"
		args = append(args, limit)
	}

	rows, queryError := repository.pool.Query(ctx, query, args...)
	if queryError != nil {
		return nil, fmt.Errorf("query active unacked alarms: %w", queryError)
	}
	defer rows.Close()

	items := make([]domain.AlarmStateRecord, 0)
	for rows.Next() {
		record, scanError := scanAlarmStateRow(rows)
		if scanError != nil {
			return nil, scanError
		}
		items = append(items, record)
	}
	if rowsError := rows.Err(); rowsError != nil {
		return nil, fmt.Errorf("iterate active unacked alarms: %w", rowsError)
	}

	return items, nil
}

func (repository *AlarmRepository) ListCommLossCandidates(
	ctx context.Context,
	threshold time.Time,
	limit int,
) ([]domain.AlarmStateRecord, error) {
	const query = `
SELECT tag_id, state, last_value, last_quality, entered_at, last_seen_at, suppressed, updated_at
FROM alarms.states
WHERE last_seen_at < $1
  AND state NOT IN ($2, $3)
ORDER BY last_seen_at ASC
LIMIT $4
`

	rows, queryError := repository.pool.Query(
		ctx,
		query,
		threshold.UTC(),
		int16(domain.AlarmStateCommLoss),
		int16(domain.AlarmStateOffline),
		limit,
	)
	if queryError != nil {
		return nil, fmt.Errorf("query comm loss candidates: %w", queryError)
	}
	defer rows.Close()

	records := make([]domain.AlarmStateRecord, 0)
	for rows.Next() {
		var (
			record      domain.AlarmStateRecord
			stateCode   int16
			qualityCode int16
		)
		if scanError := rows.Scan(
			&record.TagID,
			&stateCode,
			&record.LastValue,
			&qualityCode,
			&record.EnteredAt,
			&record.LastSeenAt,
			&record.Suppressed,
			&record.UpdatedAt,
		); scanError != nil {
			return nil, fmt.Errorf("scan comm loss candidate: %w", scanError)
		}
		record.State = domain.AlarmState(stateCode)
		record.LastQuality = domain.QualityFromCode(qualityCode)
		records = append(records, record)
	}
	if rowsError := rows.Err(); rowsError != nil {
		return nil, fmt.Errorf("iterate comm loss candidates: %w", rowsError)
	}

	return records, nil
}

func (repository *AlarmRepository) DeleteByTagID(ctx context.Context, tagID uuid.UUID) error {
	const deleteAcksQuery = `DELETE FROM alarms.acks WHERE tag_id = $1`
	const deleteStateQuery = `DELETE FROM alarms.states WHERE tag_id = $1`
	if _, execError := repository.pool.Exec(ctx, deleteAcksQuery, tagID); execError != nil {
		return fmt.Errorf("delete alarm acks by tag: %w", execError)
	}
	if _, execError := repository.pool.Exec(ctx, deleteStateQuery, tagID); execError != nil {
		return fmt.Errorf("delete alarm state by tag: %w", execError)
	}
	return nil
}

func buildAlarmStateWhere(query domain.AlarmListQuery) string {
	conditions := []string{"WHERE 1=1"}

	switch query.Status {
	case domain.AlarmListStatusAcked:
		conditions = append(conditions, "AND s.state > 0", "AND a.tag_id IS NOT NULL")
	default:
		conditions = append(conditions, "AND s.state > 0", "AND a.tag_id IS NULL")
	}

	if query.Severity == domain.AlarmSeverityWarn {
		conditions = append(conditions, "AND s.state IN (1, 2)")
	}
	if query.Severity == domain.AlarmSeverityAlarm {
		conditions = append(conditions, "AND s.state IN (3, 4)")
	}

	if query.ObjectID != uuid.Nil {
		conditions = append(conditions, "AND d.object_id = $1")
	}

	return "\n" + strings.Join(conditions, "\n")
}

func buildAlarmStateArgs(query domain.AlarmListQuery) []any {
	args := make([]any, 0, 1)
	if query.ObjectID != uuid.Nil {
		args = append(args, query.ObjectID)
	}
	return args
}

func buildClearedAlarmsWhere(query domain.AlarmListQuery) (string, []any) {
	conditions := []string{
		"WHERE e.event_type = 1",
	}
	args := make([]any, 0, 4)
	position := 1

	if query.ObjectID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("AND d.object_id = $%d", position))
		args = append(args, query.ObjectID)
		position++
	}
	if query.From != nil {
		conditions = append(conditions, fmt.Sprintf("AND e.ts >= $%d", position))
		args = append(args, query.From.UTC())
		position++
	}
	if query.To != nil {
		conditions = append(conditions, fmt.Sprintf("AND e.ts <= $%d", position))
		args = append(args, query.To.UTC())
		position++
	}
	if query.Severity == domain.AlarmSeverityWarn {
		conditions = append(conditions, "AND e.state_from IN (1, 2)")
	}
	if query.Severity == domain.AlarmSeverityAlarm {
		conditions = append(conditions, "AND e.state_from IN (3, 4)")
	}

	return "\n" + strings.Join(conditions, "\n"), args
}

func scanAlarmStateRow(rows pgx.Rows) (domain.AlarmStateRecord, error) {
	var (
		record       domain.AlarmStateRecord
		stateCode    int16
		qualityCode  int16
		ackStateCode *int16
		ackActorID   *string
		ackNote      *string
		ackAt        *time.Time
	)

	if scanError := rows.Scan(
		&record.TagID,
		&stateCode,
		&record.LastValue,
		&qualityCode,
		&record.EnteredAt,
		&record.LastSeenAt,
		&record.Suppressed,
		&record.UpdatedAt,
		&record.TagName,
		&record.DeviceID,
		&record.DeviceName,
		&record.ObjectID,
		&record.ObjectName,
		&ackStateCode,
		&ackActorID,
		&ackNote,
		&ackAt,
	); scanError != nil {
		return domain.AlarmStateRecord{}, fmt.Errorf("scan alarm row: %w", scanError)
	}

	record.State = domain.AlarmState(stateCode)
	record.LastQuality = domain.QualityFromCode(qualityCode)
	if ackStateCode != nil && ackActorID != nil && ackAt != nil {
		record.Ack = &domain.AlarmAck{
			State:   domain.AlarmState(*ackStateCode),
			ActorID: *ackActorID,
			Note:    ackNote,
			AckedAt: ackAt.UTC(),
		}
	}

	return record, nil
}
