package realtime

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
)

const (
	ClientTypeSubscribe   = "subscribe"
	ClientTypeUnsubscribe = "unsubscribe"
	ClientTypePing        = "ping"
)

const TopicAlarms = "alarms"

const (
	TopicPrefixTag     = "tag"
	TopicPrefixObject  = "object"
	TopicPrefixDiagram = "diagram"
)

const (
	ServerTypeWelcome      = "welcome"
	ServerTypeSubscribed   = "subscribed"
	ServerTypeUnsubscribed = "unsubscribed"
	ServerTypeValue        = "value"
	ServerTypeSnapshot     = "snapshot"
	ServerTypePong         = "pong"
	ServerTypeError         = "error"
	ServerTypeConfigChanged = "config.changed"
	ServerTypeAlarm         = "alarm"
	ServerTypeAlarmsBatch    = "alarms_batch"
	ServerTypeAlarmsSnapshot = "alarms_snapshot"
	ServerTypeTopicsChanged  = "topics_changed"
)

const (
	ErrorCodeLimitExceeded = "limit-exceeded"
	ErrorCodeInvalidTopic  = "invalid-topic"
	ErrorCodeAuthRequired  = "auth-required"
	ErrorCodeUnauthorized  = "unauthorized"
	ErrorCodeRateLimited   = "rate-limited"
)

type SnapshotReason string

const (
	SnapshotReasonSubscribe SnapshotReason = "subscribe"
	SnapshotReasonHeartbeat SnapshotReason = "heartbeat"
)

type ClientMessage struct {
	T      string   `json:"t"`
	Topics []string `json:"topics,omitempty"`
}

type WelcomeLimits struct {
	MaxSubscriptions int `json:"max_subscriptions"`
	DebounceMS       int `json:"debounce_ms"`
}

type WelcomeMessage struct {
	T          string        `json:"t"`
	SessionID  string        `json:"session_id"`
	ServerTime string        `json:"server_time"`
	Limits     WelcomeLimits `json:"limits"`
}

type TopicsMessage struct {
	T      string   `json:"t"`
	Topics []string `json:"topics"`
}

type ValueMessage struct {
	T     string   `json:"t"`
	TagID string   `json:"tag_id"`
	TS    string   `json:"ts"`
	V     *float64 `json:"v"`
	Q     string   `json:"q"`
}

type SnapshotMessage struct {
	T      string         `json:"t"`
	TagID  string         `json:"tag_id"`
	TS     string         `json:"ts"`
	V      *float64       `json:"v"`
	Q      string         `json:"q"`
	Reason SnapshotReason `json:"reason"`
}

type PongMessage struct {
	T string `json:"t"`
}

type ErrorMessage struct {
	T      string   `json:"t"`
	Code   string   `json:"code"`
	Detail string   `json:"detail"`
	Topics []string `json:"topics,omitempty"`
}

type ConfigChangedMessage struct {
	T          string          `json:"t"`
	EntityType string          `json:"entity_type"`
	EntityID   string          `json:"entity_id"`
	Operation  string          `json:"operation"`
	Timestamp  string          `json:"timestamp"`
	Payload    json.RawMessage `json:"payload,omitempty"`
}

type AlarmMessage struct {
	T         string   `json:"t"`
	EventType string   `json:"event_type"`
	TagID     string   `json:"tag_id"`
	StateFrom string   `json:"state_from"`
	StateTo   string   `json:"state_to"`
	Value     *float64 `json:"value"`
	Quality   string   `json:"quality"`
	TS        string   `json:"ts"`
	ActorID   *string  `json:"actor_id"`
	Note      *string  `json:"note"`
}

type AlarmsBatchMessage struct {
	T      string         `json:"t"`
	TS     string         `json:"ts"`
	Events []AlarmMessage `json:"events"`
}

type AlarmsSnapshotMessage struct {
	T     string              `json:"t"`
	TS    string              `json:"ts"`
	Items []AlarmRecordMessage `json:"items"`
}

