package http

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"gitlab.controlitix.ru/controlitix/ms-editor/internal/domain"
	"gitlab.controlitix.ru/controlitix/ms-editor/internal/usecase"
)

type Handler struct {
	monitoringObjectUseCase *usecase.MonitoringObjectUseCase
	diagramUseCase          *usecase.DiagramUseCase
	figureUseCase           *usecase.FigureUseCase
}

func NewHandler(
	monitoringObjectUseCase *usecase.MonitoringObjectUseCase,
	diagramUseCase *usecase.DiagramUseCase,
	figureUseCase *usecase.FigureUseCase,
) *Handler {
	return &Handler{
		monitoringObjectUseCase: monitoringObjectUseCase,
		diagramUseCase:          diagramUseCase,
		figureUseCase:           figureUseCase,
	}
}

func (handler *Handler) RegisterRoutes(httpServeMux *http.ServeMux) {
	httpServeMux.HandleFunc("/health", handler.handleHealth)
	httpServeMux.HandleFunc("/objects", handler.handleObjects)
	httpServeMux.HandleFunc("/objects/", handler.handleObjects)
	httpServeMux.HandleFunc("/diagrams/", handler.handleDiagrams)
	httpServeMux.HandleFunc("/figures/", handler.handleFigures)
	registerSwaggerRoutes(httpServeMux)
}

