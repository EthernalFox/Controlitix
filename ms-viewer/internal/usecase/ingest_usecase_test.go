package usecase

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
)

type fakeIngestRepository struct {
	store map[string]domain.IngestRecord
}

func (repository *fakeIngestRepository) UpsertRawValues(
	_ context.Context,
	records []domain.IngestRecord,
) error {
	if repository.store == nil {
		repository.store = make(map[string]domain.IngestRecord)
	}

	for _, record := range records {
		key := fmt.Sprintf("%s:%s", record.TagID.String(), record.Timestamp.UTC().Format(time.RFC3339Nano))
		repository.store[key] = record
	}

	return nil
}

type fakeLastValueCache struct {
	values map[string]domain.IngestRecord
}

func (cache *fakeLastValueCache) SetLastValue(
	_ context.Context,
	record domain.IngestRecord,
) error {
	if cache.values == nil {
		cache.values = make(map[string]domain.IngestRecord)
	}
	cache.values[record.TagID.String()] = record
	return nil
}

func TestIngestUseCaseFlushesByBatchSize(t *testing.T) {
	repository := &fakeIngestRepository{}
	cache := &fakeLastValueCache{}
	ingestUseCase := NewIngestUseCase(repository, cache, nil, nil, 2, nil)

	tagID := uuid.New()
	now := time.Now().UTC()

	outcome, ingestError := ingestUseCase.IngestTagValue(context.Background(), domain.IngestRecord{
		TagID:     tagID,
		Timestamp: now,
		Quality:   domain.QualityOK,
	})
	if ingestError != nil {
		t.Fatalf("IngestTagValue returned error: %v", ingestError)
	}
	if outcome.Flushed != 0 {
		t.Fatalf("expected no flush on first record, got %d", outcome.Flushed)
	}

	outcome, ingestError = ingestUseCase.IngestTagValue(context.Background(), domain.IngestRecord{
		TagID:     tagID,
		Timestamp: now.Add(time.Second),
		Quality:   domain.QualityOK,
	})
	if ingestError != nil {
		t.Fatalf("IngestTagValue returned error: %v", ingestError)
	}
	if outcome.Flushed != 2 {
		t.Fatalf("expected 2 flushed records, got %d", outcome.Flushed)
	}
	if len(repository.store) != 2 {
		t.Fatalf("expected 2 persisted records, got %d", len(repository.store))
	}
	if _, ok := cache.values[tagID.String()]; !ok {
		t.Fatalf("expected last value cache to be updated")
	}
}

func TestIngestUseCaseIsIdempotentOnDuplicateTimestamp(t *testing.T) {
	repository := &fakeIngestRepository{}
	cache := &fakeLastValueCache{}
	ingestUseCase := NewIngestUseCase(repository, cache, nil, nil, 10, nil)

	tagID := uuid.New()
	timestamp := time.Now().UTC()
	value1 := 10.0
	value2 := 20.0

	if _, ingestError := ingestUseCase.IngestTagValue(context.Background(), domain.IngestRecord{
		TagID:     tagID,
		Timestamp: timestamp,
		Value:     &value1,
		Quality:   domain.QualityOK,
	}); ingestError != nil {
		t.Fatalf("IngestTagValue returned error: %v", ingestError)
	}
	if _, ingestError := ingestUseCase.IngestTagValue(context.Background(), domain.IngestRecord{
		TagID:     tagID,
		Timestamp: timestamp,
		Value:     &value2,
		Quality:   domain.QualityHi,
	}); ingestError != nil {
		t.Fatalf("IngestTagValue returned error: %v", ingestError)
	}

	outcome, flushError := ingestUseCase.FlushPending(context.Background())
	if flushError != nil {
		t.Fatalf("FlushPending returned error: %v", flushError)
	}
	if outcome.Flushed != 2 {
		t.Fatalf("expected 2 flushed records, got %d", outcome.Flushed)
	}

	if len(repository.store) != 1 {
		t.Fatalf("expected one unique persisted key, got %d", len(repository.store))
	}

	key := fmt.Sprintf("%s:%s", tagID.String(), timestamp.Format(time.RFC3339Nano))
	persisted := repository.store[key]
	if persisted.Value == nil || *persisted.Value != value2 {
		t.Fatalf("expected duplicate upsert to keep last value %v, got %#v", value2, persisted.Value)
	}
	if persisted.Quality != domain.QualityHi {
		t.Fatalf("expected duplicate upsert to keep last quality %q, got %q", domain.QualityHi, persisted.Quality)
	}
}
