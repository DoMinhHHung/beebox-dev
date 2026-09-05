package fielddefinition

import "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"

const InitialVersion = 1

type Schema struct {
	ProjectID string
	Version   int
	Fields    []FieldDefinition
}

func NewSchema(projectID string, fields []FieldDefinition) (Schema, error) {
	if projectID == "" {
		return Schema{}, apperror.New(apperror.CodeInvalidInput, "projectID must not be empty")
	}
	validated, err := validateFields(fields)
	if err != nil {
		return Schema{}, err
	}
	return Schema{ProjectID: projectID, Version: InitialVersion, Fields: validated}, nil
}

func (s Schema) NextVersion(fields []FieldDefinition) (Schema, error) {
	validated, err := validateFields(fields)
	if err != nil {
		return Schema{}, err
	}
	return Schema{ProjectID: s.ProjectID, Version: s.Version + 1, Fields: validated}, nil
}

func validateFields(fields []FieldDefinition) ([]FieldDefinition, error) {
	if len(fields) == 0 {
		return nil, apperror.New(apperror.CodeInvalidInput, "schema must contain at least one field")
	}

	seen := make(map[string]struct{}, len(fields))
	copied := make([]FieldDefinition, len(fields))
	for i, f := range fields {
		if err := f.Validate(); err != nil {
			return nil, err
		}
		if _, dup := seen[f.Name]; dup {
			return nil, apperror.New(apperror.CodeInvalidInput, "duplicate field name: "+f.Name)
		}
		seen[f.Name] = struct{}{}
		copied[i] = f
	}
	return copied, nil
}