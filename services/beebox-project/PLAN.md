# `beebox-project` Implementation Plan

> Scope: bootstrap the `services/beebox-project` control-plane service on the `beebox/feat-setup-docs` branch.
>
> Repository: `DoMinhHHung/beebox-dev`
>
> Observed remote baseline: `main@bf0ce9763fecc27f467c7e9b7a3f5a0f2f1f9850`.
>
> Current remote tree at that baseline contains only `.gitignore`; the requested local branch is not present on the remote yet. fileciteturn5file0
>
> Rule: mark a step `[x]` only after the implementation and its verification are complete. Keep this file updated in the same branch/commit as the work it tracks.

## Goals

- Establish `beebox-project` as a standalone Go microservice.
- Use Clean Architecture with explicit dependency direction.
- Introduce a service-local `apperror` package as the normalized application error model.
- Keep `beebox-project` focused on project/control-plane concerns, not gateway or application-business execution.
- Define boundaries for Project, Module, Capability, Configuration, Credential, and Infrastructure desired state.
- Keep PostgreSQL/Supabase and Redis/Upstash behind infrastructure adapters.
- Make the service independently buildable, testable, and CI-compatible.

## Non-goals

- No API gateway.
- No authentication implementation.
- No e-commerce/payment business logic.
- No arbitrary developer-defined backend schema.
- No direct ownership of data belonging to other services/modules.
- No premature external messaging/provisioning implementation unless required to establish a contract.

---

## Steps

### 1. Bootstrap `services/beebox-project`
- [x] Create `services/beebox-project/`.
- [x] Initialize `go.mod` for `go1.26.5`.
- [x] Keep the module independently runnable/testable.

**Verify**
```bash
cd services/beebox-project
go test ./...
```

### 2. Establish the Clean Architecture skeleton
- [x] Add `cmd/server/`.
- [x] Add `internal/domain/`.
- [x] Add `internal/application/`.
- [x] Add `internal/interfaces/`.
- [x] Add `internal/infrastructure/`.
- [x] Avoid placeholder abstractions that have no current responsibility.

**Verify**
```bash
go test ./...
go vet ./...
```

### 3. Add centralized application errors
- [x] Add `apperror/apperror.go`.
- [x] Add `apperror/apperror_test.go`.
- [x] Define stable machine-readable error identity.
- [x] Support the initial categories required by the service: validation, unauthenticated, forbidden, not found, conflict, dependency failure, and internal failure.
- [x] Preserve an underlying cause without leaking sensitive implementation details.
- [x] Keep `apperror` framework/database/Redis agnostic.
- [x] Document that every service and module follows the same centralized-error principle, while `beebox-project` owns its concrete service-local error package for now.


**Verify**
```bash
go test ./apperror -count=1
```

### 4. Lock down dependency direction
- [x] Make domain independent of HTTP, PostgreSQL, Redis, and framework packages.
- [x] Reserve application use cases for domain contracts and keep them independent of concrete adapters.
- [x] Keep concrete adapters under `internal/infrastructure/`.
- [x] Keep transport conversion under `internal/interfaces/`.

**Verify**
```bash
go vet ./...
go list -deps ./... >/dev/null
```

### 5. Define the Project domain
- [x] Introduce the `project` domain package.
- [x] Define Project identity and ownership (`organization_id` ).
- [x] Define project lifecycle states and valid transitions.
- [x] Define domain invariants without introducing auth, payment, or runtime logic.
- [x] Add domain unit tests for valid and invalid transitions.
**Verify**
```bash
go test ./internal/domain/project -count=1
```

### 6. Define project-owned domain boundaries
Group these short setup tasks together:
- [x] Define boundaries for `module`, `capability`, `configuration`, `credential`, and `infrastructure` desired state.
- [x] Keep each boundary explicit about what it owns.
- [x] Model relationships as `Project -> Module -> Capability`.
- [x] Treat infrastructure as desired state, not actual runtime state.
- [x] Prevent `beebox-project` from owning domain data of other services/modules.
**Verify**
- Domain package tests pass.
- No cross-service business tables or domain imports are introduced.

