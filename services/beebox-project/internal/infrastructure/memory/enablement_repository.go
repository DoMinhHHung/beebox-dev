package memory

import (
	"context"
	"sync"

	applicationenablement "github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/application/enablement"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/configuration"
	domainenablement "github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/enablement"
)

type enablementKey struct {
	projectID    string
	moduleID     string
	capabilityID string
}

type EnablementRepository struct {
	mu    sync.Mutex
	items map[enablementKey]domainenablement.Enablement
}

func NewEnablementRepository() *EnablementRepository {
	return &EnablementRepository{items: map[enablementKey]domainenablement.Enablement{}}
}

var _ applicationenablement.Repository = (*EnablementRepository)(nil)

func (r *EnablementRepository) Save(_ context.Context, item domainenablement.Enablement) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := enablementKey{item.ProjectID, item.ModuleID, item.CapabilityID}
	if _, exists := r.items[key]; exists {
		return applicationenablement.ErrEnablementAlreadyExists
	}
	r.items[key] = cloneEnablement(item)
	return nil
}

func (r *EnablementRepository) Update(_ context.Context, item domainenablement.Enablement) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := enablementKey{item.ProjectID, item.ModuleID, item.CapabilityID}
	if _, exists := r.items[key]; !exists {
		return applicationenablement.ErrEnablementNotFound
	}
	r.items[key] = cloneEnablement(item)
	return nil
}

func (r *EnablementRepository) Get(_ context.Context, projectID, moduleID, capabilityID string) (domainenablement.Enablement, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	item, exists := r.items[enablementKey{projectID, moduleID, capabilityID}]
	if !exists {
		return domainenablement.Enablement{}, applicationenablement.ErrEnablementNotFound
	}
	return cloneEnablement(item), nil
}

func (r *EnablementRepository) ListByProject(_ context.Context, projectID string) ([]domainenablement.Enablement, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []domainenablement.Enablement
	for key, item := range r.items {
		if key.projectID == projectID {
			out = append(out, cloneEnablement(item))
		}
	}
	return out, nil
}

func (r *EnablementRepository) Delete(_ context.Context, projectID, moduleID, capabilityID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := enablementKey{projectID, moduleID, capabilityID}
	if _, exists := r.items[key]; !exists {
		return applicationenablement.ErrEnablementNotFound
	}
	delete(r.items, key)
	return nil
}

func cloneEnablement(item domainenablement.Enablement) domainenablement.Enablement {
	fields := append([]configuration.DataFieldReference(nil), item.DataFields...)
	cloned, _ := domainenablement.New(item.ProjectID, item.ModuleID, item.ModuleVersion, item.CapabilityID, item.CapabilityVersion)
	cloned, _ = cloned.WithDataFields(fields)
	return cloned
}
