package httpapi

import (
	"strings"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/application/auth"
	"github.com/gin-gonic/gin"
)

const ownerIDContextKey = "owner_id"
const bearerPrefix = "Bearer "

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

func OwnerIDFromContext(c *gin.Context) (string, bool) {
	value, exists := c.Get(ownerIDContextKey)
	if !exists {
		return "", false
	}
	ownerID, ok := value.(string)
	return ownerID, ok
}

func parseBearerToken(header string) (string, error) {
	if !strings.HasPrefix(header, bearerPrefix) {
		return "", apperror.New(apperror.CodeUnauthenticated, "missing or malformed Authorization header")
	}
	return strings.TrimPrefix(header, bearerPrefix), nil
}
