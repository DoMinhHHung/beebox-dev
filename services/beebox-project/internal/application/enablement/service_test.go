package enablement_test

import (
	"context"
	"testing"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/apperror"
	applicationenablement "github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/application/enablement"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/catalog"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/configuration"
	domainproject "github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/project"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/infrastructure/memory"
)

type fakeAuthorizer struct {
	project domainproject.Project
	err     error
}

func (f fakeAuthorizer) GetAuthorized(context.Context, string, string) (domainproject.Project, error) {
	if f.err != nil {
		return domainproject.Project{}, f.err
	}
	return f.project, nil
}

func testCatalog(t *testing.T) catalog.Catalog {
	t.Helper()
	cat, err := catalog.New(
		[]catalog.ModuleDefinition{{ID: "beebox-auth", Version: "v1"}},
		[]catalog.CapabilityDefinition{{ModuleID: "beebox-auth", ID: "password", Version: "v1"}},
		[]catalog.DataFieldDefinition{
			{ModuleID: "beebox-auth", CapabilityID: "password", CapabilityVersion: "v1", ID: "email", Version: "v1", ChangeKind: catalog.ChangeAdditive},
		},
	)
	if err != nil {
		t.Fatalf("catalog: %v", err)
	}
	return cat
}

func TestEnable_Valid(t *testing.T) {
	repo := memory.NewEnablementRepository()
	svc := applicationenablement.NewService(repo, fakeAuthorizer{project: domainproject.Project{ID: "p1", OrganizationID: "org1"}}, testCatalog(t))
	fields := []configuration.DataFieldReference{
		{ModuleID: "beebox-auth", CapabilityID: "password", CapabilityVersion: "v1", ID: "email", Version: "v1"},
	}
	got, err := svc.Enable(context.Background(), "p1", "org1", "beebox-auth", "v1", "password", "v1", fields)
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	if got.CapabilityID != "password" || len(got.DataFields) != 1 {
		t.Fatalf("got=%+v", got)
	}
}

func TestEnable_UnknownModule(t *testing.T) {
	repo := memory.NewEnablementRepository()
	svc := applicationenablement.NewService(repo, fakeAuthorizer{project: domainproject.Project{ID: "p1", OrganizationID: "org1"}}, testCatalog(t))
	_, err := svc.Enable(context.Background(), "p1", "org1", "unknown", "v1", "password", "v1", nil)
	if !apperror.IsCode(err, apperror.CodeValidation) {
		t.Fatalf("err=%v", err)
	}
}

func TestEnable_Duplicate(t *testing.T) {
	repo := memory.NewEnablementRepository()
	svc := applicationenablement.NewService(repo, fakeAuthorizer{project: domainproject.Project{ID: "p1", OrganizationID: "org1"}}, testCatalog(t))
	_, err := svc.Enable(context.Background(), "p1", "org1", "beebox-auth", "v1", "password", "v1", nil)
	if err != nil {
		t.Fatalf("first: %v", err)
	}
	_, err = svc.Enable(context.Background(), "p1", "org1", "beebox-auth", "v1", "password", "v1", nil)
	if !apperror.IsCode(err, apperror.CodeConflict) {
		t.Fatalf("err=%v", err)
	}
}

func TestEnable_Forbidden(t *testing.T) {
	repo := memory.NewEnablementRepository()
	svc := applicationenablement.NewService(repo, fakeAuthorizer{err: apperror.New(apperror.CodeForbidden, "forbidden")}, testCatalog(t))
	_, err := svc.Enable(context.Background(), "p1", "org2", "beebox-auth", "v1", "password", "v1", nil)
	if !apperror.IsCode(err, apperror.CodeForbidden) {
		t.Fatalf("err=%v", err)
	}
}

func TestDisable_NotFound(t *testing.T) {
	repo := memory.NewEnablementRepository()
	svc := applicationenablement.NewService(repo, fakeAuthorizer{project: domainproject.Project{ID: "p1", OrganizationID: "org1"}}, testCatalog(t))
	err := svc.Disable(context.Background(), "p1", "org1", "beebox-auth", "password")
	if !apperror.IsCode(err, apperror.CodeNotFound) {
		t.Fatalf("err=%v", err)
	}
}

func TestList_Empty(t *testing.T) {
	repo := memory.NewEnablementRepository()
	svc := applicationenablement.NewService(repo, fakeAuthorizer{project: domainproject.Project{ID: "p1", OrganizationID: "org1"}}, testCatalog(t))
	items, err := svc.List(context.Background(), "p1", "org1")
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	if len(items) != 0 {
		t.Fatalf("items=%v", items)
	}
}

func TestUnknownFieldRejected(t *testing.T) {
	repo := memory.NewEnablementRepository()
	svc := applicationenablement.NewService(repo, fakeAuthorizer{project: domainproject.Project{ID: "p1", OrganizationID: "org1"}}, testCatalog(t))
	fields := []configuration.DataFieldReference{
		{ModuleID: "beebox-auth", CapabilityID: "password", CapabilityVersion: "v1", ID: "nope", Version: "v1"},
	}
	_, err := svc.Enable(context.Background(), "p1", "org1", "beebox-auth", "v1", "password", "v1", fields)
	if err == nil || !apperror.IsCode(err, apperror.CodeValidation) {
		t.Fatalf("err=%v", err)
	}
}

func TestEnable_RejectsIncompatibleField(t *testing.T) {
	repo := memory.NewEnablementRepository()
	cat, err := catalog.New(
		[]catalog.ModuleDefinition{{ID: "beebox-auth", Version: "v1"}},
		[]catalog.CapabilityDefinition{{ModuleID: "beebox-auth", ID: "password", Version: "v1"}},
		[]catalog.DataFieldDefinition{
			{ModuleID: "beebox-auth", CapabilityID: "password", CapabilityVersion: "v1", ID: "password_hash", Version: "v1", ChangeKind: catalog.ChangeIncompatible},
		},
	)
	if err != nil {
		t.Fatalf("catalog: %v", err)
	}
	svc := applicationenablement.NewService(repo, fakeAuthorizer{project: domainproject.Project{ID: "p1", OrganizationID: "org1"}}, cat)
	fields := []configuration.DataFieldReference{
		{ModuleID: "beebox-auth", CapabilityID: "password", CapabilityVersion: "v1", ID: "password_hash", Version: "v1"},
	}
	_, err = svc.Enable(context.Background(), "p1", "org1", "beebox-auth", "v1", "password", "v1", fields)
	if err == nil || !apperror.IsCode(err, apperror.CodeValidation) {
		t.Fatalf("err=%v", err)
	}
}
