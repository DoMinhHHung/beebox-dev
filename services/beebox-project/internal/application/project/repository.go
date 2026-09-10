package project

import (
	"context"
	"errors"

	domainproject "github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/project"
)

var (
	ErrProjectNotFound      = errors.New("project not found")
	ErrProjectAlreadyExists = errors.New("project already exists")
)

type Repository interface {
	Create(ctx context.Context, p domainproject.Project) error
	Get(ctx context.Context, id string) (domainproject.Project, error)
	Update(ctx context.Context, p domainproject.Project) error
}
