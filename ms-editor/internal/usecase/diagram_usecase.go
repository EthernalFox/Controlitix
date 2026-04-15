package usecase

import (
	"context"
	"strings"

	"github.com/EthernalFox/Controlitix/ms-editor/internal/domain"
)

type DiagramUseCase struct {
	diagramRepository domain.DiagramRepository
}

func NewDiagramUseCase(
	diagramRepository domain.DiagramRepository,
) *DiagramUseCase {
	return &DiagramUseCase{
		diagramRepository: diagramRepository,
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

	return useCase.diagramRepository.DeleteDiagram(
		ctx,
		diagramID,
	)
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

func (useCase *DiagramUseCase) ListDiagramsByMonitoringObject(
	ctx context.Context,
	monitoringObjectID string,
) ([]domain.Diagram, error) {
	if strings.TrimSpace(monitoringObjectID) == "" {
		return nil, domain.ErrInvalidInput
	}

	return useCase.diagramRepository.ListDiagramsByMonitoringObject(
		ctx,
		monitoringObjectID,
	)
}
