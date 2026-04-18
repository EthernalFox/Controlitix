package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/EthernalFox/Controlitix/ms-editor/internal/domain"
)

type TagUseCase struct {
	deviceRepo     domain.DeviceRepository
	tagRepo        domain.TagRepository
	tagParamsRepo  domain.TagParamsRepository
	setpointsRepo  domain.TagSetpointsRepository
	scalingRepo    domain.TagScalingRepository
	referenceRepo  domain.ReferenceRepository
	eventPublisher domain.EventPublisher
	logger         *slog.Logger
}

func NewTagUseCase(
	deviceRepo domain.DeviceRepository,
	tagRepo domain.TagRepository,
	tagParamsRepo domain.TagParamsRepository,
	setpointsRepo domain.TagSetpointsRepository,
	scalingRepo domain.TagScalingRepository,
	referenceRepo domain.ReferenceRepository,
	eventPublisher domain.EventPublisher,
	logger *slog.Logger,
) *TagUseCase {
	if logger == nil {
		logger = slog.Default()
	}

	return &TagUseCase{
		deviceRepo:     deviceRepo,
		tagRepo:        tagRepo,
		tagParamsRepo:  tagParamsRepo,
		setpointsRepo:  setpointsRepo,
		scalingRepo:    scalingRepo,
		referenceRepo:  referenceRepo,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

func (useCase *TagUseCase) CreateTag(
	ctx context.Context,
	deviceID string,
	name string,
	description *string,
	params *domain.TagParams,
	setpoints *domain.TagSetpoints,
	scaling *domain.TagScaling,
) (domain.TagFull, error) {
	if strings.TrimSpace(deviceID) == "" {
		return domain.TagFull{}, fmt.Errorf("device_id is required: %w", domain.ErrInvalidInput)
	}

	if validationError := domain.ValidateRequiredName("name", name); validationError != nil {
		return domain.TagFull{}, validationError
	}

	if (setpoints != nil || scaling != nil) && params == nil {
		return domain.TagFull{}, fmt.Errorf(
			"params are required when setpoints or scaling are provided: %w",
			domain.ErrInvalidInput,
		)
	}

	deviceWithParams, getDeviceError := useCase.deviceRepo.GetDevice(ctx, deviceID)
	if getDeviceError != nil {
		return domain.TagFull{}, getDeviceError
	}

	if validationError := useCase.validateTagParams(
		ctx,
		deviceWithParams.Device.TypeName,
		params,
	); validationError != nil {
		return domain.TagFull{}, validationError
	}

	if setpointsValidationError := domain.ValidateSetpointsValue(setpoints); setpointsValidationError != nil {
		return domain.TagFull{}, setpointsValidationError
	}

	if scalingValidationError := domain.ValidateScalingValue(scaling); scalingValidationError != nil {
		return domain.TagFull{}, scalingValidationError
	}

	tag := domain.Tag{
		DeviceID:    deviceID,
		Name:        name,
		Description: description,
	}

	var createdTag domain.Tag
	var createdParams *domain.TagParams
	var createdSetpoints *domain.TagSetpoints
	var createdScaling *domain.TagScaling

	runCreate := func(executionContext context.Context) error {
		var createError error
		createdTag, createError = useCase.tagRepo.CreateTag(executionContext, tag)
		if createError != nil {
			return createError
		}

		if params == nil {
			return nil
		}

		tagParams, upsertParamsError := useCase.tagParamsRepo.UpsertTagParams(
			executionContext,
			createdTag.ID,
			normalizeTagParams(*params),
		)
		if upsertParamsError != nil {
			return upsertParamsError
		}
		createdParams = &tagParams

		if setpoints != nil {
			tagSetpoints, upsertSetpointsError := useCase.setpointsRepo.UpsertTagSetpoints(
				executionContext,
				tagParams.ID,
				normalizeTagSetpoints(tagParams.ID, *setpoints),
			)
			if upsertSetpointsError != nil {
				return upsertSetpointsError
			}
			createdSetpoints = &tagSetpoints
		}

		if scaling != nil {
			tagScaling, upsertScalingError := useCase.scalingRepo.UpsertTagScaling(
				executionContext,
				tagParams.ID,
				normalizeTagScaling(tagParams.ID, *scaling),
			)
			if upsertScalingError != nil {
				return upsertScalingError
			}
			createdScaling = &tagScaling
		}

		return nil
	}

	if transactionManager, ok := useCase.tagRepo.(transactionRunner); ok {
		if transactionError := transactionManager.InTransaction(ctx, runCreate); transactionError != nil {
			return domain.TagFull{}, transactionError
		}
	} else if createError := runCreate(ctx); createError != nil {
		return domain.TagFull{}, createError
	}

	tagFull := domain.TagFull{
		Tag:       createdTag,
		Params:    createdParams,
		Setpoints: createdSetpoints,
		Scaling:   createdScaling,
	}

	useCase.publishTagEvent(
		ctx,
		"CreateTag",
		tagFull.Tag.ID,
		tagFull.Tag.DeviceID,
		"created",
		tagFull,
	)

	return tagFull, nil
}

func (useCase *TagUseCase) GetTag(
	ctx context.Context,
	tagID string,
) (domain.TagFull, error) {
	if strings.TrimSpace(tagID) == "" {
		return domain.TagFull{}, domain.ErrInvalidInput
	}

	return useCase.tagRepo.GetTag(ctx, tagID)
}

func (useCase *TagUseCase) UpdateTag(
	ctx context.Context,
	tagID string,
	update domain.TagUpdate,
) (domain.Tag, error) {
	if strings.TrimSpace(tagID) == "" {
		return domain.Tag{}, fmt.Errorf("tag_id is required: %w", domain.ErrInvalidInput)
	}

	if validationError := domain.ValidateOptionalName("name", update.Name); validationError != nil {
		return domain.Tag{}, validationError
	}

	tag, updateError := useCase.tagRepo.UpdateTag(ctx, tagID, update)
	if updateError != nil {
		return domain.Tag{}, updateError
	}

	tagFull, getError := useCase.tagRepo.GetTag(ctx, tagID)
	if getError != nil {
		return domain.Tag{}, getError
	}

	useCase.publishTagEvent(
		ctx,
		"UpdateTag",
		tagFull.Tag.ID,
		tagFull.Tag.DeviceID,
		"updated",
		tagFull,
	)

	return tag, nil
}

func (useCase *TagUseCase) DeleteTag(
	ctx context.Context,
	tagID string,
) error {
	if strings.TrimSpace(tagID) == "" {
		return domain.ErrInvalidInput
	}

	tagFull, getError := useCase.tagRepo.GetTag(ctx, tagID)
	if getError != nil {
		return getError
	}

	deleteError := useCase.tagRepo.DeleteTag(ctx, tagID)
	if deleteError != nil {
		return deleteError
	}

	useCase.logger.Info(
		"tag soft deleted",
		"method",
		"DeleteTag",
		"tag_id",
		tagID,
		"tags_deleted",
		1,
	)

	useCase.publishTagEvent(
		ctx,
		"DeleteTag",
		tagFull.Tag.ID,
		tagFull.Tag.DeviceID,
		"deleted",
		nil,
	)

	return nil
}

func (useCase *TagUseCase) ListTags(
	ctx context.Context,
	query domain.TagListQuery,
) (domain.ListResult[domain.Tag], error) {
	if query.DeviceID != nil && strings.TrimSpace(*query.DeviceID) == "" {
		return domain.ListResult[domain.Tag]{}, domain.ErrInvalidInput
	}

	return useCase.tagRepo.ListTags(ctx, query)
}

func (useCase *TagUseCase) UpdateTagParams(
	ctx context.Context,
	tagID string,
	update domain.TagParamsUpdate,
) (domain.TagParams, error) {
	if strings.TrimSpace(tagID) == "" {
		return domain.TagParams{}, fmt.Errorf("tag_id is required: %w", domain.ErrInvalidInput)
	}

	if validationError := validateTagParamsUpdateInput(update); validationError != nil {
		return domain.TagParams{}, validationError
	}

	tagFull, getError := useCase.tagRepo.GetTag(ctx, tagID)
	if getError != nil {
		return domain.TagParams{}, getError
	}

	effectiveUnitID := unitIDForUpdate(update.UnitID, tagFull.Params)
	normalizedAddress := normalizeRawJSON(*update.Address)
	paramsToUpsert := domain.TagParams{
		TagID:      tagID,
		DataTypeID: *update.DataTypeID,
		UnitID:     effectiveUnitID,
		Address:    normalizedAddress,
	}

	deviceWithParams, getDeviceError := useCase.deviceRepo.GetDevice(ctx, tagFull.Tag.DeviceID)
	if getDeviceError != nil {
		return domain.TagParams{}, getDeviceError
	}

	if validationError := useCase.validateTagParams(
		ctx,
		deviceWithParams.Device.TypeName,
		&paramsToUpsert,
	); validationError != nil {
		return domain.TagParams{}, validationError
	}

	tagParams, updateError := useCase.tagParamsRepo.UpsertTagParams(
		ctx,
		tagID,
		paramsToUpsert,
	)
	if updateError != nil {
		return domain.TagParams{}, updateError
	}

	tagFull, getError = useCase.tagRepo.GetTag(ctx, tagID)
	if getError != nil {
		return domain.TagParams{}, getError
	}

	useCase.publishTagEvent(
		ctx,
		"UpdateTagParams",
		tagFull.Tag.ID,
		tagFull.Tag.DeviceID,
		"updated",
		tagFull,
	)

	return tagParams, nil
}

func (useCase *TagUseCase) UpdateTagSetpoints(
	ctx context.Context,
	tagID string,
	setpoints domain.TagSetpoints,
) (domain.TagSetpoints, error) {
	if strings.TrimSpace(tagID) == "" {
		return domain.TagSetpoints{}, fmt.Errorf("tag_id is required: %w", domain.ErrInvalidInput)
	}

	if validationError := domain.ValidateSetpoints(setpoints); validationError != nil {
		return domain.TagSetpoints{}, validationError
	}

	tagFull, getError := useCase.tagRepo.GetTag(ctx, tagID)
	if getError != nil {
		return domain.TagSetpoints{}, getError
	}
	if tagFull.Params == nil {
		return domain.TagSetpoints{}, fmt.Errorf(
			"tag params must exist before updating setpoints: %w",
			domain.ErrInvalidInput,
		)
	}

	tagSetpoints, updateError := useCase.setpointsRepo.UpsertTagSetpoints(
		ctx,
		tagFull.Params.ID,
		normalizeTagSetpoints(tagFull.Params.ID, setpoints),
	)
	if updateError != nil {
		return domain.TagSetpoints{}, updateError
	}

	tagFull, getError = useCase.tagRepo.GetTag(ctx, tagID)
	if getError != nil {
		return domain.TagSetpoints{}, getError
	}

	useCase.publishTagEvent(
		ctx,
		"UpdateTagSetpoints",
		tagFull.Tag.ID,
		tagFull.Tag.DeviceID,
		"updated",
		tagFull,
	)

	return tagSetpoints, nil
}

func (useCase *TagUseCase) DeleteTagSetpoints(
	ctx context.Context,
	tagID string,
) error {
	if strings.TrimSpace(tagID) == "" {
		return domain.ErrInvalidInput
	}

	tagFull, getError := useCase.tagRepo.GetTag(ctx, tagID)
	if getError != nil {
		return getError
	}
	if tagFull.Params == nil {
		return domain.ErrInvalidInput
	}

	deleteError := useCase.setpointsRepo.DeleteTagSetpoints(ctx, tagFull.Params.ID)
	if deleteError != nil {
		return deleteError
	}

	tagFull, getError = useCase.tagRepo.GetTag(ctx, tagID)
	if getError != nil {
		return getError
	}

	useCase.publishTagEvent(
		ctx,
		"DeleteTagSetpoints",
		tagFull.Tag.ID,
		tagFull.Tag.DeviceID,
		"updated",
		tagFull,
	)

	return nil
}

func (useCase *TagUseCase) UpdateTagScaling(
	ctx context.Context,
	tagID string,
	scaling domain.TagScaling,
) (domain.TagScaling, error) {
	if strings.TrimSpace(tagID) == "" {
		return domain.TagScaling{}, fmt.Errorf("tag_id is required: %w", domain.ErrInvalidInput)
	}

	if validationError := domain.ValidateScaling(scaling); validationError != nil {
		return domain.TagScaling{}, validationError
	}

	tagFull, getError := useCase.tagRepo.GetTag(ctx, tagID)
	if getError != nil {
		return domain.TagScaling{}, getError
	}
	if tagFull.Params == nil {
		return domain.TagScaling{}, fmt.Errorf(
			"tag params must exist before updating scaling: %w",
			domain.ErrInvalidInput,
		)
	}

	tagScaling, updateError := useCase.scalingRepo.UpsertTagScaling(
		ctx,
		tagFull.Params.ID,
		normalizeTagScaling(tagFull.Params.ID, scaling),
	)
	if updateError != nil {
		return domain.TagScaling{}, updateError
	}

	tagFull, getError = useCase.tagRepo.GetTag(ctx, tagID)
	if getError != nil {
		return domain.TagScaling{}, getError
	}

	useCase.publishTagEvent(
		ctx,
		"UpdateTagScaling",
		tagFull.Tag.ID,
		tagFull.Tag.DeviceID,
		"updated",
		tagFull,
	)

	return tagScaling, nil
}

func (useCase *TagUseCase) DeleteTagScaling(
	ctx context.Context,
	tagID string,
) error {
	if strings.TrimSpace(tagID) == "" {
		return domain.ErrInvalidInput
	}

	tagFull, getError := useCase.tagRepo.GetTag(ctx, tagID)
	if getError != nil {
		return getError
	}
	if tagFull.Params == nil {
		return domain.ErrInvalidInput
	}

	deleteError := useCase.scalingRepo.DeleteTagScaling(ctx, tagFull.Params.ID)
	if deleteError != nil {
		return deleteError
	}

	tagFull, getError = useCase.tagRepo.GetTag(ctx, tagID)
	if getError != nil {
		return getError
	}

	useCase.publishTagEvent(
		ctx,
		"DeleteTagScaling",
		tagFull.Tag.ID,
		tagFull.Tag.DeviceID,
		"updated",
		tagFull,
	)

	return nil
}

func (useCase *TagUseCase) ListDataTypes(
	ctx context.Context,
) ([]domain.DataType, error) {
	return useCase.referenceRepo.ListDataTypes(ctx)
}

func (useCase *TagUseCase) ListUnits(
	ctx context.Context,
	category string,
) ([]domain.Unit, error) {
	trimmedCategory := strings.TrimSpace(category)
	if trimmedCategory == "" {
		return useCase.referenceRepo.ListUnits(ctx)
	}

	return useCase.referenceRepo.ListUnitsByCategory(ctx, trimmedCategory)
}

func (useCase *TagUseCase) publishTagEvent(
	ctx context.Context,
	method string,
	tagID string,
	deviceID string,
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
			"tag_id",
			tagID,
			"device_id",
			deviceID,
			"error",
			payloadError,
		)
		return
	}

	publishError := useCase.eventPublisher.Publish(ctx, domain.ConfigChangedEvent{
		EntityType: "tag",
		EntityID:   tagID,
		Operation:  operation,
		Timestamp:  time.Now().UTC(),
		Payload:    eventPayload,
	})
	if publishError != nil {
		useCase.logger.Error(
			"failed to publish config.changed event",
			"method",
			method,
			"tag_id",
			tagID,
			"device_id",
			deviceID,
			"error",
			publishError,
		)
	}
}

