package domain

import "time"

type IdentitySource struct {
	ID        int
	Type      string
	Name      string
	Config    map[string]any
	IsEnabled bool
}

type User struct {
	ID           string
	Subject      string
	SourceID     int
	Username     string
	DisplayName  string
	Email        string
	PasswordHash string
	IsActive     bool
	LastLoginAt  *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Role struct {
	Code        string
	Name        string
	Description string
}

type Credentials struct {
	Username string
	Password string
	Extra    map[string]any
}

type ExternalIdentity struct {
	Subject     string
	Username    string
	DisplayName string
	Email       string
	Groups      []string
	Raw         map[string]any
}

type AuthenticatedUser struct {
	User   User
	Source IdentitySource
	Roles  []string
}

var AllowedRoles = map[string]struct{}{
	"admin":    {},
	"engineer": {},
	"operator": {},
}

func IsAllowedRole(role string) bool {
	_, ok := AllowedRoles[role]
	return ok
}
