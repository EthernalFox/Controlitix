package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
)

type stubTrendUseCase struct {
	getTrendFunc func(ctx context.Context, query domain.TrendQuery) (domain.TrendSeries, error)
}

func (stub *stubTrendUseCase) GetTrend(
	ctx context.Context,
	query domain.TrendQuery,
) (domain.TrendSeries, error) {
	return stub.getTrendFunc(ctx, query)
}

func (stub *stubTrendUseCase) GetTrendsBatch(
	_ context.Context,
	_ domain.TrendBatchQuery,
) (domain.TrendBatchResult, error) {
	return domain.TrendBatchResult{}, nil
}

func TestGetTrendReturnsSeriesContract(t *testing.T) {
	tagID := uuid.New()
	deviceID := uuid.New()
	now := time.Date(2026, 4, 25, 10, 0, 0, 0, time.UTC)
	value := 75.2

	handler := NewTrendsHandler(&stubTrendUseCase{
		getTrendFunc: func(_ context.Context, _ domain.TrendQuery) (domain.TrendSeries, error) {
			return domain.TrendSeries{
				TagMeta: domain.TagMeta{
					TagID:      tagID,
					TagName:    "boiler_1.t_out",
					DeviceID:   deviceID,
					DeviceName: "Boiler-1",
					Unit:       domain.Unit{ID: 7, Name: "°C", Symbol: "°C", Category: "temperature"},
					DataType:   domain.DataType{ID: 2, Name: "float32"},
				},
				From:       now,
				To:         now.Add(10 * time.Minute),
				Step:       10 * time.Second,
				Aggregator: domain.AggregatorAvg,
				Source:     domain.TrendSourceRaw,
				Points: []domain.TrendPoint{
					{Timestamp: now, Value: &value, Quality: domain.QualityOK},
				},
			}, nil
		},
	})

	router := chi.NewRouter()
	router.Get("/api/trends/{tagId}", handler.GetTrend)

	request := httptest.NewRequest(http.MethodGet, fmt.Sprintf(
		"/api/trends/%s?from=2026-04-25T10:00:00Z&to=2026-04-25T11:00:00Z&step=10s&agg=avg&limit=1000",
		tagID.String(),
	), nil)
	responseRecorder := httptest.NewRecorder()

	router.ServeHTTP(responseRecorder, request)
	if responseRecorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, responseRecorder.Code)
	}

	var payload map[string]any
	if decodeError := json.NewDecoder(responseRecorder.Body).Decode(&payload); decodeError != nil {
		t.Fatalf("decode response: %v", decodeError)
	}

	if payload["tag_id"] != tagID.String() {
		t.Fatalf("expected tag_id %s, got %v", tagID.String(), payload["tag_id"])
	}
	if payload["source"] != string(domain.TrendSourceRaw) {
		t.Fatalf("expected source %s, got %v", domain.TrendSourceRaw, payload["source"])
	}
}

func TestGetTrendReturnsInvalidRangeProblem(t *testing.T) {
	handler := NewTrendsHandler(&stubTrendUseCase{
		getTrendFunc: func(_ context.Context, _ domain.TrendQuery) (domain.TrendSeries, error) {
			return domain.TrendSeries{}, nil
		},
	})

	router := chi.NewRouter()
	router.Get("/api/trends/{tagId}", handler.GetTrend)

	request := httptest.NewRequest(http.MethodGet, "/api/trends/not-uuid?from=2026-04-25T10:00:00Z&to=2026-04-25T11:00:00Z", nil)
	responseRecorder := httptest.NewRecorder()

	router.ServeHTTP(responseRecorder, request)
	if responseRecorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, responseRecorder.Code)
	}

	var problem Problem
	if decodeError := json.NewDecoder(responseRecorder.Body).Decode(&problem); decodeError != nil {
		t.Fatalf("decode problem: %v", decodeError)
	}
	if problem.Type != "/errors/trends/invalid-range" {
		t.Fatalf("expected problem type /errors/trends/invalid-range, got %s", problem.Type)
	}
}

func TestGetTrendReturnsNotFoundProblem(t *testing.T) {
	tagID := uuid.New()
	handler := NewTrendsHandler(&stubTrendUseCase{
		getTrendFunc: func(_ context.Context, _ domain.TrendQuery) (domain.TrendSeries, error) {
			return domain.TrendSeries{}, fmt.Errorf("tag not found: %w", domain.ErrNotFound)
		},
	})

	router := chi.NewRouter()
	router.Get("/api/trends/{tagId}", handler.GetTrend)

	request := httptest.NewRequest(http.MethodGet, fmt.Sprintf(
		"/api/trends/%s?from=2026-04-25T10:00:00Z&to=2026-04-25T11:00:00Z",
		tagID.String(),
	), nil)
	responseRecorder := httptest.NewRecorder()

	router.ServeHTTP(responseRecorder, request)
	if responseRecorder.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, responseRecorder.Code)
	}

	var problem Problem
	if decodeError := json.NewDecoder(responseRecorder.Body).Decode(&problem); decodeError != nil {
		t.Fatalf("decode problem: %v", decodeError)
	}
	if problem.Type != "/errors/trends/tag-not-found" {
		t.Fatalf("expected problem type /errors/trends/tag-not-found, got %s", problem.Type)
	}
}
