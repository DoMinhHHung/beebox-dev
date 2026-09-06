package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthHandler trả về trạng thái hoạt động của dịch vụ với mã HTTP 200 và nội dung JSON {"status": "ok"}.
func HealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
