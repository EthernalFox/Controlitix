package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/EthernalFox/Controlitix/ms-poll/internal/domain"
)

type mockConfigEventRepository struct {
	loadDevice func(
		ctx context.Context,
		id string,
	) (*domain.DeviceSnapshot, error)
	loadTag func(
		ctx context.Context,
		id string,
	) (*domain.TagSnapshot, error)
}

func (repository *mockConfigEventRepository) LoadDevice(
	ctx context.Context,
	id string,
) (*domain.DeviceSnapshot, error) {
	if repository.loadDevice == nil {
		return nil, nil
	}

	return repository.loadDevice(ctx, id)
}

func (repository *mockConfigEventRepository) LoadTag(
	ctx context.Context,
	id string,
) (*domain.TagSnapshot, error) {
	if repository.loadTag == nil {
		return nil, nil
	}

	return repository.loadTag(ctx, id)
}

func TestConfigEventHandlerDeviceCreateUpsertsDevice(t *testing.T) {
	store := NewSnapshotStore()
	store.Replace(buildSnapshot(nil, nil))

	mockRepository := &mockConfigEventRepository{
		loadDevice: func(
			_ context.Context,
			id string,
		) (*domain.DeviceSnapshot, error) {
			return &domain.DeviceSnapshot{
				ID:        id,
				TypeID:    1,
				TypeName:  "modbus_tcp",
				Name:      "Device from event",
				UpdatedAt: time.Now().UTC(),
			}, nil
		},
	}

	handler := newConfigEventHandler(mockRepository, store, nil)
	handleError := handler.Handle(
		context.Background(),
		domain.ConfigChangedEvent{
			EntityType: "device",
			EntityID:   "device-1",
			Operation:  "create",
		},
	)
	if handleError != nil {
		t.Fatalf("unexpected handler error: %v", handleError)
	}

	currentSnapshot := store.Get()
	if _, exists := currentSnapshot.Devices["device-1"]; !exists {
		t.Fatal("expected device to be present after device.create event")
	}
}

func TestConfigEventHandlerDeviceDeleteRemovesDeviceAndTags(t *testing.T) {
	store := NewSnapshotStore()
	store.Replace(buildSnapshot(
		[]*domain.DeviceSnapshot{
			{
				ID:        "device-1",
				TypeID:    1,
				TypeName:  "modbus_tcp",
				Name:      "Device 1",
				UpdatedAt: time.Now().UTC(),
			},
		},
		[]*domain.TagSnapshot{
			{
				ID:        "tag-1",
				DeviceID:  "device-1",
				Name:      "Tag 1",
				DataType:  "float64",
				UpdatedAt: time.Now().UTC(),
			},
			{
				ID:        "tag-2",
				DeviceID:  "device-1",
				Name:      "Tag 2",
				DataType:  "float64",
				UpdatedAt: time.Now().UTC(),
			},
		},
	))

	handler := newConfigEventHandler(&mockConfigEventRepository{}, store, nil)
	handleError := handler.Handle(
		context.Background(),
		domain.ConfigChangedEvent{
			EntityType: "device",
			EntityID:   "device-1",
			Operation:  "delete",
		},
	)
	if handleError != nil {
		t.Fatalf("unexpected handler error: %v", handleError)
	}

	currentSnapshot := store.Get()
	if _, exists := currentSnapshot.Devices["device-1"]; exists {
		t.Fatal("expected device to be removed after device.delete event")
	}

	if _, exists := currentSnapshot.Tags["tag-1"]; exists {
		t.Fatal("expected tag-1 to be removed after device.delete event")
	}

	if _, exists := currentSnapshot.Tags["tag-2"]; exists {
		t.Fatal("expected tag-2 to be removed after device.delete event")
	}
}

func TestConfigEventHandlerTagUpdateWithMissingTagRemovesTag(t *testing.T) {
	store := NewSnapshotStore()
	store.Replace(buildSnapshot(
		[]*domain.DeviceSnapshot{
			{
				ID:        "device-1",
				TypeID:    1,
				TypeName:  "modbus_tcp",
				Name:      "Device 1",
				UpdatedAt: time.Now().UTC(),
			},
		},
		[]*domain.TagSnapshot{
			{
				ID:        "tag-1",
				DeviceID:  "device-1",
				Name:      "Tag 1",
				DataType:  "float64",
				UpdatedAt: time.Now().UTC(),
			},
		},
	))

	mockRepository := &mockConfigEventRepository{
		loadTag: func(_ context.Context, _ string) (*domain.TagSnapshot, error) {
			return nil, nil
		},
	}

	handler := newConfigEventHandler(mockRepository, store, nil)
	handleError := handler.Handle(
		context.Background(),
		domain.ConfigChangedEvent{
			EntityType: "tag",
			EntityID:   "tag-1",
			Operation:  "update",
		},
	)
	if handleError != nil {
		t.Fatalf("unexpected handler error: %v", handleError)
	}

	currentSnapshot := store.Get()
	if _, exists := currentSnapshot.Tags["tag-1"]; exists {
		t.Fatal("expected tag to be removed when LoadTag returns nil")
	}
}

func TestConfigEventHandlerUnknownEntityTypeIsNoOp(t *testing.T) {
	store := NewSnapshotStore()
	store.Replace(buildSnapshot(
		[]*domain.DeviceSnapshot{
			{
				ID:        "device-1",
				TypeID:    1,
				TypeName:  "modbus_tcp",
				Name:      "Device 1",
				UpdatedAt: time.Now().UTC(),
			},
		},
		nil,
	))

	handler := newConfigEventHandler(&mockConfigEventRepository{}, store, nil)
	handleError := handler.Handle(
		context.Background(),
		domain.ConfigChangedEvent{
			EntityType: "monitoring_object",
			EntityID:   "object-1",
			Operation:  "update",
		},
	)
	if handleError != nil {
		t.Fatalf("unexpected handler error: %v", handleError)
	}

	currentSnapshot := store.Get()
	if _, exists := currentSnapshot.Devices["device-1"]; !exists {
		t.Fatal("expected snapshot to remain unchanged for unknown entity type")
	}
}
