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
- [ ] Add `cmd/server/`.
- [ ] Add `internal/domain/`.
- [ ] Add `internal/application/`.
- [ ] Add `internal/interfaces/`.
- [ ] Add `internal/infrastructure/`.
- [ ] Avoid placeholder abstractions that have no current responsibility.

**Verify**
```bash
go test ./...
go vet ./...
```

### 3. Add centralized application errors
- [ ] Add `apperror/apperror.go`.
- [ ] Add `apperror/apperror_test.go`.
- [ ] Define stable machine-readable error identity.
- [ ] Support the initial categories required by the service: validation, unauthenticated, forbidden, not found, conflict, dependency failure, and internal failure.
- [ ] Preserve an underlying cause without leaking sensitive implementation details.
- [ ] Keep `apperror` framework/database/Redis agnostic.
- [ ] Document that every service and module follows the same centralized-error principle, while `beebox-project` owns its concrete service-local error package for now.

**Verify**
```bash
go test ./apperror -count=1
```

### 4. Lock down dependency direction
- [ ] Make domain independent of HTTP, PostgreSQL, Redis, and framework packages.
- [ ] Make application use cases depend on domain contracts rather than concrete adapters.
- [ ] Keep concrete adapters under `internal/infrastructure/`.
- [ ] Keep transport conversion under `internal/interfaces/`.

**Verify**
```bash
go vet ./...
go list -deps ./... >/dev/null
```

### 5. Define the Project domain
- [ ] Introduce the `project` domain package.
- [ ] Define Project identity and ownership (`organization_id`).
- [ ] Define project lifecycle states and valid transitions.
- [ ] Define domain invariants without introducing auth, payment, or runtime logic.
- [ ] Add domain unit tests for valid and invalid transitions.

**Verify**
```bash
go test ./internal/domain/project -count=1
```

### 6. Define project-owned domain boundaries
Group these short setup tasks together:

- [ ] Define boundaries for `module`, `capability`, `configuration`, `credential`, and `infrastructure` desired state.
- [ ] Keep each boundary explicit about what it owns.
- [ ] Model relationships as `Project -> Module -> Capability`.
- [ ] Treat infrastructure as desired state, not actual runtime state.
- [ ] Prevent `beebox-project` from owning domain data of other services.

**Verify**
- Domain package tests pass.
- No cross-service business tables or domain imports are introduced.

### 7. Define module/capability catalog contracts
- [ ] Define module identity and version.
- [ ] Define capability identity and version.
- [ ] Define how project configuration references enabled modules/capabilities.
- [ ] Reject unknown identifiers/versions deterministically.
- [ ] Keep catalog metadata separate from module runtime implementation.

**Verify**
- Add contract/unit tests for valid, missing, and unknown module/capability references.

### 8. Define controlled data-field configuration
- [ ] Define project configuration for enabled data fields.
- [ ] Treat fields as references to a BeeBox-controlled catalog.
- [ ] Reject arbitrary/custom field identifiers.
- [ ] Make field configuration version-aware.
- [ ] Define additive vs incompatible/destructive change rules.
- [ ] Add tests for supported field, unsupported field, duplicate field, incompatible change, and version mismatch cases.

**Verify**
```bash
go test ./... -count=1
```

### 9. Define configuration lifecycle and desired state
- [ ] Model configuration versions explicitly.
- [ ] Support a lifecycle such as `DRAFT -> VALIDATED -> PUBLISHED -> APPLIED`.
- [ ] Ensure invalid configuration cannot become active.
- [ ] Keep current/desired configuration separate from provisioning/actual state.
- [ ] Record enough state for later rollback/audit without implementing a full deployment engine yet.

**Verify**
- Unit tests cover valid transitions and blocked transitions.

### 10. Define Project credentials
- [ ] Define credential identity and lifecycle.
- [ ] Separate project identity from end-user identity.
- [ ] Support creation/revocation/expiry state where required.
- [ ] Never persist or log raw secret material unintentionally.
- [ ] Keep browser-safe/public credentials distinct from privileged server credentials.

**Verify**
- Unit tests cover creation, revocation, expiry, and invalid transitions.
- Error/log paths do not expose secret material.

### 11. Define HTTP/API boundary
- [ ] Add versioned HTTP routing under `internal/interfaces/http/`.
- [ ] Define request/response adapters for the Project service.
- [ ] Map `apperror` categories to stable transport errors.
- [ ] Keep handlers free of domain/persistence logic.
- [ ] Explicitly document that `beebox-project` is not an API gateway.

**Verify**
- Add handler tests for success, validation, not found, forbidden, conflict, and internal error paths.

### 12. Add persistence/cache adapter boundaries
Group these infrastructure setup tasks together:

- [ ] Define PostgreSQL repository interfaces and adapters for Project-owned data only.
- [ ] Keep PostgreSQL/Supabase as the durable source of truth.
- [ ] Define Redis/Upstash adapter boundaries for cache, rate limiting, idempotency, and other explicitly ephemeral state.
- [ ] Do not let Redis become the source of truth for project configuration.
- [ ] Keep external clients behind infrastructure ports/interfaces.

**Verify**
- Repository tests use fakes/mocks for unit scope.
- External integration tests are isolated from unit tests.

### 13. Add migrations and service configuration
- [ ] Add `migrations/` for `beebox-project` owned tables only.
- [ ] Define environment/config loading and validation.
- [ ] Document required local configuration without committing secrets.
- [ ] Fail fast on invalid required configuration.

**Verify**
- Migration files are deterministic.
- Config validation has unit tests for missing/invalid/valid values.

### 14. Integrate with repository CI
- [ ] Ensure formatting checks pass.
- [ ] Ensure `go mod verify` passes.
- [ ] Ensure `go vet ./...` passes.
- [ ] Ensure `go test -race -count=1 ./...` passes.
- [ ] Keep external Supabase/Upstash dependencies out of pure unit tests.
- [ ] Add integration-test execution separately when those tests are introduced.

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
| 2 | ⬜ | |
| 3 | ⬜ | |
| 4 | ⬜ | |
| 5 | ⬜ | |
| 6 | ⬜ | |
| 7 | ⬜ | |
| 8 | ⬜ | |
| 9 | ⬜ | |
| 10 | ⬜ | |
| 11 | ⬜ | |
| 12 | ⬜ | |
| 13 | ⬜ | |
| 14 | ⬜ | |
| 15 | ⬜ | |

## Completion Rule

A step is complete only when:

1. Its checklist items are implemented.
2. Its stated verification passes.
3. Any required tests are present and passing.
4. The step is marked `[x]` in this plan.
5. The corresponding evidence is recorded in the Progress Tracker when useful.
