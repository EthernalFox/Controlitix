package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
)

type LastValueCache interface {
	SetLastValue(ctx context.Context, record domain.IngestRecord) error
}

type RealtimeDispatcher interface {
	Publish(record domain.IngestRecord)
}

type AlarmIngestUseCase interface {
	IngestValue(ctx context.Context, record domain.IngestRecord) error
}

type IngestOutcome struct {
	Flushed int
}

type IngestUseCase struct {
	repository domain.ValueIngestRepository
	cache      LastValueCache
	dispatcher RealtimeDispatcher
	alarmUseCase AlarmIngestUseCase
	batchSize  int
	logger     *slog.Logger

	mutex  sync.Mutex
	buffer []domain.IngestRecord
}

func NewIngestUseCase(
	repository domain.ValueIngestRepository,
	cache LastValueCache,
	dispatcher RealtimeDispatcher,
	alarmUseCase AlarmIngestUseCase,
	batchSize int,
	logger *slog.Logger,
) *IngestUseCase {
	if logger == nil {
		logger = slog.Default()
	}

	return &IngestUseCase{
		repository: repository,
		cache:      cache,
		dispatcher: dispatcher,
		alarmUseCase: alarmUseCase,
		batchSize:  batchSize,
		logger:     logger,
		buffer:     make([]domain.IngestRecord, 0, batchSize),
	}
}

func (useCase *IngestUseCase) IngestTagValue(
	ctx context.Context,
	record domain.IngestRecord,
) (IngestOutcome, error) {
	useCase.mutex.Lock()
	defer useCase.mutex.Unlock()

	useCase.buffer = append(useCase.buffer, record)
	if len(useCase.buffer) < useCase.batchSize {
		return IngestOutcome{}, nil
	}

	flushed, flushError := useCase.flushLocked(ctx)
	if flushError != nil {
		return IngestOutcome{}, flushError
	}

	return IngestOutcome{Flushed: flushed}, nil
}

func (useCase *IngestUseCase) FlushPending(ctx context.Context) (IngestOutcome, error) {
	useCase.mutex.Lock()
	defer useCase.mutex.Unlock()

	flushed, flushError := useCase.flushLocked(ctx)
	if flushError != nil {
		return IngestOutcome{}, flushError
	}

	return IngestOutcome{Flushed: flushed}, nil
}

func (useCase *IngestUseCase) flushLocked(ctx context.Context) (int, error) {
	if len(useCase.buffer) == 0 {
		return 0, nil
	}

	records := append([]domain.IngestRecord(nil), useCase.buffer...)
	if upsertError := useCase.repository.UpsertRawValues(ctx, records); upsertError != nil {
		return 0, fmt.Errorf("upsert raw values: %w", upsertError)
	}

	useCase.updateLastValueCacheBestEffort(ctx, records)
	if alarmError := useCase.ingestAlarms(ctx, records); alarmError != nil {
		return 0, alarmError
	}
	useCase.publishRealtimeBestEffort(records)
	flushed := len(useCase.buffer)
	useCase.buffer = useCase.buffer[:0]
	return flushed, nil
}

func (useCase *IngestUseCase) updateLastValueCacheBestEffort(
	ctx context.Context,
	records []domain.IngestRecord,
) {
	if useCase.cache == nil {
		return
	}

	latestByTag := make(map[uuid.UUID]domain.IngestRecord, len(records))
	for _, record := range records {
		current, exists := latestByTag[record.TagID]
		if !exists || record.Timestamp.After(current.Timestamp) {
			latestByTag[record.TagID] = record
		}
	}

	for _, record := range latestByTag {
		if cacheError := useCase.cache.SetLastValue(ctx, record); cacheError != nil {
			useCase.logger.Warn(
				"failed to update redis last value cache",
				"method",
				"IngestUseCase.updateLastValueCacheBestEffort",
				"tag_id",
				record.TagID.String(),
				"error",
				cacheError,
			)
		}
	}
}

func (useCase *IngestUseCase) publishRealtimeBestEffort(
	records []domain.IngestRecord,
) {
	if useCase.dispatcher == nil {
		return
	}

	latestByTag := make(map[uuid.UUID]domain.IngestRecord, len(records))
	for _, record := range records {
		current, exists := latestByTag[record.TagID]
		if !exists || record.Timestamp.After(current.Timestamp) {
			latestByTag[record.TagID] = record
		}
	}

	for _, record := range latestByTag {
		useCase.dispatcher.Publish(record)
	}
}

func (useCase *IngestUseCase) ingestAlarms(
	ctx context.Context,
	records []domain.IngestRecord,
) error {
	if useCase.alarmUseCase == nil {
		return nil
	}

	latestByTag := make(map[uuid.UUID]domain.IngestRecord, len(records))
	for _, record := range records {
		current, exists := latestByTag[record.TagID]
		if !exists || record.Timestamp.After(current.Timestamp) {
			latestByTag[record.TagID] = record
		}
	}

	for _, record := range latestByTag {
		if ingestError := useCase.alarmUseCase.IngestValue(ctx, record); ingestError != nil {
			return fmt.Errorf("ingest alarm value: %w", ingestError)
		}
	}

	return nil
}

func IngestFlushInterval(flushMilliseconds int) time.Duration {
	if flushMilliseconds <= 0 {
		return time.Second
	}

	return time.Duration(flushMilliseconds) * time.Millisecond
}
