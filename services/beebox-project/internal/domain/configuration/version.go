package configuration

import "errors"

type LifecycleStatus string

const (
	StatusDraft     LifecycleStatus = "DRAFT"
	StatusValidated LifecycleStatus = "VALIDATED"
	StatusPublished LifecycleStatus = "PUBLISHED"
	StatusApplied   LifecycleStatus = "APPLIED"
)

var (
	ErrInvalidVersion             = errors.New("invalid configuration version")
	ErrInvalidLifecycleTransition = errors.New("invalid configuration lifecycle transition")
)

type Version struct {
	ProjectID     string
	Number        int
	Configuration Configuration
	Status        LifecycleStatus
}

func NewVersion(
	projectID string,
	number int,
	config Configuration,
) (Version, error) {
	if projectID == "" || number < 1 {
		return Version{}, ErrInvalidVersion
	}

	if config.ProjectID != projectID {
		return Version{}, ErrInvalidVersion
	}

	return Version{
		ProjectID:     projectID,
		Number:        number,
		Configuration: config,
		Status:        StatusDraft,
	}, nil
}

func (v Version) Transition(to LifecycleStatus) (Version, error) {
	if !canTransitionLifecycle(v.Status, to) {
		return Version{}, ErrInvalidLifecycleTransition
	}

	v.Status = to
	return v, nil
}

func canTransitionLifecycle(from LifecycleStatus, to LifecycleStatus) bool {
	switch from {
	case StatusDraft:
		return to == StatusValidated
	case StatusValidated:
		return to == StatusPublished
	case StatusPublished:
		return to == StatusApplied
	case StatusApplied:
		return false
	default:
		return false
	}
}