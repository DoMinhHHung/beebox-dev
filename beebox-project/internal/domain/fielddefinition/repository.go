package fielddefinition

import "context"

type Repository interface {
	Save(ctx context.Context, s Schema) error
	FindLatest(ctx context.Context, projectID string) (Schema, error)
	FindVersion(ctx context.Context, projectID string, version int) (Schema, error)
}