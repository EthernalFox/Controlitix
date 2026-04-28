package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
)

type DiagramRepository struct {
	pool *pgxpool.Pool
}

func NewDiagramRepository(pool *pgxpool.Pool) *DiagramRepository {
	return &DiagramRepository{pool: pool}
}

func (repository *DiagramRepository) ListObjectsWithPublishedDiagrams(
	ctx context.Context,
	query domain.ObjectListQuery,
) (domain.ObjectListResult, error) {
	const countQuery = `
SELECT count(*)
FROM (
    SELECT o.id
    FROM public.objects o
    JOIN public.mimic m ON m.object_id = o.id
    WHERE o.deleted_at IS NULL
      AND m.deleted_at IS NULL
      AND m.published_at IS NOT NULL
    GROUP BY o.id
) objects_with_diagrams
`

	var total int
	if countError := repository.pool.QueryRow(ctx, countQuery).Scan(&total); countError != nil {
		return domain.ObjectListResult{}, fmt.Errorf("count objects with published diagrams: %w", countError)
	}

	const itemsQuery = `
SELECT
    o.id,
    o.name,
    COALESCE(o.description, '') AS description,
    count(m.id) AS published_diagram_count,
    (array_agg(m.id ORDER BY m.name ASC, m.id ASC))[1] AS first_published_diagram_id
FROM public.objects o
JOIN public.mimic m ON m.object_id = o.id
WHERE o.deleted_at IS NULL
  AND m.deleted_at IS NULL
  AND m.published_at IS NOT NULL
GROUP BY o.id, o.name, o.description
ORDER BY o.name ASC, o.id ASC
LIMIT $1
OFFSET $2
`

	rows, queryError := repository.pool.Query(ctx, itemsQuery, query.Limit, query.Offset)
	if queryError != nil {
		return domain.ObjectListResult{}, fmt.Errorf("query objects with published diagrams: %w", queryError)
	}
	defer rows.Close()

	items := make([]domain.ObjectListItem, 0)
	for rows.Next() {
		var item domain.ObjectListItem
		if scanError := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Description,
			&item.PublishedDiagramCount,
			&item.FirstPublishedDiagramID,
		); scanError != nil {
			return domain.ObjectListResult{}, fmt.Errorf("scan object with diagrams: %w", scanError)
		}

		items = append(items, item)
	}
	if rowsError := rows.Err(); rowsError != nil {
		return domain.ObjectListResult{}, fmt.Errorf("iterate objects with diagrams: %w", rowsError)
	}

	return domain.ObjectListResult{
		Items:  items,
		Total:  total,
		Limit:  query.Limit,
		Offset: query.Offset,
	}, nil
}

func (repository *DiagramRepository) ListPublishedDiagrams(
	ctx context.Context,
	query domain.DiagramListQuery,
) (domain.DiagramListResult, error) {
	const countQuery = `
SELECT count(*)
FROM public.mimic m
WHERE m.object_id = $1
  AND m.deleted_at IS NULL
  AND m.published_at IS NOT NULL
`

	var total int
	if countError := repository.pool.QueryRow(ctx, countQuery, query.ObjectID).Scan(&total); countError != nil {
		return domain.DiagramListResult{}, fmt.Errorf("count published diagrams: %w", countError)
	}

	const itemsQuery = `
SELECT
    m.id,
    m.object_id,
    COALESCE(m.name, '') AS name,
    COALESCE(m.description, '') AS description,
    m.published_at,
    (
      SELECT count(*)
      FROM public.figures f
      WHERE f.diagram_id = m.id
        AND f.deleted_at IS NULL
    ) AS figure_count,
    (
      SELECT count(DISTINCT f.tag_id)
      FROM public.figures f
      WHERE f.diagram_id = m.id
        AND f.deleted_at IS NULL
        AND f.tag_id IS NOT NULL
    ) AS bound_tag_count
FROM public.mimic m
WHERE m.object_id = $1
  AND m.deleted_at IS NULL
  AND m.published_at IS NOT NULL
ORDER BY m.name ASC, m.id ASC
LIMIT $2
OFFSET $3
`

	rows, queryError := repository.pool.Query(
		ctx,
		itemsQuery,
		query.ObjectID,
		query.Limit,
		query.Offset,
	)
	if queryError != nil {
		return domain.DiagramListResult{}, fmt.Errorf("query published diagrams: %w", queryError)
	}
	defer rows.Close()

	items := make([]domain.DiagramListItem, 0)
	for rows.Next() {
		var item domain.DiagramListItem
		if scanError := rows.Scan(
			&item.ID,
			&item.ObjectID,
			&item.Name,
			&item.Description,
			&item.PublishedAt,
			&item.FigureCount,
			&item.BoundTagCount,
		); scanError != nil {
			return domain.DiagramListResult{}, fmt.Errorf("scan published diagram: %w", scanError)
		}

		items = append(items, item)
	}
	if rowsError := rows.Err(); rowsError != nil {
		return domain.DiagramListResult{}, fmt.Errorf("iterate published diagrams: %w", rowsError)
	}

	return domain.DiagramListResult{
		Items:  items,
		Total:  total,
		Limit:  query.Limit,
		Offset: query.Offset,
	}, nil
}

