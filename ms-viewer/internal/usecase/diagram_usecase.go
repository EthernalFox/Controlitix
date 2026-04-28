package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
)

const (
	defaultDiagramListLimit = 50
	maxDiagramListLimit     = 200
	defaultObjectListLimit  = 50
	maxObjectListLimit      = 200
)

type DiagramSnapshotCache interface {
	GetLastValues(
		ctx context.Context,
		tagIDs []uuid.UUID,
	) (map[uuid.UUID]domain.IngestRecord, error)
}

type DiagramUseCase struct {
	repository        domain.DiagramRepository
	tagMetaRepository domain.TrendRepository
	tagMetaCache      TrendTagMetaCache
	cache             DiagramSnapshotCache
	logger            *slog.Logger
}

func NewDiagramUseCase(
	repository domain.DiagramRepository,
	tagMetaRepository domain.TrendRepository,
	tagMetaCache TrendTagMetaCache,
	cache DiagramSnapshotCache,
	logger *slog.Logger,
) *DiagramUseCase {
	if logger == nil {
		logger = slog.Default()
	}

	return &DiagramUseCase{
		repository:        repository,
		tagMetaRepository: tagMetaRepository,
		tagMetaCache:      tagMetaCache,
		cache:             cache,
		logger:            logger,
	}
}

func (useCase *DiagramUseCase) ListObjectsWithPublishedDiagrams(
	ctx context.Context,
	query domain.ObjectListQuery,
) (domain.ObjectListResult, error) {
	normalizedQuery := query
	if normalizedQuery.Limit <= 0 {
		normalizedQuery.Limit = defaultObjectListLimit
	}
	if normalizedQuery.Limit > maxObjectListLimit {
		return domain.ObjectListResult{}, fmt.Errorf("limit exceeds %d: %w", maxObjectListLimit, domain.ErrInvalidInput)
	}
	if normalizedQuery.Offset < 0 {
		return domain.ObjectListResult{}, fmt.Errorf("offset must be non-negative: %w", domain.ErrInvalidInput)
	}

	return useCase.repository.ListObjectsWithPublishedDiagrams(ctx, normalizedQuery)
}

func (useCase *DiagramUseCase) ListPublishedDiagrams(
	ctx context.Context,
	query domain.DiagramListQuery,
) (domain.DiagramListResult, error) {
	if query.ObjectID == uuid.Nil {
		return domain.DiagramListResult{}, fmt.Errorf("object id is required: %w", domain.ErrInvalidInput)
	}

	normalizedQuery := query
	if normalizedQuery.Limit <= 0 {
		normalizedQuery.Limit = defaultDiagramListLimit
	}
	if normalizedQuery.Limit > maxDiagramListLimit {
		return domain.DiagramListResult{}, fmt.Errorf("limit exceeds %d: %w", maxDiagramListLimit, domain.ErrInvalidInput)
	}
	if normalizedQuery.Offset < 0 {
		return domain.DiagramListResult{}, fmt.Errorf("offset must be non-negative: %w", domain.ErrInvalidInput)
	}

	return useCase.repository.ListPublishedDiagrams(ctx, normalizedQuery)
}

func (useCase *DiagramUseCase) GetPublishedDiagram(
	ctx context.Context,
	diagramID uuid.UUID,
) (domain.Diagram, error) {
	if diagramID == uuid.Nil {
		return domain.Diagram{}, fmt.Errorf("diagram id is required: %w", domain.ErrInvalidInput)
	}

	diagram, fetchError := useCase.repository.GetPublishedDiagram(ctx, diagramID)
	if fetchError != nil {
		return domain.Diagram{}, fetchError
	}

	if tagError := useCase.enrichFigureTags(ctx, &diagram); tagError != nil {
		return domain.Diagram{}, tagError
	}

	if diagram.Canvas.Width <= 0 {
		diagram.Canvas.Width = 1920
	}
	if diagram.Canvas.Height <= 0 {
		diagram.Canvas.Height = 1080
	}
	if diagram.Canvas.Background == "" {
		diagram.Canvas.Background = "#F5F5F5"
	}

	return diagram, nil
}

