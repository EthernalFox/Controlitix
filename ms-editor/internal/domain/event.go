package domain

import (
	"context"
	"encoding/json"
	"time"
)

type ConfigChangedEvent struct {
	EntityType string          `json:"entity_type"`
	EntityID   string          `json:"entity_id"`
	Operation  string          `json:"operation"`
	Timestamp  time.Time       `json:"timestamp"`
	ActorID    *string         `json:"actor_id"`
	Payload    json.RawMessage `json:"payload"`
}

type EventPublisher interface {
	Publish(ctx context.Context, event ConfigChangedEvent) error
}
