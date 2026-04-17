package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/EthernalFox/Controlitix/ms-editor/internal/domain"
	segmentkafka "github.com/segmentio/kafka-go"
)

type KafkaEventPublisher struct {
	writer *segmentkafka.Writer
	topic  string
	logger *slog.Logger
}

func NewKafkaEventPublisher(
	brokers []string,
	topic string,
	logger *slog.Logger,
) *KafkaEventPublisher {
	if logger == nil {
		logger = slog.Default()
	}

	if len(brokers) == 0 {
		return &KafkaEventPublisher{
			topic:  topic,
			logger: logger,
		}
	}

	return &KafkaEventPublisher{
		writer: &segmentkafka.Writer{
			Addr:  segmentkafka.TCP(brokers...),
			Topic: topic,
		},
		topic:  topic,
		logger: logger,
	}
}

func (publisher *KafkaEventPublisher) Publish(
	ctx context.Context,
	event domain.ConfigChangedEvent,
) error {
	if publisher.writer == nil {
		return errors.New("kafka writer is not configured")
	}

	eventValue, marshalError := json.Marshal(event)
	if marshalError != nil {
		publisher.logger.Error(
			"failed to marshal kafka event",
			"topic",
			publisher.topic,
			"entity_type",
			event.EntityType,
			"entity_id",
			event.EntityID,
			"error",
			marshalError,
		)
		return fmt.Errorf("marshal kafka event: %w", marshalError)
	}

	message := segmentkafka.Message{
		Key:   []byte(event.EntityType + ":" + event.EntityID),
		Value: eventValue,
	}

	if publishError := publisher.writer.WriteMessages(ctx, message); publishError != nil {
		publisher.logger.Error(
			"failed to publish kafka event",
			"topic",
			publisher.topic,
			"entity_type",
			event.EntityType,
			"entity_id",
			event.EntityID,
			"error",
			publishError,
		)
		return fmt.Errorf("publish kafka event: %w", publishError)
	}

	return nil
}

func (publisher *KafkaEventPublisher) Close() error {
	if publisher.writer == nil {
		return nil
	}

	if closeError := publisher.writer.Close(); closeError != nil {
		return fmt.Errorf("close kafka writer: %w", closeError)
	}

	return nil
}
