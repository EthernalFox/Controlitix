package domain

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ConfigChangedEvent struct {
	EntityType string          `json:"entity_type"`
	EntityID   string          `json:"entity_id"`
	Operation  string          `json:"operation"`
	Timestamp  time.Time       `json:"timestamp"`
	ActorID    *string         `json:"actor_id"`
	Payload    json.RawMessage `json:"payload"`
}

func (event ConfigChangedEvent) EntityUUID() uuid.UUID {
	parsedID, parseError := uuid.Parse(strings.TrimSpace(event.EntityID))
	if parseError != nil {
		return uuid.Nil
	}

	return parsedID
}
