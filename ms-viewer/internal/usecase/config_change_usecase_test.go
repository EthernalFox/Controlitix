package usecase

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
)

type fakeTagMetaInvalidator struct {
	invalidatedTags    []string
	invalidatedDevices []string
	invalidatedAll     int
}

func (invalidator *fakeTagMetaInvalidator) Invalidate(tagID string) {
	invalidator.invalidatedTags = append(invalidator.invalidatedTags, tagID)
}

func (invalidator *fakeTagMetaInvalidator) InvalidateByDevice(deviceID string) {
	invalidator.invalidatedDevices = append(invalidator.invalidatedDevices, deviceID)
}

func (invalidator *fakeTagMetaInvalidator) InvalidateAll() {
	invalidator.invalidatedAll++
}

type fakeConfigChangeRealtime struct {
	broadcastedEvents []domain.ConfigChangedEvent
	unsubscribedTags  []uuid.UUID
	handledEvents     []domain.ConfigChangedEvent
}

func (realtime *fakeConfigChangeRealtime) HandleConfigChanged(event domain.ConfigChangedEvent) {
	realtime.handledEvents = append(realtime.handledEvents, event)
}

func (realtime *fakeConfigChangeRealtime) BroadcastConfigChanged(event domain.ConfigChangedEvent) {
	realtime.broadcastedEvents = append(realtime.broadcastedEvents, event)
}

func (realtime *fakeConfigChangeRealtime) UnsubscribeTagFromAll(tagID uuid.UUID) {
	realtime.unsubscribedTags = append(realtime.unsubscribedTags, tagID)
}

func TestConfigChangeUseCaseTagDeleted(t *testing.T) {
	invalidator := &fakeTagMetaInvalidator{}
	realtime := &fakeConfigChangeRealtime{}
	useCase := NewConfigChangeUseCase(invalidator, realtime, nil, nil)

	tagID := uuid.New()
	handleError := useCase.HandleConfigChanged(context.Background(), domain.ConfigChangedEvent{
		EntityType: "tag",
		EntityID:   tagID.String(),
		Operation:  "deleted",
		Timestamp:  time.Now().UTC(),
	})
	if handleError != nil {
		t.Fatalf("HandleConfigChanged returned error: %v", handleError)
	}

	if len(invalidator.invalidatedTags) != 1 || invalidator.invalidatedTags[0] != tagID.String() {
		t.Fatalf("unexpected invalidated tags: %v", invalidator.invalidatedTags)
	}
	if len(realtime.unsubscribedTags) != 1 || realtime.unsubscribedTags[0] != tagID {
		t.Fatalf("unexpected unsubscribed tags: %v", realtime.unsubscribedTags)
	}
	if len(realtime.broadcastedEvents) != 1 {
		t.Fatalf("expected one broadcast event, got %d", len(realtime.broadcastedEvents))
	}
	if len(realtime.handledEvents) != 1 {
		t.Fatalf("expected one handled event, got %d", len(realtime.handledEvents))
	}
}

func TestConfigChangeUseCaseDevicePayloadInvalidatesTags(t *testing.T) {
	invalidator := &fakeTagMetaInvalidator{}
	realtime := &fakeConfigChangeRealtime{}
	useCase := NewConfigChangeUseCase(invalidator, realtime, nil, nil)

	firstTag := uuid.NewString()
	secondTag := uuid.NewString()
	payload, _ := json.Marshal(map[string]any{
		"tag_ids": []string{firstTag, secondTag},
	})

	handleError := useCase.HandleConfigChanged(context.Background(), domain.ConfigChangedEvent{
		EntityType: "device",
		EntityID:   uuid.NewString(),
		Operation:  "updated",
		Timestamp:  time.Now().UTC(),
		Payload:    payload,
	})
	if handleError != nil {
		t.Fatalf("HandleConfigChanged returned error: %v", handleError)
	}

	if len(invalidator.invalidatedTags) != 2 {
		t.Fatalf("expected two invalidated tags, got %d", len(invalidator.invalidatedTags))
	}
	if len(invalidator.invalidatedDevices) != 0 {
		t.Fatalf("expected no device invalidation fallback, got %v", invalidator.invalidatedDevices)
	}
	if len(realtime.handledEvents) != 1 {
		t.Fatalf("expected one handled event, got %d", len(realtime.handledEvents))
	}
}

func TestConfigChangeUseCaseUnknownEntityType(t *testing.T) {
	invalidator := &fakeTagMetaInvalidator{}
	realtime := &fakeConfigChangeRealtime{}
	useCase := NewConfigChangeUseCase(invalidator, realtime, nil, nil)

	handleError := useCase.HandleConfigChanged(context.Background(), domain.ConfigChangedEvent{
		EntityType: "unknown",
		EntityID:   uuid.NewString(),
		Operation:  "updated",
		Timestamp:  time.Now().UTC(),
	})
	if handleError != nil {
		t.Fatalf("HandleConfigChanged returned error: %v", handleError)
	}
	if len(realtime.broadcastedEvents) != 0 {
		t.Fatalf("expected no broadcast for unknown type, got %d", len(realtime.broadcastedEvents))
	}
}
