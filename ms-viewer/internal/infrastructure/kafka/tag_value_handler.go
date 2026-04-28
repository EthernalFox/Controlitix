package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	segmentkafka "github.com/segmentio/kafka-go"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
	"github.com/EthernalFox/Controlitix/ms-viewer/internal/usecase"
)

type TagValueHandler struct {
	consumer      *Consumer
	ingestUseCase *usecase.IngestUseCase
	flushInterval time.Duration
	logger        *slog.Logger
}

type kafkaTagValueMessage struct {
	TagID string   `json:"tag_id"`
	TS    string   `json:"ts"`
	V     *float64 `json:"v"`
	Q     string   `json:"q"`
}

type fetchedMessage struct {
	message segmentkafka.Message
	err     error
}

func NewTagValueHandler(
	consumer *Consumer,
	ingestUseCase *usecase.IngestUseCase,
	flushInterval time.Duration,
	logger *slog.Logger,
) *TagValueHandler {
	if logger == nil {
		logger = slog.Default()
	}
	if flushInterval <= 0 {
		flushInterval = time.Second
	}

	return &TagValueHandler{
		consumer:      consumer,
		ingestUseCase: ingestUseCase,
		flushInterval: flushInterval,
		logger:        logger,
	}
}

func (handler *TagValueHandler) Run(ctx context.Context) error {
	if handler.consumer == nil {
		return errors.New("kafka consumer is not configured")
	}
	if handler.ingestUseCase == nil {
		return errors.New("ingest use case is not configured")
	}

	handler.logger.Info(
		"kafka consumer started",
		"method",
		"TagValueHandler.Run",
	)
	defer handler.logger.Info(
		"kafka consumer stopped",
		"method",
		"TagValueHandler.Run",
	)

	messagesChannel := make(chan fetchedMessage, 1)
	go handler.fetchMessages(ctx, messagesChannel)

	ticker := time.NewTicker(handler.flushInterval)
	defer ticker.Stop()

	pendingMessages := make([]segmentkafka.Message, 0)
	for {
		select {
		case <-ctx.Done():
			flushed, flushError := handler.flushWithRetry(ctx)
			if flushError != nil {
				return flushError
			}
			if commitError := handler.commitFirstPending(ctx, &pendingMessages, flushed); commitError != nil {
				return commitError
			}
			return nil
		case <-ticker.C:
			flushed, flushError := handler.flushWithRetry(ctx)
			if flushError != nil {
				return flushError
			}
			if commitError := handler.commitFirstPending(ctx, &pendingMessages, flushed); commitError != nil {
				return commitError
			}
		case fetched, ok := <-messagesChannel:
			if !ok {
				return nil
			}
			if fetched.err != nil {
				if errors.Is(fetched.err, context.Canceled) || errors.Is(fetched.err, context.DeadlineExceeded) {
					continue
				}
				return fetched.err
			}

			record, parseError := parseTagValueMessage(fetched.message.Value)
			if parseError != nil {
				handler.logger.Error(
					"failed to parse tags.values message",
					"method",
					"TagValueHandler.Run",
					"offset",
					fetched.message.Offset,
					"partition",
					fetched.message.Partition,
					"error",
					parseError,
				)
				if commitError := handler.consumer.CommitMessages(ctx, fetched.message); commitError != nil {
					return commitError
				}
				continue
			}

			pendingMessages = append(pendingMessages, fetched.message)
			outcome, ingestError := handler.ingestUseCase.IngestTagValue(ctx, record)
			if ingestError != nil {
				handler.logger.Error(
					"failed to flush ingest batch",
					"method",
					"TagValueHandler.Run",
					"error",
					ingestError,
				)
				flushed, flushError := handler.flushWithRetry(ctx)
				if flushError != nil {
					return flushError
				}
				if commitError := handler.commitFirstPending(ctx, &pendingMessages, flushed); commitError != nil {
					return commitError
				}
				continue
			}

			if commitError := handler.commitFirstPending(ctx, &pendingMessages, outcome.Flushed); commitError != nil {
				return commitError
			}
		}
	}
}

func (handler *TagValueHandler) fetchMessages(
	ctx context.Context,
	messagesChannel chan<- fetchedMessage,
) {
	defer close(messagesChannel)

	for {
		message, fetchError := handler.consumer.FetchMessage(ctx)
		if fetchError != nil {
			select {
			case <-ctx.Done():
				return
			case messagesChannel <- fetchedMessage{err: fetchError}:
			}
			return
		}

		select {
		case <-ctx.Done():
			return
		case messagesChannel <- fetchedMessage{message: message}:
		}
	}
}

func (handler *TagValueHandler) flushWithRetry(
	ctx context.Context,
) (int, error) {
	if handler.ingestUseCase == nil {
		return 0, errors.New("ingest use case is not configured")
	}

	backoff := handler.flushInterval
	maxBackoff := handler.flushInterval * 10

	for {
		outcome, flushError := handler.ingestUseCase.FlushPending(ctx)
		if flushError == nil {
			return outcome.Flushed, nil
		}

		handler.logger.Error(
			"failed to flush ingest buffer, retrying",
			"method",
			"TagValueHandler.flushWithRetry",
			"error",
			flushError,
		)

		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-time.After(backoff):
		}

		backoff *= 2
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
	}
}

func (handler *TagValueHandler) commitFirstPending(
	ctx context.Context,
	pendingMessages *[]segmentkafka.Message,
	count int,
) error {
	if count <= 0 {
		return nil
	}
	if count > len(*pendingMessages) {
		return fmt.Errorf("commit count %d exceeds pending messages %d", count, len(*pendingMessages))
	}

	messagesToCommit := append([]segmentkafka.Message(nil), (*pendingMessages)[:count]...)
	if commitError := handler.consumer.CommitMessages(ctx, messagesToCommit...); commitError != nil {
		return commitError
	}

	*pendingMessages = (*pendingMessages)[count:]
	return nil
}

func parseTagValueMessage(payload []byte) (domain.IngestRecord, error) {
	var message kafkaTagValueMessage
	if unmarshalError := json.Unmarshal(payload, &message); unmarshalError != nil {
		return domain.IngestRecord{}, fmt.Errorf("decode json: %w", unmarshalError)
	}

	tagID, parseUUIDError := uuid.Parse(message.TagID)
	if parseUUIDError != nil {
		return domain.IngestRecord{}, fmt.Errorf("parse tag_id: %w", parseUUIDError)
	}

	timestamp, parseTimeError := time.Parse(time.RFC3339Nano, message.TS)
	if parseTimeError != nil {
		return domain.IngestRecord{}, fmt.Errorf("parse ts: %w", parseTimeError)
	}

	quality, parseQualityError := domain.ParseQuality(message.Q)
	if parseQualityError != nil {
		return domain.IngestRecord{}, fmt.Errorf("parse quality: %w", parseQualityError)
	}

	return domain.IngestRecord{
		TagID:     tagID,
		Timestamp: timestamp.UTC(),
		Value:     message.V,
		Quality:   quality,
	}, nil
}
