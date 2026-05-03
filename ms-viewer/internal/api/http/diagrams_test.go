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

type stubDiagramUseCase struct {
	listObjectsFunc  func(ctx context.Context, query domain.ObjectListQuery) (domain.ObjectListResult, error)
	listDiagramsFunc func(ctx context.Context, query domain.DiagramListQuery) (domain.DiagramListResult, error)
	getDiagramFunc   func(ctx context.Context, diagramID uuid.UUID) (domain.Diagram, error)
	getSnapshotFunc  func(ctx context.Context, diagramID uuid.UUID) (domain.DiagramSnapshot, error)
}

func (stub *stubDiagramUseCase) ListObjectsWithPublishedDiagrams(
	ctx context.Context,
	query domain.ObjectListQuery,
) (domain.ObjectListResult, error) {
	return stub.listObjectsFunc(ctx, query)
}

func (stub *stubDiagramUseCase) ListPublishedDiagrams(
	ctx context.Context,
	query domain.DiagramListQuery,
) (domain.DiagramListResult, error) {
	return stub.listDiagramsFunc(ctx, query)
}

func (stub *stubDiagramUseCase) GetPublishedDiagram(
	ctx context.Context,
	diagramID uuid.UUID,
) (domain.Diagram, error) {
	return stub.getDiagramFunc(ctx, diagramID)
}

func (stub *stubDiagramUseCase) GetDiagramSnapshot(
	ctx context.Context,
	diagramID uuid.UUID,
) (domain.DiagramSnapshot, error) {
	return stub.getSnapshotFunc(ctx, diagramID)
}

func TestGetDiagramReturnsNotPublishedProblem(t *testing.T) {
	handler := NewDiagramsHandler(&stubDiagramUseCase{
		listObjectsFunc: func(_ context.Context, _ domain.ObjectListQuery) (domain.ObjectListResult, error) {
			return domain.ObjectListResult{}, nil
		},
		listDiagramsFunc: func(_ context.Context, _ domain.DiagramListQuery) (domain.DiagramListResult, error) {
			return domain.DiagramListResult{}, nil
		},
		getDiagramFunc: func(_ context.Context, _ uuid.UUID) (domain.Diagram, error) {
			return domain.Diagram{}, fmt.Errorf("diagram unpublished: %w", domain.ErrNotPublished)
		},
		getSnapshotFunc: func(_ context.Context, _ uuid.UUID) (domain.DiagramSnapshot, error) {
			return domain.DiagramSnapshot{}, nil
		},
	}, nil)

	router := chi.NewRouter()
	router.Get("/api/diagrams/{diagramId}", handler.GetDiagram)

	request := httptest.NewRequest(http.MethodGet, "/api/diagrams/"+uuid.NewString(), nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, response.Code)
	}

	var problem Problem
	if decodeError := json.NewDecoder(response.Body).Decode(&problem); decodeError != nil {
		t.Fatalf("decode problem: %v", decodeError)
	}
	if problem.Type != "/errors/diagrams/not-published" {
		t.Fatalf("unexpected problem type: %s", problem.Type)
	}
}

func TestGetDiagramSnapshotReturnsCacheUnavailableProblem(t *testing.T) {
	handler := NewDiagramsHandler(&stubDiagramUseCase{
		listObjectsFunc: func(_ context.Context, _ domain.ObjectListQuery) (domain.ObjectListResult, error) {
			return domain.ObjectListResult{}, nil
		},
		listDiagramsFunc: func(_ context.Context, _ domain.DiagramListQuery) (domain.DiagramListResult, error) {
			return domain.DiagramListResult{}, nil
		},
		getDiagramFunc: func(_ context.Context, _ uuid.UUID) (domain.Diagram, error) {
			return domain.Diagram{}, nil
		},
		getSnapshotFunc: func(_ context.Context, _ uuid.UUID) (domain.DiagramSnapshot, error) {
			return domain.DiagramSnapshot{}, fmt.Errorf("redis unavailable: %w", domain.ErrUnavailable)
		},
	}, nil)

	router := chi.NewRouter()
	router.Get("/api/diagrams/{diagramId}/snapshot", handler.GetDiagramSnapshot)

	request := httptest.NewRequest(http.MethodGet, "/api/diagrams/"+uuid.NewString()+"/snapshot", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, response.Code)
	}

	var problem Problem
	if decodeError := json.NewDecoder(response.Body).Decode(&problem); decodeError != nil {
		t.Fatalf("decode problem: %v", decodeError)
	}
	if problem.Type != "/errors/realtime/cache-unavailable" {
		t.Fatalf("unexpected problem type: %s", problem.Type)
	}
}

func TestGetDiagramReturnsCacheControlHeader(t *testing.T) {
	diagramID := uuid.New()
	handler := NewDiagramsHandler(&stubDiagramUseCase{
		listObjectsFunc: func(_ context.Context, _ domain.ObjectListQuery) (domain.ObjectListResult, error) {
			return domain.ObjectListResult{}, nil
		},
		listDiagramsFunc: func(_ context.Context, _ domain.DiagramListQuery) (domain.DiagramListResult, error) {
			return domain.DiagramListResult{}, nil
		},
		getDiagramFunc: func(_ context.Context, _ uuid.UUID) (domain.Diagram, error) {
			return domain.Diagram{
				ID:          diagramID,
				ObjectID:    uuid.New(),
				PublishedAt: time.Now().UTC(),
				Canvas: domain.DiagramCanvas{
					Width:      1920,
					Height:     1080,
					Background: "#F5F5F5",
				},
				Figures: []domain.DiagramFigure{},
			}, nil
		},
		getSnapshotFunc: func(_ context.Context, _ uuid.UUID) (domain.DiagramSnapshot, error) {
			return domain.DiagramSnapshot{}, nil
		},
	}, nil)

	router := chi.NewRouter()
	router.Get("/api/diagrams/{diagramId}", handler.GetDiagram)

	request := httptest.NewRequest(http.MethodGet, "/api/diagrams/"+diagramID.String(), nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
	if response.Header().Get("Cache-Control") != "private, max-age=10" {
		t.Fatalf("unexpected cache-control header: %s", response.Header().Get("Cache-Control"))
	}
}

func TestGetObjectDiagramsReturnsInvalidRequestProblem(t *testing.T) {
	handler := NewDiagramsHandler(&stubDiagramUseCase{
		listObjectsFunc: func(_ context.Context, _ domain.ObjectListQuery) (domain.ObjectListResult, error) {
			return domain.ObjectListResult{}, nil
		},
		listDiagramsFunc: func(_ context.Context, _ domain.DiagramListQuery) (domain.DiagramListResult, error) {
			return domain.DiagramListResult{}, nil
		},
		getDiagramFunc: func(_ context.Context, _ uuid.UUID) (domain.Diagram, error) {
			return domain.Diagram{}, nil
		},
		getSnapshotFunc: func(_ context.Context, _ uuid.UUID) (domain.DiagramSnapshot, error) {
			return domain.DiagramSnapshot{}, nil
		},
	}, nil)

	router := chi.NewRouter()
	router.Get("/api/objects/{objectId}/diagrams", handler.GetObjectDiagrams)

	request := httptest.NewRequest(http.MethodGet, "/api/objects/not-uuid/diagrams", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}
