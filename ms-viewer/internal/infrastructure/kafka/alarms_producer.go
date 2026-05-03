package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	segmentkafka "github.com/segmentio/kafka-go"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
)

type AlarmsProducer struct {
	writer *segmentkafka.Writer
	topic  string
	logger *slog.Logger

	retryMutex sync.Mutex
	retryQueue []segmentkafka.Message
}

type alarmEventPayload struct {
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

func NewAlarmsProducer(brokers []string, topic string, logger *slog.Logger) *AlarmsProducer {
	if logger == nil {
		logger = slog.Default()
	}

	if len(brokers) == 0 {
		return &AlarmsProducer{
			topic:  topic,
			logger: logger,
		}
	}

	return &AlarmsProducer{
		writer: &segmentkafka.Writer{
			Addr:  segmentkafka.TCP(brokers...),
			Topic: topic,
		},
		topic:  topic,
		logger: logger,
	}
}

func (producer *AlarmsProducer) Publish(ctx context.Context, event domain.AlarmEvent) error {
	return producer.PublishBatch(ctx, []domain.AlarmEvent{event})
}

func (producer *AlarmsProducer) PublishBatch(ctx context.Context, events []domain.AlarmEvent) error {
	if producer.writer == nil {
		return errors.New("kafka writer is not configured")
	}
	if len(events) == 0 {
		return nil
	}

	messages := make([]segmentkafka.Message, 0, len(events))
	for _, event := range events {
		message, buildError := producer.buildMessage(event)
		if buildError != nil {
			return buildError
		}
		messages = append(messages, message)
	}

	return producer.writeWithRetryQueue(ctx, messages)
}

func (producer *AlarmsProducer) buildMessage(event domain.AlarmEvent) (segmentkafka.Message, error) {
	payload, marshalError := json.Marshal(alarmEventPayload{
		ID:        event.ID.String(),
		TagID:     event.TagID.String(),
		EventType: event.EventType.String(),
		StateFrom: event.StateFrom.String(),
		StateTo:   event.StateTo.String(),
		Value:     event.Value,
		Quality:   string(event.Quality),
		TS:        event.TS.UTC().Format(time.RFC3339Nano),
		ActorID:   event.ActorID,
		Note:      event.Note,
	})
	if marshalError != nil {
		return segmentkafka.Message{}, fmt.Errorf("marshal alarm event: %w", marshalError)
	}

	return segmentkafka.Message{
		Key:   []byte(event.TagID.String()),
		Value: payload,
	}, nil
}

func (producer *AlarmsProducer) writeWithRetryQueue(
	ctx context.Context,
	messages []segmentkafka.Message,
) error {
	producer.retryMutex.Lock()
	payload := append(append([]segmentkafka.Message{}, producer.retryQueue...), messages...)
	producer.retryQueue = nil
	producer.retryMutex.Unlock()

	if len(payload) == 0 {
		return nil
	}

	if publishError := producer.writer.WriteMessages(ctx, payload...); publishError != nil {
		producer.retryMutex.Lock()
		producer.retryQueue = append(producer.retryQueue, payload...)
		producer.retryMutex.Unlock()

		producer.logger.Error(
			"failed to publish alarm events batch",
			"method",
			"AlarmsProducer.writeWithRetryQueue",
			"events_n",
			len(payload),
			"error",
			publishError,
		)

		return fmt.Errorf("publish alarm events batch: %w", publishError)
	}

	return nil
}

func (producer *AlarmsProducer) Close() error {
	if producer.writer == nil {
		return nil
	}
	if closeError := producer.writer.Close(); closeError != nil {
		return fmt.Errorf("close alarms writer: %w", closeError)
	}
	return nil
}
