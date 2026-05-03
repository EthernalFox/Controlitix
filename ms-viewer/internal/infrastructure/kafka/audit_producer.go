package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
	segmentkafka "github.com/segmentio/kafka-go"
)

type AuditProducer struct {
	writer *segmentkafka.Writer
	topic  string
	logger *slog.Logger

	queue chan domain.AuditRecord
	done  chan struct{}

	closeOnce sync.Once
}

func NewAuditProducer(
	brokers []string,
	topic string,
	bufferSize int,
	logger *slog.Logger,
) *AuditProducer {
	if logger == nil {
		logger = slog.Default()
	}
	if bufferSize <= 0 {
		bufferSize = 1000
	}

	filteredBrokers := make([]string, 0, len(brokers))
	for _, broker := range brokers {
		trimmed := strings.TrimSpace(broker)
		if trimmed != "" {
			filteredBrokers = append(filteredBrokers, trimmed)
		}
	}

	producer := &AuditProducer{
		topic: strings.TrimSpace(topic),
		logger: logger,
		queue: make(chan domain.AuditRecord, bufferSize),
		done:  make(chan struct{}),
	}

	if len(filteredBrokers) > 0 && producer.topic != "" {
		producer.writer = &segmentkafka.Writer{
			Addr:  segmentkafka.TCP(filteredBrokers...),
			Topic: producer.topic,
		}
		go producer.run()
	} else {
		close(producer.done)
	}

	return producer
}

func (producer *AuditProducer) Publish(_ context.Context, record domain.AuditRecord) {
	if producer == nil {
		return
	}
	if producer.writer == nil {
		producer.logger.Debug(
			"audit producer disabled",
			"method",
			"AuditProducer.Publish",
			"action",
			record.Action,
		)
		return
	}

	select {
	case producer.queue <- record:
		return
	default:
	}

	select {
	case <-producer.queue:
	default:
	}

	select {
	case producer.queue <- record:
		producer.logger.Warn(
			"audit.dropped",
			"method",
			"AuditProducer.Publish",
			"action",
			record.Action,
			"reason",
			"buffer_overflow_drop_oldest",
		)
	default:
		producer.logger.Warn(
			"audit.dropped",
			"method",
			"AuditProducer.Publish",
			"action",
			record.Action,
			"reason",
			"buffer_overflow_drop_new",
		)
	}
}

func (producer *AuditProducer) Close() error {
	if producer == nil {
		return nil
	}

	producer.closeOnce.Do(func() {
		if producer.writer != nil {
			close(producer.queue)
			<-producer.done
		}
	})

	if producer.writer == nil {
		return nil
	}

	if closeError := producer.writer.Close(); closeError != nil {
		return fmt.Errorf("close audit writer: %w", closeError)
	}
	return nil
}

func (producer *AuditProducer) run() {
	defer close(producer.done)
	for record := range producer.queue {
		payload, marshalError := json.Marshal(record)
		if marshalError != nil {
			producer.logger.Warn(
				"failed to marshal audit payload",
				"method",
				"AuditProducer.run",
				"action",
				record.Action,
				"error",
				marshalError,
			)
			continue
		}

		key := "system"
		if record.ActorID != nil && strings.TrimSpace(*record.ActorID) != "" {
			key = strings.TrimSpace(*record.ActorID)
		}

		writeError := producer.writer.WriteMessages(context.Background(), segmentkafka.Message{
			Key:   []byte(key),
			Value: payload,
		})
		if writeError != nil {
			producer.logger.Warn(
				"failed to publish audit message",
				"method",
				"AuditProducer.run",
				"action",
				record.Action,
				"error",
				writeError,
			)
		}
	}
}
