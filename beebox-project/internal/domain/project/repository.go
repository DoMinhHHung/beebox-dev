package project

import "context"

type Repository interface {
	Save(ctx context.Context, p Project) error
	FindByID(ctx context.Context, id string) (Project, error)
}
