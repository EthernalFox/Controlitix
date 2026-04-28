package realtime

import (
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
)

const defaultHeartbeatInterval = 30 * time.Second

type Debouncer struct {
	mutex             sync.Mutex
	debounceInterval  time.Duration
	heartbeatInterval time.Duration
	states            map[debounceKey]*debounceState
	logger            *slog.Logger
}

type debounceKey struct {
	connection *Connection
	tagID      uuid.UUID
}

type debounceState struct {
	key            debounceKey
	latest         domain.IngestRecord
	hasLatest      bool
	pending        *domain.IngestRecord
	lastSentAt     time.Time
	flushTimer     *time.Timer
	heartbeatTimer *time.Timer
}

func NewDebouncer(
	debounceInterval time.Duration,
	heartbeatInterval time.Duration,
	logger *slog.Logger,
) *Debouncer {
	if debounceInterval <= 0 {
		debounceInterval = 100 * time.Millisecond
	}
	if heartbeatInterval <= 0 {
		heartbeatInterval = defaultHeartbeatInterval
	}
	if logger == nil {
		logger = slog.Default()
	}

	return &Debouncer{
		debounceInterval:  debounceInterval,
		heartbeatInterval: heartbeatInterval,
		states:            make(map[debounceKey]*debounceState),
		logger:            logger,
	}
}

func (debouncer *Debouncer) Ensure(connection *Connection, tagID uuid.UUID) {
	if connection == nil || tagID == uuid.Nil {
		return
	}

	debouncer.mutex.Lock()
	defer debouncer.mutex.Unlock()

	key := debounceKey{connection: connection, tagID: tagID}
	if _, exists := debouncer.states[key]; exists {
		return
	}

	debouncer.states[key] = &debounceState{key: key}
}

func (debouncer *Debouncer) Touch(connection *Connection, record domain.IngestRecord) {
	if connection == nil || record.TagID == uuid.Nil {
		return
	}

	debouncer.mutex.Lock()
	state := debouncer.getOrCreateStateLocked(connection, record.TagID)
	state.latest = record
	state.hasLatest = true
	debouncer.resetHeartbeatTimerLocked(state)
	debouncer.mutex.Unlock()
}

func (debouncer *Debouncer) Publish(connection *Connection, record domain.IngestRecord) {
	if connection == nil || record.TagID == uuid.Nil {
		return
	}

	var immediate *domain.IngestRecord

	debouncer.mutex.Lock()
	state := debouncer.getOrCreateStateLocked(connection, record.TagID)
	state.latest = record
	state.hasLatest = true
	debouncer.resetHeartbeatTimerLocked(state)

	now := time.Now().UTC()
	if state.lastSentAt.IsZero() || now.Sub(state.lastSentAt) >= debouncer.debounceInterval {
		recordCopy := record
		immediate = &recordCopy
		state.pending = nil
		state.lastSentAt = now
	} else {
		recordCopy := record
		state.pending = &recordCopy

		wait := debouncer.debounceInterval - now.Sub(state.lastSentAt)
		if wait < 0 {
			wait = 0
		}
		debouncer.scheduleFlushLocked(state, wait)
	}
	debouncer.mutex.Unlock()

	if immediate != nil {
		connection.enqueueValue(*immediate)
	}
}

func (debouncer *Debouncer) Remove(connection *Connection, tagID uuid.UUID) {
	if connection == nil || tagID == uuid.Nil {
		return
	}

	debouncer.mutex.Lock()
	defer debouncer.mutex.Unlock()

	debouncer.removeStateLocked(debounceKey{connection: connection, tagID: tagID})
}

func (debouncer *Debouncer) StopConnection(connection *Connection) {
	if connection == nil {
		return
	}

	debouncer.mutex.Lock()
	defer debouncer.mutex.Unlock()

	keys := make([]debounceKey, 0)
	for key := range debouncer.states {
		if key.connection == connection {
			keys = append(keys, key)
		}
	}

	for _, key := range keys {
		debouncer.removeStateLocked(key)
	}
}

func (debouncer *Debouncer) StateCount() int {
	debouncer.mutex.Lock()
	defer debouncer.mutex.Unlock()

	return len(debouncer.states)
}

func (debouncer *Debouncer) getOrCreateStateLocked(
	connection *Connection,
	tagID uuid.UUID,
) *debounceState {
	key := debounceKey{connection: connection, tagID: tagID}
	state, exists := debouncer.states[key]
	if exists {
		return state
	}

	state = &debounceState{key: key}
	debouncer.states[key] = state
	return state
}

func (debouncer *Debouncer) scheduleFlushLocked(state *debounceState, wait time.Duration) {
	if state.flushTimer == nil {
		state.flushTimer = time.AfterFunc(wait, func() {
			debouncer.flushPending(state.key)
		})
		return
	}

	if !state.flushTimer.Stop() {
		debouncer.logger.Debug(
			"debounce timer already fired",
			"method",
			"Debouncer.scheduleFlushLocked",
			"tag_id",
			state.key.tagID.String(),
		)
	}
	state.flushTimer.Reset(wait)
}

func (debouncer *Debouncer) resetHeartbeatTimerLocked(state *debounceState) {
	if state.heartbeatTimer == nil {
		state.heartbeatTimer = time.AfterFunc(debouncer.heartbeatInterval, func() {
			debouncer.emitHeartbeat(state.key)
		})
		return
	}

	if !state.heartbeatTimer.Stop() {
		debouncer.logger.Debug(
			"heartbeat timer already fired",
			"method",
			"Debouncer.resetHeartbeatTimerLocked",
			"tag_id",
			state.key.tagID.String(),
		)
	}
	state.heartbeatTimer.Reset(debouncer.heartbeatInterval)
}

func (debouncer *Debouncer) flushPending(key debounceKey) {
	var (
		connection *Connection
		record     *domain.IngestRecord
	)

	debouncer.mutex.Lock()
	state, exists := debouncer.states[key]
	if !exists {
		debouncer.mutex.Unlock()
		return
	}

	state.flushTimer = nil
	if state.pending != nil {
		pendingCopy := *state.pending
		record = &pendingCopy
		state.pending = nil
		state.lastSentAt = time.Now().UTC()
		debouncer.resetHeartbeatTimerLocked(state)
		connection = key.connection
	}
	debouncer.mutex.Unlock()

	if record != nil && connection != nil {
		connection.enqueueValue(*record)
	}
}

func (debouncer *Debouncer) emitHeartbeat(key debounceKey) {
	var (
		connection *Connection
		record     *domain.IngestRecord
	)

	debouncer.mutex.Lock()
	state, exists := debouncer.states[key]
	if !exists {
		debouncer.mutex.Unlock()
		return
	}

	state.heartbeatTimer = nil
	if state.hasLatest {
		recordCopy := state.latest
		record = &recordCopy
		connection = key.connection
		debouncer.resetHeartbeatTimerLocked(state)
	}
	debouncer.mutex.Unlock()

	if record != nil && connection != nil {
		connection.enqueueSnapshot(*record, SnapshotReasonHeartbeat)
	}
}

func (debouncer *Debouncer) removeStateLocked(key debounceKey) {
	state, exists := debouncer.states[key]
	if !exists {
		return
	}

	if state.flushTimer != nil {
		state.flushTimer.Stop()
		state.flushTimer = nil
	}
	if state.heartbeatTimer != nil {
		state.heartbeatTimer.Stop()
		state.heartbeatTimer = nil
	}

	delete(debouncer.states, key)
}
