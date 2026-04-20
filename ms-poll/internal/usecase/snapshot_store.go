package usecase

import (
	"log/slog"
	"sync"

	"github.com/EthernalFox/Controlitix/ms-poll/internal/domain"
)

type SnapshotStore struct {
	mutex       sync.RWMutex
	snapshot    *domain.Snapshot
	subscribers map[uintptr]chan SnapshotEvent
	logger      *slog.Logger
}

func NewSnapshotStore() *SnapshotStore {
	return &SnapshotStore{
		subscribers: make(map[uintptr]chan SnapshotEvent),
		logger:      defaultSnapshotStoreLogger(nil),
	}
}

func (store *SnapshotStore) Get() *domain.Snapshot {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	return store.snapshot
}

func (store *SnapshotStore) Replace(next *domain.Snapshot) {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	if next == nil {
		store.snapshot = nil
		store.publishEventLocked(SnapshotEvent{Kind: "replaced"})
		return
	}

	store.snapshot = cloneSnapshot(next)
	store.publishEventLocked(SnapshotEvent{Kind: "replaced"})
}

func (store *SnapshotStore) UpsertDevice(device *domain.DeviceSnapshot) {
	if device == nil || device.ID == "" {
		return
	}

	store.mutex.Lock()
	defer store.mutex.Unlock()

	nextSnapshot := cloneSnapshot(ensureSnapshot(store.snapshot))
	nextSnapshot.Devices[device.ID] = cloneDeviceSnapshot(device)

	store.snapshot = nextSnapshot
	store.publishEventLocked(SnapshotEvent{
		Kind:     "device_upserted",
		DeviceID: device.ID,
	})
}

func (store *SnapshotStore) RemoveDevice(deviceID string) {
	if deviceID == "" {
		return
	}

	store.mutex.Lock()
	defer store.mutex.Unlock()

	if store.snapshot == nil {
		return
	}

	nextSnapshot := cloneSnapshot(store.snapshot)
	delete(nextSnapshot.Devices, deviceID)

	for tagID, tagSnapshot := range nextSnapshot.Tags {
		if tagSnapshot.DeviceID == deviceID {
			delete(nextSnapshot.Tags, tagID)
		}
	}

	delete(nextSnapshot.TagsByDevice, deviceID)
	store.snapshot = nextSnapshot
	store.publishEventLocked(SnapshotEvent{
		Kind:     "device_removed",
		DeviceID: deviceID,
	})
}

func (store *SnapshotStore) UpsertTag(tag *domain.TagSnapshot) {
	if tag == nil || tag.ID == "" || tag.DeviceID == "" {
		return
	}

	store.mutex.Lock()
	defer store.mutex.Unlock()

	nextSnapshot := cloneSnapshot(ensureSnapshot(store.snapshot))
	if existingTag, exists := nextSnapshot.Tags[tag.ID]; exists {
		nextSnapshot.TagsByDevice[existingTag.DeviceID] = removeTagByID(
			nextSnapshot.TagsByDevice[existingTag.DeviceID],
			tag.ID,
		)
		if len(nextSnapshot.TagsByDevice[existingTag.DeviceID]) == 0 {
			delete(nextSnapshot.TagsByDevice, existingTag.DeviceID)
		}
	}

	tagSnapshot := cloneTagSnapshot(tag)
	nextSnapshot.Tags[tag.ID] = tagSnapshot
	nextSnapshot.TagsByDevice[tag.DeviceID] = removeTagByID(
		nextSnapshot.TagsByDevice[tag.DeviceID],
		tag.ID,
	)
	nextSnapshot.TagsByDevice[tag.DeviceID] = append(
		nextSnapshot.TagsByDevice[tag.DeviceID],
		tagSnapshot,
	)

	store.snapshot = nextSnapshot
	store.publishEventLocked(SnapshotEvent{
		Kind:     "tag_upserted",
		DeviceID: tag.DeviceID,
		TagID:    tag.ID,
	})
}