type AlarmRecordMessage struct {
	TagID      string   `json:"tag_id"`
	TagName    string   `json:"tag_name"`
	DeviceID   string   `json:"device_id"`
	DeviceName string   `json:"device_name"`
	ObjectID   string   `json:"object_id"`
	ObjectName string   `json:"object_name"`
	State      string   `json:"state"`
	Value      *float64 `json:"value"`
	Quality    string   `json:"quality"`
	EnteredAt  string   `json:"entered_at"`
	LastSeenAt string   `json:"last_seen_at"`
	Acked      bool     `json:"acked"`
	Ack        *AlarmAckMessage `json:"ack"`
}

type AlarmAckMessage struct {
	ActorID string  `json:"actor_id"`
	AckedAt string  `json:"acked_at"`
	Note    *string `json:"note"`
}

type TopicsChangedMessage struct {
	T            string `json:"t"`
	Topic        string `json:"topic"`
	AddedCount   int    `json:"added_count"`
	RemovedCount int    `json:"removed_count"`
}

func ParseClientMessage(payload []byte) (ClientMessage, error) {
	var message ClientMessage
	if unmarshalError := json.Unmarshal(payload, &message); unmarshalError != nil {
		return ClientMessage{}, fmt.Errorf("decode ws message: %w", unmarshalError)
	}

	message.T = strings.TrimSpace(message.T)
	if message.T == "" {
		return ClientMessage{}, fmt.Errorf("decode ws message: empty type")
	}

	return message, nil
}

func ParseTagTopic(topic string) (uuid.UUID, error) {
	topic = strings.TrimSpace(topic)
	if topic == "" {
		return uuid.Nil, fmt.Errorf("topic is empty")
	}

	parts := strings.Split(topic, ":")
	if len(parts) != 2 {
		return uuid.Nil, fmt.Errorf("topic has invalid format")
	}
	if parts[0] != TopicPrefixTag {
		return uuid.Nil, fmt.Errorf("topic prefix is unsupported")
	}

	tagID, parseError := uuid.Parse(parts[1])
	if parseError != nil {
		return uuid.Nil, fmt.Errorf("parse tag topic: %w", parseError)
	}

	return tagID, nil
}

func ParseGroupTopic(topic string) (string, uuid.UUID, error) {
	topic = strings.TrimSpace(topic)
	if topic == "" {
		return "", uuid.Nil, fmt.Errorf("topic is empty")
	}

	parts := strings.Split(topic, ":")
	if len(parts) != 2 {
		return "", uuid.Nil, fmt.Errorf("topic has invalid format")
	}

	prefix := strings.ToLower(strings.TrimSpace(parts[0]))
	if prefix != TopicPrefixObject && prefix != TopicPrefixDiagram {
		return "", uuid.Nil, fmt.Errorf("topic prefix is unsupported")
	}

	entityID, parseError := uuid.Parse(parts[1])
	if parseError != nil {
		return "", uuid.Nil, fmt.Errorf("parse group topic: %w", parseError)
	}

	return prefix, entityID, nil
}

func IsAlarmsTopic(topic string) bool {
	return strings.EqualFold(strings.TrimSpace(topic), TopicAlarms)
}

func BuildWelcomeMessage(sessionID string, maxSubscriptions int, debounceMS int) WelcomeMessage {
	return WelcomeMessage{
		T:          ServerTypeWelcome,
		SessionID:  strings.TrimSpace(sessionID),
		ServerTime: time.Now().UTC().Format(time.RFC3339Nano),
		Limits: WelcomeLimits{
			MaxSubscriptions: maxSubscriptions,
			DebounceMS:       debounceMS,
		},
	}
}

func BuildSubscribedMessage(topics []string) TopicsMessage {
	return TopicsMessage{
		T:      ServerTypeSubscribed,
		Topics: normalizeTopics(topics),
	}
}

func BuildUnsubscribedMessage(topics []string) TopicsMessage {
	return TopicsMessage{
		T:      ServerTypeUnsubscribed,
		Topics: normalizeTopics(topics),
	}
}

func BuildValueMessage(record domain.IngestRecord) ValueMessage {
	return ValueMessage{
		T:     ServerTypeValue,
		TagID: record.TagID.String(),
		TS:    record.Timestamp.UTC().Format(time.RFC3339Nano),
		V:     record.Value,
		Q:     string(record.Quality),
	}
}

