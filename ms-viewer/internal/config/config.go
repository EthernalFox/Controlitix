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
	httpAddressEnvironmentKey          = "MS_VIEWER_HTTP_ADDRESS"
	databaseURLEnvironmentKey          = "MS_VIEWER_DATABASE_URL"
	redisURLEnvironmentKey             = "MS_VIEWER_REDIS_URL"
	kafkaBrokersEnvironmentKey         = "MS_VIEWER_KAFKA_BROKERS"
	kafkaTopicEnvironmentKey           = "MS_VIEWER_KAFKA_TAGS_VALUES_TOPIC"
	kafkaConsumerGroupEnvironmentKey   = "MS_VIEWER_KAFKA_CONSUMER_GROUP"
	kafkaConfigChangedTopicKey         = "MS_VIEWER_KAFKA_CONFIG_CHANGED_TOPIC"
	kafkaConfigChangedGroupKey         = "MS_VIEWER_KAFKA_CONFIG_CHANGED_GROUP"
	kafkaAlarmsTopicEnvironmentKey     = "MS_VIEWER_KAFKA_ALARMS_TOPIC"
	logLevelEnvironmentKey             = "MS_VIEWER_LOG_LEVEL"
	shutdownTimeoutEnvironmentKey      = "MS_VIEWER_SHUTDOWN_TIMEOUT_SEC"
	lastValueTTLEnvironmentKey         = "MS_VIEWER_LAST_VALUE_CACHE_TTL_SEC"
	trendsMaxLimitEnvironmentKey       = "MS_VIEWER_TRENDS_MAX_LIMIT"
	trendsMaxRangeDaysEnvironmentKey   = "MS_VIEWER_TRENDS_MAX_RANGE_DAYS"
	ingestBatchSizeEnvironmentKey      = "MS_VIEWER_INGEST_BATCH_SIZE"
	ingestFlushMsEnvironmentKey        = "MS_VIEWER_INGEST_FLUSH_MS"
	wsPingIntervalSecEnvironmentKey    = "MS_VIEWER_WS_PING_INTERVAL_SEC"
	wsPongTimeoutSecEnvironmentKey     = "MS_VIEWER_WS_PONG_TIMEOUT_SEC"
	wsWriteBufferSizeEnvironmentKey    = "MS_VIEWER_WS_WRITE_BUFFER_SIZE"
	wsDebounceMsEnvironmentKey         = "MS_VIEWER_WS_DEBOUNCE_MS"
	wsMaxSubscriptionsEnvironmentKey   = "MS_VIEWER_WS_MAX_SUBSCRIPTIONS_PER_CONN"
	wsMaxConnectionsEnvironmentKey     = "MS_VIEWER_WS_MAX_CONNECTIONS_PER_USER"
	wsGroupResolveLimitEnvironmentKey  = "MS_VIEWER_WS_GROUP_RESOLVE_LIMIT"
	jwksURLEnvironmentKey              = "MS_VIEWER_JWKS_URL"
	jwtIssuerEnvironmentKey            = "MS_VIEWER_JWT_ISSUER"
	jwtAudienceEnvironmentKey          = "MS_VIEWER_JWT_AUDIENCE_USER"
	tagMetaCacheSizeEnvironmentKey     = "MS_VIEWER_TAG_META_CACHE_SIZE"
	tagMetaCacheTTLSecEnvironmentKey   = "MS_VIEWER_TAG_META_CACHE_TTL_SEC"
	alarmHysteresisPercentEnvironmentKey = "MS_VIEWER_ALARM_HYSTERESIS_PERCENT"
	alarmCommLossTimeoutSecEnvironmentKey = "MS_VIEWER_ALARM_COMM_LOSS_TIMEOUT_SEC"

	defaultHTTPAddress         = ":8084"
	defaultKafkaTopic          = "tags.values"
	defaultKafkaConsumerGroup  = "ms-viewer-tags-values"
	defaultKafkaConfigChangedTopic = "config.changed"
	defaultKafkaConfigChangedGroup = "ms-viewer-config-changed"
	defaultKafkaAlarmsTopic    = "alarms.events"
	defaultLogLevel            = "info"
	defaultShutdownTimeout     = 10
	defaultLastValueTTL        = 86400
	defaultTrendsMaxLimit      = 5000
	defaultTrendsMaxRangeDays  = 90
	defaultIngestBatchSize     = 500
	defaultIngestFlushMs       = 250
	defaultWSPingIntervalSec   = 20
	defaultWSPongTimeoutSec    = 30
	defaultWSWriteBufferSize   = 256
	defaultWSDebounceMs        = 100
	defaultWSMaxSubscriptions  = 200
	defaultWSMaxConnections    = 5
	defaultWSGroupResolveLimit = 500
	defaultJWTIssuer           = "controlitix-auth"
	defaultJWTAudience         = "controlitix-api"
	defaultTagMetaCacheSize    = 10000
	defaultTagMetaCacheTTLSec  = 600
	defaultAlarmHysteresisPercent = 2
	defaultAlarmCommLossTimeoutSec = 60
)

