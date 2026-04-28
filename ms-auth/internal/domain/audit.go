package domain

import "time"

type AuditEvent struct {
	OccurredAt   time.Time      `json:"occurred_at"`
	ActorSubject string         `json:"actor_subject"`
	Action       string         `json:"action"`
	Target       string         `json:"target"`
	Result       string         `json:"result"`
	Reason       string         `json:"reason"`
	IP           string         `json:"ip"`
	UserAgent    string         `json:"user_agent"`
	Metadata     map[string]any `json:"metadata"`
}

const (
	AuditLoginSuccess       = "login.success"
	AuditLoginFailure       = "login.failure"
	AuditTokenRefresh       = "token.refresh"
	AuditTokenReuseDetected = "token.reuse_detected"
	AuditTokenRevoke        = "token.revoke"
	AuditLogout             = "logout"
	AuditRateLimited        = "rate_limited"
	AuditServiceTokenSuccess = "service_token.success"
	AuditServiceTokenFailure = "service_token.failure"
	AuditAdminUserCreate = "admin.user.create"
	AuditAdminUserUpdate = "admin.user.update"
	AuditAdminUserDeactivate = "admin.user.deactivate"
	AuditAdminUserSetRoles = "admin.user.set_roles"
	AuditAdminUserRevokeSessions = "admin.user.revoke_sessions"
	AuditAdminServiceCreate = "admin.service_account.create"
	AuditAdminServiceRotateSecret = "admin.service_account.rotate_secret"
	AuditAdminServiceUpdate = "admin.service_account.update"
	AuditAdminServiceDelete = "admin.service_account.delete"
)

const (
	AuditResultSuccess = "success"
	AuditResultFailure = "failure"
)
