package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/EthernalFox/Controlitix/ms-editor/internal/domain"
)

func TestFigureUseCaseBulkUpsertFiguresPublishesEvents(t *testing.T) {
	repo := &figureRepositoryMock{
		bulkResult: domain.FigureBulkResult{
			Created:    1,
			Updated:    2,
			Deleted:    1,
			CreatedIDs: []string{"550e8400-e29b-41d4-a716-446655440020"},
			UpdatedIDs: []string{"550e8400-e29b-41d4-a716-446655440010", "550e8400-e29b-41d4-a716-446655440011"},
			DeletedIDs: []string{"550e8400-e29b-41d4-a716-446655440012"},
		},
	}
	publisher := &eventPublisherMock{}
	figureUseCase := NewFigureUseCase(
		repo,
		publisher,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)

	items := []domain.FigureBulkItem{
		{
			ID:         stringPointer("550E8400-E29B-41D4-A716-446655440010"),
			FigureType: domain.FigureTypeRect,
			Parameters: json.RawMessage(`{"x":10}`),
		},
		{
			FigureType: domain.FigureTypeCircle,
			Parameters: json.RawMessage(`{"r":5}`),
		},
	}

	result, bulkError := figureUseCase.BulkUpsertFigures(
		context.Background(),
		"diagram-1",
		items,
	)
	if bulkError != nil {
		t.Fatalf("BulkUpsertFigures returned error: %v", bulkError)
	}

	if result.Created != 1 || result.Updated != 2 || result.Deleted != 1 {
		t.Fatalf("unexpected summary: %#v", result)
	}

	if !repo.bulkCalled {
		t.Fatal("expected repository BulkUpsertFigures to be called")
	}

	if len(repo.bulkItems) != len(items) {
		t.Fatalf("unexpected number of items passed to repository: got %d want %d", len(repo.bulkItems), len(items))
	}

	if repo.bulkItems[0].ID == nil || *repo.bulkItems[0].ID != "550e8400-e29b-41d4-a716-446655440010" {
		t.Fatalf("expected first item ID to be normalized UUID, got %#v", repo.bulkItems[0].ID)
	}

	if len(publisher.events) != 4 {
		t.Fatalf("expected 4 events, got %d", len(publisher.events))
	}

	expectedEvents := []struct {
		id        string
		operation string
	}{
		{id: "550e8400-e29b-41d4-a716-446655440020", operation: "created"},
		{id: "550e8400-e29b-41d4-a716-446655440010", operation: "updated"},
		{id: "550e8400-e29b-41d4-a716-446655440011", operation: "updated"},
		{id: "550e8400-e29b-41d4-a716-446655440012", operation: "deleted"},
	}

	for index, expectedEvent := range expectedEvents {
		event := publisher.events[index]
		if event.EntityType != "figure" || event.EntityID != expectedEvent.id || event.Operation != expectedEvent.operation {
			t.Fatalf("unexpected event at index %d: %#v", index, event)
		}
		if string(event.Payload) != "null" {
			t.Fatalf("expected null payload, got %s", string(event.Payload))
		}
	}
}