func validateTagParamsUpdateInput(update domain.TagParamsUpdate) error {
	if update.DataTypeID == nil || update.Address == nil {
		return fmt.Errorf("data_type_id and address are required: %w", domain.ErrInvalidInput)
	}

	if *update.DataTypeID <= 0 {
		return fmt.Errorf("data_type_id is required: %w", domain.ErrInvalidInput)
	}

	return nil
}

func normalizeTagParams(params domain.TagParams) domain.TagParams {
	params.Address = normalizeRawJSON(params.Address)
	params.UnitID = normalizeUnitID(params.UnitID)
	return params
}

func normalizeTagSetpoints(
	paramID string,
	setpoints domain.TagSetpoints,
) domain.TagSetpoints {
	setpoints.ParamID = paramID
	return setpoints
}

func normalizeTagScaling(
	paramID string,
	scaling domain.TagScaling,
) domain.TagScaling {
	scaling.ParamID = paramID
	return scaling
}

func normalizeRawJSON(value json.RawMessage) json.RawMessage {
	trimmedValue := strings.TrimSpace(string(value))
	if trimmedValue == "" || trimmedValue == "null" {
		return json.RawMessage("{}")
	}

	return value
}

func normalizeUnitID(unitID *int) *int {
	if unitID == nil || *unitID == 0 {
		return nil
	}

	return unitID
}

