package http

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

const readinessTimeout = 1 * time.Second

type HealthHandler struct {
	databaseConnection *sql.DB
}

func NewHealthHandler(databaseConnection *sql.DB) *HealthHandler {
	return &HealthHandler{
		databaseConnection: databaseConnection,
	}
}

func (handler *HealthHandler) RegisterRoutes(serveMux *http.ServeMux) {
	serveMux.HandleFunc("GET /healthz", handler.Healthz)
	serveMux.HandleFunc("GET /readyz", handler.Readyz)
}

func (handler *HealthHandler) Healthz(
	responseWriter http.ResponseWriter,
	_ *http.Request,
) {
	writeJSON(
		responseWriter,
		http.StatusOK,
		map[string]string{"status": "ok"},
	)
}

func (handler *HealthHandler) Readyz(
	responseWriter http.ResponseWriter,
	request *http.Request,
) {
	if handler.databaseConnection == nil {
		writeJSON(
			responseWriter,
			http.StatusServiceUnavailable,
			map[string]string{
				"status": "degraded",
				"reason": "database unreachable",
			},
		)
		return
	}

	readinessContext, cancelReadiness := context.WithTimeout(
		request.Context(),
		readinessTimeout,
	)
	defer cancelReadiness()

	if pingError := handler.databaseConnection.PingContext(readinessContext); pingError != nil {
		writeJSON(
			responseWriter,
			http.StatusServiceUnavailable,
			map[string]string{
				"status": "degraded",
				"reason": "database unreachable",
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
