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

	transporthttp "github.com/EthernalFox/Controlitix/ms-viewer/internal/api/http"
	"github.com/EthernalFox/Controlitix/ms-viewer/internal/config"
	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
	"github.com/EthernalFox/Controlitix/ms-viewer/internal/infrastructure/cache"
	"github.com/EthernalFox/Controlitix/ms-viewer/internal/infrastructure/database"
	kafkainfra "github.com/EthernalFox/Controlitix/ms-viewer/internal/infrastructure/kafka"
	"github.com/EthernalFox/Controlitix/ms-viewer/internal/infrastructure/repository"
	telegraminfra "github.com/EthernalFox/Controlitix/ms-viewer/internal/infrastructure/telegram"
	"github.com/EthernalFox/Controlitix/ms-viewer/internal/realtime"
	"github.com/EthernalFox/Controlitix/ms-viewer/internal/usecase"
)

type alarmBroadcasterRef struct {
	hub *realtime.Hub
}

func (reference *alarmBroadcasterRef) BroadcastAlarm(event domain.AlarmEvent) {
	if reference == nil || reference.hub == nil {
		return
	}

	reference.hub.BroadcastAlarm(event)
}

func (reference *alarmBroadcasterRef) BroadcastAlarmBatch(events []domain.AlarmEvent) {
	if reference == nil || reference.hub == nil {
		return
	}

	reference.hub.BroadcastAlarmBatch(events)
}

type alarmSnapshotProviderRef struct {
	alarmUseCase *usecase.AlarmUseCase
}

func (reference *alarmSnapshotProviderRef) GetActiveAlarmSnapshot(
	ctx context.Context,
) ([]domain.AlarmStateRecord, error) {
	if reference == nil || reference.alarmUseCase == nil {
		return nil, nil
	}

	return reference.alarmUseCase.GetActiveAlarmSnapshot(ctx)
}

