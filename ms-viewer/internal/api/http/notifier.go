package http

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
)

type NotifierHandler struct {
	notifierUseCase NotifierUseCase
}

func NewNotifierHandler(notifierUseCase NotifierUseCase) *NotifierHandler {
	return &NotifierHandler{notifierUseCase: notifierUseCase}
}

func (handler *NotifierHandler) GetChats(responseWriter http.ResponseWriter, request *http.Request) {
	if handler.notifierUseCase == nil {
		writeProblem(responseWriter, Problem{
			Type:   "/problems/internal-error",
			Title:  "Internal error",
			Status: http.StatusInternalServerError,
			Detail: "notifier use case is not configured",
		})
		return
	}

	filter, parseError := parseTelegramChatFilter(request)
	if parseError != nil {
		writeProblem(responseWriter, Problem{
			Type:   "/errors/notifier/invalid-request",
			Title:  "Invalid request",
			Status: http.StatusBadRequest,
			Detail: parseError.Error(),
		})
		return
	}

	result, useCaseError := handler.notifierUseCase.ListChats(request.Context(), filter)
	if useCaseError != nil {
		writeProblem(responseWriter, Problem{
			Type:   "/problems/internal-error",
			Title:  "Internal error",
			Status: http.StatusInternalServerError,
			Detail: "internal error",
		})
		return
	}

	writeJSON(responseWriter, http.StatusOK, mapNotifierChatsResponse(result))
}

func parseTelegramChatFilter(request *http.Request) (domain.TelegramChatFilter, error) {
	queryValues := request.URL.Query()
	filter := domain.TelegramChatFilter{}

	enabledRaw := strings.TrimSpace(queryValues.Get("enabled"))
	if enabledRaw != "" {
		enabled, parseError := strconv.ParseBool(enabledRaw)
		if parseError != nil {
			return domain.TelegramChatFilter{}, fmt.Errorf("invalid enabled")
		}
		filter.Enabled = &enabled
	}

	filter.Role = strings.TrimSpace(queryValues.Get("role"))

	objectIDRaw := strings.TrimSpace(queryValues.Get("object_id"))
	if objectIDRaw != "" {
		objectID, parseError := uuid.Parse(objectIDRaw)
		if parseError != nil {
			return domain.TelegramChatFilter{}, fmt.Errorf("invalid object_id")
		}
		filter.ObjectID = objectID
	}

	return filter, nil
}
