package fielddefinition

import (
	"regexp"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
)

var fieldNamePattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

type FieldDefinition struct {
	Name     string
	Kind     FieldKind
	Required bool
}

// NewFieldDefinition creates a validated field definition with the specified name, kind, and required status.
// It returns an empty definition and an error when the provided values are invalid.
func NewFieldDefinition(name string, kind FieldKind, required bool) (FieldDefinition, error) {
	f := FieldDefinition{Name: name, Kind: kind, Required: required}
	if err := f.Validate(); err != nil {
		return FieldDefinition{}, err
	}
	return f, nil
}

func (f FieldDefinition) Validate() error {
	if !fieldNamePattern.MatchString(f.Name) {
		return apperror.New(apperror.CodeInvalidInput, "field name must match ^[a-zA-Z_][a-zA-Z0-9_]*$: "+f.Name)
	}
	if !f.Kind.Valid() {
		return apperror.New(apperror.CodeInvalidInput, "unknown field kind: "+string(f.Kind))
	}
	return nil
}
