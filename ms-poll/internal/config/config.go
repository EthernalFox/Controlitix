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
	httpAddressEnvironmentKey         = "MS_POLL_HTTP_ADDRESS"
	databaseURLEnvironmentKey         = "MS_POLL_DATABASE_URL"
	logLevelEnvironmentKey            = "MS_POLL_LOG_LEVEL"
	kafkaBrokersEnvironmentKey        = "MS_POLL_KAFKA_BROKERS"
	kafkaConfigTopicEnvironmentKey    = "MS_POLL_KAFKA_CONFIG_TOPIC"
	kafkaConfigGroupIDEnvironmentKey  = "MS_POLL_KAFKA_CONFIG_GROUP_ID"
	kafkaTagValuesTopicEnvironmentKey = "MS_POLL_KAFKA_TAG_VALUES_TOPIC"
	kafkaRawReadingsTopicEnvironmentKey = "MS_POLL_KAFKA_RAW_READINGS_TOPIC"
	publishRawReadingsEnvironmentKey   = "MS_POLL_PUBLISH_RAW_READINGS"
	shutdownTimeoutEnvironmentKey     = "MS_POLL_SHUTDOWN_TIMEOUT_SEC"

	defaultHTTPAddress        = ":8081"
	defaultLogLevel           = "info"
	defaultKafkaConfigTopic   = "config.changed"
	defaultKafkaConfigGroupID = "ms-poll"
	defaultKafkaTagValues     = "tags.values"
	defaultKafkaRawReadings   = "raw.readings"
	defaultPublishRawReadings = false
	defaultShutdownTimeoutSec = 10
)

type Config struct {
	HTTPAddress         string
	DatabaseURL         string
	LogLevel            string
	KafkaBrokers        []string
	KafkaConfigTopic    string
	KafkaConfigGroupID  string
	KafkaTagValuesTopic string
	KafkaRawReadingsTopic string
	PublishRawReadings bool
	ShutdownTimeoutSec  int
}

func LoadConfig() (Config, error) {
	_ = godotenv.Load()

	httpAddress := readEnvironmentValue(
		httpAddressEnvironmentKey,
		defaultHTTPAddress,
	)
	databaseURL := readEnvironmentValue(databaseURLEnvironmentKey, "")
	logLevel := readEnvironmentValue(logLevelEnvironmentKey, defaultLogLevel)
	kafkaBrokers := parseEnvironmentList(kafkaBrokersEnvironmentKey)
	kafkaConfigTopic := readEnvironmentValue(
		kafkaConfigTopicEnvironmentKey,
		defaultKafkaConfigTopic,
	)
	kafkaConfigGroupID := readEnvironmentValue(
		kafkaConfigGroupIDEnvironmentKey,
		defaultKafkaConfigGroupID,
	)
	kafkaTagValuesTopic := readEnvironmentValue(
		kafkaTagValuesTopicEnvironmentKey,
		defaultKafkaTagValues,
	)
	kafkaRawReadingsTopic := readEnvironmentValue(
		kafkaRawReadingsTopicEnvironmentKey,
		defaultKafkaRawReadings,
	)
	publishRawReadings, publishRawReadingsError := parseEnvironmentBool(
		publishRawReadingsEnvironmentKey,
		defaultPublishRawReadings,
	)
	if publishRawReadingsError != nil {
		return Config{}, publishRawReadingsError
	}
	shutdownTimeoutSec, shutdownTimeoutError := parseEnvironmentInt(
		shutdownTimeoutEnvironmentKey,
		defaultShutdownTimeoutSec,
	)
	if shutdownTimeoutError != nil {
		return Config{}, shutdownTimeoutError
	}

	if databaseURL == "" {
		return Config{}, errors.New("database connection is not configured")
	}

	if len(kafkaBrokers) == 0 {
		return Config{}, errors.New("kafka brokers are not configured")
	}

	if shutdownTimeoutSec <= 0 {
		return Config{}, errors.New("shutdown timeout must be positive")
	}

	return Config{
		HTTPAddress:         httpAddress,
		DatabaseURL:         databaseURL,
		LogLevel:            logLevel,
		KafkaBrokers:        kafkaBrokers,
		KafkaConfigTopic:    kafkaConfigTopic,
		KafkaConfigGroupID:  kafkaConfigGroupID,
		KafkaTagValuesTopic: kafkaTagValuesTopic,
		KafkaRawReadingsTopic: kafkaRawReadingsTopic,
		PublishRawReadings: publishRawReadings,
		ShutdownTimeoutSec:  shutdownTimeoutSec,
	}, nil
}

func readEnvironmentValue(environmentKey string, defaultValue string) string {
	value := strings.TrimSpace(os.Getenv(environmentKey))
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

func parseEnvironmentInt(environmentKey string, defaultValue int) (int, error) {
	value := strings.TrimSpace(os.Getenv(environmentKey))
	if value == "" {
		return defaultValue, nil
	}

	parsedValue, parseError := strconv.Atoi(value)
	if parseError != nil {
		return 0, fmt.Errorf(
			"parse environment variable %s: %w",
			environmentKey,
			parseError,
		)
	}

	return parsedValue, nil
}

func parseEnvironmentBool(
	environmentKey string,
	defaultValue bool,
) (bool, error) {
	value := strings.TrimSpace(strings.ToLower(os.Getenv(environmentKey)))
	if value == "" {
		return defaultValue, nil
	}

	parsedValue, parseError := strconv.ParseBool(value)
	if parseError != nil {
		return false, fmt.Errorf(
			"parse environment variable %s: %w",
			environmentKey,
			parseError,
		)
	}

	return parsedValue, nil
}
