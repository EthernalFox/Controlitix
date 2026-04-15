package http

import (
	"encoding/json"
	"net/http"

	"github.com/EthernalFox/Controlitix/ms-editor/internal/domain"
)

func writeJSON(responseWriter http.ResponseWriter, statusCode int, payload any) {
	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(statusCode)

	if payload == nil {
		return
	}

	encoder := json.NewEncoder(responseWriter)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(payload)
}

func writeError(responseWriter http.ResponseWriter, statusCode int, message string) {
	writeJSON(responseWriter, statusCode, errorResponse{Message: message})
}

func mapDomainError(domainError error) (int, string) {
	switch {
	case domainError == nil:
		return http.StatusOK, ""
	case domainError == domain.ErrInvalidInput:
		return http.StatusBadRequest, "invalid request"
	case domainError == domain.ErrNotFound:
		return http.StatusNotFound, "not found"
	case domainError == domain.ErrConflict:
		return http.StatusConflict, "conflict"
	case domainError == domain.ErrNotImplemented:
		return http.StatusNotImplemented, "not implemented"
	default:
		return http.StatusInternalServerError, "internal error"
	}
}
