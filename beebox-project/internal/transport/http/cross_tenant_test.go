package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/application/auth"
	credentialapp "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/application/credential"
	fielddefinitionapp "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/application/fielddefinition"
	projectapp "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/application/project"
	credentialmemory "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/infrastructure/credential/memory"
	fielddefinitionmemory "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/infrastructure/fielddefinition/memory"
	ownermemory "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/infrastructure/owner/memory"
	sessionmemory "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/infrastructure/ownersession/memory"
	projectmemory "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/infrastructure/project/memory"
	"github.com/gin-gonic/gin"
)

func newFullTestEngine() *gin.Engine {
	authSvc := auth.NewService(ownermemory.NewOwnerRepository(), sessionmemory.NewSessionRepository(), time.Hour)
	projects := projectmemory.NewProjectRepository()
	projectSvc := projectapp.NewService(projects)
	credentialSvc := credentialapp.NewService(credentialmemory.New(), projects)
	fieldSvc := fielddefinitionapp.NewService(fielddefinitionmemory.New(), projects)

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(ErrorHandler())

	authGroup := engine.Group("/auth")
	RegisterAuthRoutes(authGroup, authSvc)

	dashboard := engine.Group("/dashboard")
	dashboard.Use(RequireOwnerSession(authSvc))
	RegisterProjectRoutes(dashboard, projectSvc)
	RegisterCredentialRoutes(dashboard, credentialSvc)
	RegisterFieldDefinitionRoutes(dashboard, fieldSvc)

	return engine
}

func mustUnmarshal(t *testing.T, data []byte, v interface{}) {
	t.Helper()
	if err := json.Unmarshal(data, v); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
}

func TestCrossTenant_OwnerWithOwnProjectCannotAccessAnothersResources(t *testing.T) {
	engine := newFullTestEngine()

	_, tokenA := signUpAndSignIn(t, engine, "ownerA@example.com", "password123")
	projA := createProjectViaHTTP(t, engine, tokenA)

	createCredBody := `{"environment":"test"}`
	credReq := httptest.NewRequest(http.MethodPost, "/dashboard/projects/"+projA.ID+"/credentials", strings.NewReader(createCredBody))
	credReq.Header.Set("Content-Type", "application/json")
	credReq.Header.Set("Authorization", "Bearer "+tokenA)
	credRec := httptest.NewRecorder()
	engine.ServeHTTP(credRec, credReq)
	if credRec.Code != http.StatusCreated {
		t.Fatalf("expected credential creation to succeed, got %d, body: %s", credRec.Code, credRec.Body.String())
	}
	var cred credentialWithSecretResponse
	mustUnmarshal(t, credRec.Body.Bytes(), &cred)

	defineBody := `{"fields":[{"name":"fullName","kind":"STRING","required":true}]}`
	defineRec := defineFieldsViaHTTP(t, engine, tokenA, projA.ID, defineBody)
	if defineRec.Code != http.StatusOK {
		t.Fatalf("expected define to succeed, got %d, body: %s", defineRec.Code, defineRec.Body.String())
	}

	_, tokenB := signUpAndSignIn(t, engine, "ownerB@example.com", "password123")
	createProjectViaHTTP(t, engine, tokenB)

	cases := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{"get credential of another owner's project", http.MethodGet, "/dashboard/credentials/" + cred.ID, ""},
		{"rotate credential of another owner's project", http.MethodPost, "/dashboard/credentials/" + cred.ID + "/rotate", ""},
		{"revoke credential of another owner's project", http.MethodPost, "/dashboard/credentials/" + cred.ID + "/revoke", ""},
		{"create credential under another owner's project", http.MethodPost, "/dashboard/projects/" + projA.ID + "/credentials", `{"environment":"test"}`},
		{"get fields of another owner's project", http.MethodGet, "/dashboard/projects/" + projA.ID + "/fields", ""},
		{"define fields on another owner's project", http.MethodPut, "/dashboard/projects/" + projA.ID + "/fields", defineBody},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var req *http.Request
			if tc.body != "" {
				req = httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tc.method, tc.path, nil)
			}
			req.Header.Set("Authorization", "Bearer "+tokenB)
			rec := httptest.NewRecorder()
			engine.ServeHTTP(rec, req)

			if rec.Code != http.StatusForbidden {
				t.Fatalf("expected status %d, got %d, body: %s", http.StatusForbidden, rec.Code, rec.Body.String())
			}
		})
	}
}
