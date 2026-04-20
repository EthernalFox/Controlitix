package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/EthernalFox/Controlitix/ms-poll/internal/domain"
	segmentkafka "github.com/segmentio/kafka-go"
)

type ConfigChangedHandler func(
	ctx context.Context,
	event domain.ConfigChangedEvent,
) error

type ConfigChangedConsumer struct {
	reader  *segmentkafka.Reader
	topic   string
	groupID string
	logger  *slog.Logger
}

func NewConfigChangedConsumer(
	brokers []string,
	topic string,
	groupID string,
	logger *slog.Logger,
) *ConfigChangedConsumer {
	if logger == nil {
		logger = slog.Default()
	}

	if len(brokers) == 0 {
		return &ConfigChangedConsumer{
			topic:   topic,
			groupID: groupID,
			logger:  logger,
		}
	}

	return &ConfigChangedConsumer{
		reader: segmentkafka.NewReader(segmentkafka.ReaderConfig{
			Brokers: brokers,
			Topic:   topic,
			GroupID: groupID,
		}),
		topic:   topic,
		groupID: groupID,
		logger:  logger,
	}
}

func (consumer *ConfigChangedConsumer) Run(
	ctx context.Context,
	handler ConfigChangedHandler,
) error {
	if consumer.reader == nil {
		return errors.New("kafka reader is not configured")
	}

	if handler == nil {
		return errors.New("config changed handler is not configured")
	}

	consumer.logger.Info(
		"kafka consumer started",
		"topic",
		consumer.topic,
		"group",
		consumer.groupID,
	)
	defer consumer.logger.Info(
		"kafka consumer stopped",
		"topic",
		consumer.topic,
		"group",
		consumer.groupID,
	)

	for {
		message, fetchError := consumer.reader.FetchMessage(ctx)
		if fetchError != nil {
			if errors.Is(fetchError, context.Canceled) || ctx.Err() != nil {
				return nil
			}

			return fmt.Errorf("fetch kafka message: %w", fetchError)
		}

		var event domain.ConfigChangedEvent
		if unmarshalError := json.Unmarshal(message.Value, &event); unmarshalError != nil {
			consumer.logger.Error(
				"failed to decode kafka message",
				"topic",
				consumer.topic,
				"partition",
				message.Partition,
				"offset",
				message.Offset,
				"error",
				unmarshalError,
			)
			continue
		}

		handlerError := handler(ctx, event)
		if handlerError != nil {
			consumer.logger.Error(
				"config.changed handler failed",
				"topic",
				consumer.topic,
				"partition",
				message.Partition,
				"offset",
				message.Offset,
				"error",
				handlerError,
			)
			continue
		}

		commitError := consumer.reader.CommitMessages(ctx, message)
		if commitError != nil {
			if errors.Is(commitError, context.Canceled) || ctx.Err() != nil {
				return nil
			}

			return fmt.Errorf("commit kafka message: %w", commitError)
		}
	}
}

func (consumer *ConfigChangedConsumer) Close() error {
	if consumer.reader == nil {
		return nil
	}

	closeError := consumer.reader.Close()
	if closeError != nil {
		return fmt.Errorf("close kafka reader: %w", closeError)
	}

	return nil
}
