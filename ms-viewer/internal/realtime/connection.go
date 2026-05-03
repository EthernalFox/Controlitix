package realtime

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/google/uuid"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
)

const defaultWriteTimeout = 5 * time.Second

type ConnectionOptions struct {
	Hub             *Hub
	Socket          *websocket.Conn
	Subject         string
	Roles           []string
	SessionID       string
	Logger          *slog.Logger
	WriteBufferSize int
	PingInterval    time.Duration
	PongTimeout     time.Duration
}

type Connection struct {
	hub       *Hub
	socket    *websocket.Conn
	subject   string
	roles     []string
	sessionID string
	logger    *slog.Logger

	outbox chan outboundMessage
	done   chan struct{}

	openedAt time.Time

	pingInterval time.Duration
	pongTimeout  time.Duration

	closeOnce   sync.Once
	closeCode   websocket.StatusCode
	closeReason string
}

type outboundMessage struct {
	payload []byte
	tagID   string
}

func NewConnection(options ConnectionOptions) *Connection {
	logger := options.Logger
	if logger == nil {
		logger = slog.Default()
	}
	writeBufferSize := options.WriteBufferSize
	if writeBufferSize <= 0 {
		writeBufferSize = 256
	}
	pingInterval := options.PingInterval
	if pingInterval <= 0 {
		pingInterval = 20 * time.Second
	}
	pongTimeout := options.PongTimeout
	if pongTimeout <= 0 {
		pongTimeout = 30 * time.Second
	}

	return &Connection{
		hub:          options.Hub,
		socket:       options.Socket,
		subject:      options.Subject,
		roles:        append([]string(nil), options.Roles...),
		sessionID:    options.SessionID,
		logger:       logger,
		outbox:       make(chan outboundMessage, writeBufferSize),
		done:         make(chan struct{}),
		openedAt:     time.Now().UTC(),
		pingInterval: pingInterval,
		pongTimeout:  pongTimeout,
	}
}

func (connection *Connection) Run(ctx context.Context) {
	connection.logger.Info(
		"connection.opened",
		"method",
		"Connection.Run",
		"session_id",
		connection.sessionID,
		"subject",
		connection.subject,
	)

	writerDone := make(chan struct{})
	go connection.writeLoop(ctx, writerDone)

	connection.readLoop(ctx)

	select {
	case <-writerDone:
	case <-time.After(time.Second):
	}
}

func (connection *Connection) readLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			connection.Close(websocket.StatusGoingAway, "server_shutdown")
			return
		case <-connection.done:
			return
		default:
		}

		messageType, payload, readError := connection.socket.Read(ctx)
		if readError != nil {
			statusCode := websocket.CloseStatus(readError)
			if statusCode == -1 {
				connection.Close(websocket.StatusInternalError, "read_error")
			} else {
				connection.Close(statusCode, "client_closed")
			}
			return
		}
		if messageType != websocket.MessageText {
			continue
		}

		message, parseError := ParseClientMessage(payload)
		if parseError != nil {
			connection.enqueueError(
				ErrorCodeInvalidTopic,
				"invalid message payload",
				nil,
			)
			continue
		}

		switch message.T {
		case ClientTypeSubscribe:
			connection.handleSubscribe(message.Topics)
		case ClientTypeUnsubscribe:
			connection.handleUnsubscribe(message.Topics)
		case ClientTypePing:
			connection.enqueuePong()
		default:
			connection.enqueueError(
				ErrorCodeInvalidTopic,
				"unsupported message type",
				nil,
			)
		}
	}
}

