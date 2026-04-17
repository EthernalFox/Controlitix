package usecase

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"github.com/EthernalFox/Controlitix/ms-editor/internal/domain"
)

type transactionRunner interface {
	InTransaction(ctx context.Context, operation func(context.Context) error) error
}

type DeviceUseCase struct {
	deviceRepo       domain.DeviceRepository
	deviceParamsRepo domain.DeviceParamsRepository
	deviceTypeRepo   domain.DeviceTypeRepository
	eventPublisher   domain.EventPublisher
	logger           *slog.Logger
}

func NewDeviceUseCase(
	deviceRepo domain.DeviceRepository,
	deviceParamsRepo domain.DeviceParamsRepository,
	deviceTypeRepo domain.DeviceTypeRepository,
	eventPublisher domain.EventPublisher,
	logger *slog.Logger,
) *DeviceUseCase {
	if logger == nil {
		logger = slog.Default()
	}

	return &DeviceUseCase{
		deviceRepo:       deviceRepo,
		deviceParamsRepo: deviceParamsRepo,
		deviceTypeRepo:   deviceTypeRepo,
		eventPublisher:   eventPublisher,
		logger:           logger,
	}
}

func (useCase *DeviceUseCase) CreateDevice(
	ctx context.Context,
	objectID *string,
	typeID int,
	name string,
	description *string,
	settings json.RawMessage,
) (domain.DeviceWithParams, error) {
	if !isValidOptionalIdentifier(objectID) || typeID <= 0 || strings.TrimSpace(name) == "" {
		return domain.DeviceWithParams{}, domain.ErrInvalidInput
	}

	device := domain.Device{
		ObjectID:    normalizeOptionalIdentifier(objectID),
		TypeID:      typeID,
		Name:        name,
		Description: description,
	}

	normalizedSettings := normalizeSettings(settings)

	var createdDevice domain.Device
	var deviceParams domain.DeviceParams

	runCreate := func(executionContext context.Context) error {
		var createError error
		createdDevice, createError = useCase.deviceRepo.CreateDevice(executionContext, device)
		if createError != nil {
			return createError
		}

		deviceParams, createError = useCase.deviceParamsRepo.UpsertDeviceParams(
			executionContext,
			createdDevice.ID,
			normalizedSettings,
		)
		if createError != nil {
			return createError
		}

		return nil
	}

	if transactionManager, ok := useCase.deviceRepo.(transactionRunner); ok {
		if transactionError := transactionManager.InTransaction(ctx, runCreate); transactionError != nil {
			return domain.DeviceWithParams{}, transactionError
		}
	} else if createError := runCreate(ctx); createError != nil {
		return domain.DeviceWithParams{}, createError
	}

	deviceWithParams := domain.DeviceWithParams{
		Device: createdDevice,
		Params: &deviceParams,
	}

	useCase.publishDeviceEvent(
		ctx,
		"CreateDevice",
		deviceWithParams.Device.ID,
		deviceWithParams.Device.ObjectID,
		"created",
		deviceWithParams,
	)

	return deviceWithParams, nil
}

func (useCase *DeviceUseCase) GetDevice(
	ctx context.Context,
	deviceID string,
) (domain.DeviceWithParams, error) {
	if strings.TrimSpace(deviceID) == "" {
		return domain.DeviceWithParams{}, domain.ErrInvalidInput
	}

	return useCase.deviceRepo.GetDevice(ctx, deviceID)
}

func (useCase *DeviceUseCase) UpdateDevice(
	ctx context.Context,
	deviceID string,
	update domain.DeviceUpdate,
) (domain.Device, error) {
	if strings.TrimSpace(deviceID) == "" || !isValidDeviceUpdate(update) {
		return domain.Device{}, domain.ErrInvalidInput
	}

	device, updateError := useCase.deviceRepo.UpdateDevice(ctx, deviceID, update)
	if updateError != nil {
		return domain.Device{}, updateError
	}

	deviceWithParams, getError := useCase.deviceRepo.GetDevice(ctx, deviceID)
	if getError != nil {
		return domain.Device{}, getError
	}

	useCase.publishDeviceEvent(
		ctx,
		"UpdateDevice",
		device.ID,
		device.ObjectID,
		"updated",
		deviceWithParams,
	)

	return device, nil
}

func (useCase *DeviceUseCase) DeleteDevice(
	ctx context.Context,
	deviceID string,
) error {
	if strings.TrimSpace(deviceID) == "" {
		return domain.ErrInvalidInput
	}

	deviceWithParams, getError := useCase.deviceRepo.GetDevice(ctx, deviceID)
	if getError != nil {
		return getError
	}

	deleteStats, deleteError := useCase.deviceRepo.DeleteDevice(ctx, deviceID)
	if deleteError != nil {
		return deleteError
	}

	useCase.logger.Info(
		"device cascade soft deleted",
		"method",
		"DeleteDevice",
		"device_id",
		deviceID,
		"devices_deleted",
		deleteStats.DevicesDeleted,
		"tags_deleted",
		deleteStats.TagsDeleted,
	)

	useCase.publishDeviceEvent(
		ctx,
		"DeleteDevice",
		deviceWithParams.Device.ID,
		deviceWithParams.Device.ObjectID,
		"deleted",
		nil,
	)

	return nil
}

