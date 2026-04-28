package kafka

import (
	"context"
	"fmt"
	"log/slog"

	segmentkafka "github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader  *segmentkafka.Reader
	topic   string
	groupID string
	logger  *slog.Logger
}

func NewConsumer(
	brokers []string,
	topic string,
	groupID string,
	logger *slog.Logger,
) (*Consumer, error) {
	if logger == nil {
		logger = slog.Default()
	}
	if len(brokers) == 0 {
		return nil, fmt.Errorf("kafka brokers are not configured")
	}

	return &Consumer{
		reader: segmentkafka.NewReader(segmentkafka.ReaderConfig{
			Brokers: brokers,
			Topic:   topic,
			GroupID: groupID,
		}),
		topic:   topic,
		groupID: groupID,
		logger:  logger,
	}, nil
}

func (consumer *Consumer) FetchMessage(ctx context.Context) (segmentkafka.Message, error) {
	message, fetchError := consumer.reader.FetchMessage(ctx)
	if fetchError != nil {
		return segmentkafka.Message{}, fmt.Errorf("fetch kafka message: %w", fetchError)
	}

	return message, nil
}

func (consumer *Consumer) CommitMessages(
	ctx context.Context,
	messages ...segmentkafka.Message,
) error {
	if len(messages) == 0 {
		return nil
	}

	if commitError := consumer.reader.CommitMessages(ctx, messages...); commitError != nil {
		return fmt.Errorf("commit kafka messages: %w", commitError)
	}

	return nil
}

func (consumer *Consumer) Close() error {
	if closeError := consumer.reader.Close(); closeError != nil {
		return fmt.Errorf("close kafka reader: %w", closeError)
	}

	return nil
}
