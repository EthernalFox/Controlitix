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
	httpAddressEnvironmentKey  = "MS_EDITOR_HTTP_ADDRESS"
	databaseURLEnvironmentKey  = "MS_EDITOR_DATABASE_URL"
	logLevelEnvironmentKey     = "MS_EDITOR_LOG_LEVEL"
	kafkaBrokersEnvironmentKey = "MS_EDITOR_KAFKA_BROKERS"
	kafkaTopicEnvironmentKey   = "MS_EDITOR_KAFKA_TOPIC"
	authJWKSURLEnvironmentKey  = "MS_EDITOR_AUTH_JWKS_URL"
	authIssuerEnvironmentKey   = "MS_EDITOR_AUTH_ISSUER"
	authAudienceEnvironmentKey = "MS_EDITOR_AUTH_AUDIENCE"
	authDisabledEnvironmentKey = "MS_EDITOR_AUTH_DISABLED"
	defaultHTTPAddress         = ":8080"
	defaultLogLevel            = "info"
	defaultKafkaTopic          = "config.changed"
	defaultAuthIssuer          = "controlitix-auth"
	defaultAuthAudience        = "controlitix-api"
)

type Config struct {
	HTTPAddress  string
	DatabaseURL  string
	LogLevel     string
	KafkaBrokers []string
	KafkaTopic   string
	AuthJWKSURL  string
	AuthIssuer   string
	AuthAudience string
	AuthDisabled bool
}

func LoadConfig() (Config, error) {
	_ = godotenv.Load()

	httpAddress := readEnvironmentValue(httpAddressEnvironmentKey, defaultHTTPAddress)
	databaseURL := readEnvironmentValue(databaseURLEnvironmentKey, "")
	logLevel := readEnvironmentValue(logLevelEnvironmentKey, defaultLogLevel)
	kafkaBrokers := parseEnvironmentList(kafkaBrokersEnvironmentKey)
	kafkaTopic := readEnvironmentValue(kafkaTopicEnvironmentKey, defaultKafkaTopic)
	authJWKSURL := readEnvironmentValue(authJWKSURLEnvironmentKey, "")
	authIssuer := readEnvironmentValue(authIssuerEnvironmentKey, defaultAuthIssuer)
	authAudience := readEnvironmentValue(authAudienceEnvironmentKey, defaultAuthAudience)
	authDisabled, parseBoolError := readEnvironmentBool(authDisabledEnvironmentKey, false)
	if parseBoolError != nil {
		return Config{}, parseBoolError
	}

	if databaseURL == "" {
		return Config{}, errors.New("database connection is not configured")
	}
	if !authDisabled && strings.TrimSpace(authJWKSURL) == "" {
		return Config{}, errors.New("auth jwks url is not configured")
	}

	return Config{
		HTTPAddress:  httpAddress,
		DatabaseURL:  databaseURL,
		LogLevel:     logLevel,
		KafkaBrokers: kafkaBrokers,
		KafkaTopic:   kafkaTopic,
		AuthJWKSURL:  authJWKSURL,
		AuthIssuer:   authIssuer,
		AuthAudience: authAudience,
		AuthDisabled: authDisabled,
	}, nil
}

func readEnvironmentValue(environmentKey string, defaultValue string) string {
	value := os.Getenv(environmentKey)
	if value == "" {
		return defaultValue
	}

	return value
}

func parseEnvironmentList(environmentKey string) []string {
	value := os.Getenv(environmentKey)
	if value == "" {
		return nil
	}

	rawValues := strings.Split(value, ",")
	parsedValues := make([]string, 0, len(rawValues))
	for _, rawValue := range rawValues {
		trimmedValue := strings.TrimSpace(rawValue)
		if trimmedValue == "" {
			continue
		}

		parsedValues = append(parsedValues, trimmedValue)
	}

	return parsedValues
}

func readEnvironmentBool(environmentKey string, defaultValue bool) (bool, error) {
	rawValue := strings.TrimSpace(os.Getenv(environmentKey))
	if rawValue == "" {
		return defaultValue, nil
	}

	parsedValue, parseError := strconv.ParseBool(rawValue)
	if parseError != nil {
		return false, fmt.Errorf("invalid boolean value for %s: %w", environmentKey, parseError)
	}

	return parsedValue, nil
}
