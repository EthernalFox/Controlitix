package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/EthernalFox/Controlitix/ms-poll/internal/domain"
	"github.com/EthernalFox/Controlitix/ms-poll/internal/infrastructure/repository"
)

type ConfigEventHandler struct {
	repository configEventRepository
	store      *SnapshotStore
	logger     *slog.Logger
}

func NewConfigEventHandler(
	repository *repository.ConfigRepository,
	store *SnapshotStore,
	logger *slog.Logger,
) *ConfigEventHandler {
	return newConfigEventHandler(repository, store, logger)
}

func newConfigEventHandler(
	repository configEventRepository,
	store *SnapshotStore,
	logger *slog.Logger,
) *ConfigEventHandler {
	if logger == nil {
		logger = slog.Default()
	}

	return &ConfigEventHandler{
		repository: repository,
		store:      store,
		logger:     logger,
	}
}

func (handler *ConfigEventHandler) Handle(
	ctx context.Context,
	event domain.ConfigChangedEvent,
) error {
	entityType := strings.ToLower(event.EntityType)
	operation := strings.ToLower(event.Operation)

	switch entityType {
	case "device":
		return handler.handleDeviceEvent(ctx, event.EntityID, operation)
	case "tag":
		return handler.handleTagEvent(ctx, event.EntityID, operation)
	default:
		handler.logger.Debug(
			"ignored entity_type",
			"entity_type",
			event.EntityType,
			"entity_id",
			event.EntityID,
			"operation",
			event.Operation,
		)
		return nil
	}
}

func (handler *ConfigEventHandler) handleDeviceEvent(
	ctx context.Context,
	deviceID string,
	operation string,
) error {
	switch operation {
	case "create", "update":
		deviceSnapshot, loadError := handler.repository.LoadDevice(ctx, deviceID)
		if loadError != nil {
			return fmt.Errorf("load device %s: %w", deviceID, loadError)
		}

		if deviceSnapshot == nil {
			cascadedTags := handler.cascadedTagsCount(deviceID)
			handler.store.RemoveDevice(deviceID)
			handler.logger.Info(
				"device removed",
				"device_id",
				deviceID,
				"cascaded_tags",
				cascadedTags,
			)
			return nil
		}

		handler.store.UpsertDevice(deviceSnapshot)
		handler.logger.Info(
			"device upserted",
			"device_id",
			deviceSnapshot.ID,
			"type",
			deviceSnapshot.TypeName,
		)
		return nil
	case "delete":
		cascadedTags := handler.cascadedTagsCount(deviceID)
		handler.store.RemoveDevice(deviceID)
		handler.logger.Info(
			"device removed",
			"device_id",
			deviceID,
			"cascaded_tags",
			cascadedTags,
		)
		return nil
	default:
		handler.logger.Debug(
			"ignored device operation",
			"device_id",
			deviceID,
			"operation",
			operation,
		)
		return nil
	}
}

func (handler *ConfigEventHandler) handleTagEvent(
	ctx context.Context,
	tagID string,
	operation string,
) error {
	switch operation {
	case "create", "update":
		tagSnapshot, loadError := handler.repository.LoadTag(ctx, tagID)
		if loadError != nil {
			return fmt.Errorf("load tag %s: %w", tagID, loadError)
		}

		if tagSnapshot == nil {
			handler.store.RemoveTag(tagID)
			handler.logger.Info("tag removed", "tag_id", tagID)
			return nil
		}

		handler.store.UpsertTag(tagSnapshot)
		handler.logger.Info(
			"tag upserted",
			"tag_id",
			tagSnapshot.ID,
			"device_id",
			tagSnapshot.DeviceID,
		)
		return nil
	case "delete":
		handler.store.RemoveTag(tagID)
		handler.logger.Info("tag removed", "tag_id", tagID)
		return nil
	default:
		handler.logger.Debug(
			"ignored tag operation",
			"tag_id",
			tagID,
			"operation",
			operation,
		)
		return nil
	}
}

func (handler *ConfigEventHandler) cascadedTagsCount(deviceID string) int {
	snapshot := handler.store.Get()
	if snapshot == nil {
		return 0
	}

	return len(snapshot.TagsByDevice[deviceID])
}
