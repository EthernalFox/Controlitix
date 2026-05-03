package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
	"github.com/EthernalFox/Controlitix/shared/authctx"
)

type AlarmsHandler struct {
	alarmUseCase AlarmUseCase
	auditUseCase AuditRecorder
}

func NewAlarmsHandler(alarmUseCase AlarmUseCase, auditUseCase AuditRecorder) *AlarmsHandler {
	return &AlarmsHandler{
		alarmUseCase: alarmUseCase,
		auditUseCase: auditUseCase,
	}
}

func (handler *AlarmsHandler) GetAlarms(responseWriter http.ResponseWriter, request *http.Request) {
	if handler.alarmUseCase == nil {
		writeProblem(responseWriter, Problem{
			Type:   "/problems/internal-error",
			Title:  "Internal error",
			Status: http.StatusInternalServerError,
			Detail: "alarm use case is not configured",
		})
		return
	}

	query, parseError := parseAlarmListQuery(request)
	if parseError != nil {
		writeAlarmProblem(responseWriter, parseError)
		return
	}

	result, useCaseError := handler.alarmUseCase.ListAlarms(request.Context(), query)
	if useCaseError != nil {
		writeAlarmProblem(responseWriter, useCaseError)
		return
	}

	writeJSON(responseWriter, http.StatusOK, mapAlarmsListResponse(result))
}

func (handler *AlarmsHandler) GetAlarm(responseWriter http.ResponseWriter, request *http.Request) {
	if handler.alarmUseCase == nil {
		writeProblem(responseWriter, Problem{
			Type:   "/problems/internal-error",
			Title:  "Internal error",
			Status: http.StatusInternalServerError,
			Detail: "alarm use case is not configured",
		})
		return
	}

	tagID, parseError := uuid.Parse(strings.TrimSpace(chi.URLParam(request, "tagId")))
	if parseError != nil {
		writeAlarmProblem(responseWriter, fmt.Errorf("invalid tag id: %w", domain.ErrInvalidInput))
		return
	}

	detail, useCaseError := handler.alarmUseCase.GetAlarm(request.Context(), tagID)
	if useCaseError != nil {
		writeAlarmProblem(responseWriter, useCaseError)
		return
	}

	writeJSON(responseWriter, http.StatusOK, mapAlarmDetailResponse(detail))
}

func (handler *AlarmsHandler) Acknowledge(responseWriter http.ResponseWriter, request *http.Request) {
	if handler.alarmUseCase == nil {
		writeProblem(responseWriter, Problem{
			Type:   "/problems/internal-error",
			Title:  "Internal error",
			Status: http.StatusInternalServerError,
			Detail: "alarm use case is not configured",
		})
		return
	}

	tagID, parseError := uuid.Parse(strings.TrimSpace(chi.URLParam(request, "tagId")))
	if parseError != nil {
		writeAlarmProblem(responseWriter, fmt.Errorf("invalid tag id: %w", domain.ErrInvalidInput))
		return
	}

	var body acknowledgeRequest
	if request.Body != nil {
		decoder := json.NewDecoder(request.Body)
		decoder.DisallowUnknownFields()
		if decodeError := decoder.Decode(&body); decodeError != nil &&
			!errors.Is(decodeError, io.EOF) &&
			!errors.Is(decodeError, http.ErrBodyReadAfterClose) {
			writeAlarmProblem(responseWriter, fmt.Errorf("invalid request body: %w", domain.ErrInvalidInput))
			return
		}
	}

	principal, ok := authctx.FromContext(request.Context())
	if !ok || strings.TrimSpace(principal.Subject) == "" {
		writeProblem(responseWriter, Problem{
			Type:   "/errors/auth/invalid-token",
			Title:  "Invalid token",
			Status: http.StatusUnauthorized,
		})
		return
	}

	result, acknowledgeError := handler.alarmUseCase.Acknowledge(
		request.Context(),
		tagID,
		principal.Subject,
		body.Note,
	)
	if acknowledgeError != nil {
		writeAlarmProblem(responseWriter, acknowledgeError)
		return
	}
	if handler.auditUseCase != nil {
		handler.auditUseCase.Record(request.Context(), domain.AuditEvent{
			Action: "alarm.acknowledged",
			Target: domain.AuditTarget{
				Type:  "alarm",
				ID:    result.TagID.String(),
				State: result.State.String(),
			},
			Details: map[string]any{
				"note": body.Note,
			},
			Result: domain.AuditResultSuccess,
		})
	}

	response := alarmRecordResponse{
		TagID: result.TagID.String(),
		State: result.State.String(),
		Acked: true,
		Ack: &alarmAckDTOResponse{
			ActorID: result.Ack.ActorID,
			AckedAt: result.Ack.AckedAt.UTC().Format(time.RFC3339Nano),
			Note:    result.Ack.Note,
		},
	}
	writeJSON(responseWriter, http.StatusOK, response)
}