func unitIDForUpdate(
	updateUnitID *int,
	currentParams *domain.TagParams,
) *int {
	if updateUnitID == nil {
		if currentParams == nil || currentParams.UnitID == nil {
			return nil
		}

		currentUnitID := *currentParams.UnitID
		return &currentUnitID
	}

	return normalizeUnitID(updateUnitID)
}

func (useCase *TagUseCase) validateTagParams(
	ctx context.Context,
	deviceTypeName string,
	params *domain.TagParams,
) error {
	if params == nil {
		return nil
	}

	if params.DataTypeID <= 0 {
		return fmt.Errorf("data_type_id is required: %w", domain.ErrInvalidInput)
	}

	dataTypes, listDataTypesError := useCase.referenceRepo.ListDataTypes(ctx)
	if listDataTypesError != nil {
		return listDataTypesError
	}

	if !containsDataType(dataTypes, params.DataTypeID) {
		return domain.NewValidationError(domain.FieldError{
			Field:   "data_type_id",
			Message: "must reference an existing data type",
		})
	}

	if params.UnitID != nil {
		units, listUnitsError := useCase.referenceRepo.ListUnits(ctx)
		if listUnitsError != nil {
			return listUnitsError
		}

		if !containsUnit(units, *params.UnitID) {
			return domain.NewValidationError(domain.FieldError{
				Field:   "unit_id",
				Message: "must reference an existing unit",
			})
		}
	}

	if addressValidationError := domain.ValidateTagAddress(deviceTypeName, normalizeRawJSON(params.Address)); addressValidationError != nil {
		return addressValidationError
	}

	return nil
}

func containsDataType(dataTypes []domain.DataType, dataTypeID int) bool {
	for _, dataType := range dataTypes {
		if dataType.ID == dataTypeID {
			return true
		}
	}

	return false
}

func containsUnit(units []domain.Unit, unitID int) bool {
	for _, unit := range units {
		if unit.ID == unitID {
			return true
		}
	}

	return false
}
