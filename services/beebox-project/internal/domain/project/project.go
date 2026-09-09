package project

import "errors"

type Status string

const (
	StatusDraft     Status = "DRAFT"
	StatusActive    Status = "ACTIVE"
	StatusSuspended Status = "SUSPENDED"
	StatusArchived  Status = "ARCHIVED"
)

var (
	ErrInvalidProject    = errors.New("invalid project")
	ErrInvalidTransition = errors.New("invalid project transition")
)

type Project struct {
	ID             string
	OrganizationID string
	Status         Status
}

func New(id string, organizationID string) (Project, error) {
	if id == "" || organizationID == "" {
		return Project{}, ErrInvalidProject
	}

	return Project{
		ID:             id,
		OrganizationID: organizationID,
		Status:         StatusDraft,
	}, nil
}

func (p Project) Transition(to Status) (Project, error) {
	if !canTransition(p.Status, to) {
		return Project{}, ErrInvalidTransition
	}

	p.Status = to
	return p, nil
}

func canTransition(from Status, to Status) bool {
	switch from {
	case StatusDraft:
		return to == StatusActive
	case StatusActive:
		return to == StatusSuspended || to == StatusArchived
	case StatusSuspended:
		return to == StatusActive || to == StatusArchived
	case StatusArchived:
		return false
	default:
		return false
	}
}
