package enablement_test

import (
	"errors"
	"testing"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/configuration"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/enablement"
)

func TestNew_CreatesEnablement(t *testing.T) {
	got, err := enablement.New("p1", "beebox-auth", "v1", "password", "v1")
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	if got.ProjectID != "p1" || got.ModuleID != "beebox-auth" || got.CapabilityID != "password" {
		t.Fatalf("unexpected %+v", got)
	}
}

func TestNew_RejectsMissingFields(t *testing.T) {
	_, err := enablement.New("", "beebox-auth", "v1", "password", "v1")
	if !errors.Is(err, enablement.ErrInvalidEnablement) {
		t.Fatalf("err=%v", err)
	}
}

func TestWithDataFields_RejectsDuplicate(t *testing.T) {
	item, err := enablement.New("p1", "beebox-auth", "v1", "password", "v1")
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	field := configuration.DataFieldReference{
		ModuleID: "beebox-auth", CapabilityID: "password", CapabilityVersion: "v1", ID: "email", Version: "v1",
	}
	_, err = item.WithDataFields([]configuration.DataFieldReference{field, field})
	if !errors.Is(err, enablement.ErrDuplicateField) {
		t.Fatalf("err=%v", err)
	}
}
