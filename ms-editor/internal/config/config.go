package config

import (
	"errors"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

const (
	httpAddressEnvironmentKey  = "MS_EDITOR_HTTP_ADDRESS"
	databaseURLEnvironmentKey  = "MS_EDITOR_DATABASE_URL"
	logLevelEnvironmentKey     = "MS_EDITOR_LOG_LEVEL"
	kafkaBrokersEnvironmentKey = "MS_EDITOR_KAFKA_BROKERS"
	kafkaTopicEnvironmentKey   = "MS_EDITOR_KAFKA_TOPIC"
	defaultHTTPAddress         = ":8080"
	defaultLogLevel            = "info"
	defaultKafkaTopic          = "config.changed"
)

type Config struct {
	HTTPAddress  string
	DatabaseURL  string
	LogLevel     string
	KafkaBrokers []string
	KafkaTopic   string
}

func LoadConfig() (Config, error) {
	_ = godotenv.Load()

	httpAddress := readEnvironmentValue(httpAddressEnvironmentKey, defaultHTTPAddress)
	databaseURL := readEnvironmentValue(databaseURLEnvironmentKey, "")
	logLevel := readEnvironmentValue(logLevelEnvironmentKey, defaultLogLevel)
	kafkaBrokers := parseEnvironmentList(kafkaBrokersEnvironmentKey)
	kafkaTopic := readEnvironmentValue(kafkaTopicEnvironmentKey, defaultKafkaTopic)

	if databaseURL == "" {
		return Config{}, errors.New("database connection is not configured")
	}

	return Config{
		HTTPAddress:  httpAddress,
		DatabaseURL:  databaseURL,
		LogLevel:     logLevel,
		KafkaBrokers: kafkaBrokers,
		KafkaTopic:   kafkaTopic,
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
