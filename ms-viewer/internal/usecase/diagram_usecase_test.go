package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
)

type fakeDiagramRepository struct {
	listObjectsResult      domain.ObjectListResult
	listObjectsError       error
	listDiagramsResult     domain.DiagramListResult
	listDiagramsError      error
	getDiagramResult       domain.Diagram
	getDiagramError        error
	boundTagIDs            []uuid.UUID
	boundTagIDsError       error
}

func (repository *fakeDiagramRepository) ListObjectsWithPublishedDiagrams(
	_ context.Context,
	_ domain.ObjectListQuery,
) (domain.ObjectListResult, error) {
	return repository.listObjectsResult, repository.listObjectsError
}

func (repository *fakeDiagramRepository) ListPublishedDiagrams(
	_ context.Context,
	_ domain.DiagramListQuery,
) (domain.DiagramListResult, error) {
	return repository.listDiagramsResult, repository.listDiagramsError
}

func (repository *fakeDiagramRepository) GetPublishedDiagram(
	_ context.Context,
	_ uuid.UUID,
) (domain.Diagram, error) {
	return repository.getDiagramResult, repository.getDiagramError
}

func (repository *fakeDiagramRepository) ListDiagramBoundTagIDs(
	_ context.Context,
	_ uuid.UUID,
) ([]uuid.UUID, error) {
	return repository.boundTagIDs, repository.boundTagIDsError
}

type fakeDiagramCache struct {
	values map[uuid.UUID]domain.IngestRecord
	err    error
}

func (cache *fakeDiagramCache) GetLastValues(
	_ context.Context,
	_ []uuid.UUID,
) (map[uuid.UUID]domain.IngestRecord, error) {
	if cache.err != nil {
		return nil, cache.err
	}
	return cache.values, nil
}

type fakeTrendMetaRepository struct {
	metaByTag map[uuid.UUID]domain.TagMeta
}

func (repository *fakeTrendMetaRepository) GetTagMeta(
	_ context.Context,
	tagID uuid.UUID,
) (domain.TagMeta, error) {
	meta, exists := repository.metaByTag[tagID]
	if !exists {
		return domain.TagMeta{}, domain.ErrNotFound
	}

	return meta, nil
}

func (repository *fakeTrendMetaRepository) GetRawPoints(
	_ context.Context,
	_ uuid.UUID,
	_ time.Time,
	_ time.Time,
) ([]domain.TrendPoint, error) {
	return nil, nil
}

func (repository *fakeTrendMetaRepository) GetRawPointsByStep(
	_ context.Context,
	_ uuid.UUID,
	_ time.Time,
	_ time.Time,
	_ time.Duration,
	_ domain.Aggregator,
) ([]domain.TrendPoint, error) {
	return nil, nil
}

func (repository *fakeTrendMetaRepository) GetAggregatedPoints(
	_ context.Context,
	_ uuid.UUID,
	_ time.Time,
	_ time.Time,
	_ time.Duration,
	_ domain.Aggregator,
) ([]domain.TrendPoint, error) {
	return nil, nil
}

func TestDiagramUseCaseListPublishedDiagramsValidatesLimit(t *testing.T) {
	useCase := NewDiagramUseCase(
		&fakeDiagramRepository{},
		&fakeTrendMetaRepository{},
		nil,
		&fakeDiagramCache{},
		nil,
	)

	_, useCaseError := useCase.ListPublishedDiagrams(context.Background(), domain.DiagramListQuery{
		ObjectID: uuid.New(),
		Limit:    1000,
	})
	if !errors.Is(useCaseError, domain.ErrInvalidInput) {
		t.Fatalf("expected invalid input error, got %v", useCaseError)
	}
}

