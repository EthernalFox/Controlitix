package config

import (
	"encoding/base64"
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
	passwordPepperEnvironmentKey = "MS_AUTH_PASSWORD_PEPPER"
	jwtPrivateKeyPathEnvironmentKey = "MS_AUTH_JWT_PRIVATE_KEY_PATH"
	jwtPrivateKeyEnvironmentKey = "MS_AUTH_JWT_PRIVATE_KEY"
	jwtKeyIDEnvironmentKey = "MS_AUTH_JWT_KEY_ID"
	jwtIssuerEnvironmentKey = "MS_AUTH_JWT_ISSUER"
	jwtAudienceUserEnvironmentKey = "MS_AUTH_JWT_AUDIENCE_USER"
	jwtAudienceServiceEnvironmentKey = "MS_AUTH_JWT_AUDIENCE_SERVICE"
	jwtAccessTTLEnvironmentKey = "MS_AUTH_JWT_ACCESS_TTL_SEC"
	jwtRefreshTTLEnvironmentKey = "MS_AUTH_JWT_REFRESH_TTL_SEC"
	kafkaBrokersEnvironmentKey = "MS_AUTH_KAFKA_BROKERS"
	kafkaAuditTopicEnvironmentKey = "MS_AUTH_KAFKA_AUDIT_TOPIC"
	loginRateLimitEnvironmentKey = "MS_AUTH_LOGIN_RATE_LIMIT"
	loginRateWindowSecEnvironmentKey = "MS_AUTH_LOGIN_RATE_WINDOW_SEC"
	devModeEnvironmentKey = "MS_AUTH_DEV_MODE"
	shutdownTimeoutEnvironmentKey = "MS_AUTH_SHUTDOWN_TIMEOUT_SEC"

	defaultHTTPAddress       = ":8083"
	defaultLogLevel          = "info"
	defaultShutdownTimeoutSec = 10
	defaultJWTKeyID = "default"
	defaultJWTIssuer = "controlitix-auth"
	defaultJWTAudienceUser = "controlitix-api"
	defaultJWTAudienceService = "controlitix-internal"
	defaultJWTAccessTTLSec = 900
	defaultJWTRefreshTTLSec = 1209600
	defaultKafkaAuditTopic = "audit.logs"
	defaultLoginRateLimit = 10
	defaultLoginRateWindowSec = 60
)

type Config struct {
	HTTPAddress        string
	DatabaseURL        string
	LogLevel           string
	PasswordPepper     string
	JWTPrivateKeyPath  string
	JWTPrivateKeyPEM   string
	JWTKeyID           string
	JWTIssuer          string
	JWTAudienceUser    string
	JWTAudienceService string
	JWTAccessTTLSec    int
	JWTRefreshTTLSec   int
	KafkaBrokers       []string
	KafkaAuditTopic    string
	LoginRateLimit     int
	LoginRateWindowSec int
	DevMode            bool
	ShutdownTimeoutSec int
}

