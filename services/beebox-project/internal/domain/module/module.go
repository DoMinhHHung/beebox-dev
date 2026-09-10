package module

import "errors"

var ErrInvalidModule = errors.New("invalid module")

type Module struct {
	ProjectID string
	ID        string
}

func New(projectID string, id string) (Module, error) {
	if projectID == "" || id == "" {
		return Module{}, ErrInvalidModule
	}

	return Module{
		ProjectID: projectID,
		ID:        id,
	}, nil
}