func TestFigureUseCaseBulkUpsertFiguresValidation(t *testing.T) {
	testCases := []struct {
		name          string
		items         []domain.FigureBulkItem
		expectedField string
	}{
		{
			name: "duplicate ids",
			items: []domain.FigureBulkItem{
				{
					ID:         stringPointer("550e8400-e29b-41d4-a716-446655440001"),
					FigureType: domain.FigureTypeRect,
					Parameters: json.RawMessage(`{"x":1}`),
				},
				{
					ID:         stringPointer("550e8400-e29b-41d4-a716-446655440001"),
					FigureType: domain.FigureTypeCircle,
					Parameters: json.RawMessage(`{"x":2}`),
				},
			},
			expectedField: "figures[1].id",
		},
		{
			name: "unsupported type",
			items: []domain.FigureBulkItem{
				{
					FigureType: domain.FigureType("unknown"),
					Parameters: json.RawMessage(`{"x":1}`),
				},
			},
			expectedField: "figures[0].type",
		},
		{
			name: "invalid params",
			items: []domain.FigureBulkItem{
				{
					FigureType: domain.FigureTypeRect,
					Parameters: json.RawMessage(`[]`),
				},
			},
			expectedField: "figures[0].params",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			repo := &figureRepositoryMock{}
			figureUseCase := NewFigureUseCase(repo, nil, nil)

			_, bulkError := figureUseCase.BulkUpsertFigures(
				context.Background(),
				"diagram-1",
				testCase.items,
			)
			if bulkError == nil {
				t.Fatal("expected validation error")
			}

			var validationError *domain.ValidationError
			if !errors.As(bulkError, &validationError) {
				t.Fatalf("expected ValidationError, got %T", bulkError)
			}

			foundField := false
			for _, field := range validationError.Fields {
				if field.Field == testCase.expectedField {
					foundField = true
					break
				}
			}
			if !foundField {
				t.Fatalf("expected field %q in validation errors, got %#v", testCase.expectedField, validationError.Fields)
			}

			if repo.bulkCalled {
				t.Fatal("repository BulkUpsertFigures must not be called on validation errors")
			}
		})
	}
}

func TestFigureUseCaseBulkUpsertFiguresAllowsEmptyItems(t *testing.T) {
	repo := &figureRepositoryMock{
		bulkResult: domain.FigureBulkResult{Deleted: 2, DeletedIDs: []string{"550e8400-e29b-41d4-a716-446655440099"}},
	}
	figureUseCase := NewFigureUseCase(repo, nil, nil)

	result, bulkError := figureUseCase.BulkUpsertFigures(context.Background(), "diagram-1", []domain.FigureBulkItem{})
	if bulkError != nil {
		t.Fatalf("BulkUpsertFigures returned error: %v", bulkError)
	}

	if !repo.bulkCalled {
		t.Fatal("expected repository BulkUpsertFigures to be called")
	}

	if len(repo.bulkItems) != 0 {
		t.Fatalf("expected empty items passed to repository, got %d", len(repo.bulkItems))
	}

	if result.Deleted != 2 {
		t.Fatalf("unexpected deleted counter: %d", result.Deleted)
	}
}

type figureRepositoryMock struct {
	bulkCalled    bool
	bulkDiagramID string
	bulkItems     []domain.FigureBulkItem
	bulkResult    domain.FigureBulkResult
	bulkError     error
}

func (mock *figureRepositoryMock) CreateFigures(
	ctx context.Context,
	diagramID string,
	figures []domain.Figure,
) ([]domain.Figure, error) {
	return nil, nil
}

func (mock *figureRepositoryMock) BulkUpsertFigures(
	ctx context.Context,
	diagramID string,
	items []domain.FigureBulkItem,
) (domain.FigureBulkResult, error) {
	mock.bulkCalled = true
	mock.bulkDiagramID = diagramID
	mock.bulkItems = append([]domain.FigureBulkItem(nil), items...)
	return mock.bulkResult, mock.bulkError
}

func (mock *figureRepositoryMock) UpdateFigure(
	ctx context.Context,
	figureID string,
	update domain.FigureUpdate,
) (domain.Figure, error) {
	return domain.Figure{}, nil
}

func (mock *figureRepositoryMock) DeleteFigure(ctx context.Context, figureID string) error {
	return nil
}

func (mock *figureRepositoryMock) ListFigures(
	ctx context.Context,
	query domain.FigureListQuery,
) (domain.ListResult[domain.Figure], error) {
	return domain.ListResult[domain.Figure]{}, nil
}

type eventPublisherMock struct {
	events     []domain.ConfigChangedEvent
	publishErr error
}

func (mock *eventPublisherMock) Publish(
	ctx context.Context,
	event domain.ConfigChangedEvent,
) error {
	mock.events = append(mock.events, event)
	return mock.publishErr
}

func stringPointer(value string) *string {
	return &value
}
