package configuration_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/apperror"
	applicationconfiguration "github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/application/configuration"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/application/project"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/catalog"
	domainconfiguration "github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/configuration"
	domainproject "github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/project"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/infrastructure/memory"
)

func newService(t *testing.T) (*applicationconfiguration.Service, *project.Service) {
	t.Helper()
	projectsRepo := memory.NewProjectRepository()
	projects := project.NewService(projectsRepo)
	if _, err := projects.Create(context.Background(), "project-1", "organization-1"); err != nil {
		t.Fatalf("create project: %v", err)
	}
	cat, err := catalog.New(
		[]catalog.ModuleDefinition{{ID: "auth", Version: "v1"}},
		[]catalog.CapabilityDefinition{{ModuleID: "auth", ID: "login", Version: "v1"}},
		nil,
	)
	if err != nil {
		t.Fatalf("catalog: %v", err)
	}
	return applicationconfiguration.NewService(memory.NewConfigurationRepository(), projects, cat), projects
}

func validConfiguration(t *testing.T, projectID string) domainconfiguration.Configuration {
	t.Helper()
	config, err := domainconfiguration.New(projectID, "auth", "v1", "login", "v1")
	if err != nil {
		t.Fatalf("create config: %v", err)
	}
	return config
}

func TestCreateOrUpdate_CreatesVersions(t *testing.T) {
	svc, _ := newService(t)
	ctx := context.Background()
	first, err := svc.CreateOrUpdate(ctx, "project-1", "organization-1", validConfiguration(t, "project-1"))
	if err != nil || first.Number != 1 {
		t.Fatalf("expected version 1, got %#v, %v", first, err)
	}
	second, err := svc.CreateOrUpdate(ctx, "project-1", "organization-1", validConfiguration(t, "project-1"))
	if err != nil || second.Number != 2 {
		t.Fatalf("expected version 2, got %#v, %v", second, err)
	}
}

func TestCreateOrUpdate_RejectsInvalidConfiguration(t *testing.T) {
	svc, _ := newService(t)
	_, err := svc.CreateOrUpdate(context.Background(), "project-1", "organization-1", domainconfiguration.Configuration{ProjectID: "project-1"})
	if apperror.CodeOf(err) != apperror.CodeValidation {
		t.Fatalf("expected validation, got %v", apperror.CodeOf(err))
	}
}

func TestLifecycleAndRollout(t *testing.T) {
	svc, _ := newService(t)
	ctx := context.Background()
	version, err := svc.CreateOrUpdate(ctx, "project-1", "organization-1", validConfiguration(t, "project-1"))
	if err != nil {
		t.Fatal(err)
	}
	version, err = svc.Transition(ctx, "project-1", "organization-1", version.Number, domainconfiguration.StatusValidated)
	if err != nil {
		t.Fatal(err)
	}
	version, err = svc.Transition(ctx, "project-1", "organization-1", version.Number, domainconfiguration.StatusPublished)
	if err != nil {
		t.Fatal(err)
	}
	state, err := svc.Rollout(ctx, "project-1", "organization-1", version.Number)
	if err != nil || state.DesiredVersion != version.Number {
		t.Fatalf("unexpected rollout %#v, %v", state, err)
	}
}

func TestAccessAndNotFound(t *testing.T) {
	svc, _ := newService(t)
	_, err := svc.Current(context.Background(), "missing", "organization-1")
	if apperror.CodeOf(err) != apperror.CodeNotFound {
		t.Fatalf("expected project not found, got %v", apperror.CodeOf(err))
	}
	_, err = svc.Current(context.Background(), "project-1", "organization-2")
	if apperror.CodeOf(err) != apperror.CodeForbidden {
		t.Fatalf("expected forbidden, got %v", apperror.CodeOf(err))
	}
}

func publishVersion(t *testing.T, svc *applicationconfiguration.Service, projectID, organizationID string) domainconfiguration.Version {
	t.Helper()
	ctx := context.Background()
	version, err := svc.CreateOrUpdate(ctx, projectID, organizationID, validConfiguration(t, projectID))
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	version, err = svc.Transition(ctx, projectID, organizationID, version.Number, domainconfiguration.StatusValidated)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	version, err = svc.Transition(ctx, projectID, organizationID, version.Number, domainconfiguration.StatusPublished)
	if err != nil {
		t.Fatalf("publish: %v", err)
	}
	return version
}

