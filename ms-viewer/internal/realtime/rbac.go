package realtime

import "strings"

func CanSubscribeTopic(roles []string, topic string) bool {
	topic = strings.TrimSpace(topic)
	if topic == "" {
		return false
	}

	if !hasViewerRealtimeRole(roles) {
		return false
	}

	if IsAlarmsTopic(topic) {
		return true
	}
	if _, parseError := ParseTagTopic(topic); parseError == nil {
		return true
	}
	if _, _, parseError := ParseGroupTopic(topic); parseError == nil {
		return true
	}

	return false
}

func hasViewerRealtimeRole(roles []string) bool {
	for _, role := range roles {
		switch strings.ToLower(strings.TrimSpace(role)) {
		case "operator", "engineer", "admin":
			return true
		}
	}

	// TODO(acl): add per-object/per-diagram checks when object-level ACL arrives.
	return false
}