func (connection *Connection) writeLoop(ctx context.Context, done chan<- struct{}) {
	defer close(done)

	pingTicker := time.NewTicker(connection.pingInterval)
	defer pingTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			connection.Close(websocket.StatusGoingAway, "server_shutdown")
			return
		case <-connection.done:
			return
		case <-pingTicker.C:
			pingContext, cancelPing := context.WithTimeout(ctx, connection.pongTimeout)
			pingError := connection.socket.Ping(pingContext)
			cancelPing()
			if pingError != nil {
				connection.Close(websocket.StatusGoingAway, "pong_timeout")
				return
			}
		case message := <-connection.outbox:
			writeContext, cancelWrite := context.WithTimeout(ctx, defaultWriteTimeout)
			writeError := connection.socket.Write(
				writeContext,
				websocket.MessageText,
				message.payload,
			)
			cancelWrite()
			if writeError != nil {
				connection.Close(websocket.StatusInternalError, "write_error")
				return
			}
		}
	}
}

func (connection *Connection) handleSubscribe(topics []string) {
	if connection.hub == nil {
		connection.enqueueError(ErrorCodeRateLimited, "hub is not configured", nil)
		return
	}

	allowedTopics := make([]string, 0, len(topics))
	unauthorizedTopics := make([]string, 0)
	for _, topic := range normalizeTopics(topics) {
		if CanSubscribeTopic(connection.roles, topic) {
			allowedTopics = append(allowedTopics, topic)
			continue
		}
		unauthorizedTopics = append(unauthorizedTopics, topic)
	}

	result := connection.hub.Subscribe(connection, allowedTopics)
	if len(result.SubscribedTopics) > 0 {
		connection.logger.Info(
			"subscribe",
			"method",
			"Connection.handleSubscribe",
			"session_id",
			connection.sessionID,
			"topics",
			result.SubscribedTopics,
		)
		connection.enqueueSubscribed(result.SubscribedTopics)
		connection.enqueueSubscribeSnapshots(result.SubscribedTagIDs)
		if result.SubscribedAlarms {
			connection.enqueueAlarmsSnapshotAsync()
		}
	}
	if len(result.InvalidTopics) > 0 {
		connection.enqueueError(
			ErrorCodeInvalidTopic,
			"some topics have invalid format or unsupported prefix",
			result.InvalidTopics,
		)
	}
	if len(result.LimitedTopics) > 0 {
		connection.enqueueError(
			ErrorCodeLimitExceeded,
			"subscription limit exceeded",
			result.LimitedTopics,
		)
	}
	if len(unauthorizedTopics) > 0 {
		connection.enqueueError(
			ErrorCodeUnauthorized,
			"topic subscription is forbidden",
			unauthorizedTopics,
		)
	}
}

func (connection *Connection) handleUnsubscribe(topics []string) {
	if connection.hub == nil {
		connection.enqueueError(ErrorCodeRateLimited, "hub is not configured", nil)
		return
	}

	result := connection.hub.Unsubscribe(connection, topics)
	if len(result.UnsubscribedTopics) > 0 {
		connection.enqueueUnsubscribed(result.UnsubscribedTopics)
	}
	if len(result.InvalidTopics) > 0 {
		connection.enqueueError(
			ErrorCodeInvalidTopic,
			"some topics have invalid format or unsupported prefix",
			result.InvalidTopics,
		)
	}
}

func (connection *Connection) enqueueSubscribeSnapshots(tagIDs []uuid.UUID) {
	if connection.hub == nil || len(tagIDs) == 0 {
		return
	}

	tagIDsCopy := append([]uuid.UUID(nil), tagIDs...)
	go func() {
		loadContext, cancelLoad := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelLoad()

		recordsByTag := connection.hub.LoadLatestRecords(loadContext, tagIDsCopy)
		for _, tagID := range tagIDsCopy {
			record, exists := recordsByTag[tagID]
			if !exists {
				continue
			}

			connection.hub.debouncer.Touch(connection, record)
			connection.enqueueSnapshot(record, SnapshotReasonSubscribe)
		}
	}()
}

func (connection *Connection) EnqueueWelcome(maxSubscriptions int, debounceMS int) {
	connection.enqueuePayload(BuildWelcomeMessage(connection.sessionID, maxSubscriptions, debounceMS), "")
}