func TestApply_PublishedSetsStatusAndAppliedVersion(t *testing.T) {
	svc, _ := newService(t)
	ctx := context.Background()
	version := publishVersion(t, svc, "project-1", "organization-1")

	applied, state, err := svc.Apply(ctx, "project-1", "organization-1", version.Number)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if applied.Status != domainconfiguration.StatusApplied {
		t.Fatalf("expected APPLIED status, got %s", applied.Status)
	}
	if state.AppliedVersion != version.Number {
		t.Fatalf("expected applied_version %d, got %d", version.Number, state.AppliedVersion)
	}

	loaded, err := svc.Version(ctx, "project-1", "organization-1", version.Number)
	if err != nil {
		t.Fatalf("reload version: %v", err)
	}
	if loaded.Status != domainconfiguration.StatusApplied {
		t.Fatalf("persisted status want APPLIED, got %s", loaded.Status)
	}
}

func TestApply_IdempotentWhenAlreadyApplied(t *testing.T) {
	svc, _ := newService(t)
	ctx := context.Background()
	version := publishVersion(t, svc, "project-1", "organization-1")

	first, firstState, err := svc.Apply(ctx, "project-1", "organization-1", version.Number)
	if err != nil {
		t.Fatalf("first apply: %v", err)
	}
	second, secondState, err := svc.Apply(ctx, "project-1", "organization-1", version.Number)
	if err != nil {
		t.Fatalf("second apply: %v", err)
	}
	if second.Status != domainconfiguration.StatusApplied {
		t.Fatalf("expected APPLIED, got %s", second.Status)
	}
	if secondState.AppliedVersion != firstState.AppliedVersion || secondState.AppliedVersion != first.Number {
		t.Fatalf("unexpected rollout after idempotent apply: first=%#v second=%#v", firstState, secondState)
	}
}

func TestApply_RejectsDraftAndValidated(t *testing.T) {
	svc, _ := newService(t)
	ctx := context.Background()
	draft, err := svc.CreateOrUpdate(ctx, "project-1", "organization-1", validConfiguration(t, "project-1"))
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = svc.Apply(ctx, "project-1", "organization-1", draft.Number)
	if apperror.CodeOf(err) != apperror.CodeConflict {
		t.Fatalf("draft apply expected conflict, got %v", err)
	}

	validated, err := svc.Transition(ctx, "project-1", "organization-1", draft.Number, domainconfiguration.StatusValidated)
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = svc.Apply(ctx, "project-1", "organization-1", validated.Number)
	if apperror.CodeOf(err) != apperror.CodeConflict {
		t.Fatalf("validated apply expected conflict, got %v", err)
	}
}

func TestApply_RejectsMissingVersionAndWrongOrg(t *testing.T) {
	svc, _ := newService(t)
	ctx := context.Background()
	_, _, err := svc.Apply(ctx, "project-1", "organization-1", 99)
	if apperror.CodeOf(err) != apperror.CodeNotFound {
		t.Fatalf("missing version expected not found, got %v", err)
	}
	publishVersion(t, svc, "project-1", "organization-1")
	_, _, err = svc.Apply(ctx, "project-1", "organization-2", 1)
	if apperror.CodeOf(err) != apperror.CodeForbidden {
		t.Fatalf("wrong org expected forbidden, got %v", err)
	}
}

type applyScriptRepo struct {
	inner       applicationconfiguration.Repository
	failOnApply bool
	applyCalls  int
	lastVersion domainconfiguration.Version
	lastRollout domainconfiguration.RolloutState
}