### 7. Define module/capability catalog contracts
- [x] Define module identity and version.
- [x] Define capability identity and version.
- [x] Define how project configuration references enabled modules/capabilities.
- [x] Reject unknown identifiers/versions deterministically.
- [x] Keep catalog metadata separate from module runtime implementation.
**Verify**
- Add contract/unit tests for valid, missing, and unknown module/capability references.

### 8. Define controlled data-field configuration
- [x] Define project configuration for enabled data fields.
- [x] Treat fields as references to a BeeBox-controlled catalog.
- [x] Reject arbitrary/custom field identifiers.
- [x] Make field configuration version-aware.
- [x] Define additive vs incompatible/destructive change rules.
- [x] Add tests for supported field, unsupported field, duplicate field, incompatible change, and version mismatch cases.
**Verify**
```bash
go test ./... -count=1
```

### 9. Define configuration lifecycle and desired state
- [x] Model configuration versions explicitly.
- [x] Support a lifecycle such as `DRAFT -> VALIDATED -> PUBLISHED -> APPLIED`.
- [x] Ensure invalid configuration cannot become active.
- [x] Keep current/desired configuration separate from provisioning/actual state.
- [x] Record enough state for later rollback/audit without implementing a full deployment engine yet.

**Verify**
- Unit tests cover valid transitions and blocked transitions.

### 10. Define Project credentials
- [x] Define credential identity and lifecycle.
- [x] Separate project identity from end-user identity.
- [x] Support creation/revocation/expiry state where required.
- [x] Never persist or log raw secret material unintentionally.
- [x] Keep browser-safe/public credentials distinct from privileged server credentials.

**Verify**
- Unit tests cover creation, revocation, expiry, and invalid transitions.
- Error/log paths do not expose secret material.

### 11. Define HTTP/API boundary
- [x] Add versioned HTTP routing under `internal/interfaces/http/`.
- [x] Define request/response adapters for the Project service.
- [x] Map `apperror` categories to stable transport errors.
- [x] Keep handlers free of domain/persistence logic.
- [x] Explicitly document that `beebox-project` is not an API gateway.
- [x] Scope expanded (by decision) to include full Project CRUD (`Create`/`Get`/`Transition`/`Archive`) via a new `internal/application/project` use-case layer and a temporary `internal/infrastructure/memory` repository, both implementing the `Repository` port that Step 12 will re-implement against Postgres.

**Verify**
- Add handler tests for success, validation, not found, forbidden, conflict, and internal error paths.

### 12. Add persistence/cache adapter boundaries
Group these infrastructure setup tasks together:

