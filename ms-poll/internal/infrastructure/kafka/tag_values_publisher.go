package kafka

import (
	"context"
	"errors"
	"fmt"

	"github.com/EthernalFox/Controlitix/ms-poll/internal/domain"
	segmentkafka "github.com/segmentio/kafka-go"
)

func (producer *TagValuesProducer) Publish(
	ctx context.Context,
	readings []domain.Reading,
) error {
	if len(readings) == 0 {
		return nil
	}
	if producer.writer == nil {
		return errors.New("kafka writer is not configured")
	}

	messages := make([]segmentkafka.Message, 0, len(readings))
	for _, reading := range readings {
		payload, serializeError := serializeTagValue(reading)
		if serializeError != nil {
			return serializeError
		}

		messages = append(messages, segmentkafka.Message{
			Key:   []byte(reading.TagID),
			Value: payload,
			Time:  reading.ReadAt.UTC(),
		})
	}

	writeError := producer.writer.WriteMessages(ctx, messages...)
	if writeError != nil {
		return fmt.Errorf(
			"publish readings to topic %s: %w",
			producer.topic,
			writeError,
		)
	}

	return nil
}
