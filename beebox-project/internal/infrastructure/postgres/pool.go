package postgres

import (
	"context"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Trả về lỗi nếu DATABASE_URL không hợp lệ, không thể khởi tạo pool hoặc không thể kết nối đến PostgreSQL.
func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, apperror.New(apperror.CodeInvalidInput, "invalid DATABASE_URL format")
	}

	cfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to create postgres connection pool", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to connect to postgres", err)
	}

	return pool, nil
}
