package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

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
	problemType, title := problemTypeAndTitle(statusCode)
	writeProblem(responseWriter, problemResponse{
		Type:   problemType,
		Title:  title,
		Status: statusCode,
		Detail: message,
	})
}

func writeProblem(responseWriter http.ResponseWriter, problem problemResponse) {
	responseWriter.Header().Set("Content-Type", "application/problem+json")
	responseWriter.WriteHeader(problem.Status)

	encoder := json.NewEncoder(responseWriter)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(problem)
}

func mapDomainError(domainError error) problemResponse {
	var validationError *domain.ValidationError
	if errors.As(domainError, &validationError) {
		responseErrors := make([]fieldError, 0, len(validationError.Fields))
		for _, validationField := range validationError.Fields {
			responseErrors = append(responseErrors, fieldError{
				Field:   validationField.Field,
				Message: validationField.Message,
			})
		}

		return problemResponse{
			Type:   "/problems/validation-error",
			Title:  "Validation failed",
			Status: http.StatusUnprocessableEntity,
			Detail: "One or more fields have invalid values",
			Errors: responseErrors,
		}
	}

	switch {
	case domainError == nil:
		return problemResponse{
			Type:   "/problems/internal-error",
			Title:  "Internal error",
			Status: http.StatusOK,
		}
	case errors.Is(domainError, domain.ErrInvalidInput):
		return problemResponse{
			Type:   "/problems/invalid-request",
			Title:  "Invalid request",
			Status: http.StatusBadRequest,
			Detail: trimmedDomainErrorDetail(domainError, domain.ErrInvalidInput),
		}
	case errors.Is(domainError, domain.ErrNotFound):
		return problemResponse{
			Type:   "/problems/not-found",
			Title:  "Not found",
			Status: http.StatusNotFound,
			Detail: trimmedDomainErrorDetail(domainError, domain.ErrNotFound),
		}
	case errors.Is(domainError, domain.ErrConflict):
		return problemResponse{
			Type:   "/problems/conflict",
			Title:  "Conflict",
			Status: http.StatusConflict,
			Detail: trimmedDomainErrorDetail(domainError, domain.ErrConflict),
		}
	case errors.Is(domainError, domain.ErrNotImplemented):
		return problemResponse{
			Type:   "/problems/internal-error",
			Title:  "Internal error",
			Status: http.StatusNotImplemented,
			Detail: trimmedDomainErrorDetail(domainError, domain.ErrNotImplemented),
		}
	default:
		return problemResponse{
			Type:   "/problems/internal-error",
			Title:  "Internal error",
			Status: http.StatusInternalServerError,
			Detail: "internal error",
		}
	}
}

func problemTypeAndTitle(statusCode int) (string, string) {
	switch statusCode {
	case http.StatusBadRequest:
		return "/problems/invalid-request", "Invalid request"
	case http.StatusNotFound:
		return "/problems/not-found", "Not found"
	case http.StatusConflict:
		return "/problems/conflict", "Conflict"
	case http.StatusMethodNotAllowed:
		return "/problems/method-not-allowed", "Method not allowed"
	default:
		return "/problems/internal-error", "Internal error"
	}
}

func trimmedDomainErrorDetail(domainError error, sentinel error) string {
	if domainError == nil {
		return ""
	}

	errorMessage := domainError.Error()
	sentinelMessage := sentinel.Error()
	if errorMessage == sentinelMessage {
		return sentinelMessage
	}

	suffix := ": " + sentinelMessage
	if strings.HasSuffix(errorMessage, suffix) {
		return strings.TrimSuffix(errorMessage, suffix)
	}

	return errorMessage
}
