package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/EthernalFox/Controlitix/ms-auth/internal/domain"
	segmentkafka "github.com/segmentio/kafka-go"
)

type AuditProducer struct {
	writer *segmentkafka.Writer
	topic  string
	logger *slog.Logger
}

func NewAuditProducer(
	brokers []string,
	topic string,
	logger *slog.Logger,
) (*AuditProducer, error) {
	if logger == nil {
		logger = slog.Default()
	}

	if len(brokers) == 0 {
		return nil, nil
	}

	filteredBrokers := make([]string, 0, len(brokers))
	for _, broker := range brokers {
		broker = strings.TrimSpace(broker)
		if broker == "" {
			continue
		}
		filteredBrokers = append(filteredBrokers, broker)
	}
	if len(filteredBrokers) == 0 {
		return nil, nil
	}

	topic = strings.TrimSpace(topic)
	if topic == "" {
		return nil, errors.New("kafka topic is empty")
	}

	return &AuditProducer{
		writer: &segmentkafka.Writer{
			Addr:  segmentkafka.TCP(filteredBrokers...),
			Topic: topic,
		},
		topic:  topic,
		logger: logger,
	}, nil
}

func (producer *AuditProducer) Publish(ctx context.Context, event domain.AuditEvent) error {
	if producer == nil || producer.writer == nil {
		return nil
	}

	payload := struct {
		SchemaVersion string `json:"schema_version"`
		domain.AuditEvent
	}{
		SchemaVersion: "1",
		AuditEvent:    event,
	}

	value, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal audit event for kafka: %w", err)
	}

	key := event.ActorSubject
	if strings.TrimSpace(key) == "" {
		key = "anonymous"
	}

	if err := producer.writer.WriteMessages(ctx, segmentkafka.Message{
		Key:   []byte(key),
		Value: value,
	}); err != nil {
		return fmt.Errorf("write audit kafka message: %w", err)
	}

	return nil
}

func (producer *AuditProducer) Close() error {
	if producer == nil || producer.writer == nil {
		return nil
	}

	if err := producer.writer.Close(); err != nil {
		return fmt.Errorf("close kafka writer: %w", err)
	}

	return nil
}
