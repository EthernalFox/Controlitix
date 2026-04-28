package jwks

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"testing"
)

func TestNewKeystoreLoadsPKCS1(t *testing.T) {
	t.Parallel()

	keyPEM, key := generatePEMPrivateKey(t, 2048, false)

	keystore, err := NewKeystore(keyPEM, "test-kid")
	if err != nil {
		t.Fatalf("new keystore: %v", err)
	}

	if keystore.KeyID() != "test-kid" {
		t.Fatalf("unexpected kid: %s", keystore.KeyID())
	}
	if keystore.PrivateKey().N.Cmp(key.N) != 0 {
		t.Fatal("unexpected private key loaded from pkcs1 pem")
	}
}

func TestNewKeystoreLoadsPKCS8(t *testing.T) {
	t.Parallel()

	keyPEM, key := generatePEMPrivateKey(t, 2048, true)

	keystore, err := NewKeystore(keyPEM, "test-kid")
	if err != nil {
		t.Fatalf("new keystore: %v", err)
	}

	if keystore.KeyID() != "test-kid" {
		t.Fatalf("unexpected kid: %s", keystore.KeyID())
	}
	if keystore.PrivateKey().N.Cmp(key.N) != 0 {
		t.Fatal("unexpected private key loaded from pkcs8 pem")
	}
}

func TestNewKeystoreRejectsSmallRSAKey(t *testing.T) {
	t.Parallel()

	keyPEM, _ := generatePEMPrivateKey(t, 1024, false)
	if _, err := NewKeystore(keyPEM, "kid"); err == nil {
		t.Fatal("expected error for rsa key smaller than 2048 bits")
	}
}

func TestNewKeystoreBuildsExpectedJWKS(t *testing.T) {
	t.Parallel()

	keyPEM, key := generatePEMPrivateKey(t, 2048, false)

	keystore, err := NewKeystore(keyPEM, "kid-1")
	if err != nil {
		t.Fatalf("new keystore: %v", err)
	}

	var payload struct {
		Keys []struct {
			Kty string `json:"kty"`
			Use string `json:"use"`
			Alg string `json:"alg"`
			Kid string `json:"kid"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	if err := json.Unmarshal(keystore.JWKS(), &payload); err != nil {
		t.Fatalf("unmarshal jwks payload: %v", err)
	}

	if len(payload.Keys) != 1 {
		t.Fatalf("expected one key in jwks, got %d", len(payload.Keys))
	}

	jwkKey := payload.Keys[0]
	if jwkKey.Kid != "kid-1" {
		t.Fatalf("unexpected kid: %s", jwkKey.Kid)
	}
	if jwkKey.Kty != "RSA" || jwkKey.Use != "sig" || jwkKey.Alg != "RS256" {
		t.Fatalf("unexpected jwk metadata: %#v", jwkKey)
	}

	expectedN := base64.RawURLEncoding.EncodeToString(key.PublicKey.N.Bytes())
	if jwkKey.N != expectedN {
		t.Fatal("jwks modulus does not match public key")
	}

	expectedE := base64.RawURLEncoding.EncodeToString(intToBytes(key.PublicKey.E))
	if jwkKey.E != expectedE {
		t.Fatal("jwks exponent does not match public key")
	}
}

func generatePEMPrivateKey(t *testing.T, bits int, pkcs8 bool) ([]byte, *rsa.PrivateKey) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}

	if pkcs8 {
		keyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
		if err != nil {
			t.Fatalf("marshal pkcs8 private key: %v", err)
		}
		return pem.EncodeToMemory(&pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: keyBytes,
		}), privateKey
	}

	return pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}), privateKey
}