func (r *applyScriptRepo) CreateVersion(ctx context.Context, v domainconfiguration.Version) error {
	return r.inner.CreateVersion(ctx, v)
}
func (r *applyScriptRepo) GetVersion(ctx context.Context, projectID string, number int) (domainconfiguration.Version, error) {
	return r.inner.GetVersion(ctx, projectID, number)
}
func (r *applyScriptRepo) GetLatestVersion(ctx context.Context, projectID string) (domainconfiguration.Version, error) {
	return r.inner.GetLatestVersion(ctx, projectID)
}
func (r *applyScriptRepo) UpdateVersion(ctx context.Context, v domainconfiguration.Version) error {
	return r.inner.UpdateVersion(ctx, v)
}
func (r *applyScriptRepo) GetRollout(ctx context.Context, projectID string) (domainconfiguration.RolloutState, error) {
	return r.inner.GetRollout(ctx, projectID)
}
func (r *applyScriptRepo) SaveRollout(ctx context.Context, state domainconfiguration.RolloutState) error {
	return r.inner.SaveRollout(ctx, state)
}
func (r *applyScriptRepo) Apply(ctx context.Context, version domainconfiguration.Version, state domainconfiguration.RolloutState) error {
	r.applyCalls++
	r.lastVersion = version
	r.lastRollout = state
	if r.failOnApply {
		return errors.New("forced apply failure")
	}
	return r.inner.Apply(ctx, version, state)
}

