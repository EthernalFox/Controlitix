package usecase

import (
	"testing"
	"time"

	"github.com/EthernalFox/Controlitix/ms-poll/internal/domain"
)

func TestSnapshotStoreReplaceReplacesSnapshot(t *testing.T) {
	store := NewSnapshotStore()

	firstSnapshot := buildSnapshot(
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
	)
	store.Replace(firstSnapshot)

	secondSnapshot := buildSnapshot(
		[]*domain.DeviceSnapshot{
			{
				ID:        "device-2",
				TypeID:    2,
				TypeName:  "snmp_v2c",
				Name:      "Device 2",
				UpdatedAt: time.Now().UTC(),
			},
		},
		nil,
	)
	store.Replace(secondSnapshot)
	secondSnapshot.Devices["device-2"].Name = "Mutated outside store"

	currentSnapshot := store.Get()
	if currentSnapshot == nil {
		t.Fatal("snapshot is nil")
	}

	if _, exists := currentSnapshot.Devices["device-1"]; exists {
		t.Fatal("expected old device to be removed after replace")
	}

	deviceSnapshot, exists := currentSnapshot.Devices["device-2"]
	if !exists {
		t.Fatal("expected new device to be present after replace")
	}

	if deviceSnapshot.Name != "Device 2" {
		t.Fatalf("unexpected device name after replace: %s", deviceSnapshot.Name)
	}
}

func TestSnapshotStoreUpsertDeviceCreatesAndUpdates(t *testing.T) {
	store := NewSnapshotStore()
	store.Replace(buildSnapshot(nil, nil))

	store.UpsertDevice(&domain.DeviceSnapshot{
		ID:        "device-1",
		TypeID:    1,
		TypeName:  "modbus_tcp",
		Name:      "Initial device name",
		UpdatedAt: time.Now().UTC(),
	})

	store.UpsertDevice(&domain.DeviceSnapshot{
		ID:        "device-1",
		TypeID:    1,
		TypeName:  "modbus_tcp",
		Name:      "Updated device name",
		UpdatedAt: time.Now().UTC(),
	})

	currentSnapshot := store.Get()
	deviceSnapshot, exists := currentSnapshot.Devices["device-1"]
	if !exists {
		t.Fatal("expected device to be present after upsert")
	}

	if deviceSnapshot.Name != "Updated device name" {
		t.Fatalf("unexpected upserted device name: %s", deviceSnapshot.Name)
	}
}

func TestSnapshotStoreRemoveDeviceRemovesDeviceAndTags(t *testing.T) {
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
			{
				ID:        "device-2",
				TypeID:    2,
				TypeName:  "snmp_v2c",
				Name:      "Device 2",
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
			{
				ID:        "tag-3",
				DeviceID:  "device-2",
				Name:      "Tag 3",
				DataType:  "float64",
				UpdatedAt: time.Now().UTC(),
			},
		},
	))

	store.RemoveDevice("device-1")

	currentSnapshot := store.Get()
	if _, exists := currentSnapshot.Devices["device-1"]; exists {
		t.Fatal("expected removed device to be absent")
	}

	if _, exists := currentSnapshot.Tags["tag-1"]; exists {
		t.Fatal("expected tag-1 to be removed together with device")
	}

	if _, exists := currentSnapshot.Tags["tag-2"]; exists {
		t.Fatal("expected tag-2 to be removed together with device")
	}

	if _, exists := currentSnapshot.Tags["tag-3"]; !exists {
		t.Fatal("expected unrelated tag to remain")
	}

	if _, exists := currentSnapshot.TagsByDevice["device-1"]; exists {
		t.Fatal("expected tagsByDevice entry for removed device to be absent")
	}
}

func TestSnapshotStoreUpsertTagLinksTagWithDevice(t *testing.T) {
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

	store.UpsertTag(&domain.TagSnapshot{
		ID:        "tag-1",
		DeviceID:  "device-1",
		Name:      "Tag 1",
		DataType:  "float64",
		UpdatedAt: time.Now().UTC(),
	})

	store.UpsertTag(&domain.TagSnapshot{
		ID:        "tag-1",
		DeviceID:  "device-1",
		Name:      "Tag 1 updated",
		DataType:  "float64",
		UpdatedAt: time.Now().UTC(),
	})

	currentSnapshot := store.Get()
	tagSnapshot, exists := currentSnapshot.Tags["tag-1"]
	if !exists {
		t.Fatal("expected tag to be upserted")
	}

	if tagSnapshot.Name != "Tag 1 updated" {
		t.Fatalf("unexpected tag name after upsert: %s", tagSnapshot.Name)
	}

	deviceTags := currentSnapshot.TagsByDevice["device-1"]
	if len(deviceTags) != 1 {
		t.Fatalf("expected one tag in tagsByDevice, got %d", len(deviceTags))
	}

	if deviceTags[0].ID != "tag-1" {
		t.Fatalf("unexpected tag id in tagsByDevice: %s", deviceTags[0].ID)
	}
}

func TestSnapshotStoreRemoveTagRemovesTagAndUpdatesTagsByDevice(t *testing.T) {
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

	store.RemoveTag("tag-1")

	currentSnapshot := store.Get()
	if _, exists := currentSnapshot.Tags["tag-1"]; exists {
		t.Fatal("expected removed tag to be absent")
	}

	deviceTags := currentSnapshot.TagsByDevice["device-1"]
	if len(deviceTags) != 1 {
		t.Fatalf("expected one tag after removing tag-1, got %d", len(deviceTags))
	}

	if deviceTags[0].ID != "tag-2" {
		t.Fatalf("unexpected remaining tag id: %s", deviceTags[0].ID)
	}

	store.RemoveTag("tag-2")
	currentSnapshot = store.Get()

	if _, exists := currentSnapshot.TagsByDevice["device-1"]; exists {
		t.Fatal("expected tagsByDevice entry to be removed when no tags remain")
	}
}

func buildSnapshot(
	devices []*domain.DeviceSnapshot,
	tags []*domain.TagSnapshot,
) *domain.Snapshot {
	snapshot := &domain.Snapshot{
		Devices:      make(map[string]*domain.DeviceSnapshot),
		Tags:         make(map[string]*domain.TagSnapshot),
		TagsByDevice: make(map[string][]*domain.TagSnapshot),
		LoadedAt:     time.Now().UTC(),
	}

	for _, deviceSnapshot := range devices {
		snapshot.Devices[deviceSnapshot.ID] = deviceSnapshot
	}

	for _, tagSnapshot := range tags {
		snapshot.Tags[tagSnapshot.ID] = tagSnapshot
		snapshot.TagsByDevice[tagSnapshot.DeviceID] = append(
			snapshot.TagsByDevice[tagSnapshot.DeviceID],
			tagSnapshot,
		)
	}

	return snapshot
}
