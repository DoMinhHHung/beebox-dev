package idgen

import (
	"fmt"

	"github.com/google/uuid"
)

// New returns a UUIDv7 string (time-ordered).
// New tạo UUIDv7 có thứ tự theo thời gian và trả về dưới dạng chuỗi.
// Hàm trả về lỗi nếu không thể tạo UUID.
// Giá trị trả về là chuỗi UUIDv7 và lỗi phát sinh, nếu có.
func New() (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("idgen: uuidv7: %w", err)
	}
	return id.String(), nil
}
