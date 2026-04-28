package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	defaultMemoryKiB  = uint32(64 * 1024)
	defaultTimeCost   = uint32(3)
	defaultParallel   = uint8(2)
	defaultSaltLength = 16
	defaultKeyLength  = 32
	minPepperLength   = 32
)

type Hasher struct {
	pepper []byte
}

type encodedArgon2 struct {
	memoryKiB uint32
	timeCost  uint32
	parallel  uint8
	salt      []byte
	hash      []byte
}

func NewHasher(pepper []byte) (*Hasher, error) {
	if len(pepper) < minPepperLength {
		return nil, fmt.Errorf("pepper must be at least %d bytes", minPepperLength)
	}

	pepperCopy := make([]byte, len(pepper))
	copy(pepperCopy, pepper)

	return &Hasher{pepper: pepperCopy}, nil
}

func (h *Hasher) Hash(password string) (string, error) {
	if h == nil {
		return "", errors.New("hasher is nil")
	}

	salt := make([]byte, defaultSaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}

	derived := h.derive(password, salt, defaultTimeCost, defaultMemoryKiB, defaultParallel, defaultKeyLength)
	saltEncoded := base64.RawStdEncoding.EncodeToString(salt)
	hashEncoded := base64.RawStdEncoding.EncodeToString(derived)

	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		defaultMemoryKiB,
		defaultTimeCost,
		defaultParallel,
		saltEncoded,
		hashEncoded,
	), nil
}

func (h *Hasher) Verify(password, encoded string) (bool, error) {
	if h == nil {
		return false, errors.New("hasher is nil")
	}

	parsed, err := parseEncodedArgon2(encoded)
	if err != nil {
		return false, err
	}

	derived := h.derive(
		password,
		parsed.salt,
		parsed.timeCost,
		parsed.memoryKiB,
		parsed.parallel,
		uint32(len(parsed.hash)),
	)

	if subtle.ConstantTimeCompare(derived, parsed.hash) == 1 {
		return true, nil
	}

	return false, nil
}

func (h *Hasher) derive(
	password string,
	salt []byte,
	timeCost uint32,
	memoryKiB uint32,
	parallel uint8,
	keyLength uint32,
) []byte {
	passwordWithPepper := make([]byte, 0, len(password)+len(h.pepper))
	passwordWithPepper = append(passwordWithPepper, password...)
	passwordWithPepper = append(passwordWithPepper, h.pepper...)

	return argon2.IDKey(
		passwordWithPepper,
		salt,
		timeCost,
		memoryKiB,
		parallel,
		keyLength,
	)
}

func parseEncodedArgon2(encoded string) (*encodedArgon2, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[0] != "" {
		return nil, errors.New("invalid argon2 encoded format")
	}
	if parts[1] != "argon2id" {
		return nil, errors.New("unsupported argon2 algorithm")
	}

	version, err := parseArgonVersion(parts[2])
	if err != nil {
		return nil, err
	}
	if version != argon2.Version {
		return nil, fmt.Errorf("unsupported argon2 version: %d", version)
	}

	memoryKiB, timeCost, parallel, err := parseArgonParams(parts[3])
	if err != nil {
		return nil, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, fmt.Errorf("decode salt: %w", err)
	}
	if len(salt) == 0 {
		return nil, errors.New("salt is empty")
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, fmt.Errorf("decode hash: %w", err)
	}
	if len(hash) == 0 {
		return nil, errors.New("hash is empty")
	}

	return &encodedArgon2{
		memoryKiB: memoryKiB,
		timeCost:  timeCost,
		parallel:  parallel,
		salt:      salt,
		hash:      hash,
	}, nil
}

func parseArgonVersion(raw string) (int, error) {
	parts := strings.Split(raw, "=")
	if len(parts) != 2 || parts[0] != "v" {
		return 0, errors.New("invalid argon2 version field")
	}

	version, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, fmt.Errorf("parse argon2 version: %w", err)
	}

	return version, nil
}

func parseArgonParams(raw string) (uint32, uint32, uint8, error) {
	segments := strings.Split(raw, ",")
	if len(segments) != 3 {
		return 0, 0, 0, errors.New("invalid argon2 params field")
	}

	values := map[string]string{}
	for _, segment := range segments {
		kv := strings.Split(segment, "=")
		if len(kv) != 2 {
			return 0, 0, 0, errors.New("invalid argon2 params field")
		}
		values[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
	}

	memory, err := strconv.ParseUint(values["m"], 10, 32)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("parse argon2 memory: %w", err)
	}

	timeCost, err := strconv.ParseUint(values["t"], 10, 32)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("parse argon2 time cost: %w", err)
	}

	parallel, err := strconv.ParseUint(values["p"], 10, 8)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("parse argon2 parallelism: %w", err)
	}

	if memory == 0 || timeCost == 0 || parallel == 0 {
		return 0, 0, 0, errors.New("argon2 params must be positive")
	}

	return uint32(memory), uint32(timeCost), uint8(parallel), nil
}