func LoadConfig() (Config, error) {
	_ = godotenv.Load()

	httpAddress := readEnvironmentValue(httpAddressEnvironmentKey, defaultHTTPAddress)
	databaseURL := strings.TrimSpace(os.Getenv(databaseURLEnvironmentKey))
	logLevel := readEnvironmentValue(logLevelEnvironmentKey, defaultLogLevel)
	passwordPepper := strings.TrimSpace(os.Getenv(passwordPepperEnvironmentKey))
	jwtPrivateKeyPath := strings.TrimSpace(os.Getenv(jwtPrivateKeyPathEnvironmentKey))
	jwtPrivateKeyPEM := strings.TrimSpace(os.Getenv(jwtPrivateKeyEnvironmentKey))
	jwtKeyID := readEnvironmentValue(jwtKeyIDEnvironmentKey, defaultJWTKeyID)
	jwtIssuer := readEnvironmentValue(jwtIssuerEnvironmentKey, defaultJWTIssuer)
	jwtAudienceUser := readEnvironmentValue(jwtAudienceUserEnvironmentKey, defaultJWTAudienceUser)
	jwtAudienceService := readEnvironmentValue(
		jwtAudienceServiceEnvironmentKey,
		defaultJWTAudienceService,
	)
	kafkaBrokers := parseEnvironmentCSV(os.Getenv(kafkaBrokersEnvironmentKey))
	kafkaAuditTopic := readEnvironmentValue(
		kafkaAuditTopicEnvironmentKey,
		defaultKafkaAuditTopic,
	)
	shutdownTimeoutSec, parseError := parseEnvironmentInt(
		shutdownTimeoutEnvironmentKey,
		defaultShutdownTimeoutSec,
	)
	if parseError != nil {
		return Config{}, parseError
	}
	loginRateLimit, parseError := parseEnvironmentInt(
		loginRateLimitEnvironmentKey,
		defaultLoginRateLimit,
	)
	if parseError != nil {
		return Config{}, parseError
	}
	loginRateWindowSec, parseError := parseEnvironmentInt(
		loginRateWindowSecEnvironmentKey,
		defaultLoginRateWindowSec,
	)
	if parseError != nil {
		return Config{}, parseError
	}
	devMode, parseError := parseEnvironmentBool(devModeEnvironmentKey, false)
	if parseError != nil {
		return Config{}, parseError
	}
	jwtAccessTTLSec, parseError := parseEnvironmentInt(
		jwtAccessTTLEnvironmentKey,
		defaultJWTAccessTTLSec,
	)
	if parseError != nil {
		return Config{}, parseError
	}
	jwtRefreshTTLSec, parseError := parseEnvironmentInt(
		jwtRefreshTTLEnvironmentKey,
		defaultJWTRefreshTTLSec,
	)
	if parseError != nil {
		return Config{}, parseError
	}

	if databaseURL == "" {
		return Config{}, errors.New("database connection is not configured")
	}
	if passwordPepper == "" {
		return Config{}, errors.New("password pepper is not configured")
	}

	decodedPepper, decodeError := base64.StdEncoding.DecodeString(passwordPepper)
	if decodeError != nil {
		return Config{}, fmt.Errorf("decode password pepper: %w", decodeError)
	}
	if len(decodedPepper) < 32 {
		return Config{}, errors.New("password pepper must be at least 32 bytes")
	}
	if jwtPrivateKeyPath == "" && jwtPrivateKeyPEM == "" {
		return Config{}, errors.New("jwt private key is not configured")
	}

	if shutdownTimeoutSec <= 0 {
		return Config{}, errors.New("shutdown timeout must be positive")
	}
	if jwtAccessTTLSec <= 0 {
		return Config{}, errors.New("jwt access ttl must be positive")
	}
	if jwtRefreshTTLSec <= 0 {
		return Config{}, errors.New("jwt refresh ttl must be positive")
	}
	if loginRateLimit <= 0 {
		return Config{}, errors.New("login rate limit must be positive")
	}
	if loginRateWindowSec <= 0 {
		return Config{}, errors.New("login rate window must be positive")
	}

	return Config{
		HTTPAddress:        httpAddress,
		DatabaseURL:        databaseURL,
		LogLevel:           logLevel,
		PasswordPepper:     passwordPepper,
		JWTPrivateKeyPath:  jwtPrivateKeyPath,
		JWTPrivateKeyPEM:   jwtPrivateKeyPEM,
		JWTKeyID:           jwtKeyID,
		JWTIssuer:          jwtIssuer,
		JWTAudienceUser:    jwtAudienceUser,
		JWTAudienceService: jwtAudienceService,
		JWTAccessTTLSec:    jwtAccessTTLSec,
		JWTRefreshTTLSec:   jwtRefreshTTLSec,
		KafkaBrokers:       kafkaBrokers,
		KafkaAuditTopic:    kafkaAuditTopic,
		LoginRateLimit:     loginRateLimit,
		LoginRateWindowSec: loginRateWindowSec,
		DevMode:            devMode,
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

func parseEnvironmentBool(environmentKey string, defaultValue bool) (bool, error) {
	value := strings.TrimSpace(os.Getenv(environmentKey))
	if value == "" {
		return defaultValue, nil
	}

	parsedValue, parseError := strconv.ParseBool(value)
	if parseError != nil {
		return false, fmt.Errorf("parse environment variable %s: %w", environmentKey, parseError)
	}

	return parsedValue, nil
}

func parseEnvironmentCSV(rawValue string) []string {
	parts := strings.Split(strings.TrimSpace(rawValue), ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		values = append(values, part)
	}

	return values
}