func (handler *Handler) handleHealth(
	responseWriter http.ResponseWriter,
	request *http.Request,
) {
	if request.Method != http.MethodGet {
		writeError(responseWriter, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	writeJSON(responseWriter, http.StatusOK, map[string]string{"status": "ok"})
}

func (handler *Handler) handleObjects(
	responseWriter http.ResponseWriter,
	request *http.Request,
) {
	if request.URL.Path == "/objects" {
		switch request.Method {
		case http.MethodPost:
			handler.createMonitoringObject(responseWriter, request)
		case http.MethodGet:
			handler.listMonitoringObjects(responseWriter, request)
		default:
			writeError(responseWriter, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	pathSegments := splitPath(strings.TrimPrefix(request.URL.Path, "/objects/"))
	if len(pathSegments) == 1 {
		switch request.Method {
		case http.MethodGet:
			handler.getMonitoringObject(responseWriter, request, pathSegments[0])
		case http.MethodPatch:
			handler.updateMonitoringObject(responseWriter, request, pathSegments[0])
		case http.MethodDelete:
			handler.deleteMonitoringObject(responseWriter, request, pathSegments[0])
		default:
			writeError(responseWriter, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	if len(pathSegments) == 2 && pathSegments[1] == "diagrams" {
		switch request.Method {
		case http.MethodPost:
			handler.createDiagram(responseWriter, request, pathSegments[0])
		case http.MethodGet:
			handler.listDiagramsByMonitoringObject(
				responseWriter,
				request,
				pathSegments[0],
			)
		default:
			writeError(responseWriter, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	writeError(responseWriter, http.StatusNotFound, "not found")
}

func (handler *Handler) handleDiagrams(
	responseWriter http.ResponseWriter,
	request *http.Request,
) {
	pathSegments := splitPath(strings.TrimPrefix(request.URL.Path, "/diagrams/"))
	if len(pathSegments) == 1 {
		switch request.Method {
		case http.MethodGet:
			handler.getDiagram(responseWriter, request, pathSegments[0])
		case http.MethodPatch:
			handler.updateDiagram(responseWriter, request, pathSegments[0])
		case http.MethodDelete:
			handler.deleteDiagram(responseWriter, request, pathSegments[0])
		default:
			writeError(responseWriter, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	if len(pathSegments) == 2 && pathSegments[1] == "publish" {
		if request.Method != http.MethodPost {
			writeError(responseWriter, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		handler.publishDiagram(responseWriter, request, pathSegments[0])
		return
	}

	if len(pathSegments) == 2 && pathSegments[1] == "figures" {
		switch request.Method {
		case http.MethodPost:
			handler.createFigures(responseWriter, request, pathSegments[0])
		case http.MethodGet:
			handler.listFiguresByDiagram(responseWriter, request, pathSegments[0])
		default:
			writeError(responseWriter, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	writeError(responseWriter, http.StatusNotFound, "not found")
}

func (handler *Handler) handleFigures(
	responseWriter http.ResponseWriter,
	request *http.Request,
) {
	pathSegments := splitPath(strings.TrimPrefix(request.URL.Path, "/figures/"))
	if len(pathSegments) != 1 {
		writeError(responseWriter, http.StatusNotFound, "not found")
		return
	}

	switch request.Method {
	case http.MethodPatch:
		handler.updateFigure(responseWriter, request, pathSegments[0])
	case http.MethodDelete:
		handler.deleteFigure(responseWriter, request, pathSegments[0])
	default:
		writeError(responseWriter, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (handler *Handler) createMonitoringObject(
	responseWriter http.ResponseWriter,
	request *http.Request,
) {
	requestPayload, decodeError := decodeJSONBody[monitoringObjectRequest](request.Body)
	if decodeError != nil {
		writeError(responseWriter, http.StatusBadRequest, "invalid json")
		return
	}

	monitoringObject, useCaseError := handler.monitoringObjectUseCase.CreateMonitoringObject(
		request.Context(),
		requestPayload.Name,
		requestPayload.Description,
	)
	if useCaseError != nil {
		writeDomainError(responseWriter, useCaseError)
		return
	}

	responsePayload := mapMonitoringObjectResponse(monitoringObject)
	writeJSON(responseWriter, http.StatusCreated, responsePayload)
}

func (handler *Handler) listMonitoringObjects(
	responseWriter http.ResponseWriter,
	request *http.Request,
) {
	monitoringObjects, useCaseError := handler.monitoringObjectUseCase.ListMonitoringObjects(
		request.Context(),
	)
	if useCaseError != nil {
		writeDomainError(responseWriter, useCaseError)
		return
	}

	responsePayload := make([]monitoringObjectResponse, 0, len(monitoringObjects))
	for _, monitoringObject := range monitoringObjects {
		responsePayload = append(
			responsePayload,
			mapMonitoringObjectResponse(monitoringObject),
		)
	}

	writeJSON(responseWriter, http.StatusOK, responsePayload)
}

func (handler *Handler) getMonitoringObject(
	responseWriter http.ResponseWriter,
	request *http.Request,
	monitoringObjectID string,
) {
	monitoringObject, useCaseError := handler.monitoringObjectUseCase.GetMonitoringObject(
		request.Context(),
		monitoringObjectID,
	)
	if useCaseError != nil {
		writeDomainError(responseWriter, useCaseError)
		return
	}

	responsePayload := mapMonitoringObjectResponse(monitoringObject)
	writeJSON(responseWriter, http.StatusOK, responsePayload)
}

func (handler *Handler) updateMonitoringObject(
	responseWriter http.ResponseWriter,
	request *http.Request,
	monitoringObjectID string,
) {
	requestPayload, decodeError := decodeJSONBody[monitoringObjectRequest](request.Body)
	if decodeError != nil {
		writeError(responseWriter, http.StatusBadRequest, "invalid json")
		return
	}

	monitoringObject, useCaseError := handler.monitoringObjectUseCase.UpdateMonitoringObject(
		request.Context(),
		monitoringObjectID,
		optionalString(requestPayload.Name),
		requestPayload.Description,
	)
	if useCaseError != nil {
		writeDomainError(responseWriter, useCaseError)
		return
	}

	responsePayload := mapMonitoringObjectResponse(monitoringObject)
	writeJSON(responseWriter, http.StatusOK, responsePayload)
}

func (handler *Handler) deleteMonitoringObject(
	responseWriter http.ResponseWriter,
	request *http.Request,
	monitoringObjectID string,
) {
	useCaseError := handler.monitoringObjectUseCase.DeleteMonitoringObject(
		request.Context(),
		monitoringObjectID,
	)
	if useCaseError != nil {
		writeDomainError(responseWriter, useCaseError)
		return
	}

	writeJSON(responseWriter, http.StatusNoContent, nil)
}

func (handler *Handler) createDiagram(
	responseWriter http.ResponseWriter,
	request *http.Request,
	monitoringObjectID string,
) {
	requestPayload, decodeError := decodeJSONBody[diagramRequest](request.Body)
	if decodeError != nil {
		writeError(responseWriter, http.StatusBadRequest, "invalid json")
		return
	}

	diagram, useCaseError := handler.diagramUseCase.CreateDiagram(
		request.Context(),
		monitoringObjectID,
		requestPayload.Name,
		requestPayload.Description,
	)
	if useCaseError != nil {
		writeDomainError(responseWriter, useCaseError)
		return
	}

	responsePayload := mapDiagramResponse(diagram)
	writeJSON(responseWriter, http.StatusCreated, responsePayload)
}

func (handler *Handler) listDiagramsByMonitoringObject(
	responseWriter http.ResponseWriter,
	request *http.Request,
	monitoringObjectID string,
) {
	diagrams, useCaseError := handler.diagramUseCase.ListDiagramsByMonitoringObject(
		request.Context(),
		monitoringObjectID,
	)
	if useCaseError != nil {
		writeDomainError(responseWriter, useCaseError)
		return
	}

	responsePayload := make([]diagramResponse, 0, len(diagrams))
	for _, diagram := range diagrams {
		responsePayload = append(responsePayload, mapDiagramResponse(diagram))
	}

	writeJSON(responseWriter, http.StatusOK, responsePayload)
}

func (handler *Handler) getDiagram(
	responseWriter http.ResponseWriter,
	request *http.Request,
	diagramID string,
) {
	diagram, useCaseError := handler.diagramUseCase.GetDiagram(
		request.Context(),
		diagramID,
	)
	if useCaseError != nil {
		writeDomainError(responseWriter, useCaseError)
		return
	}

	responsePayload := mapDiagramResponse(diagram)
	writeJSON(responseWriter, http.StatusOK, responsePayload)
}

func (handler *Handler) updateDiagram(
	responseWriter http.ResponseWriter,
	request *http.Request,
	diagramID string,
) {
	requestPayload, decodeError := decodeJSONBody[diagramRequest](request.Body)
	if decodeError != nil {
		writeError(responseWriter, http.StatusBadRequest, "invalid json")
		return
	}

	diagram, useCaseError := handler.diagramUseCase.UpdateDiagram(
		request.Context(),
		diagramID,
		requestPayload.Name,
		requestPayload.Description,
	)
	if useCaseError != nil {
		writeDomainError(responseWriter, useCaseError)
		return
	}

	responsePayload := mapDiagramResponse(diagram)
	writeJSON(responseWriter, http.StatusOK, responsePayload)
}

func (handler *Handler) deleteDiagram(
	responseWriter http.ResponseWriter,
	request *http.Request,
	diagramID string,
) {
	useCaseError := handler.diagramUseCase.DeleteDiagram(
		request.Context(),
		diagramID,
	)
	if useCaseError != nil {
		writeDomainError(responseWriter, useCaseError)
		return
	}

	writeJSON(responseWriter, http.StatusNoContent, nil)
}

func (handler *Handler) publishDiagram(
	responseWriter http.ResponseWriter,
	request *http.Request,
	diagramID string,
) {
	diagram, useCaseError := handler.diagramUseCase.PublishDiagram(
		request.Context(),
		diagramID,
	)
	if useCaseError != nil {
		writeDomainError(responseWriter, useCaseError)
		return
	}

	responsePayload := mapDiagramResponse(diagram)
	writeJSON(responseWriter, http.StatusOK, responsePayload)
}

func (handler *Handler) createFigures(
	responseWriter http.ResponseWriter,
	request *http.Request,
	diagramID string,
) {
	requestPayload, decodeError := decodeJSONBody[createFiguresRequest](request.Body)
	if decodeError != nil {
		writeError(responseWriter, http.StatusBadRequest, "invalid json")
		return
	}

	if len(requestPayload.Figures) == 0 {
		writeError(responseWriter, http.StatusBadRequest, "figures list is empty")
		return
	}

	figures := make([]domain.Figure, 0, len(requestPayload.Figures))
	for _, figure := range requestPayload.Figures {
		if strings.TrimSpace(figure.FigureType) == "" {
			writeError(responseWriter, http.StatusBadRequest, "figure type is required")
			return
		}

		figures = append(figures, domain.Figure{
			DiagramID:  diagramID,
			FigureType: figure.FigureType,
			Parameters: figure.Parameters,
			TagID:      figure.TagID,
		})
	}

	createdFigures, useCaseError := handler.figureUseCase.CreateFigures(
		request.Context(),
		diagramID,
		figures,
	)
	if useCaseError != nil {
		writeDomainError(responseWriter, useCaseError)
		return
	}

	responsePayload := make([]figureResponse, 0, len(createdFigures))
	for _, figure := range createdFigures {
		responsePayload = append(responsePayload, mapFigureResponse(figure))
	}

	writeJSON(responseWriter, http.StatusCreated, responsePayload)
}

func (handler *Handler) listFiguresByDiagram(
	responseWriter http.ResponseWriter,
	request *http.Request,
	diagramID string,
) {
	figures, useCaseError := handler.figureUseCase.ListFiguresByDiagram(
		request.Context(),
		diagramID,
	)
	if useCaseError != nil {
		writeDomainError(responseWriter, useCaseError)
		return
	}

	responsePayload := make([]figureResponse, 0, len(figures))
	for _, figure := range figures {
		responsePayload = append(responsePayload, mapFigureResponse(figure))
	}

	writeJSON(responseWriter, http.StatusOK, responsePayload)
}

func (handler *Handler) updateFigure(
	responseWriter http.ResponseWriter,
	request *http.Request,
	figureID string,
) {
	requestPayload, decodeError := decodeJSONBody[updateFigureRequest](request.Body)
	if decodeError != nil {
		writeError(responseWriter, http.StatusBadRequest, "invalid json")
		return
	}

	figure, useCaseError := handler.figureUseCase.UpdateFigure(
		request.Context(),
		figureID,
		requestPayload.TagID,
		requestPayload.FigureType,
		requestPayload.Parameters,
	)
	if useCaseError != nil {
		writeDomainError(responseWriter, useCaseError)
		return
	}

	responsePayload := mapFigureResponse(figure)
	writeJSON(responseWriter, http.StatusOK, responsePayload)
}

func (handler *Handler) deleteFigure(
	responseWriter http.ResponseWriter,
	request *http.Request,
	figureID string,
) {
	useCaseError := handler.figureUseCase.DeleteFigure(
		request.Context(),
		figureID,
	)
	if useCaseError != nil {
		writeDomainError(responseWriter, useCaseError)
		return
	}

	writeJSON(responseWriter, http.StatusNoContent, nil)
}

func mapMonitoringObjectResponse(monitoringObject domain.MonitoringObject) monitoringObjectResponse {
	return monitoringObjectResponse{
		ID:          monitoringObject.ID,
		Name:        monitoringObject.Name,
		Description: monitoringObject.Description,
		CreatedAt:   monitoringObject.CreatedAt,
		UpdatedAt:   monitoringObject.UpdatedAt,
	}
}

func mapDiagramResponse(diagram domain.Diagram) diagramResponse {
	return diagramResponse{
		ID:          diagram.ID,
		ObjectID:    diagram.ObjectID,
		Name:        diagram.Name,
		Description: diagram.Description,
		PublishedAt: diagram.PublishedAt,
		CreatedAt:   diagram.CreatedAt,
		UpdatedAt:   diagram.UpdatedAt,
	}
}

func mapFigureResponse(figure domain.Figure) figureResponse {
	return figureResponse{
		ID:         figure.ID,
		DiagramID:  figure.DiagramID,
		TagID:      figure.TagID,
		FigureType: figure.FigureType,
		Parameters: figure.Parameters,
		CreatedAt:  figure.CreatedAt,
		UpdatedAt:  figure.UpdatedAt,
	}
}

func splitPath(path string) []string {
	trimmedPath := strings.Trim(path, "/")
	if trimmedPath == "" {
		return []string{}
	}

	return strings.Split(trimmedPath, "/")
}

func decodeJSONBody[T any](body io.ReadCloser) (T, error) {
	defer body.Close()

	var payload T
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()

	if decodeError := decoder.Decode(&payload); decodeError != nil {
		return payload, decodeError
	}

	return payload, nil
}

func writeDomainError(responseWriter http.ResponseWriter, domainError error) {
	statusCode, message := mapDomainError(domainError)
	writeError(responseWriter, statusCode, message)
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}

	return &value
}
