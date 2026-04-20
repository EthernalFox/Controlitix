package kafka

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/EthernalFox/Controlitix/ms-poll/internal/domain"
	"github.com/EthernalFox/Controlitix/ms-poll/internal/usecase"
	segmentkafka "github.com/segmentio/kafka-go"
)

type RawReadingsPublisher struct {
	writer *segmentkafka.Writer
	topic  string
	logger *slog.Logger
}

func NewRawReadingsPublisher(
	brokers []string,
	topic string,
	logger *slog.Logger,
) *RawReadingsPublisher {
	if logger == nil {
		logger = slog.Default()
	}

	if len(brokers) == 0 {
		return &RawReadingsPublisher{
			topic:  topic,
			logger: logger,
		}
	}

	return &RawReadingsPublisher{
		writer: &segmentkafka.Writer{
			Addr:  segmentkafka.TCP(brokers...),
			Topic: topic,
		},
		topic:  topic,
		logger: logger,
	}
}

func (publisher *RawReadingsPublisher) Publish(
	ctx context.Context,
	readings []domain.Reading,
) error {
	if len(readings) == 0 {
		return nil
	}
	if publisher.writer == nil {
		return errors.New("kafka writer is not configured")
	}

	messages := make([]segmentkafka.Message, 0, len(readings))
	for _, reading := range readings {
		payload, serializeError := serializeRawReading(reading)
		if serializeError != nil {
			return serializeError
		}

		messages = append(messages, segmentkafka.Message{
			Key:   []byte(reading.TagID),
			Value: payload,
			Time:  reading.ReadAt.UTC(),
		})
	}

	writeError := publisher.writer.WriteMessages(ctx, messages...)
	if writeError != nil {
		return fmt.Errorf(
			"publish raw readings to topic %s: %w",
			publisher.topic,
			writeError,
		)
	}

	return nil
}

func (publisher *RawReadingsPublisher) Close() error {
	if publisher.writer == nil {
		return nil
	}

	closeError := publisher.writer.Close()
	if closeError != nil {
		return fmt.Errorf("close kafka raw readings writer: %w", closeError)
	}

	return nil
}

type compositePublisher struct {
	primary    usecase.TagValuesPublisher
	bestEffort usecase.TagValuesPublisher
	logger     *slog.Logger
}

func NewCompositePublisher(
	primary usecase.TagValuesPublisher,
	bestEffort usecase.TagValuesPublisher,
	logger *slog.Logger,
) usecase.TagValuesPublisher {
	if logger == nil {
		logger = slog.Default()
	}

	return &compositePublisher{
		primary:    primary,
		bestEffort: bestEffort,
		logger:     logger,
	}
}

func (publisher *compositePublisher) Publish(
	ctx context.Context,
	readings []domain.Reading,
) error {
	if len(readings) == 0 {
		return nil
	}
	if publisher.primary == nil {
		return errors.New("primary publisher is not configured")
	}

	primaryError := publisher.primary.Publish(ctx, readings)
	if primaryError != nil {
		return primaryError
	}

	if publisher.bestEffort == nil {
		return nil
	}

	bestEffortError := publisher.bestEffort.Publish(ctx, readings)
	if bestEffortError != nil {
		publisher.logger.Warn(
			"best-effort publisher failed",
			"method",
			"compositePublisher.Publish",
			"error",
			bestEffortError,
		)
	}

	return nil
}