func (store *SnapshotStore) RemoveTag(tagID string) {
	if tagID == "" {
		return
	}

	store.mutex.Lock()
	defer store.mutex.Unlock()

	if store.snapshot == nil {
		return
	}

	nextSnapshot := cloneSnapshot(store.snapshot)
	tagSnapshot, exists := nextSnapshot.Tags[tagID]
	if !exists {
		return
	}

	delete(nextSnapshot.Tags, tagID)
	nextSnapshot.TagsByDevice[tagSnapshot.DeviceID] = removeTagByID(
		nextSnapshot.TagsByDevice[tagSnapshot.DeviceID],
		tagID,
	)
	if len(nextSnapshot.TagsByDevice[tagSnapshot.DeviceID]) == 0 {
		delete(nextSnapshot.TagsByDevice, tagSnapshot.DeviceID)
	}

	store.snapshot = nextSnapshot
	store.publishEventLocked(SnapshotEvent{
		Kind:     "tag_removed",
		DeviceID: tagSnapshot.DeviceID,
		TagID:    tagID,
	})
}

func ensureSnapshot(snapshot *domain.Snapshot) *domain.Snapshot {
	if snapshot != nil {
		return snapshot
	}

	return &domain.Snapshot{
		Devices:      make(map[string]*domain.DeviceSnapshot),
		Tags:         make(map[string]*domain.TagSnapshot),
		TagsByDevice: make(map[string][]*domain.TagSnapshot),
	}
}

func cloneSnapshot(snapshot *domain.Snapshot) *domain.Snapshot {
	if snapshot == nil {
		return &domain.Snapshot{
			Devices:      make(map[string]*domain.DeviceSnapshot),
			Tags:         make(map[string]*domain.TagSnapshot),
			TagsByDevice: make(map[string][]*domain.TagSnapshot),
		}
	}

	clonedSnapshot := &domain.Snapshot{
		Devices:      make(map[string]*domain.DeviceSnapshot, len(snapshot.Devices)),
		Tags:         make(map[string]*domain.TagSnapshot, len(snapshot.Tags)),
		TagsByDevice: make(map[string][]*domain.TagSnapshot, len(snapshot.TagsByDevice)),
		LoadedAt:     snapshot.LoadedAt,
	}

	for deviceID, deviceSnapshot := range snapshot.Devices {
		clonedSnapshot.Devices[deviceID] = cloneDeviceSnapshot(deviceSnapshot)
	}

	for tagID, tagSnapshot := range snapshot.Tags {
		clonedSnapshot.Tags[tagID] = cloneTagSnapshot(tagSnapshot)
	}

	for deviceID, tagSnapshots := range snapshot.TagsByDevice {
		clonedTagSnapshots := make([]*domain.TagSnapshot, 0, len(tagSnapshots))
		for _, tagSnapshot := range tagSnapshots {
			if tagSnapshot == nil {
				continue
			}

			if clonedTagSnapshot, exists := clonedSnapshot.Tags[tagSnapshot.ID]; exists {
				clonedTagSnapshots = append(clonedTagSnapshots, clonedTagSnapshot)
				continue
			}

			clonedTagSnapshot := cloneTagSnapshot(tagSnapshot)
			clonedSnapshot.Tags[clonedTagSnapshot.ID] = clonedTagSnapshot
			clonedTagSnapshots = append(clonedTagSnapshots, clonedTagSnapshot)
		}

		if len(clonedTagSnapshots) == 0 {
			continue
		}

		clonedSnapshot.TagsByDevice[deviceID] = clonedTagSnapshots
	}

	for _, tagSnapshot := range clonedSnapshot.Tags {
		clonedSnapshot.TagsByDevice[tagSnapshot.DeviceID] = upsertTagToSlice(
			clonedSnapshot.TagsByDevice[tagSnapshot.DeviceID],
			tagSnapshot,
		)
	}

	return clonedSnapshot
}

