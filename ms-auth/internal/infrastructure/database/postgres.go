package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const defaultConnectionTimeout = 5 * time.Second

func OpenPostgres(
	ctx context.Context,
	databaseURL string,
) (*sql.DB, func() error, error) {
	if databaseURL == "" {
		return nil, nil, errors.New("database is not configured")
	}

	databaseConnection, openError := sql.Open("pgx", databaseURL)
	if openError != nil {
		return nil, nil, fmt.Errorf("open postgres connection: %w", openError)
	}

	pingContext, cancelPing := context.WithTimeout(ctx, defaultConnectionTimeout)
	defer cancelPing()

	if pingError := databaseConnection.PingContext(pingContext); pingError != nil {
		if closeError := databaseConnection.Close(); closeError != nil {
			return nil, nil, fmt.Errorf(
				"ping postgres connection: %w",
				errors.Join(pingError, closeError),
			)
		}

		return nil, nil, fmt.Errorf("ping postgres connection: %w", pingError)
	}

	return databaseConnection, databaseConnection.Close, nil
}
