package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
	//https://github.com/ilyakaznacheev/cleanenv - менеджер энв
)

const (
	httpAddressEnvironmentKey = "MS_EDITOR_HTTP_ADDRESS"
	databaseURLEnvironmentKey = "MS_EDITOR_DATABASE_URL"
	logLevelEnvironmentKey    = "MS_EDITOR_LOG_LEVEL"
	defaultHTTPAddress        = ":8080"
	defaultLogLevel           = "info"
)

type Config struct {
	HTTPAddress string
	DatabaseURL string
	LogLevel    string
}

func LoadConfig() (Config, error) {
	_ = godotenv.Load()
	

	httpAddress := readEnvironmentValue(httpAddressEnvironmentKey, defaultHTTPAddress)
	databaseURL := readEnvironmentValue(databaseURLEnvironmentKey, "")
	logLevel := readEnvironmentValue(logLevelEnvironmentKey, defaultLogLevel)

	if databaseURL == "" {
		return Config{}, errors.New("Подключение не сконфигурировано")
	}

	return Config{
		HTTPAddress: httpAddress,
		DatabaseURL: databaseURL,
		LogLevel:    logLevel,
	}, nil
}

func readEnvironmentValue(environmentKey string, defaultValue string) string {
	value := os.Getenv(environmentKey)
	if value == "" {
		return defaultValue
	}

	return value
}
