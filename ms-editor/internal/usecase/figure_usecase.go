package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/EthernalFox/Controlitix/ms-editor/internal/domain"
	"github.com/google/uuid"
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

func (useCase *FigureUseCase) BulkUpsertFigures(
	ctx context.Context,
	diagramID string,
	items []domain.FigureBulkItem,
) (domain.FigureBulkResult, error) {
	if strings.TrimSpace(diagramID) == "" {
		return domain.FigureBulkResult{}, fmt.Errorf("diagram_id is required: %w", domain.ErrInvalidInput)
	}

	validationFields := make([]domain.FieldError, 0)
	normalizedItems := make([]domain.FigureBulkItem, 0, len(items))
	seenIDs := make(map[string]struct{}, len(items))

	for index, item := range items {
		normalizedItem, fieldErrors := validateFigureBulkItem(index, item, seenIDs)
		if len(fieldErrors) > 0 {
			validationFields = append(validationFields, fieldErrors...)
		}

		normalizedItems = append(normalizedItems, normalizedItem)
	}

	if len(validationFields) > 0 {
		return domain.FigureBulkResult{}, domain.NewValidationError(validationFields...)
	}

	bulkResult, bulkError := useCase.figureRepository.BulkUpsertFigures(
		ctx,
		diagramID,
		normalizedItems,
	)
	if bulkError != nil {
		return domain.FigureBulkResult{}, bulkError
	}

	for _, figureID := range bulkResult.CreatedIDs {
		useCase.publishFigureEvent(
			ctx,
			"BulkUpsertFigures",
			figureID,
			"created",
			nil,
		)
	}

	for _, figureID := range bulkResult.UpdatedIDs {
		useCase.publishFigureEvent(
			ctx,
			"BulkUpsertFigures",
			figureID,
			"updated",
			nil,
		)
	}

	for _, figureID := range bulkResult.DeletedIDs {
		useCase.publishFigureEvent(
			ctx,
			"BulkUpsertFigures",
			figureID,
			"deleted",
			nil,
		)
	}

	return bulkResult, nil
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

func validateFigureBulkItem(
	index int,
	item domain.FigureBulkItem,
	seenIDs map[string]struct{},
) (domain.FigureBulkItem, []domain.FieldError) {
	validationFields := make([]domain.FieldError, 0)
	fieldPrefix := fmt.Sprintf("figures[%d]", index)

	normalizedItem := domain.FigureBulkItem{
		Parameters: item.Parameters,
	}

	figureType, isSupportedFigureType := domain.ParseFigureType(string(item.FigureType))
	if !isSupportedFigureType {
		validationFields = append(validationFields, domain.FieldError{
			Field:   fieldPrefix + ".type",
			Message: "must be one of supported figure types",
		})
	} else {
		normalizedItem.FigureType = figureType
	}

	if item.ID != nil {
		figureID, parseFigureIDError := parseUUIDValue(*item.ID)
		if parseFigureIDError != nil {
			validationFields = append(validationFields, domain.FieldError{
				Field:   fieldPrefix + ".id",
				Message: parseFigureIDError.Error(),
			})
		} else {
			if _, alreadyExists := seenIDs[figureID]; alreadyExists {
				validationFields = append(validationFields, domain.FieldError{
					Field:   fieldPrefix + ".id",
					Message: "must be unique",
				})
			}
			seenIDs[figureID] = struct{}{}
			normalizedItem.ID = &figureID
		}
	}

	if item.TagID != nil {
		tagID, parseTagIDError := parseUUIDValue(*item.TagID)
		if parseTagIDError != nil {
			validationFields = append(validationFields, domain.FieldError{
				Field:   fieldPrefix + ".tag_id",
				Message: parseTagIDError.Error(),
			})
		} else {
			normalizedItem.TagID = &tagID
		}
	}

	if validationError := validateFigureParameters(item.Parameters); validationError != nil {
		validationFields = append(validationFields, domain.FieldError{
			Field:   fieldPrefix + ".params",
			Message: validationError.Error(),
		})
	}

	return normalizedItem, validationFields
}

func parseUUIDValue(rawValue string) (string, error) {
	trimmedValue := strings.TrimSpace(rawValue)
	if trimmedValue == "" {
		return "", errors.New("must not be empty")
	}

	parsedValue, parseError := uuid.Parse(trimmedValue)
	if parseError != nil {
		return "", errors.New("must be a valid UUID")
	}

	return parsedValue.String(), nil
}

func validateFigureParameters(parameters json.RawMessage) error {
	trimmedParameters := strings.TrimSpace(string(parameters))
	if trimmedParameters == "" {
		return errors.New("must be a valid JSON object")
	}

	var parametersObject map[string]json.RawMessage
	if unmarshalError := json.Unmarshal([]byte(trimmedParameters), &parametersObject); unmarshalError != nil {
		return errors.New("must be a valid JSON object")
	}

	if parametersObject == nil {
		return errors.New("must be a valid JSON object")
	}

	return nil
}
