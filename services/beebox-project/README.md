# beebox-project

`beebox-project` is BeeBox's project control-plane service. It owns project
lifecycle, module/capability enablement, controlled configuration, and
project credentials. It does not implement runtime business logic for any
module, and it is not an API gateway.

## Responsibility

- Organization/project ownership and lifecycle (`DRAFT` → `ACTIVE` →
  `SUSPENDED`/`ARCHIVED`).
- Enabled modules and capabilities per project, backed by a BeeBox-controlled
  catalog (unknown/duplicate/incompatible entries are rejected at the domain
  layer).
- Controlled configuration lifecycle: `DRAFT` → `VALIDATED` → `PUBLISHED` →
  `APPLIED`, with desired state (`ConfigurationVersion`) modeled separately
  from applied/actual state (`RolloutState`) to support retries, partial
  failure, and rollback.
- Project credentials (public and secret), including issuance, secret
  hashing, revocation, and expiry — independent from end-user identity.
- Infrastructure desired-state records for a project.

## Non-responsibilities

- No API gateway behavior.
- No authentication implementation (user signup/signin, JWT issuance).
- No e-commerce/payment business logic.
- No arbitrary developer-defined backend schema generation.
- No ownership of another service's private data.
- No premature external messaging/provisioning.

## Architecture
```
transport (internal/interfaces/http)
                ↓
application (internal/application)
                ↓
domain (internal/domain)
                ↑
infrastructure (internal/infrastructure)
```

- **domain** — pure Go, no framework or infrastructure imports. Contains
  `project`, `module`, `capability`, `catalog`, `configuration` (with
  `version` and `rollout`), `credential`, and `infrastructure` (desired
  state). Domain errors are plain sentinel `error` values, not `apperror`.
- **application** — orchestrates use cases against domain types and ports
  it owns (e.g. `internal/application/project.Repository`). Translates
  domain/repository errors into `apperror` codes. No transport concerns
  (no HTTP status codes).
- **interfaces/http** — parses requests, calls application services, maps
  `apperror.Code` to HTTP status and the public JSON error shape. Contains
  no business rules.
- **infrastructure** — concrete adapters implementing application ports:
  `memory` (in-process, used for local dev and unit tests) and `postgres`
  (Supabase-backed, used in production). Both implement the same
  `project.Repository` interface, so application/transport code never
  changes when the backing store changes.

Redis/Upstash adapter boundaries are intentionally **not yet implemented** —
no concrete requirement (cache, rate limiting, idempotency) exists in this
service yet. It will be added when a real need appears, per the project's
Redis policy.

## Error model (`apperror`)

All application-facing errors carry a stable `Code`:

| Code | HTTP status |
|---|---|
| `VALIDATION` | 400 |
| `UNAUTHENTICATED` | 401 |
| `FORBIDDEN` | 403 |
| `NOT_FOUND` | 404 |
| `CONFLICT` | 409 |
| `DEPENDENCY_FAILURE` | 502 |
| `INTERNAL` | 500 |

Public HTTP error responses are always:

```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "project not found"
  }
}
```

Errors that are not a recognized `*apperror.Error` (e.g. an unwrapped
infrastructure error) are always reported as `INTERNAL` with a generic
message — the underlying cause is never exposed through the public API.

## Domain ownership

- **Project** — top-level tenant unit; owns lifecycle transitions.
- **Module / Capability** — catalog of what a project has enabled;
  `catalog` enforces valid module→capability relationships and versioning.
- **Configuration** — controlled, versioned project configuration;
  `rollout` tracks desired vs. applied state separately.
- **Credential** — project-level public/secret credentials; secrets are
  stored only as hashes, raw secret material is redacted from any string
  representation (`IssuedCredential.String()`), and never logged.
- **Infrastructure** — desired infrastructure state for a project.

## Running locally

```bash
cp .env.example .env
# edit .env with a real DATABASE_URL

psql "$DATABASE_URL" -f migrations/0001_create_projects.sql
psql "$DATABASE_URL" -f migrations/0002_create_project_configuration.sql
psql "$DATABASE_URL" -f migrations/0003_create_project_capabilities.sql

export $(cat .env | xargs)
go run ./cmd/server
```

Required environment variables:

| Variable | Required | Default |
|---|---|---|
| `DATABASE_URL` | yes | — |
| `PORT` | no | `8080` |
| `IDENTITY_URL` | no | `http://localhost:8081` |

The service fails fast with a clear error if `DATABASE_URL` is missing or
`PORT` is not a valid port number.

## Testing

```bash
gofmt -l .
go mod verify
go vet ./...
go test -race -count=1 ./...
```

Integration tests against a real Postgres instance are excluded from the
default test run (build tag `integration`):

```bash
BEEBOX_PROJECT_TEST_DATABASE_URL="postgres://..." \
  go test -tags=integration -race -count=1 ./internal/infrastructure/postgres/...
```

They also run manually via the `Integration` GitHub Actions workflow
(`workflow_dispatch`), gated behind the `BEEBOX_PROJECT_TEST_DATABASE_URL`
secret.

## API

| Method | Path | Description |
|---|---|---|
| `GET` | `/healthz` | Liveness check |
| `POST` | `/v1/projects` | Create a project |
| `GET` | `/v1/projects/{id}` | Get a project |
| `PATCH` | `/v1/projects/{id}` | Transition project status |
| `DELETE` | `/v1/projects/{id}` | Archive a project (soft; not physical deletion) |
| `PUT` | `/v1/projects/{id}/configuration` | Create a new project configuration version |
| `GET` | `/v1/projects/{id}/configuration` | Get the current project configuration version |
| `GET` | `/v1/projects/{id}/configuration/versions/{version}` | Get a specific configuration version |
| `PATCH` | `/v1/projects/{id}/configuration/versions/{version}` | Transition a configuration version lifecycle |
| `POST` | `/v1/projects/{id}/configuration/versions/{version}/rollout` | Set the desired rollout version |

All project and configuration endpoints require the existing identity session
authentication. Configuration access is authorized through the project's
organization ownership boundary.
