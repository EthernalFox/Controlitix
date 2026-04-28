package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argonTimeCost    = uint32(3)
	argonMemoryKiB   = uint32(64 * 1024)
	argonParallelism = uint8(2)
	argonSaltLength  = 16
	argonHashLength  = 32
)

func hashPasswordFromStdin() (string, error) {
	pepperValue := strings.TrimSpace(os.Getenv("MS_AUTH_PASSWORD_PEPPER"))
	if pepperValue == "" {
		return "", errors.New("MS_AUTH_PASSWORD_PEPPER is required")
	}

	pepper, decodeError := base64.StdEncoding.DecodeString(pepperValue)
	if decodeError != nil {
		return "", fmt.Errorf("decode MS_AUTH_PASSWORD_PEPPER: %w", decodeError)
	}
	if len(pepper) < 32 {
		return "", errors.New("MS_AUTH_PASSWORD_PEPPER must be at least 32 bytes after base64 decode")
	}

	passwordBytes, readError := io.ReadAll(os.Stdin)
	if readError != nil {
		return "", fmt.Errorf("read password from stdin: %w", readError)
	}

	password := strings.TrimRight(string(passwordBytes), "\r\n")
	if password == "" {
		return "", errors.New("password is empty")
	}

	salt := make([]byte, argonSaltLength)
	if _, randomError := rand.Read(salt); randomError != nil {
		return "", fmt.Errorf("generate salt: %w", randomError)
	}

	passwordWithPepper := make([]byte, 0, len(password)+len(pepper))
	passwordWithPepper = append(passwordWithPepper, []byte(password)...)
	passwordWithPepper = append(passwordWithPepper, pepper...)

	hash := argon2.IDKey(
		passwordWithPepper,
		salt,
		argonTimeCost,
		argonMemoryKiB,
		argonParallelism,
		argonHashLength,
	)

	saltEncoded := base64.RawStdEncoding.EncodeToString(salt)
	hashEncoded := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		argonMemoryKiB,
		argonTimeCost,
		argonParallelism,
		saltEncoded,
		hashEncoded,
	), nil
}