func (repository *DiagramRepository) GetPublishedDiagram(
	ctx context.Context,
	diagramID uuid.UUID,
) (domain.Diagram, error) {
	const diagramQuery = `
SELECT
    m.id,
    m.object_id,
    COALESCE(o.name, '') AS object_name,
    COALESCE(m.name, '') AS name,
    COALESCE(m.description, '') AS description,
    m.published_at
FROM public.mimic m
JOIN public.objects o ON o.id = m.object_id
WHERE m.id = $1
  AND m.deleted_at IS NULL
`

	var diagram domain.Diagram
	var publishedAt *time.Time
	if scanError := repository.pool.QueryRow(ctx, diagramQuery, diagramID).Scan(
		&diagram.ID,
		&diagram.ObjectID,
		&diagram.ObjectName,
		&diagram.Name,
		&diagram.Description,
		&publishedAt,
	); scanError != nil {
		if errors.Is(scanError, pgx.ErrNoRows) {
			return domain.Diagram{}, fmt.Errorf("diagram %s not found: %w", diagramID.String(), domain.ErrNotFound)
		}

		return domain.Diagram{}, fmt.Errorf("query diagram: %w", scanError)
	}

	if publishedAt == nil || publishedAt.IsZero() {
		return domain.Diagram{}, fmt.Errorf("diagram %s is not published: %w", diagramID.String(), domain.ErrNotPublished)
	}
	diagram.PublishedAt = publishedAt.UTC()

	diagram.Canvas = parseCanvasFromDescription(diagram.Description)
	figures, figuresError := repository.loadFigures(ctx, diagramID)
	if figuresError != nil {
		return domain.Diagram{}, figuresError
	}
	diagram.Figures = figures

	return diagram, nil
}

func (repository *DiagramRepository) loadFigures(
	ctx context.Context,
	diagramID uuid.UUID,
) ([]domain.DiagramFigure, error) {
const figuresQuery = `
SELECT
    f.id,
    f.type,
    f.tag_id,
    COALESCE(fp.params, '{}'::jsonb) AS params
FROM public.figures f
LEFT JOIN public.figure_params fp ON fp.figure_id = f.id
WHERE f.diagram_id = $1
  AND f.deleted_at IS NULL
ORDER BY f.created_at ASC, f.id ASC
`

	rows, queryError := repository.pool.Query(ctx, figuresQuery, diagramID)
	if queryError != nil {
		return nil, fmt.Errorf("query diagram figures: %w", queryError)
	}
	defer rows.Close()

	figures := make([]domain.DiagramFigure, 0)
	for rows.Next() {
		var (
			figure     domain.DiagramFigure
			figureType string
			tagID      *uuid.UUID
			params     []byte
		)

		if scanError := rows.Scan(
			&figure.ID,
			&figureType,
			&tagID,
			&params,
		); scanError != nil {
			return nil, fmt.Errorf("scan diagram figure: %w", scanError)
		}

		figure.Type = domain.FigureType(strings.TrimSpace(figureType))
		figure.TagID = tagID
		figure.Params = json.RawMessage(params)

		figures = append(figures, figure)
	}
	if rowsError := rows.Err(); rowsError != nil {
		return nil, fmt.Errorf("iterate diagram figures: %w", rowsError)
	}

	return figures, nil
}

