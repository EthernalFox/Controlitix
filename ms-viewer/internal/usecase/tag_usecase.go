package usecase

import (
	"context"
	"fmt"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
)

const maxTagSearchLimit = 20

type TagUseCase struct {
	tagRepository domain.TagRepository
}

func NewTagUseCase(tagRepository domain.TagRepository) *TagUseCase {
	return &TagUseCase{
		tagRepository: tagRepository,
	}
}

func (useCase *TagUseCase) SearchTags(
	ctx context.Context,
	query domain.TagSearchQuery,
) (domain.TagSearchResult, error) {
	normalizedQuery := query
	if normalizedQuery.Limit <= 0 {
		normalizedQuery.Limit = maxTagSearchLimit
	}
	if normalizedQuery.Limit > maxTagSearchLimit {
		return domain.TagSearchResult{}, fmt.Errorf("limit exceeds %d: %w", maxTagSearchLimit, domain.ErrInvalidInput)
	}

	return useCase.tagRepository.SearchTags(ctx, normalizedQuery)
}
