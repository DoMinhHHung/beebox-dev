package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrMissingConnectionString = errors.New("postgres connection string is required")

func NewPool(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	if connString == "" {
		return nil, ErrMissingConnectionString
	}

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}
