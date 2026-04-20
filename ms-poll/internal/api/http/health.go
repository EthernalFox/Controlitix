package http

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/EthernalFox/Controlitix/ms-poll/internal/usecase"
)

const readinessTimeout = 1 * time.Second

type HealthHandler struct {
	databaseConnection *sql.DB
	kafkaBrokers       []string
	snapshotStore      *usecase.SnapshotStore
}

func NewHealthHandler(
	databaseConnection *sql.DB,
	kafkaBrokers []string,
	snapshotStore *usecase.SnapshotStore,
) *HealthHandler {
	return &HealthHandler{
		databaseConnection: databaseConnection,
		kafkaBrokers:       kafkaBrokers,
		snapshotStore:      snapshotStore,
	}
}

func (handler *HealthHandler) RegisterRoutes(serveMux *http.ServeMux) {
	serveMux.HandleFunc("GET /healthz", handler.handleHealthz)
	serveMux.HandleFunc("GET /readyz", handler.handleReadyz)
}

func (handler *HealthHandler) handleHealthz(
	responseWriter http.ResponseWriter,
	_ *http.Request,
) {
	writeJSON(
		responseWriter,
		http.StatusOK,
		map[string]string{"status": "ok"},
	)
}

func (handler *HealthHandler) handleReadyz(
	responseWriter http.ResponseWriter,
	request *http.Request,
) {
	if len(handler.kafkaBrokers) == 0 {
		writeJSON(
			responseWriter,
			http.StatusServiceUnavailable,
			map[string]string{
				"status": "degraded",
				"reason": "kafka brokers are not configured",
			},
		)
		return
	}

	if handler.databaseConnection == nil {
		writeJSON(
			responseWriter,
			http.StatusServiceUnavailable,
			map[string]string{
				"status": "degraded",
				"reason": "database is not configured",
			},
		)
		return
	}

	readinessContext, cancelReadiness := context.WithTimeout(
		request.Context(),
		readinessTimeout,
	)
	defer cancelReadiness()

	pingError := handler.databaseConnection.PingContext(readinessContext)
	if pingError != nil {
		writeJSON(
			responseWriter,
			http.StatusServiceUnavailable,
			map[string]string{
				"status": "degraded",
				"reason": "database ping failed",
			},
		)
		return
	}

	if handler.snapshotStore == nil {
		writeJSON(
			responseWriter,
			http.StatusServiceUnavailable,
			map[string]string{
				"status": "degraded",
				"reason": "snapshot store is not configured",
			},
		)
		return
	}

	snapshot := handler.snapshotStore.Get()
	if snapshot == nil || snapshot.LoadedAt.IsZero() {
		writeJSON(
			responseWriter,
			http.StatusServiceUnavailable,
			map[string]string{
				"status": "degraded",
				"reason": "snapshot not loaded",
			},
		)
		return
	}

	writeJSON(
		responseWriter,
		http.StatusOK,
		map[string]string{"status": "ok"},
	)
}

func writeJSON(
	responseWriter http.ResponseWriter,
	statusCode int,
	payload any,
) {
	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(statusCode)
	_ = json.NewEncoder(responseWriter).Encode(payload)
}
