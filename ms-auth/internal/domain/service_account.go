package domain

import "time"

type ServiceAccount struct {
	ID               string
	ClientID         string
	ClientSecretHash string
	DisplayName      string
	Scopes           []string
	IsActive         bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
	LastUsedAt       *time.Time
}
