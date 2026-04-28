package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
)

type TrendRepository struct {
	pool *pgxpool.Pool
}

func NewTrendRepository(pool *pgxpool.Pool) *TrendRepository {
	return &TrendRepository{pool: pool}
}

func (repository *TrendRepository) GetTagMeta(
	ctx context.Context,
	tagID uuid.UUID,
) (domain.TagMeta, error) {
	const query = `
SELECT
    t.id,
    t.name,
    t.device_id,
    d.name,
    COALESCE(u.id, 0) AS unit_id,
    COALESCE(u.name, '') AS unit_name,
    COALESCE(u.symbol, '') AS unit_symbol,
    COALESCE(u.category, '') AS unit_category,
    COALESCE(dt.id, 0) AS data_type_id,
    COALESCE(dt.name, '') AS data_type_name
FROM tags.tags t
JOIN devices.devices d ON d.id = t.device_id
LEFT JOIN tags.units u ON u.id = t.unit_id
LEFT JOIN tags.data_types dt ON dt.id = t.data_type_id
WHERE t.id = $1
  AND t.deleted_at IS NULL
`

	var tagMeta domain.TagMeta
	row := repository.pool.QueryRow(ctx, query, tagID)
	if scanError := row.Scan(
		&tagMeta.TagID,
		&tagMeta.TagName,
		&tagMeta.DeviceID,
		&tagMeta.DeviceName,
		&tagMeta.Unit.ID,
		&tagMeta.Unit.Name,
		&tagMeta.Unit.Symbol,
		&tagMeta.Unit.Category,
		&tagMeta.DataType.ID,
		&tagMeta.DataType.Name,
	); scanError != nil {
		if errors.Is(scanError, pgx.ErrNoRows) {
			return domain.TagMeta{}, fmt.Errorf("tag %s not found: %w", tagID.String(), domain.ErrNotFound)
		}
		return domain.TagMeta{}, fmt.Errorf("query tag metadata: %w", scanError)
	}

	return tagMeta, nil
}

func (repository *TrendRepository) GetRawPoints(
	ctx context.Context,
	tagID uuid.UUID,
	from time.Time,
	to time.Time,
) ([]domain.TrendPoint, error) {
	const query = `
SELECT ts, v, q
FROM history.tag_values_raw
WHERE tag_id = $1
  AND ts >= $2
  AND ts < $3
ORDER BY ts ASC
`

	rows, queryError := repository.pool.Query(ctx, query, tagID, from, to)
	if queryError != nil {
		return nil, fmt.Errorf("query raw points: %w", queryError)
	}
	defer rows.Close()

	points, scanError := scanTrendPoints(rows)
	if scanError != nil {
		return nil, scanError
	}

	return points, nil
}

func (repository *TrendRepository) GetRawPointsByStep(
	ctx context.Context,
	tagID uuid.UUID,
	from time.Time,
	to time.Time,
	step time.Duration,
	aggregator domain.Aggregator,
) ([]domain.TrendPoint, error) {
	aggregationExpression, expressionError := rawAggregationExpression(aggregator)
	if expressionError != nil {
		return nil, expressionError
	}

	query := fmt.Sprintf(`
SELECT
    time_bucket($4::interval, ts) AS bucket,
    %s AS v,
    last(q, ts) AS q
FROM history.tag_values_raw
WHERE tag_id = $1
  AND ts >= $2
  AND ts < $3
GROUP BY 1
ORDER BY 1 ASC
`, aggregationExpression)

	rows, queryError := repository.pool.Query(
		ctx,
		query,
		tagID,
		from,
		to,
		durationToInterval(step),
	)
	if queryError != nil {
		return nil, fmt.Errorf("query raw points by step: %w", queryError)
	}
	defer rows.Close()

	points, scanError := scanTrendPoints(rows)
	if scanError != nil {
		return nil, scanError
	}

	return points, nil
}

