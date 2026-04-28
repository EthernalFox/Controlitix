package cache

import (
	"container/list"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
)

type CacheStats struct {
	Hits   uint64
	Misses uint64
	Size   int
}

type TagMetaCache interface {
	Get(tagID string) (domain.TagMeta, bool)
	Put(meta domain.TagMeta)
	Invalidate(tagID string)
	InvalidateByDevice(deviceID string)
	InvalidateAll()
	Stats() CacheStats
}

type tagMetaEntry struct {
	tagID      string
	deviceID   string
	meta       domain.TagMeta
	loadedAt   time.Time
	lruElement *list.Element
}

type InMemoryTagMetaCache struct {
	mutex sync.Mutex

	maxSize int
	ttl     time.Duration

	entriesByTag    map[string]*tagMetaEntry
	tagsByDevice    map[string]map[string]struct{}
	lru             *list.List
	hitsCounter     atomic.Uint64
	missesCounter   atomic.Uint64
	nowFn           func() time.Time
}

func NewInMemoryTagMetaCache(maxSize int, ttl time.Duration) *InMemoryTagMetaCache {
	if maxSize <= 0 {
		maxSize = 10000
	}
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}

	return &InMemoryTagMetaCache{
		maxSize:      maxSize,
		ttl:          ttl,
		entriesByTag: make(map[string]*tagMetaEntry, maxSize),
		tagsByDevice: make(map[string]map[string]struct{}),
		lru:          list.New(),
		nowFn: func() time.Time {
			return time.Now().UTC()
		},
	}
}

func (cache *InMemoryTagMetaCache) Get(tagID string) (domain.TagMeta, bool) {
	normalizedTagID := strings.TrimSpace(tagID)
	if normalizedTagID == "" {
		cache.missesCounter.Add(1)
		return domain.TagMeta{}, false
	}

	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	entry, exists := cache.entriesByTag[normalizedTagID]
	if !exists {
		cache.missesCounter.Add(1)
		return domain.TagMeta{}, false
	}

	now := cache.nowFn()
	if now.Sub(entry.loadedAt) > cache.ttl {
		cache.removeEntryLocked(entry)
		cache.missesCounter.Add(1)
		return domain.TagMeta{}, false
	}

	cache.lru.MoveToFront(entry.lruElement)
	cache.hitsCounter.Add(1)
	return entry.meta, true
}

func (cache *InMemoryTagMetaCache) Put(meta domain.TagMeta) {
	normalizedTagID := strings.TrimSpace(meta.TagID.String())
	if normalizedTagID == "" {
		return
	}

	normalizedDeviceID := strings.TrimSpace(meta.DeviceID.String())
	now := cache.nowFn()

	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	if existing, exists := cache.entriesByTag[normalizedTagID]; exists {
		cache.unlinkTagFromDeviceLocked(existing.deviceID, existing.tagID)
		existing.meta = meta
		existing.deviceID = normalizedDeviceID
		existing.loadedAt = now
		cache.linkTagToDeviceLocked(normalizedDeviceID, normalizedTagID)
		cache.lru.MoveToFront(existing.lruElement)
		return
	}

	entry := &tagMetaEntry{
		tagID:    normalizedTagID,
		deviceID: normalizedDeviceID,
		meta:     meta,
		loadedAt: now,
	}
	entry.lruElement = cache.lru.PushFront(entry)
	cache.entriesByTag[normalizedTagID] = entry
	cache.linkTagToDeviceLocked(normalizedDeviceID, normalizedTagID)

	cache.evictOverflowLocked()
}

func (cache *InMemoryTagMetaCache) Invalidate(tagID string) {
	normalizedTagID := strings.TrimSpace(tagID)
	if normalizedTagID == "" {
		return
	}

	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	entry, exists := cache.entriesByTag[normalizedTagID]
	if !exists {
		return
	}

	cache.removeEntryLocked(entry)
}

func (cache *InMemoryTagMetaCache) InvalidateByDevice(deviceID string) {
	normalizedDeviceID := strings.TrimSpace(deviceID)
	if normalizedDeviceID == "" {
		return
	}

	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	tags := cache.tagsByDevice[normalizedDeviceID]
	if len(tags) == 0 {
		return
	}

	for tagID := range tags {
		entry, exists := cache.entriesByTag[tagID]
		if !exists {
			continue
		}
		cache.removeEntryLocked(entry)
	}
}

func (cache *InMemoryTagMetaCache) InvalidateAll() {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	cache.entriesByTag = make(map[string]*tagMetaEntry, cache.maxSize)
	cache.tagsByDevice = make(map[string]map[string]struct{})
	cache.lru.Init()
}

func (cache *InMemoryTagMetaCache) Stats() CacheStats {
	cache.mutex.Lock()
	size := len(cache.entriesByTag)
	cache.mutex.Unlock()

	return CacheStats{
		Hits:   cache.hitsCounter.Load(),
		Misses: cache.missesCounter.Load(),
		Size:   size,
	}
}

func (cache *InMemoryTagMetaCache) evictOverflowLocked() {
	for len(cache.entriesByTag) > cache.maxSize {
		lastElement := cache.lru.Back()
		if lastElement == nil {
			return
		}

		entry, ok := lastElement.Value.(*tagMetaEntry)
		if !ok || entry == nil {
			cache.lru.Remove(lastElement)
			continue
		}

		cache.removeEntryLocked(entry)
	}
}

func (cache *InMemoryTagMetaCache) removeEntryLocked(entry *tagMetaEntry) {
	if entry == nil {
		return
	}

	delete(cache.entriesByTag, entry.tagID)
	cache.unlinkTagFromDeviceLocked(entry.deviceID, entry.tagID)
	if entry.lruElement != nil {
		cache.lru.Remove(entry.lruElement)
	}
}

func (cache *InMemoryTagMetaCache) linkTagToDeviceLocked(deviceID string, tagID string) {
	if deviceID == "" || tagID == "" {
		return
	}

	deviceTags := cache.tagsByDevice[deviceID]
	if deviceTags == nil {
		deviceTags = make(map[string]struct{})
		cache.tagsByDevice[deviceID] = deviceTags
	}
	deviceTags[tagID] = struct{}{}
}

func (cache *InMemoryTagMetaCache) unlinkTagFromDeviceLocked(deviceID string, tagID string) {
	if deviceID == "" || tagID == "" {
		return
	}

	deviceTags := cache.tagsByDevice[deviceID]
	if len(deviceTags) == 0 {
		return
	}

	delete(deviceTags, tagID)
	if len(deviceTags) == 0 {
		delete(cache.tagsByDevice, deviceID)
	}
}
