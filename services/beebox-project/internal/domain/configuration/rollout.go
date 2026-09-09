package configuration

import "errors"

var ErrInvalidRolloutState = errors.New("invalid configuration rollout state")

type RolloutState struct {
	ProjectID      string
	DesiredVersion int
	AppliedVersion int
}

func NewRolloutState(projectID string) (RolloutState, error) {
	if projectID == "" {
		return RolloutState{}, ErrInvalidRolloutState
	}

	return RolloutState{ProjectID: projectID}, nil
}

func (s RolloutState) WithDesiredVersion(version Version) (RolloutState, error) {
	if version.ProjectID != s.ProjectID {
		return RolloutState{}, ErrInvalidRolloutState
	}

	if version.Status != StatusPublished && version.Status != StatusApplied {
		return RolloutState{}, ErrInvalidLifecycleTransition
	}

	s.DesiredVersion = version.Number
	return s, nil
}

func (s RolloutState) WithAppliedVersion(version Version) (RolloutState, error) {
	if version.ProjectID != s.ProjectID {
		return RolloutState{}, ErrInvalidRolloutState
	}

	if version.Status != StatusApplied {
		return RolloutState{}, ErrInvalidLifecycleTransition
	}

	s.AppliedVersion = version.Number
	return s, nil
}