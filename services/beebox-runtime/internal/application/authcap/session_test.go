package authcap_test

import (
	"context"
	"testing"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-runtime/apperror"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-runtime/internal/application/authcap"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-runtime/internal/domain"
)

type fakeIdentity struct {
	session authcap.Session
	err     error
	token   string
}

func (f *fakeIdentity) GetSession(_ context.Context, token string) (authcap.Session, error) {
	f.token = token
	return f.session, f.err
}

func TestCurrentSession_HappyPath(t *testing.T) {
	identity := &fakeIdentity{session: authcap.Session{UserID: "u1", SessionID: "s1", OrganizationID: "o1"}}
	svc := authcap.NewService(identity)
	cfg := domain.AppliedConfiguration{
		AppliedVersion: 1, ModuleID: "beebox-auth", ModuleVersion: "v1",
		CapabilityID: "session", CapabilityVersion: "v1",
	}
	got, err := svc.CurrentSession(context.Background(), domain.ProjectContext{ProjectID: "p1"}, cfg, "Bearer user-token")
	if err != nil {
		t.Fatal(err)
	}
	if got.ProjectID != "p1" || got.UserID != "u1" || got.SessionID != "s1" {
		t.Fatalf("%#v", got)
	}
	if identity.token != "user-token" {
		t.Fatalf("token %q", identity.token)
	}
}

func TestCurrentSession_RejectsMissingBearerAndWrongCapability(t *testing.T) {
	svc := authcap.NewService(&fakeIdentity{})
	cfg := domain.AppliedConfiguration{
		AppliedVersion: 1, ModuleID: "beebox-auth", ModuleVersion: "v1",
		CapabilityID: "session", CapabilityVersion: "v1",
	}
	_, err := svc.CurrentSession(context.Background(), domain.ProjectContext{ProjectID: "p1"}, cfg, "")
	if apperror.CodeOf(err) != apperror.CodeUnauthenticated {
		t.Fatalf("got %v", err)
	}
	cfg.CapabilityID = "password"
	_, err = svc.CurrentSession(context.Background(), domain.ProjectContext{ProjectID: "p1"}, cfg, "Bearer t")
	if apperror.CodeOf(err) != apperror.CodeForbidden {
		t.Fatalf("got %v", err)
	}
}
