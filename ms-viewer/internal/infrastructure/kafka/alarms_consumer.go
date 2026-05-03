package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
)

type AlarmEventHandler interface {
	RouteAndSend(ctx context.Context, event domain.AlarmEvent) error
}

type AlarmsConsumer struct {
	consumer *Consumer
	handler  AlarmEventHandler
	logger   *slog.Logger
}

type alarmEventMessage struct {
	ID        string   `json:"id"`
	TagID     string   `json:"tag_id"`
	EventType string   `json:"event_type"`
	StateFrom string   `json:"state_from"`
	StateTo   string   `json:"state_to"`
	Value     *float64 `json:"value"`
	Quality   string   `json:"quality"`
	TS        string   `json:"ts"`
	ActorID   *string  `json:"actor_id"`
	Note      *string  `json:"note"`
}

func NewAlarmsConsumer(
	consumer *Consumer,
	handler AlarmEventHandler,
	logger *slog.Logger,
) *AlarmsConsumer {
	if logger == nil {
		logger = slog.Default()
	}
	return &AlarmsConsumer{consumer: consumer, handler: handler, logger: logger}
}

func (consumer *AlarmsConsumer) Run(ctx context.Context) error {
	if consumer.consumer == nil {
		return errors.New("kafka consumer is not configured")
	}
	if consumer.handler == nil {
		return errors.New("alarm event handler is not configured")
	}

	consumer.logger.Info(
		"kafka alarms.events consumer started",
		"method",
		"AlarmsConsumer.Run",
	)
	defer consumer.logger.Info(
		"kafka alarms.events consumer stopped",
		"method",
		"AlarmsConsumer.Run",
	)

	for {
		message, fetchError := consumer.consumer.FetchMessage(ctx)
		if fetchError != nil {
			if errors.Is(fetchError, context.Canceled) || errors.Is(fetchError, context.DeadlineExceeded) {
				return nil
			}
			return fetchError
		}

		event, parseError := parseAlarmEventMessage(message.Value)
		if parseError != nil {
			consumer.logger.Warn(
				"failed to parse alarms.events message",
				"method",
				"AlarmsConsumer.Run",
				"partition",
				message.Partition,
				"offset",
				message.Offset,
				"error",
				parseError,
			)
			if commitError := consumer.consumer.CommitMessages(ctx, message); commitError != nil {
				return commitError
			}
			continue
		}

		handleError := consumer.handler.RouteAndSend(ctx, event)
		if handleError != nil {
			consumer.logger.Error(
				"failed to handle alarm event",
				"method",
				"AlarmsConsumer.Run",
				"alarm_event_id",
				event.ID.String(),
				"tag_id",
				event.TagID.String(),
				"error",
				handleError,
			)
			continue
		}

		if commitError := consumer.consumer.CommitMessages(ctx, message); commitError != nil {
			return commitError
		}
	}
}

func parseAlarmEventMessage(payload []byte) (domain.AlarmEvent, error) {
	var message alarmEventMessage
	if unmarshalError := json.Unmarshal(payload, &message); unmarshalError != nil {
		return domain.AlarmEvent{}, fmt.Errorf("decode json: %w", unmarshalError)
	}

	eventID, parseIDError := uuid.Parse(strings.TrimSpace(message.ID))
	if parseIDError != nil {
		return domain.AlarmEvent{}, fmt.Errorf("parse id: %w", parseIDError)
	}
	tagID, parseTagIDError := uuid.Parse(strings.TrimSpace(message.TagID))
	if parseTagIDError != nil {
		return domain.AlarmEvent{}, fmt.Errorf("parse tag_id: %w", parseTagIDError)
	}
	eventType, parseEventTypeError := parseAlarmEventType(message.EventType)
	if parseEventTypeError != nil {
		return domain.AlarmEvent{}, parseEventTypeError
	}
	stateFrom, parseStateFromError := parseAlarmState(message.StateFrom)
	if parseStateFromError != nil {
		return domain.AlarmEvent{}, fmt.Errorf("parse state_from: %w", parseStateFromError)
	}
	stateTo, parseStateToError := parseAlarmState(message.StateTo)
	if parseStateToError != nil {
		return domain.AlarmEvent{}, fmt.Errorf("parse state_to: %w", parseStateToError)
	}

	timestamp, parseTimeError := time.Parse(time.RFC3339Nano, strings.TrimSpace(message.TS))
	if parseTimeError != nil {
		return domain.AlarmEvent{}, fmt.Errorf("parse ts: %w", parseTimeError)
	}
	quality, parseQualityError := domain.ParseQuality(message.Quality)
	if parseQualityError != nil {
		return domain.AlarmEvent{}, fmt.Errorf("parse quality: %w", parseQualityError)
	}

	return domain.AlarmEvent{
		ID:        eventID,
		TagID:     tagID,
		EventType: eventType,
		StateFrom: stateFrom,
		StateTo:   stateTo,
		Value:     message.Value,
		Quality:   quality,
		TS:        timestamp.UTC(),
		ActorID:   message.ActorID,
		Note:      message.Note,
	}, nil
}

func parseAlarmEventType(value string) (domain.AlarmEventType, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "raised":
		return domain.AlarmEventRaised, nil
	case "cleared":
		return domain.AlarmEventCleared, nil
	case "acked":
		return domain.AlarmEventAcked, nil
	case "suppressed":
		return domain.AlarmEventSuppressed, nil
	case "unsuppressed":
		return domain.AlarmEventUnsuppressed, nil
	default:
		return 0, fmt.Errorf("invalid event_type %q", value)
	}
}

func parseAlarmState(value string) (domain.AlarmState, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "ok":
		return domain.AlarmStateOK, nil
	case "lo":
		return domain.AlarmStateLo, nil
	case "hi":
		return domain.AlarmStateHi, nil
	case "lolo":
		return domain.AlarmStateLoLo, nil
	case "hihi":
		return domain.AlarmStateHiHi, nil
	case "uncertain":
		return domain.AlarmStateUncertain, nil
	case "bad":
		return domain.AlarmStateBad, nil
	case "comm_loss":
		return domain.AlarmStateCommLoss, nil
	case "offline":
		return domain.AlarmStateOffline, nil
	default:
		return 0, fmt.Errorf("invalid state %q", value)
	}
}
