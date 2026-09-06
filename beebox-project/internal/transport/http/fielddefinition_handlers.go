package httpapi

import (
	"net/http"
	"strconv"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
	fielddefinitionapp "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/application/fielddefinition"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/domain/fielddefinition"
	"github.com/gin-gonic/gin"
)

type fieldRequest struct {
	Name     string `json:"name"`
	Kind     string `json:"kind"`
	Required bool   `json:"required"`
}

type defineFieldsRequest struct {
	Fields []fieldRequest `json:"fields"`
}

type fieldResponse struct {
	Name     string `json:"name"`
	Kind     string `json:"kind"`
	Required bool   `json:"required"`
}

type schemaResponse struct {
	ProjectID string          `json:"project_id"`
	Version   int             `json:"version"`
	Fields    []fieldResponse `json:"fields"`
}

func toSchemaResponse(s fielddefinition.Schema) schemaResponse {
	fields := make([]fieldResponse, len(s.Fields))
	for i, f := range s.Fields {
		fields[i] = fieldResponse{Name: f.Name, Kind: string(f.Kind), Required: f.Required}
	}
	return schemaResponse{ProjectID: s.ProjectID, Version: s.Version, Fields: fields}
}

func RegisterFieldDefinitionRoutes(rg *gin.RouterGroup, svc *fielddefinitionapp.Service) {
	rg.PUT("/projects/:projectID/fields", defineFieldsHandler(svc))
	rg.GET("/projects/:projectID/fields", getLatestFieldsHandler(svc))
	rg.GET("/projects/:projectID/fields/:version", getVersionFieldsHandler(svc))
}

func defineFieldsHandler(svc *fielddefinitionapp.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req defineFieldsRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(apperror.New(apperror.CodeInvalidInput, "invalid request body"))
			return
		}

		ownerID, ok := OwnerIDFromContext(c)
		if !ok {
			c.Error(apperror.New(apperror.CodeInternal, "owner context missing"))
			return
		}

		fields := make([]fielddefinition.FieldDefinition, len(req.Fields))
		for i, f := range req.Fields {
			built, err := fielddefinition.NewFieldDefinition(f.Name, fielddefinition.FieldKind(f.Kind), f.Required)
			if err != nil {
				c.Error(err)
				return
			}
			fields[i] = built
		}

		projectID := c.Param("projectID")

		schema, err := svc.Define(c.Request.Context(), ownerID, projectID, fields)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, toSchemaResponse(schema))
	}
}

func getLatestFieldsHandler(svc *fielddefinitionapp.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		ownerID, ok := OwnerIDFromContext(c)
		if !ok {
			c.Error(apperror.New(apperror.CodeInternal, "owner context missing"))
			return
		}

		projectID := c.Param("projectID")

		schema, err := svc.GetLatest(c.Request.Context(), ownerID, projectID)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, toSchemaResponse(schema))
	}
}

func getVersionFieldsHandler(svc *fielddefinitionapp.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		ownerID, ok := OwnerIDFromContext(c)
		if !ok {
			c.Error(apperror.New(apperror.CodeInternal, "owner context missing"))
			return
		}

		projectID := c.Param("projectID")

		version, err := strconv.Atoi(c.Param("version"))
		if err != nil {
			c.Error(apperror.New(apperror.CodeInvalidInput, "version must be a valid integer"))
			return
		}

		schema, err := svc.GetVersion(c.Request.Context(), ownerID, projectID, version)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, toSchemaResponse(schema))
	}
}
