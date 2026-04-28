package http

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
)

type TrendsHandler struct {
	trendUseCase TrendUseCase
}

func NewTrendsHandler(trendUseCase TrendUseCase) *TrendsHandler {
	return &TrendsHandler{trendUseCase: trendUseCase}
}

func (handler *TrendsHandler) GetTrend(
	responseWriter http.ResponseWriter,
	request *http.Request,
) {
	tagID, parseError := uuid.Parse(strings.TrimSpace(chi.URLParam(request, "tagId")))
	if parseError != nil {
		writeProblem(responseWriter, Problem{
			Type:   "/errors/trends/invalid-range",
			Title:  "Invalid range",
			Status: http.StatusBadRequest,
			Detail: "tag id is invalid",
		})
		return
	}

	query, queryError := parseSingleTrendQuery(request, tagID)
	if queryError != nil {
		writeDomainError(responseWriter, queryError)
		return
	}

	series, useCaseError := handler.trendUseCase.GetTrend(request.Context(), query)
	if useCaseError != nil {
		writeDomainError(responseWriter, useCaseError)
		return
	}

	writeJSON(responseWriter, http.StatusOK, mapTrendSeriesResponse(series))
}

func (handler *TrendsHandler) GetTrends(
	responseWriter http.ResponseWriter,
	request *http.Request,
) {
	query, queryError := parseBatchTrendQuery(request)
	if queryError != nil {
		writeDomainError(responseWriter, queryError)
		return
	}

	result, useCaseError := handler.trendUseCase.GetTrendsBatch(request.Context(), query)
	if useCaseError != nil {
		writeDomainError(responseWriter, useCaseError)
		return
	}

	writeJSON(responseWriter, http.StatusOK, mapTrendsBatchResponse(result))
}

func parseSingleTrendQuery(
	request *http.Request,
	tagID uuid.UUID,
) (domain.TrendQuery, error) {
	from, to, step, aggregator, limit, parseError := parseCommonTrendQuery(request)
	if parseError != nil {
		return domain.TrendQuery{}, parseError
	}

	return domain.TrendQuery{
		TagID:      tagID,
		From:       from,
		To:         to,
		Step:       step,
		Aggregator: aggregator,
		Limit:      limit,
	}, nil
}

func parseBatchTrendQuery(request *http.Request) (domain.TrendBatchQuery, error) {
	tagIDsRaw := strings.TrimSpace(request.URL.Query().Get("tag_ids"))
	if tagIDsRaw == "" {
		return domain.TrendBatchQuery{}, fmt.Errorf("tag_ids is required: %w", domain.ErrInvalidInput)
	}

	parts := strings.Split(tagIDsRaw, ",")
	if len(parts) > 10 {
		return domain.TrendBatchQuery{}, fmt.Errorf("tag_ids exceeds 10: %w", domain.ErrInvalidInput)
	}

	tagIDs := make([]uuid.UUID, 0, len(parts))
	for _, part := range parts {
		tagID, parseError := uuid.Parse(strings.TrimSpace(part))
		if parseError != nil {
			return domain.TrendBatchQuery{}, fmt.Errorf("invalid tag_id: %w", domain.ErrInvalidInput)
		}
		tagIDs = append(tagIDs, tagID)
	}

	from, to, step, aggregator, limit, parseError := parseCommonTrendQuery(request)
	if parseError != nil {
		return domain.TrendBatchQuery{}, parseError
	}

	return domain.TrendBatchQuery{
		TagIDs:     tagIDs,
		From:       from,
		To:         to,
		Step:       step,
		Aggregator: aggregator,
		Limit:      limit,
	}, nil
}

func parseCommonTrendQuery(
	request *http.Request,
) (time.Time, time.Time, time.Duration, domain.Aggregator, int, error) {
	from, fromError := time.Parse(time.RFC3339, request.URL.Query().Get("from"))
	if fromError != nil {
		return time.Time{}, time.Time{}, 0, "", 0, fmt.Errorf("invalid from: %w", domain.ErrInvalidInput)
	}

	to, toError := time.Parse(time.RFC3339, request.URL.Query().Get("to"))
	if toError != nil {
		return time.Time{}, time.Time{}, 0, "", 0, fmt.Errorf("invalid to: %w", domain.ErrInvalidInput)
	}

	stepRaw := strings.TrimSpace(request.URL.Query().Get("step"))
	var step time.Duration
	if stepRaw != "" {
		parsedStep, parseStepError := time.ParseDuration(stepRaw)
		if parseStepError != nil {
			return time.Time{}, time.Time{}, 0, "", 0, fmt.Errorf("invalid step: %w", domain.ErrInvalidInput)
		}
		step = parsedStep
	}

	aggregator := domain.AggregatorAvg
	aggregatorRaw := strings.TrimSpace(request.URL.Query().Get("agg"))
	if aggregatorRaw != "" {
		parsedAggregator, parseAggregatorError := domain.ParseAggregator(aggregatorRaw)
		if parseAggregatorError != nil {
			return time.Time{}, time.Time{}, 0, "", 0, fmt.Errorf("invalid agg: %w", domain.ErrInvalidInput)
		}
		aggregator = parsedAggregator
	}

	limit := defaultTrendLimit
	limitRaw := strings.TrimSpace(request.URL.Query().Get("limit"))
	if limitRaw != "" {
		parsedLimit, parseLimitError := strconv.Atoi(limitRaw)
		if parseLimitError != nil {
			return time.Time{}, time.Time{}, 0, "", 0, fmt.Errorf("invalid limit: %w", domain.ErrInvalidInput)
		}
		limit = parsedLimit
	}
	if limit <= 0 {
		return time.Time{}, time.Time{}, 0, "", 0, fmt.Errorf("limit must be positive: %w", domain.ErrInvalidInput)
	}

	return from.UTC(), to.UTC(), step, aggregator, limit, nil
}
