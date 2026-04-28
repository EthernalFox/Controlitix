package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
)

type WSTopicRepository struct {
	pool *pgxpool.Pool
}

func NewWSTopicRepository(pool *pgxpool.Pool) *WSTopicRepository {
	return &WSTopicRepository{pool: pool}
}

func (repository *WSTopicRepository) ResolveDiagramTagIDs(
	ctx context.Context,
	diagramID uuid.UUID,
	limit int,
) ([]uuid.UUID, error) {
	const existsQuery = `
SELECT 1
FROM public.mimic
WHERE id = $1
  AND deleted_at IS NULL
`

	var marker int
	scanError := repository.pool.QueryRow(ctx, existsQuery, diagramID).Scan(&marker)
	if scanError != nil {
		if errors.Is(scanError, pgx.ErrNoRows) {
			return nil, fmt.Errorf("diagram %s not found: %w", diagramID.String(), domain.ErrNotFound)
		}
		return nil, fmt.Errorf("check diagram exists: %w", scanError)
	}

	if limit <= 0 {
		limit = 500
	}

	const query = `
SELECT DISTINCT f.tag_id
FROM public.figures f
WHERE f.diagram_id = $1
  AND f.deleted_at IS NULL
  AND f.tag_id IS NOT NULL
ORDER BY f.tag_id ASC
LIMIT $2
`

	rows, queryError := repository.pool.Query(ctx, query, diagramID, limit)
	if queryError != nil {
		return nil, fmt.Errorf("query diagram tag ids: %w", queryError)
	}
	defer rows.Close()

	tagIDs := make([]uuid.UUID, 0)
	for rows.Next() {
		var tagID uuid.UUID
		if scanTagError := rows.Scan(&tagID); scanTagError != nil {
			return nil, fmt.Errorf("scan diagram tag id: %w", scanTagError)
		}
		tagIDs = append(tagIDs, tagID)
	}
	if rowsError := rows.Err(); rowsError != nil {
		return nil, fmt.Errorf("iterate diagram tag ids: %w", rowsError)
	}

	return tagIDs, nil
}

func (repository *WSTopicRepository) ResolveObjectTagIDs(
	ctx context.Context,
	objectID uuid.UUID,
	limit int,
) ([]uuid.UUID, error) {
	const existsQuery = `
SELECT 1
FROM public.objects
WHERE id = $1
  AND deleted_at IS NULL
`

	var marker int
	scanError := repository.pool.QueryRow(ctx, existsQuery, objectID).Scan(&marker)
	if scanError != nil {
		if errors.Is(scanError, pgx.ErrNoRows) {
			return nil, fmt.Errorf("object %s not found: %w", objectID.String(), domain.ErrNotFound)
		}
		return nil, fmt.Errorf("check object exists: %w", scanError)
	}

	if limit <= 0 {
		limit = 500
	}

	const query = `
SELECT t.id
FROM tags.tags t
JOIN devices.devices d ON d.id = t.device_id
WHERE d.object_id = $1
  AND t.deleted_at IS NULL
  AND d.deleted_at IS NULL
ORDER BY t.id ASC
LIMIT $2
`

	rows, queryError := repository.pool.Query(ctx, query, objectID, limit)
	if queryError != nil {
		return nil, fmt.Errorf("query object tag ids: %w", queryError)
	}
	defer rows.Close()

	tagIDs := make([]uuid.UUID, 0)
	for rows.Next() {
		var tagID uuid.UUID
		if scanTagError := rows.Scan(&tagID); scanTagError != nil {
			return nil, fmt.Errorf("scan object tag id: %w", scanTagError)
		}
		tagIDs = append(tagIDs, tagID)
	}
	if rowsError := rows.Err(); rowsError != nil {
		return nil, fmt.Errorf("iterate object tag ids: %w", rowsError)
	}

	return tagIDs, nil
}
