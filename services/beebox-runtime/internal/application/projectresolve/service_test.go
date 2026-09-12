package projectresolve_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-runtime/apperror"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-runtime/internal/application/projectresolve"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-runtime/internal/domain"
)

type fakeCreds struct {
	ctx domain.ProjectContext
	err error
}

func (f fakeCreds) VerifyPublic(context.Context, string, string) (domain.ProjectContext, error) {
	return f.ctx, f.err
}

type fakeConfigs struct {
	cfg domain.AppliedConfiguration
	err error
}

func (f fakeConfigs) GetAppliedConfiguration(context.Context, string) (domain.AppliedConfiguration, error) {
	return f.cfg, f.err
}

func TestResolve_Success(t *testing.T) {
	svc := projectresolve.NewService(
		fakeCreds{ctx: domain.ProjectContext{ProjectID: "p1", CredentialID: "c1"}},
		fakeConfigs{},
	)
	got, err := svc.Resolve(context.Background(), "p1", "raw")
	if err != nil {
		t.Fatal(err)
	}
	if got.ProjectID != "p1" || got.CredentialID != "c1" {
		t.Fatalf("%#v", got)
	}
}

func TestResolve_RejectsMissingCredential(t *testing.T) {
	svc := projectresolve.NewService(fakeCreds{}, fakeConfigs{})
	_, err := svc.Resolve(context.Background(), "p1", "")
	if apperror.CodeOf(err) != apperror.CodeUnauthenticated {
		t.Fatalf("got %v", err)
	}
}

func TestResolve_RejectsProjectMismatch(t *testing.T) {
	svc := projectresolve.NewService(
		fakeCreds{ctx: domain.ProjectContext{ProjectID: "other", CredentialID: "c1"}},
		fakeConfigs{},
	)
	_, err := svc.Resolve(context.Background(), "p1", "raw")
	if apperror.CodeOf(err) != apperror.CodeUnauthenticated {
		t.Fatalf("got %v", err)
	}
}

func TestResolve_PropagatesDependencyFailure(t *testing.T) {
	svc := projectresolve.NewService(
		fakeCreds{err: apperror.Wrap(apperror.CodeDependencyFailure, "project service unavailable", errors.New("down"))},
		fakeConfigs{},
	)
	_, err := svc.Resolve(context.Background(), "p1", "raw")
	if apperror.CodeOf(err) != apperror.CodeDependencyFailure {
		t.Fatalf("got %v", err)
	}
}
