package http

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
)

type TagsHandler struct {
	tagUseCase TagUseCase
}

func NewTagsHandler(tagUseCase TagUseCase) *TagsHandler {
	return &TagsHandler{tagUseCase: tagUseCase}
}

func (handler *TagsHandler) GetTags(
	responseWriter http.ResponseWriter,
	request *http.Request,
) {
	query := domain.TagSearchQuery{
		Search: strings.TrimSpace(request.URL.Query().Get("search")),
		Limit:  20,
	}

	if limitRaw := strings.TrimSpace(request.URL.Query().Get("limit")); limitRaw != "" {
		limit, parseError := strconv.Atoi(limitRaw)
		if parseError != nil {
			writeDomainError(
				responseWriter,
				fmt.Errorf("invalid limit: %w", domain.ErrInvalidInput),
			)
			return
		}
		query.Limit = limit
	}

	result, useCaseError := handler.tagUseCase.SearchTags(request.Context(), query)
	if useCaseError != nil {
		writeDomainError(responseWriter, useCaseError)
		return
	}

	writeJSON(responseWriter, http.StatusOK, mapTagsResponse(result))
}
