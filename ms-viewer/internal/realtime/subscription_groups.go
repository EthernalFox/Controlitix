package realtime

import (
	"encoding/json"
	"strings"

	"github.com/google/uuid"
)

type connectionSubscriptionGroups struct {
	directTags   map[uuid.UUID]struct{}
	groupedTopics map[string]map[uuid.UUID]struct{}
	tagRefs      map[uuid.UUID]int
}

func newConnectionSubscriptionGroups() *connectionSubscriptionGroups {
	return &connectionSubscriptionGroups{
		directTags:    make(map[uuid.UUID]struct{}),
		groupedTopics: make(map[string]map[uuid.UUID]struct{}),
		tagRefs:       make(map[uuid.UUID]int),
	}
}

func extractStringFromPayload(payload json.RawMessage, field string) string {
	if len(payload) == 0 {
		return ""
	}

	var decoded map[string]any
	if unmarshalError := json.Unmarshal(payload, &decoded); unmarshalError != nil {
		return ""
	}

	value, exists := decoded[field]
	if !exists {
		return ""
	}

	stringValue, ok := value.(string)
	if !ok {
		return ""
	}

	return strings.TrimSpace(stringValue)
}
