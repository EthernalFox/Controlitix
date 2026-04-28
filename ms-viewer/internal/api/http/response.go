package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
)

type Problem struct {
	Type     string         `json:"type"`
	Title    string         `json:"title"`
	Status   int            `json:"status"`
	Detail   string         `json:"detail,omitempty"`
	Instance string         `json:"instance,omitempty"`
	Extra    map[string]any `json:"-"`
}

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

func writeProblem(responseWriter http.ResponseWriter, problem Problem) {
	if problem.Status == 0 {
		problem.Status = http.StatusInternalServerError
	}
	if strings.TrimSpace(problem.Type) == "" {
		problem.Type = "/problems/internal-error"
	}
	if strings.TrimSpace(problem.Title) == "" {
		problem.Title = "Internal error"
	}

	payload := map[string]any{
		"type":   problem.Type,
		"title":  problem.Title,
		"status": problem.Status,
	}
	if problem.Detail != "" {
		payload["detail"] = problem.Detail
	}
	if problem.Instance != "" {
		payload["instance"] = problem.Instance
	}
	for key, value := range problem.Extra {
		payload[key] = value
	}

	responseWriter.Header().Set("Content-Type", "application/problem+json")
	responseWriter.WriteHeader(problem.Status)

	encoder := json.NewEncoder(responseWriter)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(payload)
}

func writeDomainError(responseWriter http.ResponseWriter, domainError error) {
	problem := mapDomainError(domainError)
	writeProblem(responseWriter, problem)
}

func mapDomainError(domainError error) Problem {
	switch {
	case errors.Is(domainError, domain.ErrInvalidInput):
		return Problem{
			Type:   "/errors/trends/invalid-range",
			Title:  "Invalid range",
			Status: http.StatusBadRequest,
			Detail: trimSentinelSuffix(domainError, domain.ErrInvalidInput),
		}
	case errors.Is(domainError, domain.ErrNotFound):
		return Problem{
			Type:   "/errors/trends/tag-not-found",
			Title:  "Tag not found",
			Status: http.StatusNotFound,
			Detail: trimSentinelSuffix(domainError, domain.ErrNotFound),
		}
	case errors.Is(domainError, domain.ErrForbidden):
		return Problem{
			Type:   "/errors/trends/forbidden",
			Title:  "Forbidden",
			Status: http.StatusForbidden,
			Detail: trimSentinelSuffix(domainError, domain.ErrForbidden),
		}
	default:
		return Problem{
			Type:   "/problems/internal-error",
			Title:  "Internal error",
			Status: http.StatusInternalServerError,
			Detail: "internal error",
		}
	}
}

func trimSentinelSuffix(err error, sentinel error) string {
	if err == nil {
		return ""
	}
	if err.Error() == sentinel.Error() {
		return sentinel.Error()
	}

	suffix := ": " + sentinel.Error()
	if strings.HasSuffix(err.Error(), suffix) {
		return strings.TrimSuffix(err.Error(), suffix)
	}

	return err.Error()
}