func main() {
	applicationConfig, configError := config.LoadConfig()
	if configError != nil {
		slog.Error("failed to load config", "error", configError)
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: parseLogLevel(applicationConfig.LogLevel),
	}))
	slog.SetDefault(logger)

	applicationContext, cancelApplication := context.WithCancel(context.Background())
	defer cancelApplication()

	postgresPool, closeDatabase, databaseError := database.OpenPostgresPool(
		applicationContext,
		applicationConfig.DatabaseURL,
	)
	if databaseError != nil {
		logger.Error("failed to connect to postgres", "error", databaseError)
		os.Exit(1)
	}
	defer closeDatabase()

	redisCache, cacheError := cache.NewRedisCache(
		applicationConfig.RedisURL,
		time.Duration(applicationConfig.LastValueCacheTTLSec)*time.Second,
	)
	if cacheError != nil {
		logger.Error("failed to initialize redis", "error", cacheError)
		os.Exit(1)
	}
	defer func() {
		if closeError := redisCache.Close(); closeError != nil {
			logger.Error("failed to close redis", "error", closeError)
		}
	}()

	authMiddleware, authValidator, stopJWKSRefresh, authError := transporthttp.BuildAuthMiddleware(
		applicationContext,
		applicationConfig.JWKSURL,
		applicationConfig.JWTIssuer,
		applicationConfig.JWTAudienceUser,
		logger,
	)
	if authError != nil {
		logger.Error("failed to initialize auth middleware", "error", authError)
		os.Exit(1)
	}
	defer stopJWKSRefresh()

	trendRepository := repository.NewTrendRepository(postgresPool)
	tagRepository := repository.NewTagRepository(postgresPool)
	ingestRepository := repository.NewIngestRepository(postgresPool)
	diagramRepository := repository.NewDiagramRepository(postgresPool)
	wsTopicRepository := repository.NewWSTopicRepository(postgresPool)
	alarmRepository := repository.NewAlarmRepository(postgresPool)
	telegramChatsRepository := repository.NewTelegramChatsRepository(postgresPool)
	notificationRepository := repository.NewNotificationRepository(postgresPool)
	alarmsProducer := kafkainfra.NewAlarmsProducer(
		applicationConfig.KafkaBrokers,
		applicationConfig.KafkaAlarmsTopic,
		logger,
	)
	defer func() {
		if closeError := alarmsProducer.Close(); closeError != nil {
			logger.Error("failed to close alarms producer", "error", closeError)
		}
	}()
	auditProducer := kafkainfra.NewAuditProducer(
		applicationConfig.KafkaBrokers,
		applicationConfig.KafkaAuditTopic,
		applicationConfig.AuditBufferSize,
		logger,
	)
	defer func() {
		if closeError := auditProducer.Close(); closeError != nil {
			logger.Error("failed to close audit producer", "error", closeError)
		}
	}()
	auditUseCase := usecase.NewAuditUseCase(auditProducer, logger)
	tagMetaCache := cache.NewInMemoryTagMetaCache(
		applicationConfig.TagMetaCacheSize,
		time.Duration(applicationConfig.TagMetaCacheTTLSec)*time.Second,
	)
	alarmBroadcaster := &alarmBroadcasterRef{}
	alarmSnapshotProvider := &alarmSnapshotProviderRef{}
	alarmUseCase := usecase.NewAlarmUseCase(
		alarmRepository,
		alarmsProducer,
		alarmBroadcaster,
		auditUseCase,
		applicationConfig.AlarmHysteresisPercent,
		time.Duration(applicationConfig.AlarmCommLossTimeoutSec)*time.Second,
		logger,
	)
	if initializeError := alarmUseCase.Initialize(applicationContext); initializeError != nil {
		logger.Error("failed to initialize alarm use case", "error", initializeError)
		os.Exit(1)
	}
	alarmSnapshotProvider.alarmUseCase = alarmUseCase
	debouncer := realtime.NewDebouncer(
		time.Duration(applicationConfig.WSDebounceMs)*time.Millisecond,
		30*time.Second,
		logger,
	)
	membershipCache := realtime.NewMembershipCache(600 * time.Second)
	topicResolver := realtime.NewTopicResolver(
		wsTopicRepository,
		membershipCache,
		applicationConfig.WSGroupResolveLimit,
		logger,
	)
	hub := realtime.NewHub(realtime.HubOptions{
		Debouncer:               debouncer,
		LastValueProvider:       redisCache,
		MaxSubscriptionsPerConn: applicationConfig.WSMaxSubscriptionsPerConn,
		MaxConnectionsPerUser:   applicationConfig.WSMaxConnectionsPerUser,
		AlarmSnapshotProvider:   alarmSnapshotProvider,
		TopicResolver:           topicResolver,
		Logger:                  logger,
	})
	alarmBroadcaster.hub = hub
	dispatcher := realtime.NewDispatcher(hub)
	wsHandler := transporthttp.NewWSHandler(transporthttp.WSHandlerOptions{
		Hub:                hub,
		Validator:          authValidator,
		AuditUseCase:       auditUseCase,
		Logger:             logger,
		WSPingIntervalSec:  applicationConfig.WSPingIntervalSec,
		WSPongTimeoutSec:   applicationConfig.WSPongTimeoutSec,
		WSWriteBufferSize:  applicationConfig.WSWriteBufferSize,
		WSMaxSubscriptions: applicationConfig.WSMaxSubscriptionsPerConn,
		WSDebounceMS:       applicationConfig.WSDebounceMs,
	})

	trendUseCase := usecase.NewTrendUseCase(
		trendRepository,
		tagMetaCache,
		applicationConfig.TrendsMaxLimit,
		applicationConfig.TrendsMaxRangeDays,
		logger,
	)
	tagUseCase := usecase.NewTagUseCase(tagRepository)
	diagramUseCase := usecase.NewDiagramUseCase(
		diagramRepository,
		trendRepository,
		tagMetaCache,
		redisCache,
		logger,
	)
	configChangeUseCase := usecase.NewConfigChangeUseCase(tagMetaCache, hub, alarmUseCase, logger)
	ingestUseCase := usecase.NewIngestUseCase(
		ingestRepository,
		redisCache,
		dispatcher,
		alarmUseCase,
		applicationConfig.IngestBatchSize,
		logger,
	)
	notifierDryRun := strings.TrimSpace(applicationConfig.TelegramBotToken) == ""
	var notifierClient domain.TelegramNotifierClient
	if notifierDryRun {
		notifierClient = telegraminfra.NewDryRunClient(logger)
		logger.Warn(
			"telegram notifier is running in dry-run mode",
			"method",
			"main",
			"notifier",
			"dry-run",
			"telegram_bot_token",
			"unset",
		)
	} else {
		notifierClient = telegraminfra.NewClient(
			applicationConfig.TelegramAPIURL,
			applicationConfig.TelegramBotToken,
			time.Duration(applicationConfig.TelegramRequestTimeoutSec)*time.Second,
			logger,
		)
		logger.Info(
			"telegram notifier is enabled",
			"method",
			"main",
			"notifier",
			"enabled",
			"telegram_bot_token",
			"set",
		)
	}
	notifierUseCase := usecase.NewNotifierUseCase(
		telegramChatsRepository,
		notificationRepository,
		notifierClient,
		telegraminfra.NewRenderer(),
		auditUseCase,
		usecase.NotifierUseCaseOptions{
			MaxAttempts:     applicationConfig.NotifierMaxAttempts,
			BackoffBase:     time.Duration(applicationConfig.NotifierBackoffSec) * time.Second,
			EscalationDelay: time.Duration(applicationConfig.NotifierEscalationDelaySec) * time.Second,
			BatchLimit:      100,
			PollInterval:    time.Duration(applicationConfig.NotifierBatchPollSec) * time.Second,
			DryRun:          notifierDryRun,
		},
		logger,
	)

	consumer, consumerError := kafkainfra.NewConsumer(
		applicationConfig.KafkaBrokers,
		applicationConfig.KafkaTagsValuesTopic,
		applicationConfig.KafkaConsumerGroup,
		logger,
	)
	if consumerError != nil {
		logger.Error("failed to initialize kafka consumer", "error", consumerError)
		os.Exit(1)
	}
	defer func() {
		if closeError := consumer.Close(); closeError != nil {
			logger.Error("failed to close kafka consumer", "error", closeError)
		}
	}()

	configChangedConsumer, configChangedConsumerError := kafkainfra.NewConsumer(
		applicationConfig.KafkaBrokers,
		applicationConfig.KafkaConfigChangedTopic,
		applicationConfig.KafkaConfigChangedGroup,
		logger,
	)
	if configChangedConsumerError != nil {
		logger.Error("failed to initialize config.changed consumer", "error", configChangedConsumerError)
		os.Exit(1)
	}
	defer func() {
		if closeError := configChangedConsumer.Close(); closeError != nil {
			logger.Error("failed to close config.changed consumer", "error", closeError)
		}
	}()
	alarmsConsumer, alarmsConsumerError := kafkainfra.NewConsumer(
		applicationConfig.KafkaBrokers,
		applicationConfig.KafkaAlarmsTopic,
		applicationConfig.KafkaAlarmsGroup,
		logger,
	)
	if alarmsConsumerError != nil {
		logger.Error("failed to initialize alarms.events consumer", "error", alarmsConsumerError)
		os.Exit(1)
	}
	defer func() {
		if closeError := alarmsConsumer.Close(); closeError != nil {
			logger.Error("failed to close alarms.events consumer", "error", closeError)
		}
	}()

	tagValueHandler := kafkainfra.NewTagValueHandler(
		consumer,
		ingestUseCase,
		usecase.IngestFlushInterval(applicationConfig.IngestFlushMs),
		logger,
	)
	configChangedHandler := kafkainfra.NewConfigChangedConsumer(
		configChangedConsumer,
		configChangeUseCase,
		logger,
	)
	alarmsHandler := kafkainfra.NewAlarmsConsumer(
		alarmsConsumer,
		notifierUseCase,
		logger,
	)

	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-applicationContext.Done():
				return
			case <-ticker.C:
				stats := tagMetaCache.Stats()
				logger.Debug(
					"tag meta cache stats",
					"method",
					"main.cacheStats",
					"hits",
					stats.Hits,
					"misses",
					stats.Misses,
					"size",
					stats.Size,
				)
			}
		}
	}()

	router := transporthttp.NewRouter(transporthttp.RouterOptions{
		TrendUseCase:    trendUseCase,
		TagUseCase:      tagUseCase,
		DiagramUseCase:  diagramUseCase,
		AlarmUseCase:    alarmUseCase,
		NotifierUseCase: notifierUseCase,
		AuditUseCase:    auditUseCase,
		AuthMiddleware:  authMiddleware,
		WSHandler:       wsHandler,
		DatabaseChecker: postgresPool,
		RedisChecker:    redisCache,
		Logger:          logger,
	})

	httpServer := transporthttp.NewServer(applicationConfig.HTTPAddress, router)

	runtimeErrors := make(chan error, 6)
	go func() {
		logger.Info("http server started", "address", applicationConfig.HTTPAddress)
		serverError := httpServer.Start()
		if serverError != nil && !errors.Is(serverError, http.ErrServerClosed) {
			runtimeErrors <- fmt.Errorf("start http server: %w", serverError)
		}
	}()

	go func() {
		handlerError := tagValueHandler.Run(applicationContext)
		if handlerError != nil && !errors.Is(handlerError, context.Canceled) {
			runtimeErrors <- fmt.Errorf("run tag value handler: %w", handlerError)
		}
	}()

	go func() {
		handlerError := configChangedHandler.Run(applicationContext)
		if handlerError != nil && !errors.Is(handlerError, context.Canceled) {
			runtimeErrors <- fmt.Errorf("run config changed handler: %w", handlerError)
		}
	}()
	go func() {
		handlerError := alarmsHandler.Run(applicationContext)
		if handlerError != nil && !errors.Is(handlerError, context.Canceled) {
			runtimeErrors <- fmt.Errorf("run alarms consumer handler: %w", handlerError)
		}
	}()

	go func() {
		scannerError := alarmUseCase.RunCommLossScanner(applicationContext, 30*time.Second)
		if scannerError != nil && !errors.Is(scannerError, context.Canceled) {
			runtimeErrors <- fmt.Errorf("run comm loss scanner: %w", scannerError)
		}
	}()
	go func() {
		schedulerError := notifierUseCase.RunScheduler(applicationContext)
		if schedulerError != nil && !errors.Is(schedulerError, context.Canceled) {
			runtimeErrors <- fmt.Errorf("run notifier scheduler: %w", schedulerError)
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

	logger.Info("http server shutting down")
	if shutdownError := httpServer.Shutdown(shutdownContext); shutdownError != nil {
		logger.Error("http server shutdown failed", "error", shutdownError)
	}

	logger.Info("ms-viewer stopped")
}

func parseLogLevel(logLevel string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(logLevel)) {
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
