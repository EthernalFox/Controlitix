package jwks

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/EthernalFox/Controlitix/ms-auth/internal/config"
)

const (
	defaultKeyID   = "default"
	minRSABitSize  = 2048
	privateKeyType = "PRIVATE KEY"
)

type Keystore struct {
	privateKey *rsa.PrivateKey
	keyID      string
	jwkBytes   []byte
}

type jwkSet struct {
	Keys []jwk `json:"keys"`
}

type jwk struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

func NewKeystore(privateKeyPEM []byte, keyID string) (*Keystore, error) {
	privateKey, err := parseRSAPrivateKey(privateKeyPEM)
	if err != nil {
		return nil, err
	}
	if privateKey.N.BitLen() < minRSABitSize {
		return nil, fmt.Errorf("rsa key size must be at least %d bits", minRSABitSize)
	}

	keyID = strings.TrimSpace(keyID)
	if keyID == "" {
		keyID = defaultKeyID
	}

	jwkBytes, err := buildJWKS(privateKey, keyID)
	if err != nil {
		return nil, fmt.Errorf("build jwks: %w", err)
	}

	return &Keystore{
		privateKey: privateKey,
		keyID:      keyID,
		jwkBytes:   jwkBytes,
	}, nil
}

func LoadFromConfig(cfg config.Config) (*Keystore, error) {
	keyPEM, err := readPrivateKeyFromConfig(cfg)
	if err != nil {
		return nil, err
	}

	keystore, err := NewKeystore(keyPEM, cfg.JWTKeyID)
	if err != nil {
		return nil, fmt.Errorf("initialize jwt keystore: %w", err)
	}

	return keystore, nil
}

func (k *Keystore) PrivateKey() *rsa.PrivateKey {
	return k.privateKey
}

func (k *Keystore) KeyID() string {
	return k.keyID
}

func (k *Keystore) JWKS() []byte {
	jwksCopy := make([]byte, len(k.jwkBytes))
	copy(jwksCopy, k.jwkBytes)
	return jwksCopy
}

func readPrivateKeyFromConfig(cfg config.Config) ([]byte, error) {
	if strings.TrimSpace(cfg.JWTPrivateKeyPath) != "" {
		keyPEM, err := os.ReadFile(strings.TrimSpace(cfg.JWTPrivateKeyPath))
		if err != nil {
			return nil, fmt.Errorf("read jwt private key file: %w", err)
		}
		return keyPEM, nil
	}

	inlinePEM := strings.TrimSpace(cfg.JWTPrivateKeyPEM)
	if inlinePEM == "" {
		return nil, errors.New("jwt private key is not configured")
	}

	inlinePEM = strings.ReplaceAll(inlinePEM, "\\n", "\n")
	return []byte(inlinePEM), nil
}

func parseRSAPrivateKey(privateKeyPEM []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(privateKeyPEM)
	if block == nil {
		return nil, errors.New("invalid private key pem")
	}

	if x509.IsEncryptedPEMBlock(block) ||
		strings.Contains(strings.ToUpper(block.Type), "ENCRYPTED") {
		return nil, errors.New("encrypted private keys are not supported")
	}

	if block.Type == "RSA PRIVATE KEY" {
		key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("parse pkcs1 private key: %w", err)
		}
		return key, nil
	}

	if block.Type == privateKeyType || strings.HasSuffix(block.Type, "PRIVATE KEY") {
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("parse pkcs8 private key: %w", err)
		}
		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("private key is not RSA")
		}
		return rsaKey, nil
	}

	return nil, fmt.Errorf("unsupported private key type: %s", block.Type)
}

func buildJWKS(privateKey *rsa.PrivateKey, keyID string) ([]byte, error) {
	publicKey := privateKey.PublicKey
	modulus := base64.RawURLEncoding.EncodeToString(publicKey.N.Bytes())
	exponentBytes := intToBytes(publicKey.E)
	exponent := base64.RawURLEncoding.EncodeToString(exponentBytes)

	payload := jwkSet{
		Keys: []jwk{
			{
				Kty: "RSA",
				Use: "sig",
				Alg: "RS256",
				Kid: keyID,
				N:   modulus,
				E:   exponent,
			},
		},
	}

	jwksBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal jwks payload: %w", err)
	}

	return jwksBytes, nil
}

func intToBytes(value int) []byte {
	if value == 0 {
		return []byte{0}
	}

	result := make([]byte, 0, 8)
	for value > 0 {
		result = append([]byte{byte(value & 0xff)}, result...)
		value >>= 8
	}

	return result
}
