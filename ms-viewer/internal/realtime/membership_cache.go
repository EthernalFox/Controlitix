package realtime

import (
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type membershipCacheEntry struct {
	tagIDs    []uuid.UUID
	expiresAt time.Time
}

type MembershipCache struct {
	mutex        sync.RWMutex
	ttl          time.Duration
	byKey        map[string]membershipCacheEntry
	entitiesByTag map[uuid.UUID]map[string]struct{}
}

func NewMembershipCache(ttl time.Duration) *MembershipCache {
	if ttl <= 0 {
		ttl = 600 * time.Second
	}

	return &MembershipCache{
		ttl:           ttl,
		byKey:         make(map[string]membershipCacheEntry),
		entitiesByTag: make(map[uuid.UUID]map[string]struct{}),
	}
}

func (cache *MembershipCache) Get(topic string) ([]uuid.UUID, bool) {
	if cache == nil {
		return nil, false
	}

	topic = normalizeTopicKey(topic)
	if topic == "" {
		return nil, false
	}

	cache.mutex.RLock()
	entry, exists := cache.byKey[topic]
	cache.mutex.RUnlock()
	if !exists {
		return nil, false
	}
	if time.Now().UTC().After(entry.expiresAt) {
		cache.mutex.Lock()
		cache.deleteLocked(topic)
		cache.mutex.Unlock()
		return nil, false
	}

	copied := append([]uuid.UUID(nil), entry.tagIDs...)
	return copied, true
}

func (cache *MembershipCache) Set(topic string, tagIDs []uuid.UUID) {
	if cache == nil {
		return
	}

	topic = normalizeTopicKey(topic)
	if topic == "" {
		return
	}

	normalizedTagIDs := uniqueUUIDs(tagIDs)

	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	cache.deleteLocked(topic)
	cache.byKey[topic] = membershipCacheEntry{
		tagIDs:    normalizedTagIDs,
		expiresAt: time.Now().UTC().Add(cache.ttl),
	}
	for _, tagID := range normalizedTagIDs {
		references := cache.entitiesByTag[tagID]
		if references == nil {
			references = make(map[string]struct{})
			cache.entitiesByTag[tagID] = references
		}
		references[topic] = struct{}{}
	}
}

func (cache *MembershipCache) Invalidate(entityType string, entityID string) []string {
	if cache == nil {
		return nil
	}

	topic := buildGroupTopicKey(entityType, entityID)
	if topic == "" {
		return nil
	}

	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	if _, exists := cache.byKey[topic]; !exists {
		return nil
	}
	cache.deleteLocked(topic)
	return []string{topic}
}

func (cache *MembershipCache) InvalidateByTag(tagID string) []string {
	if cache == nil {
		return nil
	}

	parsedTagID, parseError := uuid.Parse(strings.TrimSpace(tagID))
	if parseError != nil {
		return nil
	}

	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	references := cache.entitiesByTag[parsedTagID]
	if len(references) == 0 {
		return nil
	}

	keys := make([]string, 0, len(references))
	for key := range references {
		cache.deleteLocked(key)
		keys = append(keys, key)
	}

	return uniqueStrings(keys)
}

func (cache *MembershipCache) deleteLocked(topic string) {
	entry, exists := cache.byKey[topic]
	if !exists {
		return
	}

	delete(cache.byKey, topic)
	for _, tagID := range entry.tagIDs {
		references := cache.entitiesByTag[tagID]
		if references == nil {
			continue
		}
		delete(references, topic)
		if len(references) == 0 {
			delete(cache.entitiesByTag, tagID)
		}
	}
}

func normalizeTopicKey(topic string) string {
	topic = strings.TrimSpace(strings.ToLower(topic))
	if topic == "" {
		return ""
	}
	return topic
}

func buildGroupTopicKey(entityType string, entityID string) string {
	entityType = strings.TrimSpace(strings.ToLower(entityType))
	entityID = strings.TrimSpace(strings.ToLower(entityID))
	if entityID == "" {
		return ""
	}

	switch entityType {
	case TopicPrefixObject:
		return TopicPrefixObject + ":" + entityID
	case TopicPrefixDiagram:
		return TopicPrefixDiagram + ":" + entityID
	default:
		return ""
	}
}

func uniqueUUIDs(values []uuid.UUID) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{}, len(values))
	result := make([]uuid.UUID, 0, len(values))
	for _, value := range values {
		if value == uuid.Nil {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		normalized := strings.TrimSpace(strings.ToLower(value))
		if normalized == "" {
			continue
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	return result
}
