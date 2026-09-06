package fielddefinition

import (
	"testing"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
)

func mustField(t *testing.T, name string, kind FieldKind, required bool) FieldDefinition {
	t.Helper()
	f, err := NewFieldDefinition(name, kind, required)
	if err != nil {
		t.Fatalf("unexpected error building field %q: %v", name, err)
	}
	return f
}

func TestNewSchema_Success(t *testing.T) {
	fields := []FieldDefinition{
		mustField(t, "fullName", FieldKindString, true),
		mustField(t, "email", FieldKindString, true),
	}

	s, err := NewSchema("proj-1", fields)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Version != InitialVersion {
		t.Fatalf("expected version %d, got %d", InitialVersion, s.Version)
	}
	if len(s.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(s.Fields))
	}
}

func TestNewSchema_EmptyProjectID(t *testing.T) {
	_, err := NewSchema("", []FieldDefinition{mustField(t, "email", FieldKindString, true)})
	if apperror.CodeOf(err) != apperror.CodeInvalidInput {
		t.Fatalf("expected CodeInvalidInput, got %v", apperror.CodeOf(err))
	}
}

func TestNewSchema_NoFields(t *testing.T) {
	_, err := NewSchema("proj-1", nil)
	if apperror.CodeOf(err) != apperror.CodeInvalidInput {
		t.Fatalf("expected CodeInvalidInput, got %v", apperror.CodeOf(err))
	}
}

func TestNewSchema_DuplicateFieldName(t *testing.T) {
	fields := []FieldDefinition{
		mustField(t, "email", FieldKindString, true),
		mustField(t, "email", FieldKindString, false),
	}
	_, err := NewSchema("proj-1", fields)
	if apperror.CodeOf(err) != apperror.CodeInvalidInput {
		t.Fatalf("expected CodeInvalidInput, got %v", apperror.CodeOf(err))
	}
}

func TestFieldDefinition_InvalidNames(t *testing.T) {
	cases := []string{"", "full name", "1fullName", "full-name"}
	for _, name := range cases {
		_, err := NewFieldDefinition(name, FieldKindString, true)
		if apperror.CodeOf(err) != apperror.CodeInvalidInput {
			t.Fatalf("name %q: expected CodeInvalidInput, got %v", name, apperror.CodeOf(err))
		}
	}
}

func TestFieldDefinition_UnknownKind(t *testing.T) {
	_, err := NewFieldDefinition("email", FieldKind("UNKNOWN"), true)
	if apperror.CodeOf(err) != apperror.CodeInvalidInput {
		t.Fatalf("expected CodeInvalidInput, got %v", apperror.CodeOf(err))
	}
}

func TestSchema_NextVersion_IncrementsAndPreservesOriginal(t *testing.T) {
	original, err := NewSchema("proj-1", []FieldDefinition{mustField(t, "email", FieldKindString, true)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	next, err := original.NextVersion([]FieldDefinition{
		mustField(t, "email", FieldKindString, true),
		mustField(t, "isVerify", FieldKindBoolean, true),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if original.Version != InitialVersion {
		t.Fatalf("original schema mutated: version is %d", original.Version)
	}
	if len(original.Fields) != 1 {
		t.Fatalf("original schema fields mutated: got %d fields", len(original.Fields))
	}
	if next.Version != InitialVersion+1 {
		t.Fatalf("expected next version %d, got %d", InitialVersion+1, next.Version)
	}
	if next.ProjectID != original.ProjectID {
		t.Fatalf("expected ProjectID to carry over, got %q", next.ProjectID)
	}
}

func TestSchema_NextVersion_InvalidFieldsRejected(t *testing.T) {
	original, err := NewSchema("proj-1", []FieldDefinition{mustField(t, "email", FieldKindString, true)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = original.NextVersion(nil)
	if apperror.CodeOf(err) != apperror.CodeInvalidInput {
		t.Fatalf("expected CodeInvalidInput, got %v", apperror.CodeOf(err))
	}
}
