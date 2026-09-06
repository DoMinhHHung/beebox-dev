package postgres

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

const pgErrCodeUniqueViolation = "23505"

func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == pgErrCodeUniqueViolation
	}
	return false
}
