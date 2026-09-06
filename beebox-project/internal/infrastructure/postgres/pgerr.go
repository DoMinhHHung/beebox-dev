package postgres

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

const pgErrCodeUniqueViolation = "23505"

// IsUniqueViolation xác định lỗi có phải là lỗi vi phạm ràng buộc duy nhất của PostgreSQL hay không.
// Trả về true nếu lỗi hoặc lỗi được bọc có mã lỗi 23505, false trong các trường hợp khác.
func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == pgErrCodeUniqueViolation
	}
	return false
}