func cloneDeviceSnapshot(
	deviceSnapshot *domain.DeviceSnapshot,
) *domain.DeviceSnapshot {
	if deviceSnapshot == nil {
		return nil
	}

	var objectID *string
	if deviceSnapshot.ObjectID != nil {
		value := *deviceSnapshot.ObjectID
		objectID = &value
	}

	return &domain.DeviceSnapshot{
		ID:        deviceSnapshot.ID,
		ObjectID:  objectID,
		TypeID:    deviceSnapshot.TypeID,
		TypeName:  deviceSnapshot.TypeName,
		Name:      deviceSnapshot.Name,
		Settings:  cloneRawJSON(deviceSnapshot.Settings),
		UpdatedAt: deviceSnapshot.UpdatedAt,
	}
}

func cloneTagSnapshot(tagSnapshot *domain.TagSnapshot) *domain.TagSnapshot {
	if tagSnapshot == nil {
		return nil
	}

	var unitSymbol *string
	if tagSnapshot.UnitSymbol != nil {
		value := *tagSnapshot.UnitSymbol
		unitSymbol = &value
	}

	return &domain.TagSnapshot{
		ID:         tagSnapshot.ID,
		DeviceID:   tagSnapshot.DeviceID,
		Name:       tagSnapshot.Name,
		DataType:   tagSnapshot.DataType,
		UnitSymbol: unitSymbol,
		Address:    cloneRawJSON(tagSnapshot.Address),
		Scaling:    cloneTagScaling(tagSnapshot.Scaling),
		Setpoints:  cloneTagSetpoints(tagSnapshot.Setpoints),
		UpdatedAt:  tagSnapshot.UpdatedAt,
	}
}

func cloneTagScaling(tagScaling *domain.TagScaling) *domain.TagScaling {
	if tagScaling == nil {
		return nil
	}

	return &domain.TagScaling{
		RawMin: cloneFloatPointer(tagScaling.RawMin),
		RawMax: cloneFloatPointer(tagScaling.RawMax),
		EngMin: cloneFloatPointer(tagScaling.EngMin),
		EngMax: cloneFloatPointer(tagScaling.EngMax),
		Factor: cloneFloatPointer(tagScaling.Factor),
		Offset: cloneFloatPointer(tagScaling.Offset),
	}
}

func cloneTagSetpoints(
	tagSetpoints *domain.TagSetpoints,
) *domain.TagSetpoints {
	if tagSetpoints == nil {
		return nil
	}

	return &domain.TagSetpoints{
		LoLo: cloneFloatPointer(tagSetpoints.LoLo),
		Lo:   cloneFloatPointer(tagSetpoints.Lo),
		Hi:   cloneFloatPointer(tagSetpoints.Hi),
		HiHi: cloneFloatPointer(tagSetpoints.HiHi),
	}
}

func cloneFloatPointer(value *float64) *float64 {
	if value == nil {
		return nil
	}

	copiedValue := *value
	return &copiedValue
}

func cloneRawJSON(source []byte) []byte {
	if len(source) == 0 {
		return nil
	}

	return append([]byte(nil), source...)
}

func removeTagByID(
	tagSnapshots []*domain.TagSnapshot,
	tagID string,
) []*domain.TagSnapshot {
	if len(tagSnapshots) == 0 {
		return nil
	}

	filtered := make([]*domain.TagSnapshot, 0, len(tagSnapshots))
	for _, tagSnapshot := range tagSnapshots {
		if tagSnapshot == nil || tagSnapshot.ID == tagID {
			continue
		}

		filtered = append(filtered, tagSnapshot)
	}

	return filtered
}

func upsertTagToSlice(
	tagSnapshots []*domain.TagSnapshot,
	tagSnapshot *domain.TagSnapshot,
) []*domain.TagSnapshot {
	filtered := removeTagByID(tagSnapshots, tagSnapshot.ID)
	return append(filtered, tagSnapshot)
}
