package password

import (
	"strings"
	"testing"
)

func TestHasherHashAndVerify(t *testing.T) {
	t.Parallel()

	pepper := []byte(strings.Repeat("p", 32))
	hasher, err := NewHasher(pepper)
	if err != nil {
		t.Fatalf("new hasher: %v", err)
	}

	encoded, err := hasher.Hash("secret-password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	ok, err := hasher.Verify("secret-password", encoded)
	if err != nil {
		t.Fatalf("verify password: %v", err)
	}
	if !ok {
		t.Fatal("expected password verification to succeed")
	}
}

func TestHasherVerifyWithDifferentPasswordReturnsFalse(t *testing.T) {
	t.Parallel()

	pepper := []byte(strings.Repeat("p", 32))
	hasher, err := NewHasher(pepper)
	if err != nil {
		t.Fatalf("new hasher: %v", err)
	}

	encoded, err := hasher.Hash("primary-password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	ok, err := hasher.Verify("other-password", encoded)
	if err != nil {
		t.Fatalf("verify password: %v", err)
	}
	if ok {
		t.Fatal("expected password verification to fail for another password")
	}
}

func TestHasherVerifyWithDifferentPepperReturnsFalse(t *testing.T) {
	t.Parallel()

	hasherA, err := NewHasher([]byte(strings.Repeat("a", 32)))
	if err != nil {
		t.Fatalf("new hasher A: %v", err)
	}

	hasherB, err := NewHasher([]byte(strings.Repeat("b", 32)))
	if err != nil {
		t.Fatalf("new hasher B: %v", err)
	}

	encoded, err := hasherA.Hash("same-password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	ok, err := hasherB.Verify("same-password", encoded)
	if err != nil {
		t.Fatalf("verify password: %v", err)
	}
	if ok {
		t.Fatal("expected password verification to fail for another pepper")
	}
}

func TestHasherVerifyWithCorruptedEncodedReturnsError(t *testing.T) {
	t.Parallel()

	hasher, err := NewHasher([]byte(strings.Repeat("p", 32)))
	if err != nil {
		t.Fatalf("new hasher: %v", err)
	}

	if _, err := hasher.Verify("password", "not-an-argon2-encoded-value"); err == nil {
		t.Fatal("expected error for malformed encoded hash")
	}
}

func TestNewHasherWithEmptyPepperReturnsError(t *testing.T) {
	t.Parallel()

	if _, err := NewHasher(nil); err == nil {
		t.Fatal("expected error for empty pepper")
	}
}

func TestHasherHashUsesRandomSalt(t *testing.T) {
	t.Parallel()

	hasher, err := NewHasher([]byte(strings.Repeat("p", 32)))
	if err != nil {
		t.Fatalf("new hasher: %v", err)
	}

	firstHash, err := hasher.Hash("same-password")
	if err != nil {
		t.Fatalf("first hash: %v", err)
	}

	secondHash, err := hasher.Hash("same-password")
	if err != nil {
		t.Fatalf("second hash: %v", err)
	}

	if firstHash == secondHash {
		t.Fatal("expected different hashes for equal password because of random salt")
	}
}
