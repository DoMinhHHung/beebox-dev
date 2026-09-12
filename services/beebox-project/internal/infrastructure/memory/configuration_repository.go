package memory

import (
	"context"
	"sync"

	applicationconfiguration "github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/application/configuration"
	domainconfiguration "github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/configuration"
)

type ConfigurationRepository struct {
	mu       sync.Mutex
	versions map[string]map[int]domainconfiguration.Version
	rollouts map[string]domainconfiguration.RolloutState
}

func NewConfigurationRepository() *ConfigurationRepository {
	return &ConfigurationRepository{
		versions: make(map[string]map[int]domainconfiguration.Version),
		rollouts: make(map[string]domainconfiguration.RolloutState),
	}
}

var _ applicationconfiguration.Repository = (*ConfigurationRepository)(nil)

func (r *ConfigurationRepository) CreateVersion(_ context.Context, version domainconfiguration.Version) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	versions := r.versions[version.ProjectID]
	if versions == nil {
		versions = make(map[int]domainconfiguration.Version)
		r.versions[version.ProjectID] = versions
	}
	if _, exists := versions[version.Number]; exists {
		return applicationconfiguration.ErrVersionConflict
	}
	versions[version.Number] = version
	return nil
}

func (r *ConfigurationRepository) GetVersion(_ context.Context, projectID string, number int) (domainconfiguration.Version, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	version, ok := r.versions[projectID][number]
	if !ok {
		return domainconfiguration.Version{}, applicationconfiguration.ErrVersionNotFound
	}
	return version, nil
}

func (r *ConfigurationRepository) GetLatestVersion(_ context.Context, projectID string) (domainconfiguration.Version, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	versions := r.versions[projectID]
	var latest domainconfiguration.Version
	for _, version := range versions {
		if version.Number > latest.Number {
			latest = version
		}
	}
	if latest.Number == 0 {
		return domainconfiguration.Version{}, applicationconfiguration.ErrVersionNotFound
	}
	return latest, nil
}

func (r *ConfigurationRepository) UpdateVersion(_ context.Context, version domainconfiguration.Version) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.versions[version.ProjectID][version.Number]; !ok {
		return applicationconfiguration.ErrVersionNotFound
	}
	r.versions[version.ProjectID][version.Number] = version
	return nil
}

func (r *ConfigurationRepository) GetRollout(_ context.Context, projectID string) (domainconfiguration.RolloutState, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	state, ok := r.rollouts[projectID]
	if !ok {
		return domainconfiguration.RolloutState{}, applicationconfiguration.ErrRolloutNotFound
	}
	return state, nil
}

func (r *ConfigurationRepository) SaveRollout(_ context.Context, state domainconfiguration.RolloutState) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rollouts[state.ProjectID] = state
	return nil
}
