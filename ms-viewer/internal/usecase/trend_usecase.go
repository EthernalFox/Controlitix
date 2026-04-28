package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
)

const (
	defaultTrendLimit = 1000
	maxBatchTagIDs    = 10
	batchConcurrency  = 4
)

type TrendUseCase struct {
	trendRepository    domain.TrendRepository
	tagMetaCache       TrendTagMetaCache
	trendsMaxLimit     int
	trendsMaxRangeDays int
	logger             *slog.Logger
}

type TrendTagMetaCache interface {
	Get(tagID string) (domain.TagMeta, bool)
	Put(meta domain.TagMeta)
}

func NewTrendUseCase(
	trendRepository domain.TrendRepository,
	tagMetaCache TrendTagMetaCache,
	trendsMaxLimit int,
	trendsMaxRangeDays int,
	logger *slog.Logger,
) *TrendUseCase {
	if logger == nil {
		logger = slog.Default()
	}

	return &TrendUseCase{
		trendRepository:    trendRepository,
		tagMetaCache:       tagMetaCache,
		trendsMaxLimit:     trendsMaxLimit,
		trendsMaxRangeDays: trendsMaxRangeDays,
		logger:             logger,
	}
}

func (useCase *TrendUseCase) GetTrend(
	ctx context.Context,
	query domain.TrendQuery,
) (domain.TrendSeries, error) {
	validatedQuery, validationError := useCase.normalizeAndValidateQuery(query)
	if validationError != nil {
		return domain.TrendSeries{}, validationError
	}

	tagMeta, metadataError := useCase.loadTagMeta(ctx, validatedQuery.TagID)
	if metadataError != nil {
		return domain.TrendSeries{}, metadataError
	}

	window := validatedQuery.To.Sub(validatedQuery.From)
	effectiveStep := computeEffectiveStep(window, validatedQuery.Step, validatedQuery.Limit)

	series := domain.TrendSeries{
		TagMeta:     tagMeta,
		From:        validatedQuery.From,
		To:          validatedQuery.To,
		Step:        effectiveStep,
		Aggregator:  validatedQuery.Aggregator,
	}

	if window <= 24*time.Hour {
		points, pointsError := useCase.trendRepository.GetRawPoints(
			ctx,
			validatedQuery.TagID,
			validatedQuery.From,
			validatedQuery.To,
		)
		if pointsError != nil {
			return domain.TrendSeries{}, pointsError
		}

		if len(points) > validatedQuery.Limit || validatedQuery.Step > 0 {
			points, pointsError = useCase.trendRepository.GetRawPointsByStep(
				ctx,
				validatedQuery.TagID,
				validatedQuery.From,
				validatedQuery.To,
				effectiveStep,
				validatedQuery.Aggregator,
			)
			if pointsError != nil {
				return domain.TrendSeries{}, pointsError
			}
		}

		series.Source = domain.TrendSourceRaw
		series.Points = truncatePoints(points, validatedQuery.Limit)
		return series, nil
	}

	stepForAggregate := effectiveStep
	if stepForAggregate < time.Minute {
		stepForAggregate = time.Minute
	}

	points, pointsError := useCase.trendRepository.GetAggregatedPoints(
		ctx,
		validatedQuery.TagID,
		validatedQuery.From,
		validatedQuery.To,
		stepForAggregate,
		validatedQuery.Aggregator,
	)
	if pointsError != nil {
		return domain.TrendSeries{}, pointsError
	}

	series.Source = domain.TrendSourceAgg1m
	series.Step = stepForAggregate
	series.Points = truncatePoints(points, validatedQuery.Limit)
	return series, nil
}

func (useCase *TrendUseCase) loadTagMeta(
	ctx context.Context,
	tagID uuid.UUID,
) (domain.TagMeta, error) {
	if useCase.tagMetaCache != nil {
		cachedMeta, exists := useCase.tagMetaCache.Get(tagID.String())
		if exists {
			return cachedMeta, nil
		}
	}

	meta, metadataError := useCase.trendRepository.GetTagMeta(ctx, tagID)
	if metadataError != nil {
		return domain.TagMeta{}, metadataError
	}

	if useCase.tagMetaCache != nil {
		useCase.tagMetaCache.Put(meta)
	}

	return meta, nil
}