type Config struct {
	HTTPAddress               string
	DatabaseURL               string
	RedisURL                  string
	KafkaBrokers              []string
	KafkaTagsValuesTopic      string
	KafkaConsumerGroup        string
	KafkaConfigChangedTopic   string
	KafkaConfigChangedGroup   string
	KafkaAlarmsTopic          string
	LogLevel                  string
	ShutdownTimeoutSec        int
	LastValueCacheTTLSec      int
	TrendsMaxLimit            int
	TrendsMaxRangeDays        int
	IngestBatchSize           int
	IngestFlushMs             int
	WSPingIntervalSec         int
	WSPongTimeoutSec          int
	WSWriteBufferSize         int
	WSDebounceMs              int
	WSMaxSubscriptionsPerConn int
	WSMaxConnectionsPerUser   int
	WSGroupResolveLimit       int
	JWKSURL                   string
	JWTIssuer                 string
	JWTAudienceUser           string
	TagMetaCacheSize          int
	TagMetaCacheTTLSec        int
	AlarmHysteresisPercent    float64
	AlarmCommLossTimeoutSec   int
}

func LoadConfig() (Config, error) {
	_ = godotenv.Load()

	configuration := Config{
		HTTPAddress:          readEnvironmentValue(httpAddressEnvironmentKey, defaultHTTPAddress),
		DatabaseURL:          strings.TrimSpace(os.Getenv(databaseURLEnvironmentKey)),
		RedisURL:             strings.TrimSpace(os.Getenv(redisURLEnvironmentKey)),
		KafkaBrokers:         parseEnvironmentList(kafkaBrokersEnvironmentKey),
		KafkaTagsValuesTopic: readEnvironmentValue(kafkaTopicEnvironmentKey, defaultKafkaTopic),
		KafkaConsumerGroup:   readEnvironmentValue(kafkaConsumerGroupEnvironmentKey, defaultKafkaConsumerGroup),
		KafkaConfigChangedTopic: readEnvironmentValue(
			kafkaConfigChangedTopicKey,
			defaultKafkaConfigChangedTopic,
		),
		KafkaConfigChangedGroup: readEnvironmentValue(
			kafkaConfigChangedGroupKey,
			defaultKafkaConfigChangedGroup,
		),
		KafkaAlarmsTopic: readEnvironmentValue(
			kafkaAlarmsTopicEnvironmentKey,
			defaultKafkaAlarmsTopic,
		),
		LogLevel:             readEnvironmentValue(logLevelEnvironmentKey, defaultLogLevel),
		JWKSURL:              strings.TrimSpace(os.Getenv(jwksURLEnvironmentKey)),
		JWTIssuer:            readEnvironmentValue(jwtIssuerEnvironmentKey, defaultJWTIssuer),
		JWTAudienceUser:      readEnvironmentValue(jwtAudienceEnvironmentKey, defaultJWTAudience),
	}

	var parseError error
	configuration.ShutdownTimeoutSec, parseError = parseEnvironmentInt(
		shutdownTimeoutEnvironmentKey,
		defaultShutdownTimeout,
	)
	if parseError != nil {
		return Config{}, parseError
	}
	configuration.LastValueCacheTTLSec, parseError = parseEnvironmentInt(
		lastValueTTLEnvironmentKey,
		defaultLastValueTTL,
	)
	if parseError != nil {
		return Config{}, parseError
	}
	configuration.TrendsMaxLimit, parseError = parseEnvironmentInt(
		trendsMaxLimitEnvironmentKey,
		defaultTrendsMaxLimit,
	)
	if parseError != nil {
		return Config{}, parseError
	}
	configuration.TrendsMaxRangeDays, parseError = parseEnvironmentInt(
		trendsMaxRangeDaysEnvironmentKey,
		defaultTrendsMaxRangeDays,
	)
	if parseError != nil {
		return Config{}, parseError
	}
	configuration.IngestBatchSize, parseError = parseEnvironmentInt(
		ingestBatchSizeEnvironmentKey,
		defaultIngestBatchSize,
	)
	if parseError != nil {
		return Config{}, parseError
	}
	configuration.IngestFlushMs, parseError = parseEnvironmentInt(
		ingestFlushMsEnvironmentKey,
		defaultIngestFlushMs,
	)
	if parseError != nil {
		return Config{}, parseError
	}
	configuration.WSPingIntervalSec, parseError = parseEnvironmentInt(
		wsPingIntervalSecEnvironmentKey,
		defaultWSPingIntervalSec,
	)
	if parseError != nil {
		return Config{}, parseError
	}
	configuration.WSPongTimeoutSec, parseError = parseEnvironmentInt(
		wsPongTimeoutSecEnvironmentKey,
		defaultWSPongTimeoutSec,
	)
	if parseError != nil {
		return Config{}, parseError
	}
	configuration.WSWriteBufferSize, parseError = parseEnvironmentInt(
		wsWriteBufferSizeEnvironmentKey,
		defaultWSWriteBufferSize,
	)
	if parseError != nil {
		return Config{}, parseError
	}
	configuration.WSDebounceMs, parseError = parseEnvironmentInt(
		wsDebounceMsEnvironmentKey,
		defaultWSDebounceMs,
	)
	if parseError != nil {
		return Config{}, parseError
	}
	configuration.WSMaxSubscriptionsPerConn, parseError = parseEnvironmentInt(
		wsMaxSubscriptionsEnvironmentKey,
		defaultWSMaxSubscriptions,
	)
	if parseError != nil {
		return Config{}, parseError
	}
	configuration.WSMaxConnectionsPerUser, parseError = parseEnvironmentInt(
		wsMaxConnectionsEnvironmentKey,
		defaultWSMaxConnections,
	)
	if parseError != nil {
		return Config{}, parseError
	}
	configuration.WSGroupResolveLimit, parseError = parseEnvironmentInt(
		wsGroupResolveLimitEnvironmentKey,
		defaultWSGroupResolveLimit,
	)
	if parseError != nil {
		return Config{}, parseError
	}
	configuration.TagMetaCacheSize, parseError = parseEnvironmentInt(
		tagMetaCacheSizeEnvironmentKey,
		defaultTagMetaCacheSize,
	)
	if parseError != nil {
		return Config{}, parseError
	}
	configuration.TagMetaCacheTTLSec, parseError = parseEnvironmentInt(
		tagMetaCacheTTLSecEnvironmentKey,
		defaultTagMetaCacheTTLSec,
	)
	if parseError != nil {
		return Config{}, parseError
	}
	configuration.AlarmHysteresisPercent, parseError = parseEnvironmentFloat(
		alarmHysteresisPercentEnvironmentKey,
		defaultAlarmHysteresisPercent,
	)
	if parseError != nil {
		return Config{}, parseError
	}
	configuration.AlarmCommLossTimeoutSec, parseError = parseEnvironmentInt(
		alarmCommLossTimeoutSecEnvironmentKey,
		defaultAlarmCommLossTimeoutSec,
	)
	if parseError != nil {
		return Config{}, parseError
	}

	if configuration.DatabaseURL == "" {
		return Config{}, errors.New("database connection is not configured")
	}
	if configuration.RedisURL == "" {
		return Config{}, errors.New("redis connection is not configured")
	}
	if configuration.JWKSURL == "" {
		return Config{}, errors.New("jwks url is not configured")
	}
	if len(configuration.KafkaBrokers) == 0 {
		return Config{}, errors.New("kafka brokers are not configured")
	}
	if configuration.ShutdownTimeoutSec <= 0 {
		return Config{}, errors.New("shutdown timeout must be positive")
	}
	if configuration.LastValueCacheTTLSec <= 0 {
		return Config{}, errors.New("last value cache ttl must be positive")
	}
	if configuration.TrendsMaxLimit <= 0 {
		return Config{}, errors.New("trends max limit must be positive")
	}
	if configuration.TrendsMaxRangeDays <= 0 {
		return Config{}, errors.New("trends max range days must be positive")
	}
	if configuration.IngestBatchSize <= 0 {
		return Config{}, errors.New("ingest batch size must be positive")
	}
	if configuration.IngestFlushMs <= 0 {
		return Config{}, errors.New("ingest flush interval must be positive")
	}
	if configuration.WSPingIntervalSec <= 0 {
		return Config{}, errors.New("ws ping interval must be positive")
	}
	if configuration.WSPongTimeoutSec <= 0 {
		return Config{}, errors.New("ws pong timeout must be positive")
	}
	if configuration.WSWriteBufferSize <= 0 {
		return Config{}, errors.New("ws write buffer size must be positive")
	}
	if configuration.WSDebounceMs <= 0 {
		return Config{}, errors.New("ws debounce interval must be positive")
	}
	if configuration.WSMaxSubscriptionsPerConn <= 0 {
		return Config{}, errors.New("ws max subscriptions per connection must be positive")
	}
	if configuration.WSMaxConnectionsPerUser <= 0 {
		return Config{}, errors.New("ws max connections per user must be positive")
	}
	if configuration.WSGroupResolveLimit <= 0 {
		return Config{}, errors.New("ws group resolve limit must be positive")
	}
	if configuration.TagMetaCacheSize <= 0 {
		return Config{}, errors.New("tag meta cache size must be positive")
	}
	if configuration.TagMetaCacheTTLSec <= 0 {
		return Config{}, errors.New("tag meta cache ttl must be positive")
	}
	if configuration.AlarmHysteresisPercent <= 0 {
		return Config{}, errors.New("alarm hysteresis percent must be positive")
	}
	if configuration.AlarmCommLossTimeoutSec <= 0 {
		return Config{}, errors.New("alarm comm loss timeout must be positive")
	}

	return configuration, nil
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

func parseEnvironmentFloat(environmentKey string, defaultValue float64) (float64, error) {
	value := strings.TrimSpace(os.Getenv(environmentKey))
	if value == "" {
		return defaultValue, nil
	}

	parsedValue, parseError := strconv.ParseFloat(value, 64)
	if parseError != nil {
		return 0, fmt.Errorf("parse environment variable %s: %w", environmentKey, parseError)
	}

	return parsedValue, nil
}

func parseEnvironmentList(environmentKey string) []string {
	value := strings.TrimSpace(os.Getenv(environmentKey))
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
