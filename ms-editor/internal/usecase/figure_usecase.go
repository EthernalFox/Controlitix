package usecase

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/EthernalFox/Controlitix/ms-editor/internal/domain"
)

type FigureUseCase struct {
	figureRepository domain.FigureRepository
}

func NewFigureUseCase(
	figureRepository domain.FigureRepository,
) *FigureUseCase {
	return &FigureUseCase{
		figureRepository: figureRepository,
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

	return useCase.figureRepository.DeleteFigure(
		ctx,
		figureID,
	)
}

func (useCase *FigureUseCase) ListFiguresByDiagram(
	ctx context.Context,
	diagramID string,
) ([]domain.Figure, error) {
	if strings.TrimSpace(diagramID) == "" {
		return nil, domain.ErrInvalidInput
	}

	return useCase.figureRepository.ListFiguresByDiagram(
		ctx,
		diagramID,
	)
}
