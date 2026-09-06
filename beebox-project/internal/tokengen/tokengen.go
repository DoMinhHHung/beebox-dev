package tokengen

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
)

// New tạo token ngẫu nhiên an toàn về mặt mật mã với số byte được yêu cầu.
// Token được mã hóa bằng Base64 URL-safe không đệm.
//
// nBytes là số byte dữ liệu ngẫu nhiên cần tạo.
//
// New trả về token đã mã hóa hoặc lỗi nội bộ nếu không thể tạo dữ liệu ngẫu nhiên.
func New(nBytes int) (string, error) {
	buf := make([]byte, nBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", apperror.Wrap(apperror.CodeInternal, "failed to generate random token", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// Hash tạo giá trị băm SHA-256 của token và mã hóa kết quả thành chuỗi thập lục phân.
func Hash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