func (useCase *TrendUseCase) GetTrendsBatch(
	ctx context.Context,
	query domain.TrendBatchQuery,
) (domain.TrendBatchResult, error) {
	if len(query.TagIDs) == 0 {
		return domain.TrendBatchResult{}, fmt.Errorf("tag_ids must not be empty: %w", domain.ErrInvalidInput)
	}
	if len(query.TagIDs) > maxBatchTagIDs {
		return domain.TrendBatchResult{}, fmt.Errorf("tag_ids exceeds %d: %w", maxBatchTagIDs, domain.ErrInvalidInput)
	}

	validationQuery := domain.TrendQuery{
		TagID:      query.TagIDs[0],
		From:       query.From,
		To:         query.To,
		Step:       query.Step,
		Aggregator: query.Aggregator,
		Limit:      query.Limit,
	}
	if _, validationError := useCase.normalizeAndValidateQuery(validationQuery); validationError != nil {
		return domain.TrendBatchResult{}, validationError
	}

	result := domain.TrendBatchResult{
		Series: make([]domain.TrendSeries, 0, len(query.TagIDs)),
		Errors: make([]domain.TrendBatchError, 0),
	}

	group, groupContext := errgroup.WithContext(ctx)
	group.SetLimit(batchConcurrency)

	var resultMutex sync.Mutex
	for _, tagID := range query.TagIDs {
		tagID := tagID
		group.Go(func() error {
			series, seriesError := useCase.GetTrend(groupContext, domain.TrendQuery{
				TagID:      tagID,
				From:       query.From,
				To:         query.To,
				Step:       query.Step,
				Aggregator: query.Aggregator,
				Limit:      query.Limit,
			})

			resultMutex.Lock()
			defer resultMutex.Unlock()
			if seriesError != nil {
				result.Errors = append(result.Errors, domain.TrendBatchError{
					TagID:   tagID,
					Problem: mapErrorToProblem(seriesError),
				})
				return nil
			}

			result.Series = append(result.Series, series)
			return nil
		})
	}

	if groupError := group.Wait(); groupError != nil {
		return domain.TrendBatchResult{}, fmt.Errorf("wait batch workers: %w", groupError)
	}

	return result, nil
}

func (useCase *TrendUseCase) normalizeAndValidateQuery(
	query domain.TrendQuery,
) (domain.TrendQuery, error) {
	normalized := query
	if normalized.Limit <= 0 {
		normalized.Limit = defaultTrendLimit
	}
	if normalized.Aggregator == "" {
		normalized.Aggregator = domain.AggregatorAvg
	}

	if normalized.TagID == uuid.Nil {
		return domain.TrendQuery{}, fmt.Errorf("tag id is required: %w", domain.ErrInvalidInput)
	}
	if normalized.From.IsZero() || normalized.To.IsZero() {
		return domain.TrendQuery{}, fmt.Errorf("from and to are required: %w", domain.ErrInvalidInput)
	}
	if !normalized.From.Before(normalized.To) {
		return domain.TrendQuery{}, fmt.Errorf("from must be before to: %w", domain.ErrInvalidInput)
	}
	if normalized.Limit > useCase.trendsMaxLimit {
		return domain.TrendQuery{}, fmt.Errorf("limit exceeds max: %w", domain.ErrInvalidInput)
	}

	window := normalized.To.Sub(normalized.From)
	maxRange := time.Duration(useCase.trendsMaxRangeDays) * 24 * time.Hour
	if window > maxRange {
		return domain.TrendQuery{}, fmt.Errorf("range exceeds %d days: %w", useCase.trendsMaxRangeDays, domain.ErrInvalidInput)
	}

	switch normalized.Aggregator {
	case domain.AggregatorAvg,
		domain.AggregatorLast,
		domain.AggregatorMin,
		domain.AggregatorMax:
	default:
		return domain.TrendQuery{}, fmt.Errorf("aggregator is invalid: %w", domain.ErrInvalidInput)
	}

	if normalized.Step < 0 {
		return domain.TrendQuery{}, fmt.Errorf("step must be positive: %w", domain.ErrInvalidInput)
	}

	return normalized, nil
}

func computeEffectiveStep(
	window time.Duration,
	requestedStep time.Duration,
	limit int,
) time.Duration {
	if limit <= 0 {
		if requestedStep > 0 {
			return requestedStep
		}
		return window
	}

	nanoseconds := window.Nanoseconds()
	if nanoseconds <= 0 {
		if requestedStep > 0 {
			return requestedStep
		}
		return time.Second
	}

	minStepNanoseconds := (nanoseconds + int64(limit) - 1) / int64(limit)
	if minStepNanoseconds <= 0 {
		minStepNanoseconds = int64(time.Second)
	}

	minStep := time.Duration(minStepNanoseconds)
	if requestedStep > minStep {
		return requestedStep
	}

	return minStep
}

func truncatePoints(points []domain.TrendPoint, limit int) []domain.TrendPoint {
	if limit <= 0 || len(points) <= limit {
		return points
	}

	return points[:limit]
}

func mapErrorToProblem(err error) domain.TrendProblem {
	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		return domain.TrendProblem{
			Type:   "/errors/trends/invalid-range",
			Title:  "Invalid range",
			Status: 400,
			Detail: err.Error(),
		}
	case errors.Is(err, domain.ErrNotFound):
		return domain.TrendProblem{
			Type:   "/errors/trends/tag-not-found",
			Title:  "Tag not found",
			Status: 404,
			Detail: err.Error(),
		}
	case errors.Is(err, domain.ErrForbidden):
		return domain.TrendProblem{
			Type:   "/errors/trends/forbidden",
			Title:  "Forbidden",
			Status: 403,
			Detail: err.Error(),
		}
	default:
		return domain.TrendProblem{
			Type:   "/problems/internal-error",
			Title:  "Internal error",
			Status: 500,
			Detail: "internal error",
		}
	}
}
