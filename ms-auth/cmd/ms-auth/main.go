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

	transporthttp "github.com/EthernalFox/Controlitix/ms-auth/internal/api/http"
	"github.com/EthernalFox/Controlitix/ms-auth/internal/config"
	"github.com/EthernalFox/Controlitix/ms-auth/internal/infrastructure/database"
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

	healthHandler := transporthttp.NewHealthHandler(databaseConnection)
	httpServeMux := http.NewServeMux()
	healthHandler.RegisterRoutes(httpServeMux)

	httpServer := &http.Server{
		Addr:    applicationConfig.HTTPAddress,
		Handler: httpServeMux,
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
