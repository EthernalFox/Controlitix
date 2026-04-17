package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/EthernalFox/Controlitix/ms-editor/internal/domain"
)

type MonitoringObjectUseCase struct {
	monitoringObjectRepository domain.MonitoringObjectRepository
	eventPublisher             domain.EventPublisher
	logger                     *slog.Logger
}

func NewMonitoringObjectUseCase(
	monitoringObjectRepository domain.MonitoringObjectRepository,
	eventPublisher domain.EventPublisher,
	logger *slog.Logger,
) *MonitoringObjectUseCase {
	if logger == nil {
		logger = slog.Default()
	}

	return &MonitoringObjectUseCase{
		monitoringObjectRepository: monitoringObjectRepository,
		eventPublisher:             eventPublisher,
		logger:                     logger,
	}
}

func (useCase *MonitoringObjectUseCase) CreateMonitoringObject(
	ctx context.Context,
	name string,
	description *string,
) (domain.MonitoringObject, error) {
	if validationError := domain.ValidateRequiredName("name", name); validationError != nil {
		return domain.MonitoringObject{}, validationError
	}

	monitoringObject := domain.MonitoringObject{
		Name:        name,
		Description: description,
	}

	return useCase.monitoringObjectRepository.CreateMonitoringObject(
		ctx,
		monitoringObject,
	)
}

func (useCase *MonitoringObjectUseCase) GetMonitoringObject(
	ctx context.Context,
	monitoringObjectID string,
) (domain.MonitoringObject, error) {
	if strings.TrimSpace(monitoringObjectID) == "" {
		return domain.MonitoringObject{}, fmt.Errorf("object_id is required: %w", domain.ErrInvalidInput)
	}

	return useCase.monitoringObjectRepository.GetMonitoringObject(
		ctx,
		monitoringObjectID,
	)
}

func (useCase *MonitoringObjectUseCase) UpdateMonitoringObject(
	ctx context.Context,
	monitoringObjectID string,
	name *string,
	description *string,
) (domain.MonitoringObject, error) {
	if strings.TrimSpace(monitoringObjectID) == "" {
		return domain.MonitoringObject{}, fmt.Errorf("object_id is required: %w", domain.ErrInvalidInput)
	}

	if validationError := domain.ValidateOptionalName("name", name); validationError != nil {
		return domain.MonitoringObject{}, validationError
	}

	update := domain.MonitoringObjectUpdate{
		Name:        name,
		Description: description,
	}

	return useCase.monitoringObjectRepository.UpdateMonitoringObject(
		ctx,
		monitoringObjectID,
		update,
	)
}

func (useCase *MonitoringObjectUseCase) DeleteMonitoringObject(
	ctx context.Context,
	monitoringObjectID string,
) error {
	if strings.TrimSpace(monitoringObjectID) == "" {
		return domain.ErrInvalidInput
	}

	deleteStats, deleteError := useCase.monitoringObjectRepository.DeleteMonitoringObject(
		ctx,
		monitoringObjectID,
	)
	if deleteError != nil {
		return deleteError
	}

	useCase.logger.Info(
		"monitoring object cascade soft deleted",
		"method",
		"DeleteMonitoringObject",
		"object_id",
		monitoringObjectID,
		"objects_deleted",
		deleteStats.ObjectsDeleted,
		"devices_deleted",
		deleteStats.DevicesDeleted,
		"tags_deleted",
		deleteStats.TagsDeleted,
		"diagrams_deleted",
		deleteStats.DiagramsDeleted,
		"figures_deleted",
		deleteStats.FiguresDeleted,
	)

	useCase.publishMonitoringObjectEvent(
		ctx,
		"DeleteMonitoringObject",
		monitoringObjectID,
		"deleted",
		nil,
	)

	return nil
}

func (useCase *MonitoringObjectUseCase) ListMonitoringObjects(
	ctx context.Context,
	query domain.ObjectListQuery,
) (domain.ListResult[domain.MonitoringObject], error) {
	return useCase.monitoringObjectRepository.ListMonitoringObjects(ctx, query)
}

func (useCase *MonitoringObjectUseCase) publishMonitoringObjectEvent(
	ctx context.Context,
	method string,
	monitoringObjectID string,
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
			"object_id",
			monitoringObjectID,
			"error",
			payloadError,
		)
		return
	}

	publishError := useCase.eventPublisher.Publish(ctx, domain.ConfigChangedEvent{
		EntityType: "object",
		EntityID:   monitoringObjectID,
		Operation:  operation,
		Timestamp:  time.Now().UTC(),
		Payload:    eventPayload,
	})
	if publishError != nil {
		useCase.logger.Error(
			"failed to publish config.changed event",
			"method",
			method,
			"object_id",
			monitoringObjectID,
			"error",
			publishError,
		)
	}
}
