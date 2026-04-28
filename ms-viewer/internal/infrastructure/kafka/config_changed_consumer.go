package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
)

type ConfigChangeHandler interface {
	HandleConfigChanged(ctx context.Context, event domain.ConfigChangedEvent) error
}

type ConfigChangedConsumer struct {
	consumer *Consumer
	handler  ConfigChangeHandler
	logger   *slog.Logger
}

func NewConfigChangedConsumer(
	consumer *Consumer,
	handler ConfigChangeHandler,
	logger *slog.Logger,
) *ConfigChangedConsumer {
	if logger == nil {
		logger = slog.Default()
	}

	return &ConfigChangedConsumer{
		consumer: consumer,
		handler:  handler,
		logger:   logger,
	}
}

func (consumer *ConfigChangedConsumer) Run(ctx context.Context) error {
	if consumer.consumer == nil {
		return errors.New("kafka consumer is not configured")
	}
	if consumer.handler == nil {
		return errors.New("config change handler is not configured")
	}

	consumer.logger.Info(
		"kafka config.changed consumer started",
		"method",
		"ConfigChangedConsumer.Run",
	)
	defer consumer.logger.Info(
		"kafka config.changed consumer stopped",
		"method",
		"ConfigChangedConsumer.Run",
	)

	for {
		message, fetchError := consumer.consumer.FetchMessage(ctx)
		if fetchError != nil {
			if errors.Is(fetchError, context.Canceled) || errors.Is(fetchError, context.DeadlineExceeded) {
				return nil
			}
			return fetchError
		}

		event, parseError := parseConfigChangedMessage(message.Value)
		if parseError != nil {
			consumer.logger.Warn(
				"failed to parse config.changed message",
				"method",
				"ConfigChangedConsumer.Run",
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

		handleError := consumer.handler.HandleConfigChanged(ctx, event)
		if handleError != nil {
			consumer.logger.Error(
				"failed to handle config.changed event",
				"method",
				"ConfigChangedConsumer.Run",
				"entity_type",
				event.EntityType,
				"entity_id",
				event.EntityID,
				"operation",
				event.Operation,
				"error",
				handleError,
			)
			return fmt.Errorf("handle config.changed event: %w", handleError)
		}

		if commitError := consumer.consumer.CommitMessages(ctx, message); commitError != nil {
			return commitError
		}
	}
}

func parseConfigChangedMessage(payload []byte) (domain.ConfigChangedEvent, error) {
	var event domain.ConfigChangedEvent
	if unmarshalError := json.Unmarshal(payload, &event); unmarshalError != nil {
		return domain.ConfigChangedEvent{}, fmt.Errorf("decode json: %w", unmarshalError)
	}

	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	return event, nil
}
