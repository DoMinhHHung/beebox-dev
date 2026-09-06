package httpapi

import (
	"net/http"
	"time"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
	credentialapp "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/application/credential"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/domain/credential"
	"github.com/gin-gonic/gin"
)

type createCredentialRequest struct {
	Environment string `json:"environment"`
}

type credentialMetadataResponse struct {
	ID          string     `json:"id"`
	ProjectID   string     `json:"project_id"`
	Environment string     `json:"environment"`
	PublicKey   string     `json:"public_key"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	RevokedAt   *time.Time `json:"revoked_at"`
}

type credentialWithSecretResponse struct {
	credentialMetadataResponse
	SecretKey string `json:"secret_key"`
}

// toMetadataResponse converts credential metadata to its HTTP response representation.
func toMetadataResponse(md credential.Metadata) credentialMetadataResponse {
	return credentialMetadataResponse{
		ID:          md.ID,
		ProjectID:   md.ProjectID,
		Environment: string(md.Environment),
		PublicKey:   md.PublicKey,
		Status:      string(md.Status),
		CreatedAt:   md.CreatedAt,
		RevokedAt:   md.RevokedAt,
	}
}

// toSecretResponse combines credential metadata and a secret key into an API response.
func toSecretResponse(md credential.Metadata, secret string) credentialWithSecretResponse {
	return credentialWithSecretResponse{
		credentialMetadataResponse: toMetadataResponse(md),
		SecretKey:                  secret,
	}
}

// RegisterCredentialRoutes đăng ký các route HTTP để tạo, xoay vòng, thu hồi và truy vấn thông tin xác thực.
func RegisterCredentialRoutes(rg *gin.RouterGroup, svc *credentialapp.Service) {
	rg.POST("/projects/:projectID/credentials", createCredentialHandler(svc))
	rg.POST("/credentials/:credentialID/rotate", rotateCredentialHandler(svc))
	rg.POST("/credentials/:credentialID/revoke", revokeCredentialHandler(svc))
	rg.GET("/credentials/:credentialID", getCredentialHandler(svc))
}

// createCredentialHandler tạo HTTP handler để tạo thông tin xác thực cho một dự án.
func createCredentialHandler(svc *credentialapp.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req createCredentialRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(apperror.New(apperror.CodeInvalidInput, "invalid request body"))
			return
		}

		ownerID, ok := OwnerIDFromContext(c)
		if !ok {
			c.Error(apperror.New(apperror.CodeInternal, "owner context missing"))
			return
		}

		projectID := c.Param("projectID")

		md, secret, err := svc.Create(c.Request.Context(), ownerID, projectID, credential.Environment(req.Environment))
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusCreated, toSecretResponse(md, secret))
	}
}

// rotateCredentialHandler tạo trình xử lý HTTP để xoay vòng thông tin xác thực và trả về siêu dữ liệu cùng khóa bí mật mới.
// Giá trị trả về là trình xử lý thực hiện thao tác xoay vòng cho thông tin xác thực được chỉ định.
func rotateCredentialHandler(svc *credentialapp.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		ownerID, ok := OwnerIDFromContext(c)
		if !ok {
			c.Error(apperror.New(apperror.CodeInternal, "owner context missing"))
			return
		}

		credentialID := c.Param("credentialID")

		md, secret, err := svc.Rotate(c.Request.Context(), ownerID, credentialID)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, toSecretResponse(md, secret))
	}
}

// revokeCredentialHandler tạo HTTP handler thu hồi thông tin xác thực và trả về siêu dữ liệu của thông tin xác thực đó.
func revokeCredentialHandler(svc *credentialapp.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		ownerID, ok := OwnerIDFromContext(c)
		if !ok {
			c.Error(apperror.New(apperror.CodeInternal, "owner context missing"))
			return
		}

		credentialID := c.Param("credentialID")

		md, err := svc.Revoke(c.Request.Context(), ownerID, credentialID)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, toMetadataResponse(md))
	}
}

// getCredentialHandler trả về handler lấy siêu dữ liệu của thông tin xác thực theo mã thông tin xác thực.
// @param svc Dịch vụ quản lý thông tin xác thực.
// @returns Handler HTTP trả về siêu dữ liệu thông tin xác thực hoặc chuyển tiếp lỗi.
func getCredentialHandler(svc *credentialapp.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		ownerID, ok := OwnerIDFromContext(c)
		if !ok {
			c.Error(apperror.New(apperror.CodeInternal, "owner context missing"))
			return
		}

		credentialID := c.Param("credentialID")

		md, err := svc.GetMetadata(c.Request.Context(), ownerID, credentialID)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, toMetadataResponse(md))
	}
}