func (handler *AlarmsHandler) AcknowledgeBulk(responseWriter http.ResponseWriter, request *http.Request) {
	if handler.alarmUseCase == nil {
		writeProblem(responseWriter, Problem{
			Type:   "/problems/internal-error",
			Title:  "Internal error",
			Status: http.StatusInternalServerError,
			Detail: "alarm use case is not configured",
		})
		return
	}

	var body bulkAcknowledgeRequest
	if request.Body != nil {
		decoder := json.NewDecoder(request.Body)
		decoder.DisallowUnknownFields()
		if decodeError := decoder.Decode(&body); decodeError != nil &&
			!errors.Is(decodeError, io.EOF) &&
			!errors.Is(decodeError, http.ErrBodyReadAfterClose) {
			writeAlarmProblem(responseWriter, fmt.Errorf("invalid request body: %w", domain.ErrInvalidInput))
			return
		}
	}

	if body.Note != nil && len([]rune(strings.TrimSpace(*body.Note))) > 500 {
		writeProblem(responseWriter, Problem{
			Type:   "/errors/alarms/note-too-long",
			Title:  "Note is too long",
			Status: http.StatusUnprocessableEntity,
			Detail: "note must be 500 characters or less",
		})
		return
	}

	tagIDs := make([]uuid.UUID, 0, len(body.Items))
	seen := make(map[uuid.UUID]struct{}, len(body.Items))
	for _, item := range body.Items {
		tagID, parseError := uuid.Parse(strings.TrimSpace(item.TagID))
		if parseError != nil {
			writeAlarmProblem(responseWriter, fmt.Errorf("invalid tag id: %w", domain.ErrInvalidInput))
			return
		}
		if _, exists := seen[tagID]; exists {
			continue
		}
		seen[tagID] = struct{}{}
		tagIDs = append(tagIDs, tagID)
	}

	if len(tagIDs) == 0 || len(tagIDs) > 200 {
		writeProblem(responseWriter, Problem{
			Type:   "/errors/alarms/bulk-size",
			Title:  "Invalid bulk size",
			Status: http.StatusUnprocessableEntity,
			Detail: "items count must be between 1 and 200",
		})
		return
	}

	principal, ok := authctx.FromContext(request.Context())
	if !ok || strings.TrimSpace(principal.Subject) == "" {
		writeProblem(responseWriter, Problem{
			Type:   "/errors/auth/invalid-token",
			Title:  "Invalid token",
			Status: http.StatusUnauthorized,
		})
		return
	}

	result, acknowledgeError := handler.alarmUseCase.AcknowledgeBulk(
		request.Context(),
		tagIDs,
		principal.Subject,
		body.Note,
	)
	if acknowledgeError != nil {
		writeAlarmProblem(responseWriter, acknowledgeError)
		return
	}

	statusCode := http.StatusOK
	if result.SuccessN == 0 && result.FailedN > 0 {
		statusCode = http.StatusConflict
	} else if result.SuccessN > 0 && result.FailedN > 0 {
		statusCode = http.StatusMultiStatus
	}

	writeJSON(responseWriter, statusCode, mapBulkAcknowledgeResponse(result))
}

