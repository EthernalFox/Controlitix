package usecase

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/EthernalFox/Controlitix/ms-editor/internal/domain"
)

type DiagramUseCase struct {
	diagramRepository domain.DiagramRepository
	eventPublisher    domain.EventPublisher
	logger            *slog.Logger
}

func NewDiagramUseCase(
	diagramRepository domain.DiagramRepository,
	eventPublisher domain.EventPublisher,
	logger *slog.Logger,
) *DiagramUseCase {
	if logger == nil {
		logger = slog.Default()
	}

	return &DiagramUseCase{
		diagramRepository: diagramRepository,
		eventPublisher:    eventPublisher,
		logger:            logger,
	}
}

func (useCase *DiagramUseCase) CreateDiagram(
	ctx context.Context,
	objectID string,
	name *string,
	description *string,
) (domain.Diagram, error) {
	if strings.TrimSpace(objectID) == "" {
		return domain.Diagram{}, domain.ErrInvalidInput
	}

	diagram := domain.Diagram{
		ObjectID:    objectID,
		Name:        name,
		Description: description,
	}

	return useCase.diagramRepository.CreateDiagram(ctx, diagram)
}

func (useCase *DiagramUseCase) GetDiagram(
	ctx context.Context,
	diagramID string,
) (domain.Diagram, error) {
	if strings.TrimSpace(diagramID) == "" {
		return domain.Diagram{}, domain.ErrInvalidInput
	}

	return useCase.diagramRepository.GetDiagram(ctx, diagramID)
}

func (useCase *DiagramUseCase) UpdateDiagram(
	ctx context.Context,
	diagramID string,
	name *string,
	description *string,
) (domain.Diagram, error) {
	if strings.TrimSpace(diagramID) == "" {
		return domain.Diagram{}, domain.ErrInvalidInput
	}

	update := domain.DiagramUpdate{
		Name:        name,
		Description: description,
	}

	return useCase.diagramRepository.UpdateDiagram(
		ctx,
		diagramID,
		update,
	)
}

func (useCase *DiagramUseCase) DeleteDiagram(
	ctx context.Context,
	diagramID string,
) error {
	if strings.TrimSpace(diagramID) == "" {
		return domain.ErrInvalidInput
	}

	deleteStats, deleteError := useCase.diagramRepository.DeleteDiagram(
		ctx,
		diagramID,
	)
	if deleteError != nil {
		return deleteError
	}

	useCase.logger.Info(
		"diagram cascade soft deleted",
		"method",
		"DeleteDiagram",
		"diagram_id",
		diagramID,
		"diagrams_deleted",
		deleteStats.DiagramsDeleted,
		"figures_deleted",
		deleteStats.FiguresDeleted,
	)

	useCase.publishDiagramEvent(
		ctx,
		"DeleteDiagram",
		diagramID,
		"deleted",
		nil,
	)

	return nil
}

func (useCase *DiagramUseCase) PublishDiagram(
	ctx context.Context,
	diagramID string,
) (domain.Diagram, error) {
	if strings.TrimSpace(diagramID) == "" {
		return domain.Diagram{}, domain.ErrInvalidInput
	}

	return useCase.diagramRepository.PublishDiagram(
		ctx,
		diagramID,
	)
}

func (useCase *DiagramUseCase) ListDiagrams(
	ctx context.Context,
	query domain.DiagramListQuery,
) (domain.ListResult[domain.Diagram], error) {
	if query.ObjectID != nil && strings.TrimSpace(*query.ObjectID) == "" {
		return domain.ListResult[domain.Diagram]{}, domain.ErrInvalidInput
	}

	return useCase.diagramRepository.ListDiagrams(ctx, query)
}

func (useCase *DiagramUseCase) publishDiagramEvent(
	ctx context.Context,
	method string,
	diagramID string,
	operation string,
	payload any,
) {
	if useCase.eventPublisher == nil {
		return
	}

	eventPayload, payloadError := marshalEventPayload(payload)
	if payloadError != nil {
		useCase.logger.Error(
			"failed to marshal config.changed payload",
			"method",
			method,
			"diagram_id",
			diagramID,
			"error",
			payloadError,
		)
		return
	}

	publishError := useCase.eventPublisher.Publish(ctx, domain.ConfigChangedEvent{
		EntityType: "diagram",
		EntityID:   diagramID,
		Operation:  operation,
		Timestamp:  time.Now().UTC(),
		Payload:    eventPayload,
	})
	if publishError != nil {
		useCase.logger.Error(
			"failed to publish config.changed event",
			"method",
			method,
			"diagram_id",
			diagramID,
			"error",
			publishError,
		)
	}
}