func TestDiagramUseCaseGetPublishedDiagramAppliesCanvasFallback(t *testing.T) {
	diagramID := uuid.New()
	useCase := NewDiagramUseCase(&fakeDiagramRepository{
		getDiagramResult: domain.Diagram{
			ID:          diagramID,
			ObjectID:    uuid.New(),
			PublishedAt: time.Now().UTC(),
		},
	}, &fakeTrendMetaRepository{}, nil, &fakeDiagramCache{}, nil)

	diagram, useCaseError := useCase.GetPublishedDiagram(context.Background(), diagramID)
	if useCaseError != nil {
		t.Fatalf("expected no error, got %v", useCaseError)
	}
	if diagram.Canvas.Width != 1920 || diagram.Canvas.Height != 1080 {
		t.Fatalf("unexpected canvas fallback: %#v", diagram.Canvas)
	}
	if diagram.Canvas.Background != "#F5F5F5" {
		t.Fatalf("unexpected background fallback: %s", diagram.Canvas.Background)
	}
}

func TestDiagramUseCaseGetSnapshotCollectsMissingTags(t *testing.T) {
	tagWithValue := uuid.New()
	tagWithoutValue := uuid.New()
	now := time.Now().UTC()
	value := 42.0

	useCase := NewDiagramUseCase(
		&fakeDiagramRepository{boundTagIDs: []uuid.UUID{tagWithValue, tagWithoutValue}},
		&fakeTrendMetaRepository{},
		nil,
		&fakeDiagramCache{values: map[uuid.UUID]domain.IngestRecord{
			tagWithValue: {
				TagID:     tagWithValue,
				Timestamp: now,
				Value:     &value,
				Quality:   domain.QualityOK,
			},
		}},
		nil,
	)

	snapshot, useCaseError := useCase.GetDiagramSnapshot(context.Background(), uuid.New())
	if useCaseError != nil {
		t.Fatalf("expected no error, got %v", useCaseError)
	}
	if len(snapshot.Values) != 1 {
		t.Fatalf("expected one value, got %d", len(snapshot.Values))
	}
	if len(snapshot.MissingTagIDs) != 1 || snapshot.MissingTagIDs[0] != tagWithoutValue {
		t.Fatalf("unexpected missing tags: %v", snapshot.MissingTagIDs)
	}
}

func TestDiagramUseCaseGetSnapshotReturnsCacheUnavailable(t *testing.T) {
	useCase := NewDiagramUseCase(
		&fakeDiagramRepository{boundTagIDs: []uuid.UUID{uuid.New()}},
		&fakeTrendMetaRepository{},
		nil,
		&fakeDiagramCache{err: errors.New("redis down")},
		nil,
	)

	_, useCaseError := useCase.GetDiagramSnapshot(context.Background(), uuid.New())
	if !errors.Is(useCaseError, domain.ErrUnavailable) {
		t.Fatalf("expected unavailable error, got %v", useCaseError)
	}
}

func TestDiagramUseCaseGetPublishedDiagramEnrichesTagMeta(t *testing.T) {
	tagID := uuid.New()
	useCase := NewDiagramUseCase(
		&fakeDiagramRepository{
			getDiagramResult: domain.Diagram{
				ID:          uuid.New(),
				ObjectID:    uuid.New(),
				PublishedAt: time.Now().UTC(),
				Figures: []domain.DiagramFigure{
					{
						ID:    uuid.New(),
						Type:  domain.FigureTypeRect,
						TagID: &tagID,
					},
				},
			},
		},
		&fakeTrendMetaRepository{
			metaByTag: map[uuid.UUID]domain.TagMeta{
				tagID: {
					TagID:      tagID,
					TagName:    "tag-1",
					DeviceID:   uuid.New(),
					DeviceName: "device-1",
				},
			},
		},
		&fakeTrendTagMetaCache{},
		&fakeDiagramCache{},
		nil,
	)

	diagram, useCaseError := useCase.GetPublishedDiagram(context.Background(), uuid.New())
	if useCaseError != nil {
		t.Fatalf("expected no error, got %v", useCaseError)
	}
	if len(diagram.Figures) != 1 {
		t.Fatalf("expected one figure, got %d", len(diagram.Figures))
	}
	if diagram.Figures[0].Tag == nil {
		t.Fatal("expected figure tag meta to be present")
	}
	if diagram.Figures[0].Tag.Name != "tag-1" {
		t.Fatalf("unexpected tag name %q", diagram.Figures[0].Tag.Name)
	}
}
