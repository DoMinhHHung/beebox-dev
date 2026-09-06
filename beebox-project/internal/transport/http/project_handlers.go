package httpapi

import (
	"net/http"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
	projectapp "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/application/project"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/domain/project"
	"github.com/gin-gonic/gin"
)

type createProjectRequest struct {
	Name string `json:"name"`
	Tier string `json:"tier"`
}

type createProjectResponse struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Tier    string `json:"tier"`
	OwnerID string `json:"owner_id"`
}

func RegisterProjectRoutes(rg *gin.RouterGroup, svc *projectapp.Service) {
	rg.POST("/projects", createProjectHandler(svc))
}

func createProjectHandler(svc *projectapp.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req createProjectRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(apperror.New(apperror.CodeInvalidInput, "invalid request body"))
			return
		}

		ownerID, ok := OwnerIDFromContext(c)
		if !ok {
			c.Error(apperror.New(apperror.CodeInternal, "owner context missing"))
			return
		}

		tier := req.Tier
		if tier == "" {
			tier = string(project.TierFree)
		}

		p, err := svc.CreateProject(c.Request.Context(), ownerID, req.Name, project.Tier(tier))
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusCreated, createProjectResponse{
			ID:      p.ID,
			Name:    p.Name,
			Tier:    string(p.Tier),
			OwnerID: p.OwnerID,
		})
	}
}
