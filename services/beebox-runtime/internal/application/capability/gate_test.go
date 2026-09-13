package capability_test

import (
	"testing"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-runtime/apperror"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-runtime/internal/application/capability"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-runtime/internal/domain"
)

func TestRequireSessionAuth(t *testing.T) {
	ok := domain.AppliedConfiguration{
		AppliedVersion: 1, ModuleID: "beebox-auth", ModuleVersion: "v1",
		CapabilityID: "session", CapabilityVersion: "v1",
	}
	if err := capability.RequireSessionAuth(ok); err != nil {
		t.Fatal(err)
	}

	password := ok
	password.CapabilityID = "password"
	if apperror.CodeOf(capability.RequireSessionAuth(password)) != apperror.CodeForbidden {
		t.Fatal("expected forbidden for password capability")
	}

	v2 := ok
	v2.ModuleVersion = "v2"
	if apperror.CodeOf(capability.RequireSessionAuth(v2)) != apperror.CodeConflict {
		t.Fatal("expected conflict for unsupported module version")
	}

	none := domain.AppliedConfiguration{}
	if apperror.CodeOf(capability.RequireSessionAuth(none)) != apperror.CodeNotFound {
		t.Fatal("expected not found")
	}
}
