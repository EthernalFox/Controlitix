package repository

import (
	"database/sql"
	"testing"
	"time"
)

type serviceAccountScannerMock struct {
	scan func(dest ...any) error
}

func (m *serviceAccountScannerMock) Scan(dest ...any) error {
	return m.scan(dest...)
}

func TestNormalizeScopes(t *testing.T) {
	t.Parallel()

	normalized := normalizeScopes([]string{
		"config.read",
		" config.read ",
		"",
		"tags.values.write",
		"scope with spaces",
	})

	if len(normalized) != 2 {
		t.Fatalf("unexpected normalized scopes: %#v", normalized)
	}
	if normalized[0] != "config.read" || normalized[1] != "tags.values.write" {
		t.Fatalf("unexpected normalized scopes: %#v", normalized)
	}
}

func TestScanServiceAccount(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 25, 18, 0, 0, 0, time.UTC)
	scanner := &serviceAccountScannerMock{
		scan: func(dest ...any) error {
			*dest[0].(*string) = "service-id"
			*dest[1].(*string) = "ms-poll"
			*dest[2].(*string) = "$argon2id$hash"
			*dest[3].(*string) = "MS Poll"
			*dest[4].(*[]string) = []string{"config.read", "tags.values.write"}
			*dest[5].(*bool) = true
			*dest[6].(*time.Time) = now
			*dest[7].(*time.Time) = now
			*dest[8].(*sql.NullTime) = sql.NullTime{
				Time:  now,
				Valid: true,
			}
			return nil
		},
	}

	account, err := scanServiceAccount(scanner)
	if err != nil {
		t.Fatalf("scan service account: %v", err)
	}

	if account.ID != "service-id" || account.ClientID != "ms-poll" {
		t.Fatalf("unexpected service account: %#v", account)
	}
	if account.LastUsedAt == nil || !account.LastUsedAt.Equal(now) {
		t.Fatalf("unexpected last used at: %#v", account.LastUsedAt)
	}
	if len(account.Scopes) != 2 {
		t.Fatalf("unexpected scopes: %#v", account.Scopes)
	}
}
