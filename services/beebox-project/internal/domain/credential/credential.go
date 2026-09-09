package credential

import "errors"

var ErrInvalidCredential = errors.New("invalid credential")

type Credential struct {
	ProjectID string
	ID        string
	Name      string
}

func New(projectID string, id string, name string) (Credential, error) {
	if projectID == "" || id == "" || name == "" {
		return Credential{}, ErrInvalidCredential
	}

	return Credential{
		ProjectID: projectID,
		ID:        id,
		Name:      name,
	}, nil
}
