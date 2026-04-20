package http

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/EthernalFox/Controlitix/ms-editor/internal/domain"
	"github.com/EthernalFox/Controlitix/ms-editor/internal/usecase"
)

type Handler struct {
	monitoringObjectUseCase *usecase.MonitoringObjectUseCase
	diagramUseCase          *usecase.DiagramUseCase
	figureUseCase           *usecase.FigureUseCase
	deviceUseCase           *usecase.DeviceUseCase
	tagUseCase              *usecase.TagUseCase
}

func NewHandler(
	monitoringObjectUseCase *usecase.MonitoringObjectUseCase,
	diagramUseCase *usecase.DiagramUseCase,
	figureUseCase *usecase.FigureUseCase,
	deviceUseCase *usecase.DeviceUseCase,
	tagUseCase *usecase.TagUseCase,
) *Handler {
	return &Handler{
		monitoringObjectUseCase: monitoringObjectUseCase,
		diagramUseCase:          diagramUseCase,
		figureUseCase:           figureUseCase,
		deviceUseCase:           deviceUseCase,
		tagUseCase:              tagUseCase,
	}
}

func (handler *Handler) RegisterRoutes(httpServeMux *http.ServeMux) {
	httpServeMux.HandleFunc("/health", handler.handleHealth)
	httpServeMux.HandleFunc("/objects", handler.handleObjects)
	httpServeMux.HandleFunc("/objects/", handler.handleObjects)
	httpServeMux.HandleFunc("/diagrams/", handler.handleDiagrams)
	httpServeMux.HandleFunc("/figures/", handler.handleFigures)
	httpServeMux.HandleFunc("/devices", handler.handleDevices)
	httpServeMux.HandleFunc("/devices/", handler.handleDevices)
	httpServeMux.HandleFunc("/device-types", handler.handleDeviceTypes)
	httpServeMux.HandleFunc("/tags", handler.handleTags)
	httpServeMux.HandleFunc("/tags/", handler.handleTags)
	httpServeMux.HandleFunc("/data-types", handler.handleDataTypes)
	httpServeMux.HandleFunc("/units", handler.handleUnits)
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

	if len(pathSegments) == 2 && pathSegments[1] == "devices" {
		switch request.Method {
		case http.MethodPost:
			handler.createDevice(responseWriter, request, &pathSegments[0])
		case http.MethodGet:
			handler.listDevices(responseWriter, request, &pathSegments[0])
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

func (handler *Handler) handleDevices(
	responseWriter http.ResponseWriter,
	request *http.Request,
) {
	if request.URL.Path == "/devices" {
		switch request.Method {
		case http.MethodPost:
			handler.createDevice(responseWriter, request, nil)
		case http.MethodGet:
			handler.listDevices(responseWriter, request, nil)
		default:
			writeError(responseWriter, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	pathSegments := splitPath(strings.TrimPrefix(request.URL.Path, "/devices/"))
	if len(pathSegments) == 1 {
		switch request.Method {
		case http.MethodGet:
			handler.getDevice(responseWriter, request, pathSegments[0])
		case http.MethodPatch:
			handler.updateDevice(responseWriter, request, pathSegments[0])
		case http.MethodDelete:
			handler.deleteDevice(responseWriter, request, pathSegments[0])
		default:
			writeError(responseWriter, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	if len(pathSegments) == 2 && pathSegments[1] == "config" {
		if request.Method != http.MethodPost {
			writeError(responseWriter, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		handler.updateDeviceParams(responseWriter, request, pathSegments[0])
		return
	}

	if len(pathSegments) == 2 && pathSegments[1] == "assign" {
		if request.Method != http.MethodPatch {
			writeError(responseWriter, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		handler.assignDeviceToObject(responseWriter, request, pathSegments[0])
		return
	}

	if len(pathSegments) == 2 && pathSegments[1] == "tags" {
		switch request.Method {
		case http.MethodPost:
			handler.createTag(responseWriter, request, pathSegments[0])
		case http.MethodGet:
			handler.listTags(responseWriter, request, &pathSegments[0])
		default:
			writeError(responseWriter, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	writeError(responseWriter, http.StatusNotFound, "not found")
}

func (handler *Handler) handleDeviceTypes(
	responseWriter http.ResponseWriter,
	request *http.Request,
) {
	if request.Method != http.MethodGet {
		writeError(responseWriter, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	handler.listDeviceTypes(responseWriter, request)
}

func (handler *Handler) handleTags(
	responseWriter http.ResponseWriter,
	request *http.Request,
) {
	if request.URL.Path == "/tags" {
		if request.Method != http.MethodGet {
			writeError(responseWriter, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		handler.listTags(responseWriter, request, nil)
		return
	}

	pathSegments := splitPath(strings.TrimPrefix(request.URL.Path, "/tags/"))
	if len(pathSegments) == 1 {
		switch request.Method {
		case http.MethodGet:
			handler.getTag(responseWriter, request, pathSegments[0])
		case http.MethodPatch:
			handler.updateTag(responseWriter, request, pathSegments[0])
		case http.MethodDelete:
			handler.deleteTag(responseWriter, request, pathSegments[0])
		default:
			writeError(responseWriter, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	if len(pathSegments) == 2 && pathSegments[1] == "params" {
		if request.Method != http.MethodPut {
			writeError(responseWriter, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		handler.updateTagParams(responseWriter, request, pathSegments[0])
		return
	}

	if len(pathSegments) == 2 && pathSegments[1] == "setpoints" {
		switch request.Method {
		case http.MethodPut:
			handler.updateTagSetpoints(responseWriter, request, pathSegments[0])
		case http.MethodDelete:
			handler.deleteTagSetpoints(responseWriter, request, pathSegments[0])
		default:
			writeError(responseWriter, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	if len(pathSegments) == 2 && pathSegments[1] == "scaling" {
		switch request.Method {
		case http.MethodPut:
			handler.updateTagScaling(responseWriter, request, pathSegments[0])
		case http.MethodDelete:
			handler.deleteTagScaling(responseWriter, request, pathSegments[0])
		default:
			writeError(responseWriter, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	writeError(responseWriter, http.StatusNotFound, "not found")
}

func (handler *Handler) handleDataTypes(
	responseWriter http.ResponseWriter,
	request *http.Request,
) {
	if request.Method != http.MethodGet {
		writeError(responseWriter, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	handler.listDataTypes(responseWriter, request)
}

func (handler *Handler) handleUnits(
	responseWriter http.ResponseWriter,
	request *http.Request,
) {
	if request.Method != http.MethodGet {
		writeError(responseWriter, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	handler.listUnits(responseWriter, request)
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
	query, parseError := parseObjectListQuery(request)
	if parseError != nil {
		writeError(responseWriter, http.StatusBadRequest, parseError.Error())
		return
	}

	monitoringObjects, useCaseError := handler.monitoringObjectUseCase.ListMonitoringObjects(
		request.Context(),
		query,
	)
	if useCaseError != nil {
		writeDomainError(responseWriter, useCaseError)
		return
	}

	responsePayload := make([]monitoringObjectResponse, 0, len(monitoringObjects.Items))
	for _, monitoringObject := range monitoringObjects.Items {
		responsePayload = append(
			responsePayload,
			mapMonitoringObjectResponse(monitoringObject),
		)
	}

	writeJSON(responseWriter, http.StatusOK, mapListResponse(
		responsePayload,
		monitoringObjects.Total,
		monitoringObjects.Offset,
		monitoringObjects.Limit,
	))
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
	query, parseError := parseDiagramListQuery(request, &monitoringObjectID)
	if parseError != nil {
		writeError(responseWriter, http.StatusBadRequest, parseError.Error())
		return
	}

	diagrams, useCaseError := handler.diagramUseCase.ListDiagrams(
		request.Context(),
		query,
	)
	if useCaseError != nil {
		writeDomainError(responseWriter, useCaseError)
		return
	}

	responsePayload := make([]diagramResponse, 0, len(diagrams.Items))
	for _, diagram := range diagrams.Items {
		responsePayload = append(responsePayload, mapDiagramResponse(diagram))
	}

	writeJSON(responseWriter, http.StatusOK, mapListResponse(
		responsePayload,
		diagrams.Total,
		diagrams.Offset,
		diagrams.Limit,
	))
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
		figureType, isSupportedFigureType := domain.ParseFigureType(figure.FigureType)
		if !isSupportedFigureType {
			writeError(responseWriter, http.StatusBadRequest, "unsupported figure type")
			return
		}

		figures = append(figures, domain.Figure{
			DiagramID:  diagramID,
			FigureType: figureType,
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
	query, parseError := parseFigureListQuery(request, diagramID)
	if parseError != nil {
		writeError(responseWriter, http.StatusBadRequest, parseError.Error())
		return
	}

	figures, useCaseError := handler.figureUseCase.ListFigures(
		request.Context(),
		query,
	)
	if useCaseError != nil {
		writeDomainError(responseWriter, useCaseError)
		return
	}

	responsePayload := make([]figureResponse, 0, len(figures.Items))
	for _, figure := range figures.Items {
		responsePayload = append(responsePayload, mapFigureResponse(figure))
	}

	writeJSON(responseWriter, http.StatusOK, mapListResponse(
		responsePayload,
		figures.Total,
		figures.Offset,
		figures.Limit,
	))
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

	var figureType *domain.FigureType
	if requestPayload.FigureType != nil {
		parsedFigureType, isSupportedFigureType := domain.ParseFigureType(
			*requestPayload.FigureType,
		)
		if !isSupportedFigureType {
			writeError(responseWriter, http.StatusBadRequest, "unsupported figure type")
			return
		}

		figureType = &parsedFigureType
	}

	figure, useCaseError := handler.figureUseCase.UpdateFigure(
		request.Context(),
		figureID,
		requestPayload.TagID,
		figureType,
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

func (handler *Handler) createDevice(
	responseWriter http.ResponseWriter,
	request *http.Request,
	objectID *string,
) {
	requestPayload, decodeError := decodeJSONBody[createDeviceRequest](request.Body)
	if decodeError != nil {
		writeError(responseWriter, http.StatusBadRequest, "invalid json")
		return
	}

	requestObjectID := requestPayload.ObjectID
	if objectID != nil {
		requestObjectID = objectID
	}

	deviceWithParams, useCaseError := handler.deviceUseCase.CreateDevice(
		request.Context(),
		requestObjectID,
		requestPayload.TypeID,
		requestPayload.Name,
		requestPayload.Description,
		requestPayload.Settings,
	)
	if useCaseError != nil {
		writeDomainError(responseWriter, useCaseError)
		return
	}

	writeJSON(
		responseWriter,
		http.StatusCreated,
		mapDeviceWithParamsResponse(deviceWithParams),
	)
}

func (handler *Handler) listDevices(
	responseWriter http.ResponseWriter,
	request *http.Request,
	objectID *string,
) {
	query, parseError := parseDeviceListQuery(request, objectID)
	if parseError != nil {
		writeError(responseWriter, http.StatusBadRequest, parseError.Error())
		return
	}

	devices, useCaseError := handler.deviceUseCase.ListDevices(
		request.Context(),
		query,
	)
	if useCaseError != nil {
		writeDomainError(responseWriter, useCaseError)
		return
	}

	responsePayload := make([]deviceResponse, 0, len(devices.Items))
	for _, device := range devices.Items {
		responsePayload = append(responsePayload, mapDeviceResponse(device, nil))
	}

	writeJSON(responseWriter, http.StatusOK, mapListResponse(
		responsePayload,
		devices.Total,
		devices.Offset,
		devices.Limit,
	))
}

func (handler *Handler) getDevice(
	responseWriter http.ResponseWriter,
	request *http.Request,
	deviceID string,
) {
	deviceWithParams, useCaseError := handler.deviceUseCase.GetDevice(
		request.Context(),
		deviceID,
	)
	if useCaseError != nil {
		writeDomainError(responseWriter, useCaseError)
		return
	}

	writeJSON(
		responseWriter,
		http.StatusOK,
		mapDeviceWithParamsResponse(deviceWithParams),
	)
}

func (handler *Handler) updateDevice(
	responseWriter http.ResponseWriter,
	request *http.Request,
	deviceID string,
) {
	requestPayload, decodeError := decodeJSONBody[updateDeviceRequest](request.Body)
	if decodeError != nil {
		writeError(responseWriter, http.StatusBadRequest, "invalid json")
		return
	}

	_, useCaseError := handler.deviceUseCase.UpdateDevice(
		request.Context(),
		deviceID,
		domain.DeviceUpdate{
			TypeID:      requestPayload.TypeID,
			Name:        requestPayload.Name,
			Description: requestPayload.Description,
		},
	)
	if useCaseError != nil {
		writeDomainError(responseWriter, useCaseError)
		return
	}

	deviceWithParams, getError := handler.deviceUseCase.GetDevice(
		request.Context(),
		deviceID,
	)
	if getError != nil {
		writeDomainError(responseWriter, getError)
		return
	}

	writeJSON(
		responseWriter,
		http.StatusOK,
		mapDeviceWithParamsResponse(deviceWithParams),
	)
}

func (handler *Handler) deleteDevice(
	responseWriter http.ResponseWriter,
	request *http.Request,
	deviceID string,
) {
	useCaseError := handler.deviceUseCase.DeleteDevice(request.Context(), deviceID)
	if useCaseError != nil {
		writeDomainError(responseWriter, useCaseError)
		return
	}

	writeJSON(responseWriter, http.StatusNoContent, nil)
}

func (handler *Handler) updateDeviceParams(
	responseWriter http.ResponseWriter,
	request *http.Request,
	deviceID string,
) {
	requestPayload, decodeError := decodeJSONBody[deviceParamsRequest](request.Body)
	if decodeError != nil {
		writeError(responseWriter, http.StatusBadRequest, "invalid json")
		return
	}

	_, useCaseError := handler.deviceUseCase.UpdateDeviceParams(
		request.Context(),
		deviceID,
		requestPayload.Settings,
	)
	if useCaseError != nil {
		writeDomainError(responseWriter, useCaseError)
		return
	}

	deviceWithParams, getError := handler.deviceUseCase.GetDevice(
		request.Context(),
		deviceID,
	)
	if getError != nil {
		writeDomainError(responseWriter, getError)
		return
	}

	writeJSON(
		responseWriter,
		http.StatusOK,
		mapDeviceWithParamsResponse(deviceWithParams),
	)
}

func (handler *Handler) assignDeviceToObject(
	responseWriter http.ResponseWriter,
	request *http.Request,
	deviceID string,
) {
	requestPayload, decodeError := decodeJSONBody[assignDeviceRequest](request.Body)
	if decodeError != nil {
		writeError(responseWriter, http.StatusBadRequest, "invalid json")
		return
	}

	_, useCaseError := handler.deviceUseCase.AssignDeviceToObject(
		request.Context(),
		deviceID,
		requestPayload.ObjectID,
	)
	if useCaseError != nil {
		writeDomainError(responseWriter, useCaseError)
		return
	}

	deviceWithParams, getError := handler.deviceUseCase.GetDevice(
		request.Context(),
		deviceID,
	)
	if getError != nil {
		writeDomainError(responseWriter, getError)
		return
	}

	writeJSON(
		responseWriter,
		http.StatusOK,
		mapDeviceWithParamsResponse(deviceWithParams),
	)
}

func (handler *Handler) listDeviceTypes(
	responseWriter http.ResponseWriter,
	request *http.Request,
) {
	deviceTypes, useCaseError := handler.deviceUseCase.ListDeviceTypes(
		request.Context(),
	)
	if useCaseError != nil {
		writeDomainError(responseWriter, useCaseError)
		return
	}

	responsePayload := make([]deviceTypeResponse, 0, len(deviceTypes))
	for _, deviceType := range deviceTypes {
		responsePayload = append(responsePayload, deviceTypeResponse{
			ID:   deviceType.ID,
			Name: deviceType.Name,
		})
	}

	writeJSON(responseWriter, http.StatusOK, responsePayload)
}

func (handler *Handler) createTag(
	responseWriter http.ResponseWriter,
	request *http.Request,
	deviceID string,
) {
	requestPayload, decodeError := decodeJSONBody[createTagRequest](request.Body)
	if decodeError != nil {
		writeError(responseWriter, http.StatusBadRequest, "invalid json")
		return
	}

	tagFull, useCaseError := handler.tagUseCase.CreateTag(
		request.Context(),
		deviceID,
		requestPayload.Name,
		requestPayload.Description,
		mapTagParamsInput(requestPayload.Params),
		mapSetpoints(requestPayload.Setpoints),
		mapScaling(requestPayload.Scaling),
	)
	if useCaseError != nil {
		writeDomainError(responseWriter, useCaseError)
		return
	}

	writeJSON(responseWriter, http.StatusCreated, mapTagFullResponse(tagFull))
}

func (handler *Handler) listTags(
	responseWriter http.ResponseWriter,
	request *http.Request,
	deviceID *string,
) {
	query, parseError := parseTagListQuery(request, deviceID)
	if parseError != nil {
		writeError(responseWriter, http.StatusBadRequest, parseError.Error())
		return
	}

	tags, useCaseError := handler.tagUseCase.ListTags(
		request.Context(),
		query,
	)
	if useCaseError != nil {
		writeDomainError(responseWriter, useCaseError)
		return
	}

	responsePayload := make([]tagResponse, 0, len(tags.Items))
	for _, tag := range tags.Items {
		responsePayload = append(responsePayload, mapTagResponse(tag))
	}

	writeJSON(responseWriter, http.StatusOK, mapListResponse(
		responsePayload,
		tags.Total,
		tags.Offset,
		tags.Limit,
	))
}

func (handler *Handler) getTag(
	responseWriter http.ResponseWriter,
	request *http.Request,
	tagID string,
) {
	tagFull, useCaseError := handler.tagUseCase.GetTag(request.Context(), tagID)
	if useCaseError != nil {
		writeDomainError(responseWriter, useCaseError)
		return
	}

	writeJSON(responseWriter, http.StatusOK, mapTagFullResponse(tagFull))
}

func (handler *Handler) updateTag(
	responseWriter http.ResponseWriter,
	request *http.Request,
	tagID string,
) {
	requestPayload, decodeError := decodeJSONBody[updateTagRequest](request.Body)
	if decodeError != nil {
		writeError(responseWriter, http.StatusBadRequest, "invalid json")
		return
	}

	_, useCaseError := handler.tagUseCase.UpdateTag(
		request.Context(),
		tagID,
		domain.TagUpdate{
			Name:        requestPayload.Name,
			Description: requestPayload.Description,
		},
	)
	if useCaseError != nil {
		writeDomainError(responseWriter, useCaseError)
		return
	}

	tagFull, getError := handler.tagUseCase.GetTag(request.Context(), tagID)
	if getError != nil {
		writeDomainError(responseWriter, getError)
		return
	}

	writeJSON(responseWriter, http.StatusOK, mapTagFullResponse(tagFull))
}

func (handler *Handler) deleteTag(
	responseWriter http.ResponseWriter,
	request *http.Request,
	tagID string,
) {
	useCaseError := handler.tagUseCase.DeleteTag(request.Context(), tagID)
	if useCaseError != nil {
		writeDomainError(responseWriter, useCaseError)
		return
	}

	writeJSON(responseWriter, http.StatusNoContent, nil)
}

func (handler *Handler) updateTagParams(
	responseWriter http.ResponseWriter,
	request *http.Request,
	tagID string,
) {
	requestPayload, decodeError := decodeJSONBody[tagParamsInput](request.Body)
	if decodeError != nil {
		writeError(responseWriter, http.StatusBadRequest, "invalid json")
		return
	}

	_, useCaseError := handler.tagUseCase.UpdateTagParams(
		request.Context(),
		tagID,
		domain.TagParamsUpdate{
			DataTypeID: intPointer(requestPayload.DataTypeID),
			UnitID:     requestPayload.UnitID,
			Address:    jsonRawMessagePointer(requestPayload.Address),
		},
	)
	if useCaseError != nil {
		writeDomainError(responseWriter, useCaseError)
		return
	}

	tagFull, getError := handler.tagUseCase.GetTag(request.Context(), tagID)
	if getError != nil {
		writeDomainError(responseWriter, getError)
		return
	}

	writeJSON(responseWriter, http.StatusOK, mapTagFullResponse(tagFull))
}

func (handler *Handler) updateTagSetpoints(
	responseWriter http.ResponseWriter,
	request *http.Request,
	tagID string,
) {
	requestPayload, decodeError := decodeJSONBody[setpointsInput](request.Body)
	if decodeError != nil {
		writeError(responseWriter, http.StatusBadRequest, "invalid json")
		return
	}

	_, useCaseError := handler.tagUseCase.UpdateTagSetpoints(
		request.Context(),
		tagID,
		domain.TagSetpoints{
			LoLo: requestPayload.LoLo,
			Lo:   requestPayload.Lo,
			Hi:   requestPayload.Hi,
			HiHi: requestPayload.HiHi,
		},
	)
	if useCaseError != nil {
		writeDomainError(responseWriter, useCaseError)
		return
	}

	tagFull, getError := handler.tagUseCase.GetTag(request.Context(), tagID)
	if getError != nil {
		writeDomainError(responseWriter, getError)
		return
	}

	writeJSON(responseWriter, http.StatusOK, mapTagFullResponse(tagFull))
}

func (handler *Handler) deleteTagSetpoints(
	responseWriter http.ResponseWriter,
	request *http.Request,
	tagID string,
) {
	useCaseError := handler.tagUseCase.DeleteTagSetpoints(request.Context(), tagID)
	if useCaseError != nil {
		writeDomainError(responseWriter, useCaseError)
		return
	}

	tagFull, getError := handler.tagUseCase.GetTag(request.Context(), tagID)
	if getError != nil {
		writeDomainError(responseWriter, getError)
		return
	}

	writeJSON(responseWriter, http.StatusOK, mapTagFullResponse(tagFull))
}

func (handler *Handler) updateTagScaling(
	responseWriter http.ResponseWriter,
	request *http.Request,
	tagID string,
) {
	requestPayload, decodeError := decodeJSONBody[scalingInput](request.Body)
	if decodeError != nil {
		writeError(responseWriter, http.StatusBadRequest, "invalid json")
		return
	}

	_, useCaseError := handler.tagUseCase.UpdateTagScaling(
		request.Context(),
		tagID,
		domain.TagScaling{
			RawMin: requestPayload.RawMin,
			RawMax: requestPayload.RawMax,
			EngMin: requestPayload.EngMin,
			EngMax: requestPayload.EngMax,
			Factor: requestPayload.Factor,
			Offset: requestPayload.Offset,
		},
	)
	if useCaseError != nil {
		writeDomainError(responseWriter, useCaseError)
		return
	}

	tagFull, getError := handler.tagUseCase.GetTag(request.Context(), tagID)
	if getError != nil {
		writeDomainError(responseWriter, getError)
		return
	}

	writeJSON(responseWriter, http.StatusOK, mapTagFullResponse(tagFull))
}

func (handler *Handler) deleteTagScaling(
	responseWriter http.ResponseWriter,
	request *http.Request,
	tagID string,
) {
	useCaseError := handler.tagUseCase.DeleteTagScaling(request.Context(), tagID)
	if useCaseError != nil {
		writeDomainError(responseWriter, useCaseError)
		return
	}

	tagFull, getError := handler.tagUseCase.GetTag(request.Context(), tagID)
	if getError != nil {
		writeDomainError(responseWriter, getError)
		return
	}

	writeJSON(responseWriter, http.StatusOK, mapTagFullResponse(tagFull))
}

func (handler *Handler) listDataTypes(
	responseWriter http.ResponseWriter,
	request *http.Request,
) {
	dataTypes, useCaseError := handler.tagUseCase.ListDataTypes(request.Context())
	if useCaseError != nil {
		writeDomainError(responseWriter, useCaseError)
		return
	}

	responsePayload := make([]dataTypeResponse, 0, len(dataTypes))
	for _, dataType := range dataTypes {
		responsePayload = append(responsePayload, dataTypeResponse{
			ID:   dataType.ID,
			Name: dataType.Name,
		})
	}

	writeJSON(responseWriter, http.StatusOK, responsePayload)
}

func (handler *Handler) listUnits(
	responseWriter http.ResponseWriter,
	request *http.Request,
) {
	units, useCaseError := handler.tagUseCase.ListUnits(
		request.Context(),
		request.URL.Query().Get("category"),
	)
	if useCaseError != nil {
		writeDomainError(responseWriter, useCaseError)
		return
	}

	responsePayload := make([]unitResponse, 0, len(units))
	for _, unit := range units {
		responsePayload = append(responsePayload, unitResponse{
			ID:       unit.ID,
			Name:     unit.Name,
			Symbol:   unit.Symbol,
			Category: unit.Category,
		})
	}

	writeJSON(responseWriter, http.StatusOK, responsePayload)
}

func mapMonitoringObjectResponse(
	monitoringObject domain.MonitoringObject,
) monitoringObjectResponse {
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
		FigureType: string(figure.FigureType),
		Parameters: figure.Parameters,
		CreatedAt:  figure.CreatedAt,
		UpdatedAt:  figure.UpdatedAt,
	}
}

func mapDeviceWithParamsResponse(
	deviceWithParams domain.DeviceWithParams,
) deviceResponse {
	var settings json.RawMessage
	if deviceWithParams.Params != nil {
		settings = deviceWithParams.Params.Settings
	}

	return mapDeviceResponse(deviceWithParams.Device, settings)
}

func mapDeviceResponse(device domain.Device, settings json.RawMessage) deviceResponse {
	return deviceResponse{
		ID:          device.ID,
		ObjectID:    device.ObjectID,
		TypeID:      device.TypeID,
		TypeName:    device.TypeName,
		Name:        device.Name,
		Description: device.Description,
		Settings:    settings,
		CreatedAt:   device.CreatedAt,
		UpdatedAt:   device.UpdatedAt,
	}
}

func mapTagFullResponse(tagFull domain.TagFull) tagResponse {
	return tagResponse{
		ID:          tagFull.Tag.ID,
		DeviceID:    tagFull.Tag.DeviceID,
		Name:        tagFull.Tag.Name,
		Description: tagFull.Tag.Description,
		Params:      mapTagParamsResponse(tagFull.Params),
		Setpoints:   mapTagSetpointsResponse(tagFull.Setpoints),
		Scaling:     mapTagScalingResponse(tagFull.Scaling),
		CreatedAt:   tagFull.Tag.CreatedAt,
		UpdatedAt:   tagFull.Tag.UpdatedAt,
	}
}

func mapTagResponse(tag domain.Tag) tagResponse {
	return tagResponse{
		ID:          tag.ID,
		DeviceID:    tag.DeviceID,
		Name:        tag.Name,
		Description: tag.Description,
		CreatedAt:   tag.CreatedAt,
		UpdatedAt:   tag.UpdatedAt,
	}
}

func mapTagParamsResponse(tagParams *domain.TagParams) *tagParamsResponse {
	if tagParams == nil {
		return nil
	}

	return &tagParamsResponse{
		ID:         tagParams.ID,
		DataTypeID: tagParams.DataTypeID,
		UnitID:     tagParams.UnitID,
		Address:    tagParams.Address,
		CreatedAt:  tagParams.CreatedAt,
		UpdatedAt:  tagParams.UpdatedAt,
	}
}

func mapTagSetpointsResponse(setpoints *domain.TagSetpoints) *setpointsInput {
	if setpoints == nil {
		return nil
	}

	return &setpointsInput{
		LoLo: setpoints.LoLo,
		Lo:   setpoints.Lo,
		Hi:   setpoints.Hi,
		HiHi: setpoints.HiHi,
	}
}

func mapTagScalingResponse(scaling *domain.TagScaling) *scalingInput {
	if scaling == nil {
		return nil
	}

	return &scalingInput{
		RawMin: scaling.RawMin,
		RawMax: scaling.RawMax,
		EngMin: scaling.EngMin,
		EngMax: scaling.EngMax,
		Factor: scaling.Factor,
		Offset: scaling.Offset,
	}
}

func mapTagParamsInput(input *tagParamsInput) *domain.TagParams {
	if input == nil {
		return nil
	}

	return &domain.TagParams{
		DataTypeID: input.DataTypeID,
		UnitID:     input.UnitID,
		Address:    input.Address,
	}
}

func mapSetpoints(input *setpointsInput) *domain.TagSetpoints {
	if input == nil {
		return nil
	}

	return &domain.TagSetpoints{
		LoLo: input.LoLo,
		Lo:   input.Lo,
		Hi:   input.Hi,
		HiHi: input.HiHi,
	}
}

func mapScaling(input *scalingInput) *domain.TagScaling {
	if input == nil {
		return nil
	}

	return &domain.TagScaling{
		RawMin: input.RawMin,
		RawMax: input.RawMax,
		EngMin: input.EngMin,
		EngMax: input.EngMax,
		Factor: input.Factor,
		Offset: input.Offset,
	}
}

func intPointer(value int) *int {
	return &value
}

func jsonRawMessagePointer(value json.RawMessage) *json.RawMessage {
	return &value
}

func mapListResponse(items any, total int, offset int, limit int) listResponse {
	return listResponse{
		Items:  items,
		Total:  total,
		Offset: offset,
		Limit:  limit,
	}
}

func parseObjectListQuery(request *http.Request) (domain.ObjectListQuery, error) {
	pagination, parseError := parsePagination(request)
	if parseError != nil {
		return domain.ObjectListQuery{}, parseError
	}

	return domain.ObjectListQuery{
		Pagination: pagination,
		Search:     request.URL.Query().Get("search"),
	}, nil
}

func parseDeviceListQuery(
	request *http.Request,
	pathObjectID *string,
) (domain.DeviceListQuery, error) {
	pagination, parseError := parsePagination(request)
	if parseError != nil {
		return domain.DeviceListQuery{}, parseError
	}

	objectID := pathObjectID
	if objectID == nil {
		objectID = stringPointerOrNil(request.URL.Query().Get("object_id"))
	}

	typeID, parseTypeError := parseOptionalInt(request.URL.Query().Get("type_id"))
	if parseTypeError != nil {
		return domain.DeviceListQuery{}, parseTypeError
	}

	return domain.DeviceListQuery{
		Pagination: pagination,
		ObjectID:   objectID,
		TypeID:     typeID,
		Search:     request.URL.Query().Get("search"),
	}, nil
}

func parseTagListQuery(
	request *http.Request,
	pathDeviceID *string,
) (domain.TagListQuery, error) {
	pagination, parseError := parsePagination(request)
	if parseError != nil {
		return domain.TagListQuery{}, parseError
	}

	deviceID := pathDeviceID
	if deviceID == nil {
		deviceID = stringPointerOrNil(request.URL.Query().Get("device_id"))
	}

	objectID := stringPointerOrNil(request.URL.Query().Get("object_id"))

	dataTypeID, parseTypeError := parseOptionalInt(request.URL.Query().Get("data_type_id"))
	if parseTypeError != nil {
		return domain.TagListQuery{}, parseTypeError
	}

	unitID, parseUnitError := parseOptionalInt(request.URL.Query().Get("unit_id"))
	if parseUnitError != nil {
		return domain.TagListQuery{}, parseUnitError
	}

	return domain.TagListQuery{
		Pagination: pagination,
		ObjectID:   objectID,
		DeviceID:   deviceID,
		DataTypeID: dataTypeID,
		UnitID:     unitID,
		Search:     request.URL.Query().Get("search"),
	}, nil
}

func parseDiagramListQuery(
	request *http.Request,
	pathObjectID *string,
) (domain.DiagramListQuery, error) {
	pagination, parseError := parsePagination(request)
	if parseError != nil {
		return domain.DiagramListQuery{}, parseError
	}

	objectID := pathObjectID
	if objectID == nil {
		objectID = stringPointerOrNil(request.URL.Query().Get("object_id"))
	}

	return domain.DiagramListQuery{
		Pagination: pagination,
		ObjectID:   objectID,
		Search:     request.URL.Query().Get("search"),
	}, nil
}

func parseFigureListQuery(
	request *http.Request,
	diagramID string,
) (domain.FigureListQuery, error) {
	pagination, parseError := parsePagination(request)
	if parseError != nil {
		return domain.FigureListQuery{}, parseError
	}

	typeFilter := stringPointerOrNil(request.URL.Query().Get("type"))
	return domain.FigureListQuery{
		Pagination: pagination,
		DiagramID:  diagramID,
		TypeFilter: typeFilter,
	}, nil
}

func parsePagination(request *http.Request) (domain.Pagination, error) {
	offset, parseOffsetError := parseIntWithDefault(request.URL.Query().Get("offset"), 0)
	if parseOffsetError != nil {
		return domain.Pagination{}, parseOffsetError
	}

	limit, parseLimitError := parseIntWithDefault(request.URL.Query().Get("limit"), 50)
	if parseLimitError != nil {
		return domain.Pagination{}, parseLimitError
	}

	if offset < 0 || limit < 0 {
		return domain.Pagination{}, domain.ErrInvalidInput
	}

	if limit > 200 {
		limit = 200
	}

	return domain.Pagination{
		Offset: offset,
		Limit:  limit,
	}, nil
}

func parseIntWithDefault(rawValue string, defaultValue int) (int, error) {
	if strings.TrimSpace(rawValue) == "" {
		return defaultValue, nil
	}

	parsedValue, parseError := strconv.Atoi(rawValue)
	if parseError != nil {
		return 0, domain.ErrInvalidInput
	}

	return parsedValue, nil
}

func parseOptionalInt(rawValue string) (*int, error) {
	if strings.TrimSpace(rawValue) == "" {
		return nil, nil
	}

	parsedValue, parseError := strconv.Atoi(rawValue)
	if parseError != nil {
		return nil, domain.ErrInvalidInput
	}

	return &parsedValue, nil
}

func stringPointerOrNil(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	return &value
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
	writeProblem(responseWriter, mapDomainError(domainError))
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}

	return &value
}