func (repository *TrendRepository) GetAggregatedPoints(
	ctx context.Context,
	tagID uuid.UUID,
	from time.Time,
	to time.Time,
	step time.Duration,
	aggregator domain.Aggregator,
) ([]domain.TrendPoint, error) {
	if step <= time.Minute {
		columnExpression, expressionError := aggregateColumnExpression(aggregator)
		if expressionError != nil {
			return nil, expressionError
		}

		query := fmt.Sprintf(`
SELECT
    bucket AS ts,
    %s AS v,
    q
FROM history.tag_values_1m
WHERE tag_id = $1
  AND bucket >= $2
  AND bucket < $3
ORDER BY bucket ASC
`, columnExpression)

		rows, queryError := repository.pool.Query(ctx, query, tagID, from, to)
		if queryError != nil {
			return nil, fmt.Errorf("query aggregate points: %w", queryError)
		}
		defer rows.Close()

		points, scanError := scanTrendPoints(rows)
		if scanError != nil {
			return nil, scanError
		}

		return points, nil
	}

	aggregationExpression, expressionError := aggregateAggregationExpression(aggregator)
	if expressionError != nil {
		return nil, expressionError
	}

	query := fmt.Sprintf(`
SELECT
    time_bucket($4::interval, bucket) AS bucket,
    %s AS v,
    last(q, bucket) AS q
FROM history.tag_values_1m
WHERE tag_id = $1
  AND bucket >= $2
  AND bucket < $3
GROUP BY 1
ORDER BY 1 ASC
`, aggregationExpression)

	rows, queryError := repository.pool.Query(
		ctx,
		query,
		tagID,
		from,
		to,
		durationToInterval(step),
	)
	if queryError != nil {
		return nil, fmt.Errorf("query aggregate points by step: %w", queryError)
	}
	defer rows.Close()

	points, scanError := scanTrendPoints(rows)
	if scanError != nil {
		return nil, scanError
	}

	return points, nil
}

func scanTrendPoints(rows pgx.Rows) ([]domain.TrendPoint, error) {
	points := make([]domain.TrendPoint, 0)
	for rows.Next() {
		var point domain.TrendPoint
		var qualityCode int16
		if scanError := rows.Scan(&point.Timestamp, &point.Value, &qualityCode); scanError != nil {
			return nil, fmt.Errorf("scan trend point: %w", scanError)
		}
		point.Quality = domain.QualityFromCode(qualityCode)
		points = append(points, point)
	}

	if rowsError := rows.Err(); rowsError != nil {
		return nil, fmt.Errorf("iterate trend points: %w", rowsError)
	}

	return points, nil
}

func rawAggregationExpression(aggregator domain.Aggregator) (string, error) {
	switch aggregator {
	case domain.AggregatorLast:
		return "last(v, ts)", nil
	case domain.AggregatorAvg:
		return "avg(v)", nil
	case domain.AggregatorMin:
		return "min(v)", nil
	case domain.AggregatorMax:
		return "max(v)", nil
	default:
		return "", fmt.Errorf("invalid aggregator: %w", domain.ErrInvalidInput)
	}
}

func aggregateColumnExpression(aggregator domain.Aggregator) (string, error) {
	switch aggregator {
	case domain.AggregatorLast:
		return "\"last\"", nil
	case domain.AggregatorAvg:
		return "\"avg\"", nil
	case domain.AggregatorMin:
		return "\"min\"", nil
	case domain.AggregatorMax:
		return "\"max\"", nil
	default:
		return "", fmt.Errorf("invalid aggregator: %w", domain.ErrInvalidInput)
	}
}

func aggregateAggregationExpression(aggregator domain.Aggregator) (string, error) {
	switch aggregator {
	case domain.AggregatorLast:
		return "last(\"last\", bucket)", nil
	case domain.AggregatorAvg:
		return "sum(\"avg\" * \"count\") / NULLIF(sum(\"count\"), 0)", nil
	case domain.AggregatorMin:
		return "min(\"min\")", nil
	case domain.AggregatorMax:
		return "max(\"max\")", nil
	default:
		return "", fmt.Errorf("invalid aggregator: %w", domain.ErrInvalidInput)
	}
}

func durationToInterval(duration time.Duration) string {
	if duration <= 0 {
		return "1 second"
	}

	if duration%time.Second == 0 {
		return fmt.Sprintf("%d seconds", int64(duration/time.Second))
	}
	if duration%time.Millisecond == 0 {
		return fmt.Sprintf("%d milliseconds", int64(duration/time.Millisecond))
	}
	if duration < time.Microsecond {
		return "1 microsecond"
	}

	return fmt.Sprintf("%d microseconds", int64(duration/time.Microsecond))
}
