package project

import (
	"time"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
)

type Tier string

const (
	TierFree       Tier = "free"
	TierPro        Tier = "pro"
	TierEnterprise Tier = "enterprise"
)

type Project struct {
	ID        string
	OwnerID   string
	Name      string
	Tier      Tier
	CreatedAt time.Time
}

func NewProject(id, ownerID, name string, tier Tier, createdAt time.Time) (Project, error) {
	p := Project{
		ID:        id,
		OwnerID:   ownerID,
		Name:      name,
		Tier:      tier,
		CreatedAt: createdAt,
	}

	if err := p.validate(); err != nil {
		return Project{}, err
	}

	return p, nil
}

func (p Project) validate() error {
	if p.OwnerID == "" {
		return apperror.New(apperror.CodeInvalidInput, "owner id must not be empty")
	}
	if p.Name == "" {
		return apperror.New(apperror.CodeInvalidInput, "name must not be empty")
	}
	if len(p.Name) > 100 {
		return apperror.New(apperror.CodeInvalidInput, "name must not exceed 100 characters")
	}
	switch p.Tier {
	case TierFree, TierPro, TierEnterprise:
	default:
		return apperror.New(apperror.CodeInvalidInput, "tier must be one of free/pro/enterprise")
	}
	return nil
}