package main

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	transporthttp "github.com/EthernalFox/Controlitix/ms-auth/internal/api/http"
	"github.com/EthernalFox/Controlitix/ms-auth/internal/config"
	"github.com/EthernalFox/Controlitix/ms-auth/internal/infrastructure/database"
	"github.com/EthernalFox/Controlitix/ms-auth/internal/infrastructure/identity"
	"github.com/EthernalFox/Controlitix/ms-auth/internal/infrastructure/jwks"
	kafkainfra "github.com/EthernalFox/Controlitix/ms-auth/internal/infrastructure/kafka"
	"github.com/EthernalFox/Controlitix/ms-auth/internal/infrastructure/password"
	"github.com/EthernalFox/Controlitix/ms-auth/internal/infrastructure/repository"
	"github.com/EthernalFox/Controlitix/ms-auth/internal/usecase"
	"github.com/EthernalFox/Controlitix/ms-auth/internal/usecase/ratelimit"
	"github.com/go-chi/chi/v5"
)

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
	logger.Info("starting ms-auth")

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
	logger.Info("database connected")

	dependencies, dependenciesError := buildDependencies(
		databaseConnection,
		applicationConfig,
		logger,
	)
	if dependenciesError != nil {
		logger.Error("failed to build dependencies", "error", dependenciesError)
		os.Exit(1)
	}
	defer func() {
		if closeError := dependencies.Close(); closeError != nil {
			logger.Error("failed to close dependencies", "error", closeError)
		}
	}()

	healthHandler := transporthttp.NewHealthHandler(databaseConnection)
	router := chi.NewRouter()
	router.Use(transporthttp.RequestID, transporthttp.AccessLog(logger))
	router.Get("/healthz", healthHandler.Healthz)
	router.Get("/readyz", healthHandler.Readyz)
	router.Get("/.well-known/jwks.json", jwks.NewHandler(dependencies.keystore).ServeHTTP)

	authHandler := transporthttp.NewAuthHandler(
		dependencies.authService,
		dependencies.tokenService,
		dependencies.userRepository,
		dependencies.sourceRepository,
		dependencies.auditService,
		dependencies.limiter,
		dependencies.keystore,
		transporthttp.AuthHandlerConfig{
			RefreshCookieName:   "refresh_token",
			RefreshCookiePath:   "/api/auth",
			RefreshCookieSecure: !applicationConfig.DevMode,
			RefreshCookieTTL: time.Duration(
				applicationConfig.JWTRefreshTTLSec,
			) * time.Second,
			DevMode:         applicationConfig.DevMode,
			JWTIssuer:       applicationConfig.JWTIssuer,
			JWTAudienceUser: applicationConfig.JWTAudienceUser,
		},
		logger,
	)
	serviceTokenHandler := transporthttp.NewServiceTokenHandler(
		dependencies.serviceTokenService,
		logger,
	)
	adminHandler := transporthttp.NewAdminHandler(
		dependencies.adminService,
		dependencies.userRepository,
		dependencies.sourceRepository,
		dependencies.serviceAccountRepository,
		dependencies.auditRepository,
		logger,
	)
	router.Route("/api/auth", func(authRouter chi.Router) {
		authHandler.Register(authRouter)
		serviceTokenHandler.Register(authRouter)
		authRouter.With(
			transporthttp.RequireAuthWithIssuer(
				dependencies.keystore,
				applicationConfig.JWTIssuer,
				applicationConfig.JWTAudienceUser,
			),
			transporthttp.RequireRole("admin"),
		).Group(func(adminRouter chi.Router) {
			adminHandler.Register(adminRouter)
		})
	})

	httpServer := &http.Server{
		Addr:    applicationConfig.HTTPAddress,
		Handler: router,
	}

	runtimeErrors := make(chan error, 1)
	go func() {
		logger.Info("http server started", "address", applicationConfig.HTTPAddress)
		serverError := httpServer.ListenAndServe()
		if serverError != nil && !errors.Is(serverError, http.ErrServerClosed) {
			runtimeErrors <- fmt.Errorf("run http server: %w", serverError)
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

	if closeDatabaseError := closeDatabase(); closeDatabaseError != nil {
		logger.Error("failed to close database", "error", closeDatabaseError)
	} else {
		logger.Info("database closed")
	}

	logger.Info("ms-auth stopped")
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

type dependencies struct {
	authService      *usecase.AuthService
	tokenService     *usecase.TokenService
	serviceTokenService *usecase.ServiceTokenService
	adminService     *usecase.AdminService
	auditService     *usecase.AuditService
	auditRepository  *repository.AuditRepository
	userRepository   *repository.UserRepository
	sourceRepository *repository.IdentitySourceRepository
	serviceAccountRepository *repository.ServiceAccountRepository
	limiter          *ratelimit.Limiter
	kafkaProducer    *kafkainfra.AuditProducer
	keystore         *jwks.Keystore
}

func (deps *dependencies) Close() error {
	if deps == nil {
		return nil
	}

	if deps.limiter != nil {
		deps.limiter.Close()
	}
	if deps.kafkaProducer != nil {
		if err := deps.kafkaProducer.Close(); err != nil {
			return fmt.Errorf("close audit kafka producer: %w", err)
		}
	}

	return nil
}

func buildDependencies(
	databaseConnection *sql.DB,
	applicationConfig config.Config,
	logger *slog.Logger,
) (*dependencies, error) {
	pepper, decodeError := base64.StdEncoding.DecodeString(applicationConfig.PasswordPepper)
	if decodeError != nil {
		return nil, fmt.Errorf("decode password pepper: %w", decodeError)
	}
	if len(pepper) < 32 {
		return nil, errors.New("password pepper must be at least 32 bytes")
	}

	hasher, hasherError := password.NewHasher(pepper)
	if hasherError != nil {
		return nil, fmt.Errorf("initialize password hasher: %w", hasherError)
	}

	userRepository := repository.NewUserRepository(databaseConnection)
	sourceRepository := repository.NewIdentitySourceRepository(databaseConnection)
	localProvider := identity.NewLocalProvider(userRepository, hasher, logger)
	registry := identity.NewRegistry(localProvider)
	authService := usecase.NewAuthService(sourceRepository, userRepository, registry, logger)
	keystore, keystoreError := jwks.LoadFromConfig(applicationConfig)
	if keystoreError != nil {
		return nil, fmt.Errorf("load jwt keystore: %w", keystoreError)
	}

	refreshRepository := repository.NewRefreshTokenRepository(databaseConnection)
	userLookup := repository.NewUserLookupAdapter(userRepository, sourceRepository)
	tokenService := usecase.NewTokenService(
		keystore,
		refreshRepository,
		userLookup,
		usecase.TokenConfig{
			Issuer:          applicationConfig.JWTIssuer,
			AudienceUser:    applicationConfig.JWTAudienceUser,
			AudienceService: applicationConfig.JWTAudienceService,
			AccessTTL:       time.Duration(applicationConfig.JWTAccessTTLSec) * time.Second,
			RefreshTTL:      time.Duration(applicationConfig.JWTRefreshTTLSec) * time.Second,
		},
		logger,
	)
	auditRepository := repository.NewAuditRepository(databaseConnection)
	auditProducer, producerError := kafkainfra.NewAuditProducer(
		applicationConfig.KafkaBrokers,
		applicationConfig.KafkaAuditTopic,
		logger,
	)
	if producerError != nil {
		return nil, fmt.Errorf("initialize kafka audit producer: %w", producerError)
	}
	if len(applicationConfig.KafkaBrokers) == 0 {
		logger.Warn("kafka audit producer disabled: brokers not configured")
	}
	auditService := usecase.NewAuditService(auditRepository, auditProducer, logger)
	serviceAccountRepository := repository.NewServiceAccountRepository(databaseConnection)
	serviceTokenService := usecase.NewServiceTokenService(
		serviceAccountRepository,
		tokenService,
		hasher,
		auditService,
		logger,
	)
	adminService := usecase.NewAdminService(
		userRepository,
		serviceAccountRepository,
		sourceRepository,
		auditService,
		refreshRepository,
		hasher,
		logger,
	)
	loginLimiter := ratelimit.New(
		applicationConfig.LoginRateLimit,
		time.Duration(applicationConfig.LoginRateWindowSec)*time.Second,
		time.Now,
	)

	logger.Info(
		"jwt keystore loaded",
		"kid",
		keystore.KeyID(),
		"key_size",
		keystore.PrivateKey().N.BitLen(),
	)

	return &dependencies{
		authService:      authService,
		tokenService:     tokenService,
		serviceTokenService: serviceTokenService,
		adminService:     adminService,
		auditService:     auditService,
		auditRepository:  auditRepository,
		userRepository:   userRepository,
		sourceRepository: sourceRepository,
		serviceAccountRepository: serviceAccountRepository,
		limiter:          loginLimiter,
		kafkaProducer:    auditProducer,
		keystore:         keystore,
	}, nil
}