- [x] Define PostgreSQL repository interfaces and adapters for Project-owned data only.
- [x] Keep PostgreSQL/Supabase as the durable source of truth.
- [ ] Define Redis/Upstash adapter boundaries for cache, rate limiting, idempotency, and other explicitly ephemeral state — **deferred**: no concrete requirement exists yet in this service (rule 12/28: don't introduce Redis by default). Will be picked up when a real need appears (e.g. rate-limiting a public endpoint, idempotency key for a mutating operation).
- [ ] Do not let Redis become the source of truth for project configuration. (applies once 12's Redis item above is actually implemented)
 - [x] Keep external clients behind infrastructure ports/interfaces.

**Verify**
- Repository tests use fakes/mocks for unit scope.
- External integration tests are isolated from unit tests.

### 13. Add migrations and service configuration
- [x] Add `migrations/` for `beebox-project` owned tables only. (added in Step 12: 0001_create_projects.sql)
- [x] Define environment/config loading and validation.
- [x] Document required local configuration without committing secrets. (.env.example)
- [x] Fail fast on invalid required configuration.

**Verify**
- Migration files are deterministic.
- Config validation has unit tests for missing/invalid/valid values.

### 14. Integrate with repository CI
- [x] Ensure formatting checks pass. (existing root `.github/workflows/ci.yaml` auto-discovers this module's go.mod)
- [x] Ensure `go mod verify` passes.
- [x] Ensure `go vet ./...` passes.
- [x] Ensure `go test -race -count=1 ./...` passes.
- [x] Keep external Supabase/Upstash dependencies out of pure unit tests. (integration test is build-tag gated, memory repo used for unit tests)
- [x] Add integration-test execution separately when those tests are introduced. (`.github/workflows/integration.yaml`, manual `workflow_dispatch`, skips gracefully without a configured secret)
- [x] Fixed: `ci.yaml` push-branch pattern did not match this branch's naming convention (`beebox/**`); added so CI actually runs on push, not only on PR.

**Verify**
```bash
gofmt -l .
go mod verify
go vet ./...
go test -race -count=1 ./...
```

### 15. Finish documentation and architecture review
- [ ] Add/update `services/beebox-project/README.md`.
- [ ] Document service responsibility and non-responsibilities.
- [ ] Document package boundaries and dependency direction.
- [ ] Document the centralized `apperror` contract.
- [ ] Document Project/Module/Capability/Configuration/Credential/Infrastructure ownership.
- [ ] Confirm no gateway logic or other service business logic leaked into the service.
- [ ] Run the complete verification one final time.
- [ ] Update this plan so all completed steps are marked `[x]` before committing.

**Final verification**
```bash
gofmt -l .
go mod verify
go vet ./...
go test -race -count=1 ./...
```

---

## Progress Tracker

| Step | Status | Evidence |
|---|---|---|
| 1 | ✅ | go.mod initialized (github.com/DoMinhHHung/beebox-dev/services/beebox-project, go 1.26.5); `go test ./...` → no test files, no build errors |
| 2 | ✅ | build/vet/test green; gofmt clean; no interfaces/abstractions added |
| 3 | ✅ | `apperror` implementation and tests added; `go test ./apperror -count=1` passes; service error contract documented |
| 4 | ✅ | Dependency boundaries verified; domain/application contain no infrastructure or transport imports; `go vet ./...` and `go list -deps ./...` pass |
| 5 | ✅ | Project identity, organization ownership, lifecycle states, valid/invalid transitions, and domain tests added; domain and full verification pass |
| 6 | ✅ | Project-owned module, capability, configuration, credential, and infrastructure desired-state boundaries added; Project-to-Module-to-Capability references and domain tests pass |
| 7 | ✅ | Module and capability catalog metadata added; exact identifier/version lookup, configuration references, unknown references, mismatched versions, and duplicate definitions are covered by unit tests |
| 8 | ✅ | Controlled data-field references added with exact catalog/version validation; supported, custom, duplicate, incompatible, and version-mismatch cases covered by tests |
| 9 | ✅ | Version of configuration wall, lifecycle DRAFT→VALIDATED→PUBLISHED→APPLIED|
| 10 | ✅ | Define project credential lifecycle. Keep browser-safe/public credentials distinct from privileged server credentials|
| 11 | ✅ | HTTP boundary + full Project CRUD (create/get/transition/archive) via in-memory repository; apperror→HTTP status mapping; build/vet/test green; gofmt clean |
| 12 | ✅ (Postgres only) | PostgresProjectRepository implements project.Repository; migration 0001 added; integration test tagged `integration`, skipped without BEEBOX_PROJECT_TEST_DATABASE_URL; Redis/Upstash boundary intentionally deferred per rule 12/28 until a concrete need exists |
| 13 | ✅ | internal/infrastructure/config added with Load/validatePort, 6 unit tests (valid/missing/invalid); .env.example documents required vars without secrets; cmd/server fails fast on invalid config |
| 14 | ✅ | Root CI (gofmt/mod verify/vet/test -race) auto-covers beebox-project via module discovery; added `beebox/**` to push trigger; added manual integration workflow for postgres tests |
| 15 | ⬜ | |

## Completion Rule

A step is complete only when:

1. Its checklist items are implemented.
2. Its stated verification passes.
3. Any required tests are present and passing.
4. The step is marked `[x]` in this plan.
5. The corresponding evidence is recorded in the Progress Tracker when useful.
