package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	transporthttp "github.com/EthernalFox/Controlitix/ms-poll/internal/api/http"
	"github.com/EthernalFox/Controlitix/ms-poll/internal/config"
	"github.com/EthernalFox/Controlitix/ms-poll/internal/infrastructure/database"
	modbusdriver "github.com/EthernalFox/Controlitix/ms-poll/internal/infrastructure/driver/modbus"
	snmpdriver "github.com/EthernalFox/Controlitix/ms-poll/internal/infrastructure/driver/snmp"
	eventkafka "github.com/EthernalFox/Controlitix/ms-poll/internal/infrastructure/kafka"
	"github.com/EthernalFox/Controlitix/ms-poll/internal/infrastructure/repository"
	"github.com/EthernalFox/Controlitix/ms-poll/internal/usecase"
	"github.com/gosnmp/gosnmp"
)

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

	applicationContext, cancelApplication := context.WithCancel(context.Background())
	defer cancelApplication()

	databaseConnection, closeDatabase, databaseError := database.OpenPostgres(
		applicationContext,
		applicationConfig.DatabaseURL,
	)
	if databaseError != nil {
		logger.Error("failed to connect to database", "error", databaseError)
		os.Exit(1)
	}

	configRepository := repository.NewConfigRepository(databaseConnection)
	snapshotStore := usecase.NewSnapshotStore()
	configLoader := usecase.NewConfigLoader(configRepository, snapshotStore, logger)
	if loadError := configLoader.Load(applicationContext); loadError != nil {
		logger.Error("failed to load config snapshot", "error", loadError)
		if closeError := closeDatabase(); closeError != nil {
			logger.Error("failed to close database", "error", closeError)
		}
		os.Exit(1)
	}

	configEventHandler := usecase.NewConfigEventHandler(
		configRepository,
		snapshotStore,
		logger,
	)
	driverRegistry := usecase.NewDriverRegistry()
	serialPool := modbusdriver.NewSerialPortPool()
	driverRegistry.Register(modbusdriver.NewTCPDriverFactory())
	driverRegistry.Register(modbusdriver.NewRTUDriverFactory(serialPool))
	driverRegistry.Register(snmpdriver.NewVersionedFactory(gosnmp.Version1))
	driverRegistry.Register(snmpdriver.NewVersionedFactory(gosnmp.Version2c))
	driverRegistry.Register(snmpdriver.NewVersionedFactory(gosnmp.Version3))

	configChangedConsumer := eventkafka.NewConfigChangedConsumer(
		applicationConfig.KafkaBrokers,
		applicationConfig.KafkaConfigTopic,
		applicationConfig.KafkaConfigGroupID,
		logger,
	)
	tagValuesProducer := eventkafka.NewTagValuesProducer(
		applicationConfig.KafkaBrokers,
		applicationConfig.KafkaTagValuesTopic,
		logger,
	)
	var readingsPublisher usecase.TagValuesPublisher = tagValuesProducer
	rawReadingsPublisher := (*eventkafka.RawReadingsPublisher)(nil)
	if applicationConfig.PublishRawReadings {
		rawReadingsPublisher = eventkafka.NewRawReadingsPublisher(
			applicationConfig.KafkaBrokers,
			applicationConfig.KafkaRawReadingsTopic,
			logger,
		)
		readingsPublisher = eventkafka.NewCompositePublisher(
			tagValuesProducer,
			rawReadingsPublisher,
			logger,
		)
	}

	normalizingPublisher := usecase.NewNormalizingPublisher(
		readingsPublisher,
		snapshotStore,
		logger,
	)
	scheduler := usecase.NewScheduler(
		snapshotStore,
		driverRegistry,
		normalizingPublisher,
		logger,
	)

	healthHandler := transporthttp.NewHealthHandler(
		databaseConnection,
		applicationConfig.KafkaBrokers,
		snapshotStore,
	)

	httpServeMux := http.NewServeMux()
	healthHandler.RegisterRoutes(httpServeMux)

	httpServer := &http.Server{
		Addr:    applicationConfig.HTTPAddress,
		Handler: httpServeMux,
	}

	runtimeErrors := make(chan error, 3)

	go func() {
		logger.Info("http server started", "address", applicationConfig.HTTPAddress)
		serverError := httpServer.ListenAndServe()
		if serverError != nil && !errors.Is(serverError, http.ErrServerClosed) {
			runtimeErrors <- fmt.Errorf("run http server: %w", serverError)
		}
	}()

	go func() {
		consumerError := configChangedConsumer.Run(
			applicationContext,
			configEventHandler.Handle,
		)
		if consumerError != nil {
			runtimeErrors <- fmt.Errorf("run kafka consumer: %w", consumerError)
		}
	}()

	go func() {
		schedulerError := scheduler.Run(applicationContext)
		if schedulerError != nil &&
			!errors.Is(schedulerError, context.Canceled) {
			runtimeErrors <- fmt.Errorf("run scheduler: %w", schedulerError)
		}
	}()

	signalContext, stopSignals := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stopSignals()

	select {
	case <-signalContext.Done():
		logger.Info("shutdown signal received")
	case runtimeError := <-runtimeErrors:
		logger.Error("runtime component failed", "error", runtimeError)
	}

	cancelApplication()

	shutdownContext, cancelShutdown := context.WithTimeout(
		context.Background(),
		time.Duration(applicationConfig.ShutdownTimeoutSec)*time.Second,
	)
	defer cancelShutdown()

	if closeConsumerError := configChangedConsumer.Close(); closeConsumerError != nil {
		logger.Error("failed to close kafka consumer", "error", closeConsumerError)
	}

	logger.Info("http server shutting down")
	if shutdownError := httpServer.Shutdown(shutdownContext); shutdownError != nil {
		logger.Error("http server shutdown failed", "error", shutdownError)
	}

	if closeProducerError := tagValuesProducer.Close(); closeProducerError != nil {
		logger.Error("failed to close kafka producer", "error", closeProducerError)
	} else {
		logger.Info("kafka producer closed")
	}

	if rawReadingsPublisher != nil {
		if closeRawProducerError := rawReadingsPublisher.Close(); closeRawProducerError != nil {
			logger.Error(
				"failed to close kafka raw readings producer",
				"error",
				closeRawProducerError,
			)
		} else {
			logger.Info("kafka raw readings producer closed")
		}
	}

	if closeDatabaseError := closeDatabase(); closeDatabaseError != nil {
		logger.Error("failed to close database", "error", closeDatabaseError)
	} else {
		logger.Info("database closed")
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
