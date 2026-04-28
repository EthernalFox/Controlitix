package http

import (
	"encoding/json"
	stdhttp "net/http"
)

type Problem struct {
	Type     string         `json:"type"`
	Title    string         `json:"title"`
	Status   int            `json:"status"`
	Detail   string         `json:"detail,omitempty"`
	Instance string         `json:"instance,omitempty"`
	Extra    map[string]any `json:"-"`
}

func WriteProblem(responseWriter stdhttp.ResponseWriter, problem Problem) {
	if problem.Status == 0 {
		problem.Status = stdhttp.StatusInternalServerError
	}
	if problem.Type == "" {
		problem.Type = "/errors/internal"
	}
	if problem.Title == "" {
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
