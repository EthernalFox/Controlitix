package kafka

import (
	"fmt"
	"log/slog"

	segmentkafka "github.com/segmentio/kafka-go"
)

type TagValuesProducer struct {
	writer *segmentkafka.Writer
	topic  string
	logger *slog.Logger
}

func NewTagValuesProducer(
	brokers []string,
	topic string,
	logger *slog.Logger,
) *TagValuesProducer {
	if logger == nil {
		logger = slog.Default()
	}

	if len(brokers) == 0 {
		return &TagValuesProducer{
			topic:  topic,
			logger: logger,
		}
	}

	return &TagValuesProducer{
		writer: &segmentkafka.Writer{
			Addr:  segmentkafka.TCP(brokers...),
			Topic: topic,
		},
		topic:  topic,
		logger: logger,
	}
}

func (producer *TagValuesProducer) Close() error {
	if producer.writer == nil {
		return nil
	}

	closeError := producer.writer.Close()
	if closeError != nil {
		return fmt.Errorf("close kafka writer: %w", closeError)
	}

	return nil
}