func parseAlarmListQuery(request *http.Request) (domain.AlarmListQuery, error) {
	status := domain.ParseAlarmListStatus(request.URL.Query().Get("status"))
	severity := domain.ParseAlarmSeverity(request.URL.Query().Get("severity"))

	objectID := uuid.Nil
	objectRaw := strings.TrimSpace(request.URL.Query().Get("object_id"))
	if objectRaw != "" {
		parsedObjectID, parseObjectIDError := uuid.Parse(objectRaw)
		if parseObjectIDError != nil {
			return domain.AlarmListQuery{}, fmt.Errorf("invalid object_id: %w", domain.ErrInvalidInput)
		}
		objectID = parsedObjectID
	}

	var from *time.Time
	fromRaw := strings.TrimSpace(request.URL.Query().Get("from"))
	if fromRaw != "" {
		parsedFrom, parseFromError := time.Parse(time.RFC3339, fromRaw)
		if parseFromError != nil {
			return domain.AlarmListQuery{}, fmt.Errorf("invalid from: %w", domain.ErrInvalidInput)
		}
		parsedFrom = parsedFrom.UTC()
		from = &parsedFrom
	}

	var to *time.Time
	toRaw := strings.TrimSpace(request.URL.Query().Get("to"))
	if toRaw != "" {
		parsedTo, parseToError := time.Parse(time.RFC3339, toRaw)
		if parseToError != nil {
			return domain.AlarmListQuery{}, fmt.Errorf("invalid to: %w", domain.ErrInvalidInput)
		}
		parsedTo = parsedTo.UTC()
		to = &parsedTo
	}

	limit := 50
	limitRaw := strings.TrimSpace(request.URL.Query().Get("limit"))
	if limitRaw != "" {
		parsedLimit, parseLimitError := strconv.Atoi(limitRaw)
		if parseLimitError != nil {
			return domain.AlarmListQuery{}, fmt.Errorf("invalid limit: %w", domain.ErrInvalidInput)
		}
		limit = parsedLimit
	}

	offset := 0
	offsetRaw := strings.TrimSpace(request.URL.Query().Get("offset"))
	if offsetRaw != "" {
		parsedOffset, parseOffsetError := strconv.Atoi(offsetRaw)
		if parseOffsetError != nil {
			return domain.AlarmListQuery{}, fmt.Errorf("invalid offset: %w", domain.ErrInvalidInput)
		}
		offset = parsedOffset
	}

	return domain.AlarmListQuery{
		Status:   status,
		Severity: severity,
		ObjectID: objectID,
		From:     from,
		To:       to,
		Limit:    limit,
		Offset:   offset,
	}, nil
}

func writeAlarmProblem(responseWriter http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		writeProblem(responseWriter, Problem{
			Type:   "/errors/alarms/not-found",
			Title:  "Alarm not found",
			Status: http.StatusNotFound,
			Detail: trimSentinelSuffix(err, domain.ErrNotFound),
		})
	case errors.Is(err, domain.ErrAlarmNotActive):
		writeProblem(responseWriter, Problem{
			Type:   "/errors/alarms/not-active",
			Title:  "Alarm is not active",
			Status: http.StatusConflict,
			Detail: trimSentinelSuffix(err, domain.ErrAlarmNotActive),
		})
	case errors.Is(err, domain.ErrAlarmAlreadyAcked):
		writeProblem(responseWriter, Problem{
			Type:   "/errors/alarms/already-acked",
			Title:  "Alarm already acknowledged",
			Status: http.StatusConflict,
			Detail: trimSentinelSuffix(err, domain.ErrAlarmAlreadyAcked),
		})
	case errors.Is(err, domain.ErrInvalidInput):
		writeProblem(responseWriter, Problem{
			Type:   "/errors/alarms/invalid-request",
			Title:  "Invalid request",
			Status: http.StatusBadRequest,
			Detail: trimSentinelSuffix(err, domain.ErrInvalidInput),
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