func TestApply_RepositoryFailureDoesNotLeavePartialSuccessWhenRepoIsAtomic(t *testing.T) {
	projectsRepo := memory.NewProjectRepository()
	projects := project.NewService(projectsRepo)
	if _, err := projects.Create(context.Background(), "project-1", "organization-1"); err != nil {
		t.Fatal(err)
	}
	cat, err := catalog.New(
		[]catalog.ModuleDefinition{{ID: "auth", Version: "v1"}},
		[]catalog.CapabilityDefinition{{ModuleID: "auth", ID: "login", Version: "v1"}},
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	inner := memory.NewConfigurationRepository()
	scripted := &applyScriptRepo{inner: inner, failOnApply: true}
	svc := applicationconfiguration.NewService(scripted, projects, cat)
	ctx := context.Background()
	version := publishVersion(t, svc, "project-1", "organization-1")

	_, _, err = svc.Apply(ctx, "project-1", "organization-1", version.Number)
	if apperror.CodeOf(err) != apperror.CodeDependencyFailure {
		t.Fatalf("expected dependency failure, got %v", err)
	}
	if scripted.applyCalls != 1 {
		t.Fatalf("expected one apply call, got %d", scripted.applyCalls)
	}

	loaded, err := inner.GetVersion(ctx, "project-1", version.Number)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Status != domainconfiguration.StatusPublished {
		t.Fatalf("atomic failure must leave version unpublished, got %s", loaded.Status)
	}
	_, err = inner.GetRollout(ctx, "project-1")
	if !errors.Is(err, applicationconfiguration.ErrRolloutNotFound) {
		t.Fatalf("expected no rollout after failed apply, got %#v %v", err, err)
	}
}

func TestApply_RepairsAppliedVersionWhenStatusAlreadyApplied(t *testing.T) {
	svc, _ := newService(t)
	ctx := context.Background()
	version := publishVersion(t, svc, "project-1", "organization-1")
	legacy, err := svc.Transition(ctx, "project-1", "organization-1", version.Number, domainconfiguration.StatusApplied)
	if err != nil {
		t.Fatal(err)
	}
	if legacy.Status != domainconfiguration.StatusApplied {
		t.Fatal("expected legacy APPLIED status")
	}

	applied, state, err := svc.Apply(ctx, "project-1", "organization-1", version.Number)
	if err != nil {
		t.Fatalf("repair apply: %v", err)
	}
	if applied.Status != domainconfiguration.StatusApplied {
		t.Fatalf("status %s", applied.Status)
	}
	if state.AppliedVersion != version.Number {
		t.Fatalf("applied_version %d", state.AppliedVersion)
	}
}

func activateProject(t *testing.T, projects *project.Service, projectID string) {
	t.Helper()
	if _, err := projects.Transition(context.Background(), projectID, domainproject.StatusActive); err != nil {
		t.Fatalf("activate project: %v", err)
	}
}

func TestGetAppliedConfiguration_ReturnsAppliedSnapshot(t *testing.T) {
	svc, projects := newService(t)
	ctx := context.Background()
	activateProject(t, projects, "project-1")
	version := publishVersion(t, svc, "project-1", "organization-1")
	applied, state, err := svc.Apply(ctx, "project-1", "organization-1", version.Number)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	if _, err := svc.CreateOrUpdate(ctx, "project-1", "organization-1", validConfiguration(t, "project-1")); err != nil {
		t.Fatal(err)
	}

	snapshot, err := svc.GetAppliedConfiguration(ctx, "project-1")
	if err != nil {
		t.Fatalf("get applied: %v", err)
	}
	if snapshot.AppliedVersion != state.AppliedVersion || snapshot.AppliedVersion != applied.Number {
		t.Fatalf("applied version mismatch: snapshot=%d state=%d applied=%d", snapshot.AppliedVersion, state.AppliedVersion, applied.Number)
	}
	if snapshot.ProjectStatus != domainproject.StatusActive {
		t.Fatalf("status %s", snapshot.ProjectStatus)
	}
	if snapshot.ModuleID != "auth" || snapshot.CapabilityID != "login" {
		t.Fatalf("unexpected module/capability %#v", snapshot)
	}
}

func TestGetAppliedConfiguration_IgnoresDesiredAndLatest(t *testing.T) {
	svc, projects := newService(t)
	ctx := context.Background()
	activateProject(t, projects, "project-1")
	v1 := publishVersion(t, svc, "project-1", "organization-1")
	if _, _, err := svc.Apply(ctx, "project-1", "organization-1", v1.Number); err != nil {
		t.Fatal(err)
	}
	v2 := publishVersion(t, svc, "project-1", "organization-1")
	if _, err := svc.Rollout(ctx, "project-1", "organization-1", v2.Number); err != nil {
		t.Fatal(err)
	}

	snapshot, err := svc.GetAppliedConfiguration(ctx, "project-1")
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.AppliedVersion != v1.Number {
		t.Fatalf("expected applied %d, got %d", v1.Number, snapshot.AppliedVersion)
	}
}

func TestGetAppliedConfiguration_NotFoundWhenMissingOrZero(t *testing.T) {
	svc, projects := newService(t)
	ctx := context.Background()
	activateProject(t, projects, "project-1")

	_, err := svc.GetAppliedConfiguration(ctx, "project-1")
	if apperror.CodeOf(err) != apperror.CodeNotFound {
		t.Fatalf("expected not found without rollout, got %v", err)
	}

	version := publishVersion(t, svc, "project-1", "organization-1")
	if _, err := svc.Rollout(ctx, "project-1", "organization-1", version.Number); err != nil {
		t.Fatal(err)
	}
	_, err = svc.GetAppliedConfiguration(ctx, "project-1")
	if apperror.CodeOf(err) != apperror.CodeNotFound {
		t.Fatalf("expected not found when applied_version=0, got %v", err)
	}
}

func TestGetAppliedConfiguration_RejectsInactiveProject(t *testing.T) {
	svc, _ := newService(t)
	ctx := context.Background()
	version := publishVersion(t, svc, "project-1", "organization-1")
	if _, _, err := svc.Apply(ctx, "project-1", "organization-1", version.Number); err != nil {
		t.Fatal(err)
	}
	_, err := svc.GetAppliedConfiguration(ctx, "project-1")
	if apperror.CodeOf(err) != apperror.CodeForbidden {
		t.Fatalf("expected forbidden for draft project, got %v", err)
	}
}

func TestGetAppliedConfiguration_InconsistentAppliedVersion(t *testing.T) {
	projectsRepo := memory.NewProjectRepository()
	projects := project.NewService(projectsRepo)
	if _, err := projects.Create(context.Background(), "project-1", "organization-1"); err != nil {
		t.Fatal(err)
	}
	activateProject(t, projects, "project-1")
	cat, err := catalog.New(
		[]catalog.ModuleDefinition{{ID: "auth", Version: "v1"}},
		[]catalog.CapabilityDefinition{{ModuleID: "auth", ID: "login", Version: "v1"}},
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	inner := memory.NewConfigurationRepository()
	svc := applicationconfiguration.NewService(inner, projects, cat)
	ctx := context.Background()
	version := publishVersion(t, svc, "project-1", "organization-1")
	if _, _, err := svc.Apply(ctx, "project-1", "organization-1", version.Number); err != nil {
		t.Fatal(err)
	}
	if err := inner.SaveRollout(ctx, domainconfiguration.RolloutState{
		ProjectID:      "project-1",
		DesiredVersion: version.Number,
		AppliedVersion: 99,
	}); err != nil {
		t.Fatal(err)
	}
	_, err = svc.GetAppliedConfiguration(ctx, "project-1")
	if apperror.CodeOf(err) != apperror.CodeInternal {
		t.Fatalf("expected internal for inconsistent applied version, got %v", err)
	}
}
