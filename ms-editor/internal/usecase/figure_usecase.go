package usecase

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"github.com/EthernalFox/Controlitix/ms-editor/internal/domain"
)

type FigureUseCase struct {
	figureRepository domain.FigureRepository
	eventPublisher   domain.EventPublisher
	logger           *slog.Logger
}

func NewFigureUseCase(
	figureRepository domain.FigureRepository,
	eventPublisher domain.EventPublisher,
	logger *slog.Logger,
) *FigureUseCase {
	if logger == nil {
		logger = slog.Default()
	}

	return &FigureUseCase{
		figureRepository: figureRepository,
		eventPublisher:   eventPublisher,
		logger:           logger,
	}
}

func (useCase *FigureUseCase) CreateFigures(
	ctx context.Context,
	diagramID string,
	figures []domain.Figure,
) ([]domain.Figure, error) {
	if strings.TrimSpace(diagramID) == "" {
		return nil, domain.ErrInvalidInput
	}

	return useCase.figureRepository.CreateFigures(
		ctx,
		diagramID,
		figures,
	)
}

func (useCase *FigureUseCase) UpdateFigure(
	ctx context.Context,
	figureID string,
	tagID *string,
	figureType *domain.FigureType,
	parameters *json.RawMessage,
) (domain.Figure, error) {
	if strings.TrimSpace(figureID) == "" {
		return domain.Figure{}, domain.ErrInvalidInput
	}

	update := domain.FigureUpdate{
		TagID:      tagID,
		FigureType: figureType,
		Parameters: parameters,
	}

	return useCase.figureRepository.UpdateFigure(
		ctx,
		figureID,
		update,
	)
}

func (useCase *FigureUseCase) DeleteFigure(
	ctx context.Context,
	figureID string,
) error {
	if strings.TrimSpace(figureID) == "" {
		return domain.ErrInvalidInput
	}

	deleteError := useCase.figureRepository.DeleteFigure(
		ctx,
		figureID,
	)
	if deleteError != nil {
		return deleteError
	}

	useCase.logger.Info(
		"figure soft deleted",
		"method",
		"DeleteFigure",
		"figure_id",
		figureID,
		"figures_deleted",
		1,
	)

	useCase.publishFigureEvent(
		ctx,
		"DeleteFigure",
		figureID,
		"deleted",
		nil,
	)

	return nil
}

func (useCase *FigureUseCase) ListFigures(
	ctx context.Context,
	query domain.FigureListQuery,
) (domain.ListResult[domain.Figure], error) {
	if strings.TrimSpace(query.DiagramID) == "" {
		return domain.ListResult[domain.Figure]{}, domain.ErrInvalidInput
	}

	return useCase.figureRepository.ListFigures(ctx, query)
}

func (useCase *FigureUseCase) publishFigureEvent(
	ctx context.Context,
	method string,
	figureID string,
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
			"figure_id",
			figureID,
			"error",
			payloadError,
		)
		return
	}

	publishError := useCase.eventPublisher.Publish(ctx, domain.ConfigChangedEvent{
		EntityType: "figure",
		EntityID:   figureID,
		Operation:  operation,
		Timestamp:  time.Now().UTC(),
		Payload:    eventPayload,
	})
	if publishError != nil {
		useCase.logger.Error(
			"failed to publish config.changed event",
			"method",
			method,
			"figure_id",
			figureID,
			"error",
			publishError,
		)
	}
}