func (repository *DiagramRepository) ListDiagramBoundTagIDs(
	ctx context.Context,
	diagramID uuid.UUID,
) ([]uuid.UUID, error) {
	if visibilityError := repository.ensureDiagramIsPublished(ctx, diagramID); visibilityError != nil {
		return nil, visibilityError
	}

	const query = `
SELECT DISTINCT f.tag_id
FROM public.figures f
WHERE f.diagram_id = $1
  AND f.deleted_at IS NULL
  AND f.tag_id IS NOT NULL
ORDER BY f.tag_id ASC
`

	rows, queryError := repository.pool.Query(ctx, query, diagramID)
	if queryError != nil {
		return nil, fmt.Errorf("query bound tag ids: %w", queryError)
	}
	defer rows.Close()

	tagIDs := make([]uuid.UUID, 0)
	for rows.Next() {
		var tagID uuid.UUID
		if scanError := rows.Scan(&tagID); scanError != nil {
			return nil, fmt.Errorf("scan bound tag id: %w", scanError)
		}

		tagIDs = append(tagIDs, tagID)
	}
	if rowsError := rows.Err(); rowsError != nil {
		return nil, fmt.Errorf("iterate bound tag ids: %w", rowsError)
	}

	return tagIDs, nil
}

func (repository *DiagramRepository) ensureDiagramIsPublished(
	ctx context.Context,
	diagramID uuid.UUID,
) error {
	const query = `
SELECT m.published_at
FROM public.mimic m
WHERE m.id = $1
  AND m.deleted_at IS NULL
`

	var publishedAt *time.Time
	if scanError := repository.pool.QueryRow(ctx, query, diagramID).Scan(&publishedAt); scanError != nil {
		if errors.Is(scanError, pgx.ErrNoRows) {
			return fmt.Errorf("diagram %s not found: %w", diagramID.String(), domain.ErrNotFound)
		}
		return fmt.Errorf("query diagram visibility: %w", scanError)
	}

	if publishedAt == nil || publishedAt.IsZero() {
		return fmt.Errorf("diagram %s is not published: %w", diagramID.String(), domain.ErrNotPublished)
	}

	return nil
}

func parseCanvasFromDescription(description string) domain.DiagramCanvas {
	var payload map[string]any
	if unmarshalError := json.Unmarshal([]byte(strings.TrimSpace(description)), &payload); unmarshalError != nil {
		return domain.DiagramCanvas{}
	}

	var canvas domain.DiagramCanvas
	if rawCanvas, exists := payload["canvas"]; exists {
		canvasObject, ok := rawCanvas.(map[string]any)
		if ok {
			canvas.Width = parseCanvasInt(canvasObject["width"])
			canvas.Height = parseCanvasInt(canvasObject["height"])
			canvas.Background = parseCanvasString(canvasObject["background"])
		}
	}

	if canvas.Width == 0 {
		canvas.Width = parseCanvasInt(payload["canvas_width"])
	}
	if canvas.Height == 0 {
		canvas.Height = parseCanvasInt(payload["canvas_height"])
	}
	if canvas.Background == "" {
		canvas.Background = parseCanvasString(payload["canvas_background"])
	}

	return canvas
}

func parseCanvasInt(value any) int {
	switch typedValue := value.(type) {
	case float64:
		return int(typedValue)
	case int:
		return typedValue
	case int32:
		return int(typedValue)
	case int64:
		return int(typedValue)
	default:
		return 0
	}
}

func parseCanvasString(value any) string {
	text, ok := value.(string)
	if !ok {
		return ""
	}

	return strings.TrimSpace(text)
}
