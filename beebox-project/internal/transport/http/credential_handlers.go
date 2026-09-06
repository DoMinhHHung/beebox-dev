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

func toSecretResponse(md credential.Metadata, secret string) credentialWithSecretResponse {
	return credentialWithSecretResponse{
		credentialMetadataResponse: toMetadataResponse(md),
		SecretKey:                  secret,
	}
}

func RegisterCredentialRoutes(rg *gin.RouterGroup, svc *credentialapp.Service) {
	rg.POST("/projects/:projectID/credentials", createCredentialHandler(svc))
	rg.POST("/credentials/:credentialID/rotate", rotateCredentialHandler(svc))
	rg.POST("/credentials/:credentialID/revoke", revokeCredentialHandler(svc))
	rg.GET("/credentials/:credentialID", getCredentialHandler(svc))
}

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
