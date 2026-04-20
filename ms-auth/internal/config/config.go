package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

const (
	httpAddressEnvironmentKey    = "MS_AUTH_HTTP_ADDRESS"
	databaseURLEnvironmentKey    = "MS_AUTH_DATABASE_URL"
	logLevelEnvironmentKey       = "MS_AUTH_LOG_LEVEL"
	shutdownTimeoutEnvironmentKey = "MS_AUTH_SHUTDOWN_TIMEOUT_SEC"

	defaultHTTPAddress       = ":8083"
	defaultLogLevel          = "info"
	defaultShutdownTimeoutSec = 10
)

type Config struct {
	HTTPAddress        string
	DatabaseURL        string
	LogLevel           string
	ShutdownTimeoutSec int
}

func LoadConfig() (Config, error) {
	_ = godotenv.Load()

	httpAddress := readEnvironmentValue(httpAddressEnvironmentKey, defaultHTTPAddress)
	databaseURL := strings.TrimSpace(os.Getenv(databaseURLEnvironmentKey))
	logLevel := readEnvironmentValue(logLevelEnvironmentKey, defaultLogLevel)
	shutdownTimeoutSec, parseError := parseEnvironmentInt(
		shutdownTimeoutEnvironmentKey,
		defaultShutdownTimeoutSec,
	)
	if parseError != nil {
		return Config{}, parseError
	}

	if databaseURL == "" {
		return Config{}, errors.New("database connection is not configured")
	}
	if shutdownTimeoutSec <= 0 {
		return Config{}, errors.New("shutdown timeout must be positive")
	}

	return Config{
		HTTPAddress:        httpAddress,
		DatabaseURL:        databaseURL,
		LogLevel:           logLevel,
		ShutdownTimeoutSec: shutdownTimeoutSec,
	}, nil
}

func readEnvironmentValue(environmentKey string, defaultValue string) string {
	value := strings.TrimSpace(os.Getenv(environmentKey))
	if value == "" {
		return defaultValue
	}

	return value
}

func parseEnvironmentInt(environmentKey string, defaultValue int) (int, error) {
	value := strings.TrimSpace(os.Getenv(environmentKey))
	if value == "" {
		return defaultValue, nil
	}

	parsedValue, parseError := strconv.Atoi(value)
	if parseError != nil {
		return 0, fmt.Errorf("parse environment variable %s: %w", environmentKey, parseError)
	}

	return parsedValue, nil
}
