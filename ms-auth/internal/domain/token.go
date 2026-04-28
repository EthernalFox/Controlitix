package domain

import "time"

type AccessClaims struct {
	Issuer      string   `json:"iss"`
	Audience    []string `json:"aud"`
	Subject     string   `json:"sub"`
	IssuedAt    int64    `json:"iat"`
	ExpiresAt   int64    `json:"exp"`
	NotBefore   int64    `json:"nbf"`
	JWTID       string   `json:"jti"`
	Source      string   `json:"src"`
	Roles       []string `json:"roles"`
	Scope       string   `json:"scope"`
	Username    string   `json:"username,omitempty"`
	DisplayName string   `json:"display_name,omitempty"`
	ClientID    string   `json:"client_id,omitempty"`
}

type IssuedTokens struct {
	AccessToken      string
	RefreshToken     string
	ExpiresIn        int
	RefreshExpiresIn int
	FamilyID         string
}

type RefreshRecord struct {
	ID        string
	UserID    string
	FamilyID  string
	TokenHash string
	ParentID  *string
	IssuedAt  time.Time
	ExpiresAt time.Time
	UsedAt    *time.Time
	RevokedAt *time.Time
	UserAgent string
	IP        string
}
