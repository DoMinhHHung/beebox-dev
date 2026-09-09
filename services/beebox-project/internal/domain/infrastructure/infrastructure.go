package infrastructure

import "errors"

var ErrInvalidDesiredState = errors.New("invalid infrastructure desired state")

type DesiredState struct {
	ProjectID string
	Provider  string
	Region    string
	Plan      string
}

func NewDesiredState(
	projectID string,
	provider string,
	region string,
	plan string,
) (DesiredState, error) {
	if projectID == "" || provider == "" || region == "" || plan == "" {
		return DesiredState{}, ErrInvalidDesiredState
	}

	return DesiredState{
		ProjectID: projectID,
		Provider:  provider,
		Region:    region,
		Plan:      plan,
	}, nil
}
