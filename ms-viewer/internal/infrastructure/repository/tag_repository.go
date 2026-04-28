package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
)

type TagRepository struct {
	pool *pgxpool.Pool
}

func NewTagRepository(pool *pgxpool.Pool) *TagRepository {
	return &TagRepository{pool: pool}
}

func (repository *TagRepository) SearchTags(
	ctx context.Context,
	query domain.TagSearchQuery,
) (domain.TagSearchResult, error) {
	search := "%" + strings.TrimSpace(query.Search) + "%"

	const countQuery = `
SELECT count(*)
FROM tags.tags t
WHERE t.deleted_at IS NULL
  AND t.name ILIKE $1
`

	var total int
	if countError := repository.pool.QueryRow(ctx, countQuery, search).Scan(&total); countError != nil {
		return domain.TagSearchResult{}, fmt.Errorf("count tags: %w", countError)
	}

	const itemsQuery = `
SELECT
    t.id,
    t.name,
    t.device_id,
    d.name,
    COALESCE(u.symbol, '') AS unit_symbol
FROM tags.tags t
JOIN devices.devices d ON d.id = t.device_id
LEFT JOIN tags.units u ON u.id = t.unit_id
WHERE t.deleted_at IS NULL
  AND t.name ILIKE $1
ORDER BY t.name ASC
LIMIT $2
`

	rows, queryError := repository.pool.Query(ctx, itemsQuery, search, query.Limit)
	if queryError != nil {
		return domain.TagSearchResult{}, fmt.Errorf("query tags: %w", queryError)
	}
	defer rows.Close()

	items := make([]domain.TagSearchItem, 0)
	for rows.Next() {
		var item domain.TagSearchItem
		if scanError := rows.Scan(
			&item.ID,
			&item.Name,
			&item.DeviceID,
			&item.DeviceName,
			&item.UnitSymbol,
		); scanError != nil {
			return domain.TagSearchResult{}, fmt.Errorf("scan tag: %w", scanError)
		}
		items = append(items, item)
	}

	if rowsError := rows.Err(); rowsError != nil {
		return domain.TagSearchResult{}, fmt.Errorf("iterate tags: %w", rowsError)
	}

	return domain.TagSearchResult{
		Items: items,
		Total: total,
	}, nil
}
