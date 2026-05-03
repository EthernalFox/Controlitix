package domain

import "time"

const (
	AuditResultSuccess = "success"
	AuditResultFailure = "failure"
)

type AuditTarget struct {
	Type  string `json:"type"`
	ID    string `json:"id,omitempty"`
	State string `json:"state,omitempty"`
	Name  string `json:"name,omitempty"`
}

type AuditEvent struct {
	Action  string         `json:"action"`
	Target  AuditTarget    `json:"target"`
	Details map[string]any `json:"details,omitempty"`
	Result  string         `json:"result"`
}

type AuditRecord struct {
	Version       int            `json:"v"`
	Timestamp     time.Time      `json:"ts"`
	Service       string         `json:"service"`
	ActorID       *string        `json:"actor_id"`
	ActorUsername string         `json:"actor_username"`
	ActorRoles    []string       `json:"actor_roles,omitempty"`
	Action        string         `json:"action"`
	Target        AuditTarget    `json:"target"`
	Details       map[string]any `json:"details,omitempty"`
	Result        string         `json:"result"`
	IP            string         `json:"ip,omitempty"`
	UserAgent     string         `json:"user_agent,omitempty"`
}