func (useCase *DeviceUseCase) ListDevices(
	ctx context.Context,
	query domain.DeviceListQuery,
) (domain.ListResult[domain.Device], error) {
	if query.ObjectID != nil && strings.TrimSpace(*query.ObjectID) == "" {
		return domain.ListResult[domain.Device]{}, domain.ErrInvalidInput
	}

	return useCase.deviceRepo.ListDevices(ctx, query)
}

func (useCase *DeviceUseCase) AssignDeviceToObject(
	ctx context.Context,
	deviceID string,
	objectID *string,
) (domain.Device, error) {
	if strings.TrimSpace(deviceID) == "" || !isValidOptionalIdentifier(objectID) {
		return domain.Device{}, domain.ErrInvalidInput
	}

	device, assignError := useCase.deviceRepo.AssignDeviceToObject(
		ctx,
		deviceID,
		normalizeOptionalIdentifier(objectID),
	)
	if assignError != nil {
		return domain.Device{}, assignError
	}

	deviceWithParams, getError := useCase.deviceRepo.GetDevice(ctx, deviceID)
	if getError != nil {
		return domain.Device{}, getError
	}

	useCase.publishDeviceEvent(
		ctx,
		"AssignDeviceToObject",
		device.ID,
		device.ObjectID,
		"updated",
		deviceWithParams,
	)

	return device, nil
}

func (useCase *DeviceUseCase) UpdateDeviceParams(
	ctx context.Context,
	deviceID string,
	settings json.RawMessage,
) (domain.DeviceParams, error) {
	if strings.TrimSpace(deviceID) == "" {
		return domain.DeviceParams{}, domain.ErrInvalidInput
	}

	deviceParams, updateError := useCase.deviceParamsRepo.UpsertDeviceParams(
		ctx,
		deviceID,
		normalizeSettings(settings),
	)
	if updateError != nil {
		return domain.DeviceParams{}, updateError
	}

	deviceWithParams, getError := useCase.deviceRepo.GetDevice(ctx, deviceID)
	if getError != nil {
		return domain.DeviceParams{}, getError
	}

	useCase.publishDeviceEvent(
		ctx,
		"UpdateDeviceParams",
		deviceWithParams.Device.ID,
		deviceWithParams.Device.ObjectID,
		"updated",
		deviceWithParams,
	)

	return deviceParams, nil
}

func (useCase *DeviceUseCase) ListDeviceTypes(
	ctx context.Context,
) ([]domain.DeviceType, error) {
	return useCase.deviceTypeRepo.ListDeviceTypes(ctx)
}

func (useCase *DeviceUseCase) publishDeviceEvent(
	ctx context.Context,
	method string,
	deviceID string,
	objectID *string,
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
			"device_id",
			deviceID,
			"object_id",
			objectID,
			"error",
			payloadError,
		)
		return
	}

	publishError := useCase.eventPublisher.Publish(ctx, domain.ConfigChangedEvent{
		EntityType: "device",
		EntityID:   deviceID,
		Operation:  operation,
		Timestamp:  time.Now().UTC(),
		Payload:    eventPayload,
	})
	if publishError != nil {
		useCase.logger.Error(
			"failed to publish config.changed event",
			"method",
			method,
			"device_id",
			deviceID,
			"object_id",
			objectID,
			"error",
			publishError,
		)
	}
}

func marshalEventPayload(payload any) (json.RawMessage, error) {
	if payload == nil {
		return json.RawMessage("null"), nil
	}

	encodedPayload, marshalError := json.Marshal(payload)
	if marshalError != nil {
		return nil, marshalError
	}

	return encodedPayload, nil
}

func normalizeSettings(settings json.RawMessage) json.RawMessage {
	trimmedSettings := strings.TrimSpace(string(settings))
	if trimmedSettings == "" || trimmedSettings == "null" {
		return json.RawMessage("{}")
	}

	return settings
}

func normalizeOptionalIdentifier(identifier *string) *string {
	if identifier == nil {
		return nil
	}

	trimmedIdentifier := strings.TrimSpace(*identifier)
	return &trimmedIdentifier
}

func isValidOptionalIdentifier(identifier *string) bool {
	return identifier == nil || strings.TrimSpace(*identifier) != ""
}

func isValidDeviceUpdate(update domain.DeviceUpdate) bool {
	if update.TypeID != nil && *update.TypeID <= 0 {
		return false
	}

	if update.Name != nil && strings.TrimSpace(*update.Name) == "" {
		return false
	}

	return update.Description == nil || true
}
