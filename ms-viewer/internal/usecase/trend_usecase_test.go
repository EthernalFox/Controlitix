package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
)

type fakeTrendRepository struct {
	tagMeta domain.TagMeta

	rawPoints []domain.TrendPoint
	aggPoints []domain.TrendPoint

	rawCalls int
	aggCalls int
}

type fakeTrendTagMetaCache struct {
	values map[string]domain.TagMeta
}

func (cache *fakeTrendTagMetaCache) Get(tagID string) (domain.TagMeta, bool) {
	if cache.values == nil {
		return domain.TagMeta{}, false
	}

	meta, exists := cache.values[tagID]
	return meta, exists
}

func (cache *fakeTrendTagMetaCache) Put(meta domain.TagMeta) {
	if cache.values == nil {
		cache.values = make(map[string]domain.TagMeta)
	}

	cache.values[meta.TagID.String()] = meta
}

func (repository *fakeTrendRepository) GetTagMeta(_ context.Context, _ uuid.UUID) (domain.TagMeta, error) {
	return repository.tagMeta, nil
}

func (repository *fakeTrendRepository) GetRawPoints(
	_ context.Context,
	_ uuid.UUID,
	_ time.Time,
	_ time.Time,
) ([]domain.TrendPoint, error) {
	repository.rawCalls++
	return repository.rawPoints, nil
}

func (repository *fakeTrendRepository) GetRawPointsByStep(
	_ context.Context,
	_ uuid.UUID,
	_ time.Time,
	_ time.Time,
	_ time.Duration,
	_ domain.Aggregator,
) ([]domain.TrendPoint, error) {
	repository.rawCalls++
	return repository.rawPoints, nil
}

func (repository *fakeTrendRepository) GetAggregatedPoints(
	_ context.Context,
	_ uuid.UUID,
	_ time.Time,
	_ time.Time,
	_ time.Duration,
	_ domain.Aggregator,
) ([]domain.TrendPoint, error) {
	repository.aggCalls++
	return repository.aggPoints, nil
}

func TestGetTrendUsesRawForWindowUpTo24Hours(t *testing.T) {
	now := time.Now().UTC()
	tagID := uuid.New()
	repository := &fakeTrendRepository{
		tagMeta: domain.TagMeta{
			TagID: tagID,
		},
		rawPoints: []domain.TrendPoint{{Timestamp: now, Quality: domain.QualityOK}},
	}

	trendUseCase := NewTrendUseCase(repository, nil, 5000, 90, nil)

	series, useCaseError := trendUseCase.GetTrend(context.Background(), domain.TrendQuery{
		TagID:      tagID,
		From:       now.Add(-time.Hour),
		To:         now,
		Aggregator: domain.AggregatorAvg,
		Limit:      1000,
	})
	if useCaseError != nil {
		t.Fatalf("GetTrend returned error: %v", useCaseError)
	}

	if series.Source != domain.TrendSourceRaw {
		t.Fatalf("expected source %q, got %q", domain.TrendSourceRaw, series.Source)
	}
	if repository.rawCalls == 0 {
		t.Fatalf("expected raw repository to be called")
	}
	if repository.aggCalls != 0 {
		t.Fatalf("expected aggregate repository not to be called")
	}
}

func TestGetTrendUsesAggForWindowGreaterThan24Hours(t *testing.T) {
	now := time.Now().UTC()
	tagID := uuid.New()
	repository := &fakeTrendRepository{
		tagMeta: domain.TagMeta{
			TagID: tagID,
		},
		aggPoints: []domain.TrendPoint{{Timestamp: now, Quality: domain.QualityOK}},
	}

	trendUseCase := NewTrendUseCase(repository, nil, 5000, 90, nil)

	series, useCaseError := trendUseCase.GetTrend(context.Background(), domain.TrendQuery{
		TagID:      tagID,
		From:       now.Add(-48 * time.Hour),
		To:         now,
		Aggregator: domain.AggregatorAvg,
		Limit:      1000,
	})
	if useCaseError != nil {
		t.Fatalf("GetTrend returned error: %v", useCaseError)
	}

	if series.Source != domain.TrendSourceAgg1m {
		t.Fatalf("expected source %q, got %q", domain.TrendSourceAgg1m, series.Source)
	}
	if repository.aggCalls == 0 {
		t.Fatalf("expected aggregate repository to be called")
	}
	if repository.rawCalls != 0 {
		t.Fatalf("expected raw repository not to be called")
	}
}

func TestGetTrendReadsTagMetaFromCache(t *testing.T) {
	now := time.Now().UTC()
	tagID := uuid.New()
	repository := &fakeTrendRepository{
		tagMeta: domain.TagMeta{
			TagID: tagID,
		},
		rawPoints: []domain.TrendPoint{{Timestamp: now, Quality: domain.QualityOK}},
	}

	cache := &fakeTrendTagMetaCache{
		values: map[string]domain.TagMeta{
			tagID.String(): {
				TagID:   tagID,
				TagName: "cached-tag",
			},
		},
	}

	trendUseCase := NewTrendUseCase(repository, cache, 5000, 90, nil)

	series, useCaseError := trendUseCase.GetTrend(context.Background(), domain.TrendQuery{
		TagID:      tagID,
		From:       now.Add(-time.Hour),
		To:         now,
		Aggregator: domain.AggregatorAvg,
		Limit:      1000,
	})
	if useCaseError != nil {
		t.Fatalf("GetTrend returned error: %v", useCaseError)
	}

	if series.TagName != "cached-tag" {
		t.Fatalf("expected cached tag meta, got %q", series.TagName)
	}
}
