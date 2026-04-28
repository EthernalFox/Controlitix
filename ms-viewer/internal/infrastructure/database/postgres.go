package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultConnectionTimeout = 5 * time.Second

func OpenPostgresPool(
	ctx context.Context,
	databaseURL string,
) (*pgxpool.Pool, func(), error) {
	if databaseURL == "" {
		return nil, nil, errors.New("database connection is not configured")
	}

	configuration, parseError := pgxpool.ParseConfig(databaseURL)
	if parseError != nil {
		return nil, nil, fmt.Errorf("parse database url: %w", parseError)
	}

	pool, openError := pgxpool.NewWithConfig(ctx, configuration)
	if openError != nil {
		return nil, nil, fmt.Errorf("open postgres pool: %w", openError)
	}

	pingContext, cancelPing := context.WithTimeout(ctx, defaultConnectionTimeout)
	defer cancelPing()

	if pingError := pool.Ping(pingContext); pingError != nil {
		pool.Close()
		return nil, nil, fmt.Errorf("ping postgres pool: %w", pingError)
	}

	return pool, pool.Close, nil
}
