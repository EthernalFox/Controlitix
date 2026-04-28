package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
)

type IngestRepository struct {
	pool *pgxpool.Pool
}

func NewIngestRepository(pool *pgxpool.Pool) *IngestRepository {
	return &IngestRepository{pool: pool}
}

func (repository *IngestRepository) UpsertRawValues(
	ctx context.Context,
	records []domain.IngestRecord,
) error {
	if len(records) == 0 {
		return nil
	}

	arguments := make([]any, 0, len(records)*4)
	placeholders := make([]string, 0, len(records))
	for index, record := range records {
		base := index*4 + 1
		placeholders = append(
			placeholders,
			fmt.Sprintf("($%d, $%d, $%d, $%d)", base, base+1, base+2, base+3),
		)
		arguments = append(arguments, record.TagID, record.Timestamp, record.Value, record.Quality.ToCode())
	}

	query := `
INSERT INTO history.tag_values_raw (tag_id, ts, v, q)
VALUES ` + strings.Join(placeholders, ",") + `
ON CONFLICT (tag_id, ts)
DO UPDATE
SET
    v = EXCLUDED.v,
    q = EXCLUDED.q
`

	if _, executeError := repository.pool.Exec(ctx, query, arguments...); executeError != nil {
		return fmt.Errorf("upsert raw values: %w", executeError)
	}

	return nil
}