func (connection *Connection) enqueueSubscribed(topics []string) {
	connection.enqueuePayload(BuildSubscribedMessage(topics), "")
}

func (connection *Connection) enqueueUnsubscribed(topics []string) {
	connection.enqueuePayload(BuildUnsubscribedMessage(topics), "")
}

func (connection *Connection) enqueueValue(record domain.IngestRecord) {
	connection.enqueuePayload(BuildValueMessage(record), record.TagID.String())
}

func (connection *Connection) enqueueSnapshot(record domain.IngestRecord, reason SnapshotReason) {
	connection.enqueuePayload(BuildSnapshotMessage(record, reason), record.TagID.String())
}

func (connection *Connection) enqueuePong() {
	connection.enqueuePayload(BuildPongMessage(), "")
}

func (connection *Connection) enqueueError(code string, detail string, topics []string) {
	connection.enqueuePayload(BuildErrorMessage(code, detail, topics), "")
}

func (connection *Connection) enqueueConfigChanged(event domain.ConfigChangedEvent) {
	connection.enqueuePayload(BuildConfigChangedMessage(event), "config.changed")
}

func (connection *Connection) enqueueTopicsChanged(
	topic string,
	addedCount int,
	removedCount int,
) {
	connection.enqueuePayload(
		BuildTopicsChangedMessage(topic, addedCount, removedCount),
		"topics_changed",
	)
}

func (connection *Connection) enqueuePayload(payload any, tagID string) {
	encodedPayload, marshalError := json.Marshal(payload)
	if marshalError != nil {
		connection.logger.Error(
			"failed to encode websocket payload",
			"method",
			"Connection.enqueuePayload",
			"error",
			marshalError,
		)
		return
	}

	message := outboundMessage{payload: encodedPayload, tagID: tagID}

	select {
	case <-connection.done:
		return
	default:
	}

	select {
	case connection.outbox <- message:
		return
	default:
	}

	var dropped outboundMessage
	wasDropped := false
	select {
	case dropped = <-connection.outbox:
		wasDropped = true
	default:
	}

	if wasDropped {
		connection.logger.Warn(
			"dropped",
			"method",
			"Connection.enqueuePayload",
			"tag_id",
			dropped.tagID,
			"reason",
			"outbox_full",
		)
	}

	select {
	case connection.outbox <- message:
	default:
		connection.logger.Warn(
			"dropped",
			"method",
			"Connection.enqueuePayload",
			"tag_id",
			message.tagID,
			"reason",
			"outbox_full",
		)
	}
}

func (connection *Connection) Close(statusCode websocket.StatusCode, reason string) {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		reason = "closed"
	}

	connection.closeOnce.Do(func() {
		connection.closeCode = statusCode
		connection.closeReason = reason

		close(connection.done)
		if connection.socket != nil {
			_ = connection.socket.Close(statusCode, reason)
		}
		if connection.hub != nil {
			connection.hub.UnregisterConnection(connection)
		}

		connection.logger.Info(
			"connection.closed",
			"method",
			"Connection.Close",
			"session_id",
			connection.sessionID,
			"reason",
			reason,
			"code",
			int(statusCode),
			"duration_ms",
			time.Since(connection.openedAt).Milliseconds(),
		)
	})
}

func (connection *Connection) Subject() string {
	return connection.subject
}

func (connection *Connection) SessionID() string {
	return connection.sessionID
}

func (connection *Connection) String() string {
	return fmt.Sprintf("session=%s subject=%s", connection.sessionID, connection.subject)
}

func (connection *Connection) CloseCode() int {
	return int(connection.closeCode)
}

func (connection *Connection) CloseReason() string {
	return strings.TrimSpace(connection.closeReason)
}

func (connection *Connection) DurationMS() int64 {
	if connection.openedAt.IsZero() {
		return 0
	}
	return time.Since(connection.openedAt).Milliseconds()
}
