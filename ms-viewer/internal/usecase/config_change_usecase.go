package usecase

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/google/uuid"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
)

type TagMetaCacheInvalidator interface {
	Invalidate(tagID string)
	InvalidateByDevice(deviceID string)
	InvalidateAll()
}

type ConfigChangeRealtime interface {
	HandleConfigChanged(event domain.ConfigChangedEvent)
	BroadcastConfigChanged(event domain.ConfigChangedEvent)
	UnsubscribeTagFromAll(tagID uuid.UUID)
}

type AlarmConfigSync interface {
	HandleTagConfigChanged(ctx context.Context, tagID string, operation string)
}

type ConfigChangeUseCase struct {
	tagMetaCache TagMetaCacheInvalidator
	realtimeHub  ConfigChangeRealtime
	alarmSync    AlarmConfigSync
	logger       *slog.Logger
}

func NewConfigChangeUseCase(
	tagMetaCache TagMetaCacheInvalidator,
	realtimeHub ConfigChangeRealtime,
	alarmSync AlarmConfigSync,
	logger *slog.Logger,
) *ConfigChangeUseCase {
	if logger == nil {
		logger = slog.Default()
	}

	return &ConfigChangeUseCase{
		tagMetaCache: tagMetaCache,
		realtimeHub:  realtimeHub,
		alarmSync:    alarmSync,
		logger:       logger,
	}
}

func (useCase *ConfigChangeUseCase) HandleConfigChanged(
	ctx context.Context,
	event domain.ConfigChangedEvent,
) error {
	entityType := strings.ToLower(strings.TrimSpace(event.EntityType))
	entityID := strings.TrimSpace(event.EntityID)
	operation := strings.ToLower(strings.TrimSpace(event.Operation))

	switch entityType {
	case "tag":
		if useCase.tagMetaCache != nil && entityID != "" {
			useCase.tagMetaCache.Invalidate(entityID)
		}
		if useCase.alarmSync != nil && entityID != "" {
			useCase.alarmSync.HandleTagConfigChanged(ctx, entityID, operation)
		}
		if operation == "deleted" && useCase.realtimeHub != nil {
			tagID := event.EntityUUID()
			if tagID != uuid.Nil {
				useCase.realtimeHub.UnsubscribeTagFromAll(tagID)
			}
		}
	case "device":
		if useCase.tagMetaCache != nil {
			invalidatedByTags := useCase.invalidateByTagIDsFromPayload(event.Payload)
			if !invalidatedByTags {
				if entityID != "" {
					useCase.tagMetaCache.InvalidateByDevice(entityID)
				} else {
					useCase.tagMetaCache.InvalidateAll()
				}
			}
		}
	case "object":
		if useCase.tagMetaCache != nil {
			useCase.tagMetaCache.InvalidateAll()
		}
	case "diagram", "figure":
		// No tag meta cache action is required for diagram/figure updates.
	default:
		useCase.logger.Warn(
			"unknown config.changed entity type",
			"method",
			"ConfigChangeUseCase.HandleConfigChanged",
			"entity_type",
			event.EntityType,
			"entity_id",
			event.EntityID,
		)
		return nil
	}

	if useCase.realtimeHub != nil {
		useCase.realtimeHub.HandleConfigChanged(event)
		useCase.realtimeHub.BroadcastConfigChanged(event)
	}

	return nil
}

func (useCase *ConfigChangeUseCase) invalidateByTagIDsFromPayload(
	payload json.RawMessage,
) bool {
	if useCase.tagMetaCache == nil || len(payload) == 0 {
		return false
	}

	var raw map[string]any
	if unmarshalError := json.Unmarshal(payload, &raw); unmarshalError != nil {
		useCase.logger.Warn(
			"failed to decode config.changed payload",
			"method",
			"ConfigChangeUseCase.invalidateByTagIDsFromPayload",
			"error",
			unmarshalError,
		)
		return false
	}

	tagIDs := extractTagIDs(raw)
	if len(tagIDs) == 0 {
		return false
	}

	for _, tagID := range tagIDs {
		useCase.tagMetaCache.Invalidate(tagID)
	}

	return true
}

func extractTagIDs(raw map[string]any) []string {
	tagIDs := make([]string, 0)
	appendTagID := func(candidate string) {
		normalized := strings.TrimSpace(candidate)
		if normalized == "" {
			return
		}
		tagIDs = append(tagIDs, normalized)
	}

	if value, exists := raw["tag_id"]; exists {
		if tagID, ok := value.(string); ok {
			appendTagID(tagID)
		}
	}

	if value, exists := raw["tag_ids"]; exists {
		if list, ok := value.([]any); ok {
			for _, entry := range list {
				tagID, ok := entry.(string)
				if ok {
					appendTagID(tagID)
				}
			}
		}
	}

	if value, exists := raw["tags"]; exists {
		list, ok := value.([]any)
		if ok {
			for _, entry := range list {
				entryObject, ok := entry.(map[string]any)
				if !ok {
					continue
				}
				if tagID, ok := entryObject["id"].(string); ok {
					appendTagID(tagID)
				}
			}
		}
	}

	return uniqueStrings(tagIDs)
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	uniqueValues := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		uniqueValues = append(uniqueValues, value)
	}

	return uniqueValues
}