func (useCase *DiagramUseCase) GetDiagramSnapshot(
	ctx context.Context,
	diagramID uuid.UUID,
) (domain.DiagramSnapshot, error) {
	if diagramID == uuid.Nil {
		return domain.DiagramSnapshot{}, fmt.Errorf("diagram id is required: %w", domain.ErrInvalidInput)
	}

	tagIDs, fetchError := useCase.repository.ListDiagramBoundTagIDs(ctx, diagramID)
	if fetchError != nil {
		return domain.DiagramSnapshot{}, fetchError
	}

	snapshot := domain.DiagramSnapshot{
		DiagramID: diagramID,
		TS:        time.Now().UTC(),
		Values:    make([]domain.IngestRecord, 0),
	}

	if len(tagIDs) == 0 {
		return snapshot, nil
	}
	if useCase.cache == nil {
		return domain.DiagramSnapshot{}, fmt.Errorf("last value cache is not configured: %w", domain.ErrUnavailable)
	}

	valuesByTag, cacheError := useCase.cache.GetLastValues(ctx, tagIDs)
	if cacheError != nil {
		return domain.DiagramSnapshot{}, fmt.Errorf("read redis snapshot: %w", domain.ErrUnavailable)
	}

	missingTagIDs := make([]uuid.UUID, 0)
	for _, tagID := range tagIDs {
		record, exists := valuesByTag[tagID]
		if !exists {
			missingTagIDs = append(missingTagIDs, tagID)
			continue
		}

		snapshot.Values = append(snapshot.Values, record)
	}

	snapshot.MissingTagIDs = missingTagIDs
	sort.Slice(snapshot.Values, func(firstIndex int, secondIndex int) bool {
		return snapshot.Values[firstIndex].TagID.String() < snapshot.Values[secondIndex].TagID.String()
	})

	return snapshot, nil
}

func (useCase *DiagramUseCase) enrichFigureTags(
	ctx context.Context,
	diagram *domain.Diagram,
) error {
	if diagram == nil || len(diagram.Figures) == 0 {
		return nil
	}
	if useCase.tagMetaRepository == nil {
		return nil
	}

	resolvedTagMeta := make(map[string]domain.TagMeta)
	for figureIndex := range diagram.Figures {
		figure := &diagram.Figures[figureIndex]
		if figure.TagID == nil {
			continue
		}

		tagID := figure.TagID.String()
		meta, exists := resolvedTagMeta[tagID]
		if !exists {
			loadedMeta, loadError := useCase.loadTagMeta(ctx, *figure.TagID)
			if loadError != nil {
				if errors.Is(loadError, domain.ErrNotFound) {
					figure.Tag = nil
					continue
				}
				return loadError
			}
			meta = loadedMeta
			resolvedTagMeta[tagID] = meta
		}

		figure.Tag = &domain.TagBrief{
			ID:         meta.TagID,
			Name:       meta.TagName,
			DeviceID:   meta.DeviceID,
			DeviceName: meta.DeviceName,
			Unit:       meta.Unit,
			DataType:   meta.DataType,
		}
	}

	return nil
}

func (useCase *DiagramUseCase) loadTagMeta(
	ctx context.Context,
	tagID uuid.UUID,
) (domain.TagMeta, error) {
	if useCase.tagMetaCache != nil {
		cachedMeta, exists := useCase.tagMetaCache.Get(tagID.String())
		if exists {
			return cachedMeta, nil
		}
	}

	meta, metadataError := useCase.tagMetaRepository.GetTagMeta(ctx, tagID)
	if metadataError != nil {
		return domain.TagMeta{}, metadataError
	}

	if useCase.tagMetaCache != nil {
		useCase.tagMetaCache.Put(meta)
	}

	return meta, nil
}
