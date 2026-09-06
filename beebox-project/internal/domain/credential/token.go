package credential

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
	"golang.org/x/crypto/bcrypt"
)

// randomToken tạo token ngẫu nhiên an toàn bằng mật mã và mã hóa token bằng Base64 URL không có padding. 
// Hàm trả về lỗi nội bộ nếu không thể tạo dữ liệu ngẫu nhiên.
func randomToken(nBytes int) (string, error) {
	buf := make([]byte, nBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", apperror.Wrap(apperror.CodeInternal, "failed to generate random token", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// newPublicKey tạo khóa công khai cho môi trường được chỉ định và trả về lỗi nếu không thể tạo token.
// Giá trị trả về có dạng "pk_<môi trường>_<token>".
func newPublicKey(env Environment) (string, error) {
	token, err := randomToken(24)
	if err != nil {
		return "", err
	}
	return "pk_" + string(env) + "_" + token, nil
}

// newSecret generates a secret key for the specified environment and returns its plaintext and bcrypt hash.
// It returns an internal application error if hashing fails.
func newSecret(env Environment) (plaintext string, hash string, err error) {
	token, err := randomToken(32)
	if err != nil {
		return "", "", err
	}
	plaintext = "sk_" + string(env) + "_" + token
	hashBytes, hashErr := bcrypt.GenerateFromPassword([]byte(plaintext), bcrypt.DefaultCost)
	if hashErr != nil {
		return "", "", apperror.Wrap(apperror.CodeInternal, "failed to hash secret key", hashErr)
	}
	return plaintext, string(hashBytes), nil
}
