package database

import (
	"context"
	"database/sql"
	"errors"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const defaultConnectionTimeout = 5 * time.Second

func OpenPostgres(
	ctx context.Context,
	databaseURL string,
) (*sql.DB, func() error, error) {
	if databaseURL == "" {
		return nil, nil, errors.New("База данных не сконфигурирована")
	}

	databaseConnection, databaseError := sql.Open("pgx", databaseURL)
	if databaseError != nil {
		return nil, nil, databaseError
	}

	pingContext, cancel := context.WithTimeout(ctx, defaultConnectionTimeout)
	defer cancel()

	if pingError := databaseConnection.PingContext(pingContext); pingError != nil {
		closeError := databaseConnection.Close()
		if closeError != nil {
			return nil, nil, errors.Join(pingError, closeError)
		}

		return nil, nil, pingError
	}

	return databaseConnection, databaseConnection.Close, nil
}