func BuildSnapshotMessage(record domain.IngestRecord, reason SnapshotReason) SnapshotMessage {
	return SnapshotMessage{
		T:      ServerTypeSnapshot,
		TagID:  record.TagID.String(),
		TS:     record.Timestamp.UTC().Format(time.RFC3339Nano),
		V:      record.Value,
		Q:      string(record.Quality),
		Reason: reason,
	}
}

func BuildPongMessage() PongMessage {
	return PongMessage{T: ServerTypePong}
}

func BuildErrorMessage(code string, detail string, topics []string) ErrorMessage {
	return ErrorMessage{
		T:      ServerTypeError,
		Code:   strings.TrimSpace(code),
		Detail: strings.TrimSpace(detail),
		Topics: normalizeTopics(topics),
	}
}

func BuildConfigChangedMessage(event domain.ConfigChangedEvent) ConfigChangedMessage {
	return ConfigChangedMessage{
		T:          ServerTypeConfigChanged,
		EntityType: strings.TrimSpace(event.EntityType),
		EntityID:   strings.TrimSpace(event.EntityID),
		Operation:  strings.TrimSpace(event.Operation),
		Timestamp:  event.Timestamp.UTC().Format(time.RFC3339Nano),
		Payload:    event.Payload,
	}
}

func BuildAlarmMessage(event domain.AlarmEvent) AlarmMessage {
	return AlarmMessage{
		T:         ServerTypeAlarm,
		EventType: event.EventType.String(),
		TagID:     event.TagID.String(),
		StateFrom: event.StateFrom.String(),
		StateTo:   event.StateTo.String(),
		Value:     event.Value,
		Quality:   string(event.Quality),
		TS:        event.TS.UTC().Format(time.RFC3339Nano),
		ActorID:   event.ActorID,
		Note:      event.Note,
	}
}

func BuildAlarmsBatchMessage(events []domain.AlarmEvent) AlarmsBatchMessage {
	messages := make([]AlarmMessage, 0, len(events))
	for _, event := range events {
		messages = append(messages, BuildAlarmMessage(event))
	}

	return AlarmsBatchMessage{
		T:      ServerTypeAlarmsBatch,
		TS:     time.Now().UTC().Format(time.RFC3339Nano),
		Events: messages,
	}
}

func BuildAlarmsSnapshotMessage(records []domain.AlarmStateRecord) AlarmsSnapshotMessage {
	items := make([]AlarmRecordMessage, 0, len(records))
	for _, record := range records {
		var ack *AlarmAckMessage
		if record.Ack != nil {
			ack = &AlarmAckMessage{
				ActorID: record.Ack.ActorID,
				AckedAt: record.Ack.AckedAt.UTC().Format(time.RFC3339Nano),
				Note:    record.Ack.Note,
			}
		}

		items = append(items, AlarmRecordMessage{
			TagID:      record.TagID.String(),
			TagName:    record.TagName,
			DeviceID:   record.DeviceID.String(),
			DeviceName: record.DeviceName,
			ObjectID:   record.ObjectID.String(),
			ObjectName: record.ObjectName,
			State:      record.State.String(),
			Value:      record.LastValue,
			Quality:    string(record.LastQuality),
			EnteredAt:  record.EnteredAt.UTC().Format(time.RFC3339Nano),
			LastSeenAt: record.LastSeenAt.UTC().Format(time.RFC3339Nano),
			Acked:      record.Ack != nil,
			Ack:        ack,
		})
	}

	return AlarmsSnapshotMessage{
		T:     ServerTypeAlarmsSnapshot,
		TS:    time.Now().UTC().Format(time.RFC3339Nano),
		Items: items,
	}
}

func BuildTopicsChangedMessage(topic string, addedCount int, removedCount int) TopicsChangedMessage {
	return TopicsChangedMessage{
		T:            ServerTypeTopicsChanged,
		Topic:        strings.TrimSpace(topic),
		AddedCount:   addedCount,
		RemovedCount: removedCount,
	}
}

func normalizeTopics(topics []string) []string {
	normalized := make([]string, 0, len(topics))
	seen := make(map[string]struct{}, len(topics))
	for _, topic := range topics {
		topic = strings.TrimSpace(topic)
		if topic == "" {
			continue
		}
		if _, exists := seen[topic]; exists {
			continue
		}

		seen[topic] = struct{}{}
		normalized = append(normalized, topic)
	}

	return normalized
}
