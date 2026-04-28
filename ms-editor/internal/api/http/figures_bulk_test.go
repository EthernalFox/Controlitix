package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/EthernalFox/Controlitix/ms-editor/internal/domain"
	"github.com/EthernalFox/Controlitix/ms-editor/internal/usecase"
)

func TestBulkUpsertFiguresValidationType(t *testing.T) {
	repo := &figureRepositoryHandlerMock{}
	handler := newFigureHandlerForTests(repo)

	request := httptest.NewRequest(
		http.MethodPut,
		"/diagrams/diagram-1/figures",
		bytes.NewBufferString(`{"figures":[{"type":"unknown","tag_id":null,"params":{"x":1}}]}`),
	)
	responseRecorder := httptest.NewRecorder()

	handler.handleDiagrams(responseRecorder, request)

	if responseRecorder.Code != http.StatusUnprocessableEntity {
		t.Fatalf("unexpected status code: got %d want %d", responseRecorder.Code, http.StatusUnprocessableEntity)
	}

	var problem problemResponse
	if decodeError := json.Unmarshal(responseRecorder.Body.Bytes(), &problem); decodeError != nil {
		t.Fatalf("failed to decode response: %v", decodeError)
	}

	if problem.Type != "/problems/validation-error" {
		t.Fatalf("unexpected problem type: %s", problem.Type)
	}

	assertContainsFieldError(t, problem.Errors, "figures[0].type")

	if repo.bulkCalled {
		t.Fatal("repository BulkUpsertFigures must not be called for invalid type")
	}
}

func TestBulkUpsertFiguresValidationDuplicateID(t *testing.T) {
	repo := &figureRepositoryHandlerMock{}
	handler := newFigureHandlerForTests(repo)

	request := httptest.NewRequest(
		http.MethodPut,
		"/diagrams/diagram-1/figures",
		bytes.NewBufferString(`{"figures":[{"id":"550e8400-e29b-41d4-a716-446655440001","type":"rect","tag_id":null,"params":{"x":1}},{"id":"550e8400-e29b-41d4-a716-446655440001","type":"circle","tag_id":null,"params":{"x":2}}]}`),
	)
	responseRecorder := httptest.NewRecorder()

	handler.handleDiagrams(responseRecorder, request)

	if responseRecorder.Code != http.StatusUnprocessableEntity {
		t.Fatalf("unexpected status code: got %d want %d", responseRecorder.Code, http.StatusUnprocessableEntity)
	}

	var problem problemResponse
	if decodeError := json.Unmarshal(responseRecorder.Body.Bytes(), &problem); decodeError != nil {
		t.Fatalf("failed to decode response: %v", decodeError)
	}

	assertContainsFieldError(t, problem.Errors, "figures[1].id")

	if repo.bulkCalled {
		t.Fatal("repository BulkUpsertFigures must not be called for duplicate IDs")
	}
}

func TestBulkUpsertFiguresNotFoundForForeignFigureID(t *testing.T) {
	repo := &figureRepositoryHandlerMock{
		bulkError: fmt.Errorf(
			"figure %s does not belong to diagram %s: %w",
			"550e8400-e29b-41d4-a716-446655440777",
			"diagram-1",
			domain.ErrNotFound,
		),
	}
	handler := newFigureHandlerForTests(repo)

	request := httptest.NewRequest(
		http.MethodPut,
		"/diagrams/diagram-1/figures",
		bytes.NewBufferString(`{"figures":[{"id":"550e8400-e29b-41d4-a716-446655440777","type":"rect","tag_id":null,"params":{"x":1}}]}`),
	)
	responseRecorder := httptest.NewRecorder()

	handler.handleDiagrams(responseRecorder, request)

	if responseRecorder.Code != http.StatusNotFound {
		t.Fatalf("unexpected status code: got %d want %d", responseRecorder.Code, http.StatusNotFound)
	}

	var problem problemResponse
	if decodeError := json.Unmarshal(responseRecorder.Body.Bytes(), &problem); decodeError != nil {
		t.Fatalf("failed to decode response: %v", decodeError)
	}

	if problem.Type != "/problems/not-found" {
		t.Fatalf("unexpected problem type: %s", problem.Type)
	}

	if problem.Detail == "" {
		t.Fatal("expected not-found detail to be present")
	}

	if !repo.bulkCalled {
		t.Fatal("expected repository BulkUpsertFigures to be called")
	}
}

func newFigureHandlerForTests(repo domain.FigureRepository) *Handler {
	figureUseCase := usecase.NewFigureUseCase(
		repo,
		nil,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)

	return NewHandler(nil, nil, figureUseCase, nil, nil, nil)
}

func assertContainsFieldError(t *testing.T, errors []fieldError, field string) {
	t.Helper()

	for _, fieldError := range errors {
		if fieldError.Field == field {
			return
		}
	}

	t.Fatalf("field error %q not found in %#v", field, errors)
}

type figureRepositoryHandlerMock struct {
	bulkCalled bool
	bulkError  error
	bulkResult domain.FigureBulkResult
}

func (mock *figureRepositoryHandlerMock) CreateFigures(
	ctx context.Context,
	diagramID string,
	figures []domain.Figure,
) ([]domain.Figure, error) {
	return nil, nil
}

func (mock *figureRepositoryHandlerMock) BulkUpsertFigures(
	ctx context.Context,
	diagramID string,
	items []domain.FigureBulkItem,
) (domain.FigureBulkResult, error) {
	mock.bulkCalled = true
	return mock.bulkResult, mock.bulkError
}

func (mock *figureRepositoryHandlerMock) UpdateFigure(
	ctx context.Context,
	figureID string,
	update domain.FigureUpdate,
) (domain.Figure, error) {
	return domain.Figure{}, nil
}

func (mock *figureRepositoryHandlerMock) DeleteFigure(ctx context.Context, figureID string) error {
	return nil
}

func (mock *figureRepositoryHandlerMock) ListFigures(
	ctx context.Context,
	query domain.FigureListQuery,
) (domain.ListResult[domain.Figure], error) {
	return domain.ListResult[domain.Figure]{}, nil
}
