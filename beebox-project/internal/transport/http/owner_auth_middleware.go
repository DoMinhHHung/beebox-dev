package httpapi

import (
	"strings"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/application/auth"
	"github.com/gin-gonic/gin"
)

const ownerIDContextKey = "owner_id"
const bearerPrefix = "Bearer "

// RequireOwnerSession creates middleware that verifies the bearer token in the Authorization header and stores the authenticated owner ID in the request context.
func RequireOwnerSession(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := parseBearerToken(c.GetHeader("Authorization"))
		if err != nil {
			c.Error(err)
			c.Abort()
			return
		}

		ownerID, err := svc.VerifySession(c.Request.Context(), token)
		if err != nil {
			c.Error(err)
			c.Abort()
			return
		}

		c.Set(ownerIDContextKey, ownerID)
		c.Next()
	}
}

// OwnerIDFromContext retrieves the owner ID stored in the Gin context.
// It returns the owner ID and true when the context contains a string value
// for the owner ID; otherwise, it returns an empty string and false.
func OwnerIDFromContext(c *gin.Context) (string, bool) {
	value, exists := c.Get(ownerIDContextKey)
	if !exists {
		return "", false
	}
	ownerID, ok := value.(string)
	return ownerID, ok
}

// parseBearerToken trích xuất Bearer token từ giá trị header Authorization.
// Hàm trả về phần sau tiền tố "Bearer " hoặc lỗi xác thực nếu header không bắt đầu bằng tiền tố này.
func parseBearerToken(header string) (string, error) {
	if !strings.HasPrefix(header, bearerPrefix) {
		return "", apperror.New(apperror.CodeUnauthenticated, "missing or malformed Authorization header")
	}
	return strings.TrimPrefix(header, bearerPrefix), nil
}
