package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	redis "github.com/redis/go-redis/v9"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
)

type RedisCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisCache(redisURL string, ttl time.Duration) (*RedisCache, error) {
	options, parseError := redis.ParseURL(redisURL)
	if parseError != nil {
		return nil, fmt.Errorf("parse redis url: %w", parseError)
	}

	client := redis.NewClient(options)
	return &RedisCache{
		client: client,
		ttl:    ttl,
	}, nil
}

func (cache *RedisCache) Ping(ctx context.Context) error {
	if pingError := cache.client.Ping(ctx).Err(); pingError != nil {
		return fmt.Errorf("ping redis: %w", pingError)
	}

	return nil
}

func (cache *RedisCache) SetLastValue(
	ctx context.Context,
	record domain.IngestRecord,
) error {
	payload := map[string]any{
		"ts": record.Timestamp.UTC().Format(time.RFC3339Nano),
		"v":  record.Value,
		"q":  record.Quality,
	}

	encodedPayload, marshalError := json.Marshal(payload)
	if marshalError != nil {
		return fmt.Errorf("marshal cache payload: %w", marshalError)
	}

	key := "ms-viewer:last:" + record.TagID.String()
	if setError := cache.client.Set(ctx, key, encodedPayload, cache.ttl).Err(); setError != nil {
		return fmt.Errorf("set redis key %s: %w", key, setError)
	}

	return nil
}

func (cache *RedisCache) GetLastValue(
	ctx context.Context,
	tagID uuid.UUID,
) (domain.IngestRecord, bool, error) {
	key := "ms-viewer:last:" + tagID.String()
	payload, readError := cache.client.Get(ctx, key).Result()
	if readError != nil {
		if errors.Is(readError, redis.Nil) {
			return domain.IngestRecord{}, false, nil
		}
		return domain.IngestRecord{}, false, fmt.Errorf("get redis key %s: %w", key, readError)
	}

	record, parseError := parseCachedRecord(tagID, []byte(payload), key)
	if parseError != nil {
		return domain.IngestRecord{}, false, parseError
	}

	return record, true, nil
}

func (cache *RedisCache) GetLastValues(
	ctx context.Context,
	tagIDs []uuid.UUID,
) (map[uuid.UUID]domain.IngestRecord, error) {
	if len(tagIDs) == 0 {
		return map[uuid.UUID]domain.IngestRecord{}, nil
	}

	keys := make([]string, 0, len(tagIDs))
	tagIDsByKey := make(map[string]uuid.UUID, len(tagIDs))
	seenTagIDs := make(map[uuid.UUID]struct{}, len(tagIDs))
	for _, tagID := range tagIDs {
		if tagID == uuid.Nil {
			continue
		}
		if _, exists := seenTagIDs[tagID]; exists {
			continue
		}

		seenTagIDs[tagID] = struct{}{}
		key := "ms-viewer:last:" + tagID.String()
		keys = append(keys, key)
		tagIDsByKey[key] = tagID
	}

	values, readError := cache.client.MGet(ctx, keys...).Result()
	if readError != nil {
		return nil, fmt.Errorf("mget redis snapshot values: %w", readError)
	}

	result := make(map[uuid.UUID]domain.IngestRecord, len(values))
	for index, rawValue := range values {
		if rawValue == nil {
			continue
		}

		key := keys[index]
		tagID := tagIDsByKey[key]
		payload := []byte(strings.TrimSpace(fmt.Sprint(rawValue)))
		record, parseError := parseCachedRecord(tagID, payload, key)
		if parseError != nil {
			continue
		}

		result[tagID] = record
	}

	return result, nil
}

func (cache *RedisCache) Close() error {
	if closeError := cache.client.Close(); closeError != nil {
		return fmt.Errorf("close redis client: %w", closeError)
	}

	return nil
}

func parseCachedRecord(
	tagID uuid.UUID,
	payload []byte,
	key string,
) (domain.IngestRecord, error) {
	var cached struct {
		TS string   `json:"ts"`
		V  *float64 `json:"v"`
		Q  string   `json:"q"`
	}
	if unmarshalError := json.Unmarshal(payload, &cached); unmarshalError != nil {
		return domain.IngestRecord{}, fmt.Errorf("decode redis payload %s: %w", key, unmarshalError)
	}

	parsedTimestamp, parseTimeError := time.Parse(time.RFC3339Nano, cached.TS)
	if parseTimeError != nil {
		return domain.IngestRecord{}, fmt.Errorf("parse cached timestamp %s: %w", key, parseTimeError)
	}

	quality, parseQualityError := domain.ParseQuality(cached.Q)
	if parseQualityError != nil {
		return domain.IngestRecord{}, fmt.Errorf("parse cached quality %s: %w", key, parseQualityError)
	}

	return domain.IngestRecord{
		TagID:     tagID,
		Timestamp: parsedTimestamp.UTC(),
		Value:     cached.V,
		Quality:   quality,
	}, nil
}
