package http

import (
	"context"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-runtime/internal/domain"
)

type contextKey int

const (
	projectContextKey contextKey = 1
	appliedConfigKey  contextKey = 2
)

func withProjectContext(ctx context.Context, project domain.ProjectContext) context.Context {
	return context.WithValue(ctx, projectContextKey, project)
}

func projectContextFrom(ctx context.Context) (domain.ProjectContext, bool) {
	v, ok := ctx.Value(projectContextKey).(domain.ProjectContext)
	return v, ok
}

func withAppliedConfiguration(ctx context.Context, cfg domain.AppliedConfiguration) context.Context {
	return context.WithValue(ctx, appliedConfigKey, cfg)
}

func appliedConfigurationFrom(ctx context.Context) (domain.AppliedConfiguration, bool) {
	v, ok := ctx.Value(appliedConfigKey).(domain.AppliedConfiguration)
	return v, ok
}
