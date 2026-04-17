package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	transporthttp "github.com/EthernalFox/Controlitix/ms-editor/internal/api/http"
	"github.com/EthernalFox/Controlitix/ms-editor/internal/config"
	"github.com/EthernalFox/Controlitix/ms-editor/internal/infrastructure/database"
	eventkafka "github.com/EthernalFox/Controlitix/ms-editor/internal/infrastructure/kafka"
	"github.com/EthernalFox/Controlitix/ms-editor/internal/infrastructure/repository"
	"github.com/EthernalFox/Controlitix/ms-editor/internal/usecase"
)

const shutdownTimeout = 10 * time.Second

func main() {
	applicationConfig, configError := config.LoadConfig()
	if configError != nil {
		slog.Error("failed to load config", "error", configError)
		os.Exit(1)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: parseLogLevel(applicationConfig.LogLevel),
	}))
	slog.SetDefault(logger)

	applicationContext := context.Background()
	databaseConnection, closeDatabase, databaseError := database.OpenPostgres(
		applicationContext,
		applicationConfig.DatabaseURL,
	)
	if databaseError != nil {
		logger.Error("failed to connect to database", "error", databaseError)
		os.Exit(1)
	}
	defer func() {
		if closeError := closeDatabase(); closeError != nil {
			logger.Error("failed to close database", "error", closeError)
		}
	}()

	postgresRepository := repository.NewPostgresRepository(databaseConnection)
	kafkaEventPublisher := eventkafka.NewKafkaEventPublisher(
		applicationConfig.KafkaBrokers,
		applicationConfig.KafkaTopic,
		logger,
	)
	defer func() {
		if closeError := kafkaEventPublisher.Close(); closeError != nil {
			logger.Error("failed to close kafka publisher", "error", closeError)
		}
	}()

	monitoringObjectUseCase := usecase.NewMonitoringObjectUseCase(postgresRepository)
	diagramUseCase := usecase.NewDiagramUseCase(postgresRepository)
	figureUseCase := usecase.NewFigureUseCase(postgresRepository)
	deviceUseCase := usecase.NewDeviceUseCase(
		postgresRepository,
		postgresRepository,
		postgresRepository,
		kafkaEventPublisher,
		logger,
	)

	httpServeMux := http.NewServeMux()
	httpHandler := transporthttp.NewHandler(
		monitoringObjectUseCase,
		diagramUseCase,
		figureUseCase,
		deviceUseCase,
	)
	httpHandler.RegisterRoutes(httpServeMux)

	httpServer := transporthttp.NewServer(
		applicationConfig.HTTPAddress,
		httpServeMux,
	)

	shutdownContext, stopShutdown := signal.NotifyContext(
		applicationContext,
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stopShutdown()

	go func() {
		logger.Info("http server started", "address", applicationConfig.HTTPAddress)
		serverError := httpServer.Start()
		if serverError != nil && !errors.Is(serverError, http.ErrServerClosed) {
			logger.Error("http server failed", "error", serverError)
			stopShutdown()
		}
	}()

	<-shutdownContext.Done()

	logger.Info("http server shutting down")
	shutdownTimeoutContext, cancelShutdown := context.WithTimeout(
		context.Background(),
		shutdownTimeout,
	)
	defer cancelShutdown()

	if shutdownError := httpServer.Shutdown(shutdownTimeoutContext); shutdownError != nil {
		logger.Error("http server shutdown failed", "error", shutdownError)
	}
}

func parseLogLevel(logLevel string) slog.Level {
	switch strings.ToLower(logLevel) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
