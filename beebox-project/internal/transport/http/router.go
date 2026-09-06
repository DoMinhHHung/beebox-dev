package httpapi

import (
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/application/auth"
	credentialapp "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/application/credential"
	fielddefinitionapp "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/application/fielddefinition"
	projectapp "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/application/project"
	"github.com/gin-gonic/gin"
)

type Dependencies struct {
	AuthService            *auth.Service
	ProjectService         *projectapp.Service
	CredentialService      *credentialapp.Service
	FieldDefinitionService *fielddefinitionapp.Service
}

// New creates an HTTP engine with health, authentication, and protected dashboard routes.
func New(deps Dependencies) *gin.Engine {
	engine := gin.Default()
	engine.Use(ErrorHandler())

	engine.GET("/healthz", HealthHandler)

	authGroup := engine.Group("/auth")
	RegisterAuthRoutes(authGroup, deps.AuthService)

	dashboard := engine.Group("/dashboard")
	dashboard.Use(RequireOwnerSession(deps.AuthService))
	RegisterProjectRoutes(dashboard, deps.ProjectService)
	RegisterCredentialRoutes(dashboard, deps.CredentialService)
	RegisterFieldDefinitionRoutes(dashboard, deps.FieldDefinitionService)

	return engine
}
