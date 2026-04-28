package http

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
)

type DiagramsHandler struct {
	diagramUseCase DiagramUseCase
}

func NewDiagramsHandler(diagramUseCase DiagramUseCase) *DiagramsHandler {
	return &DiagramsHandler{diagramUseCase: diagramUseCase}
}

func (handler *DiagramsHandler) GetObjects(
	responseWriter http.ResponseWriter,
	request *http.Request,
) {
	query, parseError := parseObjectListQuery(request)
	if parseError != nil {
		writeDiagramError(responseWriter, parseError)
		return
	}

	result, useCaseError := handler.diagramUseCase.ListObjectsWithPublishedDiagrams(
		request.Context(),
		query,
	)
	if useCaseError != nil {
		writeDiagramError(responseWriter, useCaseError)
		return
	}

	writeJSON(responseWriter, http.StatusOK, mapObjectListResponse(result))
}

func (handler *DiagramsHandler) GetObjectDiagrams(
	responseWriter http.ResponseWriter,
	request *http.Request,
) {
	objectID, parseError := uuid.Parse(strings.TrimSpace(chi.URLParam(request, "objectId")))
	if parseError != nil {
		writeDiagramError(
			responseWriter,
			fmt.Errorf("object id is invalid: %w", domain.ErrInvalidInput),
		)
		return
	}

	query, queryError := parseDiagramListQuery(request)
	if queryError != nil {
		writeDiagramError(responseWriter, queryError)
		return
	}
	query.ObjectID = objectID

	result, useCaseError := handler.diagramUseCase.ListPublishedDiagrams(
		request.Context(),
		query,
	)
	if useCaseError != nil {
		writeDiagramError(responseWriter, useCaseError)
		return
	}

	writeJSON(responseWriter, http.StatusOK, mapDiagramListResponse(result))
}

func (handler *DiagramsHandler) GetDiagram(
	responseWriter http.ResponseWriter,
	request *http.Request,
) {
	diagramID, parseError := uuid.Parse(strings.TrimSpace(chi.URLParam(request, "diagramId")))
	if parseError != nil {
		writeDiagramError(
			responseWriter,
			fmt.Errorf("diagram id is invalid: %w", domain.ErrInvalidInput),
		)
		return
	}

	diagram, useCaseError := handler.diagramUseCase.GetPublishedDiagram(request.Context(), diagramID)
	if useCaseError != nil {
		writeDiagramError(responseWriter, useCaseError)
		return
	}

	responseWriter.Header().Set("Cache-Control", "private, max-age=10")
	writeJSON(responseWriter, http.StatusOK, mapDiagramResponse(diagram))
}

func (handler *DiagramsHandler) GetDiagramSnapshot(
	responseWriter http.ResponseWriter,
	request *http.Request,
) {
	diagramID, parseError := uuid.Parse(strings.TrimSpace(chi.URLParam(request, "diagramId")))
	if parseError != nil {
		writeDiagramError(
			responseWriter,
			fmt.Errorf("diagram id is invalid: %w", domain.ErrInvalidInput),
		)
		return
	}

	snapshot, useCaseError := handler.diagramUseCase.GetDiagramSnapshot(request.Context(), diagramID)
	if useCaseError != nil {
		writeDiagramError(responseWriter, useCaseError)
		return
	}

	writeJSON(responseWriter, http.StatusOK, mapDiagramSnapshotResponse(snapshot))
}

func parseObjectListQuery(request *http.Request) (domain.ObjectListQuery, error) {
	query := domain.ObjectListQuery{}

	if limitRaw := strings.TrimSpace(request.URL.Query().Get("limit")); limitRaw != "" {
		limit, parseError := strconv.Atoi(limitRaw)
		if parseError != nil {
			return domain.ObjectListQuery{}, fmt.Errorf("invalid limit: %w", domain.ErrInvalidInput)
		}
		query.Limit = limit
	}

	if offsetRaw := strings.TrimSpace(request.URL.Query().Get("offset")); offsetRaw != "" {
		offset, parseError := strconv.Atoi(offsetRaw)
		if parseError != nil {
			return domain.ObjectListQuery{}, fmt.Errorf("invalid offset: %w", domain.ErrInvalidInput)
		}
		query.Offset = offset
	}

	return query, nil
}

func parseDiagramListQuery(request *http.Request) (domain.DiagramListQuery, error) {
	query := domain.DiagramListQuery{}

	if limitRaw := strings.TrimSpace(request.URL.Query().Get("limit")); limitRaw != "" {
		limit, parseError := strconv.Atoi(limitRaw)
		if parseError != nil {
			return domain.DiagramListQuery{}, fmt.Errorf("invalid limit: %w", domain.ErrInvalidInput)
		}
		query.Limit = limit
	}

	if offsetRaw := strings.TrimSpace(request.URL.Query().Get("offset")); offsetRaw != "" {
		offset, parseError := strconv.Atoi(offsetRaw)
		if parseError != nil {
			return domain.DiagramListQuery{}, fmt.Errorf("invalid offset: %w", domain.ErrInvalidInput)
		}
		query.Offset = offset
	}

	return query, nil
}

func writeDiagramError(responseWriter http.ResponseWriter, useCaseError error) {
	switch {
	case errors.Is(useCaseError, domain.ErrInvalidInput):
		writeProblem(responseWriter, Problem{
			Type:   "/errors/diagrams/invalid-request",
			Title:  "Invalid request",
			Status: http.StatusBadRequest,
			Detail: trimSentinelSuffix(useCaseError, domain.ErrInvalidInput),
		})
	case errors.Is(useCaseError, domain.ErrNotPublished):
		writeProblem(responseWriter, Problem{
			Type:   "/errors/diagrams/not-published",
			Title:  "Diagram is not published",
			Status: http.StatusNotFound,
			Detail: trimSentinelSuffix(useCaseError, domain.ErrNotPublished),
		})
	case errors.Is(useCaseError, domain.ErrNotFound):
		writeProblem(responseWriter, Problem{
			Type:   "/errors/diagrams/not-found",
			Title:  "Diagram not found",
			Status: http.StatusNotFound,
			Detail: trimSentinelSuffix(useCaseError, domain.ErrNotFound),
		})
	case errors.Is(useCaseError, domain.ErrUnavailable):
		writeProblem(responseWriter, Problem{
			Type:   "/errors/realtime/cache-unavailable",
			Title:  "Realtime cache unavailable",
			Status: http.StatusServiceUnavailable,
			Detail: "cache unavailable",
		})
	default:
		writeProblem(responseWriter, Problem{
			Type:   "/problems/internal-error",
			Title:  "Internal error",
			Status: http.StatusInternalServerError,
			Detail: "internal error",
		})
	}
}
