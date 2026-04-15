package http

import (
	"embed"
	"net/http"
)

const (
	swaggerIndexRoute = "/swagger"
	swaggerSpecRoute  = "/swagger/openapi.yaml"
)

//go:embed swagger/index.html swagger/openapi.yaml
var swaggerFiles embed.FS

func registerSwaggerRoutes(httpServeMux *http.ServeMux) {
	httpServeMux.HandleFunc(swaggerIndexRoute, handleSwaggerIndex)
	httpServeMux.HandleFunc(swaggerSpecRoute, handleSwaggerSpec)
}

func handleSwaggerIndex(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		writeError(responseWriter, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	payload, readError := swaggerFiles.ReadFile("swagger/index.html")
	if readError != nil {
		writeError(responseWriter, http.StatusInternalServerError, "swagger not available")
		return
	}

	responseWriter.Header().Set("Content-Type", "text/html; charset=utf-8")
	responseWriter.WriteHeader(http.StatusOK)
	_, _ = responseWriter.Write(payload)
}

func handleSwaggerSpec(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		writeError(responseWriter, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	payload, readError := swaggerFiles.ReadFile("swagger/openapi.yaml")
	if readError != nil {
		writeError(responseWriter, http.StatusInternalServerError, "swagger not available")
		return
	}

	responseWriter.Header().Set("Content-Type", "application/yaml")
	responseWriter.WriteHeader(http.StatusOK)
	_, _ = responseWriter.Write(payload)
}
