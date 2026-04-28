package cache

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
)

func TestTagMetaCacheGetPutInvalidate(t *testing.T) {
	cache := NewInMemoryTagMetaCache(10, time.Minute)

	tagID := uuid.New()
	deviceID := uuid.New()
	meta := domain.TagMeta{
		TagID:      tagID,
		TagName:    "tag-1",
		DeviceID:   deviceID,
		DeviceName: "device-1",
	}

	if _, exists := cache.Get(tagID.String()); exists {
		t.Fatal("expected cache miss before put")
	}

	cache.Put(meta)

	cachedMeta, exists := cache.Get(tagID.String())
	if !exists {
		t.Fatal("expected cache hit after put")
	}
	if cachedMeta.TagName != "tag-1" {
		t.Fatalf("unexpected tag name %q", cachedMeta.TagName)
	}

	cache.Invalidate(tagID.String())

	if _, exists = cache.Get(tagID.String()); exists {
		t.Fatal("expected cache miss after invalidate")
	}
}

func TestTagMetaCacheInvalidateByDevice(t *testing.T) {
	cache := NewInMemoryTagMetaCache(10, time.Minute)

	deviceID := uuid.New()
	tagA := domain.TagMeta{TagID: uuid.New(), DeviceID: deviceID}
	tagB := domain.TagMeta{TagID: uuid.New(), DeviceID: deviceID}
	tagC := domain.TagMeta{TagID: uuid.New(), DeviceID: uuid.New()}

	cache.Put(tagA)
	cache.Put(tagB)
	cache.Put(tagC)

	cache.InvalidateByDevice(deviceID.String())

	if _, exists := cache.Get(tagA.TagID.String()); exists {
		t.Fatal("expected tag A to be invalidated")
	}
	if _, exists := cache.Get(tagB.TagID.String()); exists {
		t.Fatal("expected tag B to be invalidated")
	}
	if _, exists := cache.Get(tagC.TagID.String()); !exists {
		t.Fatal("expected tag C to stay in cache")
	}
}

func TestTagMetaCacheTTLAndLRU(t *testing.T) {
	cache := NewInMemoryTagMetaCache(2, time.Minute)

	now := time.Now().UTC()
	cache.nowFn = func() time.Time { return now }

	tagA := domain.TagMeta{TagID: uuid.New(), DeviceID: uuid.New()}
	tagB := domain.TagMeta{TagID: uuid.New(), DeviceID: uuid.New()}
	tagC := domain.TagMeta{TagID: uuid.New(), DeviceID: uuid.New()}

	cache.Put(tagA)
	cache.Put(tagB)

	if _, exists := cache.Get(tagA.TagID.String()); !exists {
		t.Fatal("expected tag A in cache")
	}

	cache.Put(tagC)

	if _, exists := cache.Get(tagB.TagID.String()); exists {
		t.Fatal("expected tag B evicted by LRU")
	}
	if _, exists := cache.Get(tagA.TagID.String()); !exists {
		t.Fatal("expected tag A preserved by recency")
	}

	now = now.Add(2 * time.Minute)
	if _, exists := cache.Get(tagA.TagID.String()); exists {
		t.Fatal("expected tag A expired by TTL")
	}
}
