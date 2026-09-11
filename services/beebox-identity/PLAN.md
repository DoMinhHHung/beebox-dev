# Plan — beebox-identity

## Goal

Build `services/beebox-identity` as the single BeeBox service responsible for end-user identity and authentication state.

It must provide the identity foundation required by later services/modules without duplicating identity storage or business logic.

## Scope

### In scope

- User identity
- Credentials
- Password authentication
- Sessions
- Authentication state
- Token/session lifecycle required by the API contract
- Email verification
- Password recovery/reset
- Authentication context for downstream services

### Out of scope

- Project lifecycle
- Project configuration
- Module catalog/configuration
- Compute/storage provisioning
- Billing
- API gateway
- Developer-facing auth capability catalog
- Enterprise SSO unless a concrete requirement is introduced

---

## Phase 0 — Preconditions

- [ ] `beebox-project` is merged into `main`
- [ ] Any committed database credential/secret is revoked or rotated
- [ ] `.env.example` contains placeholders only
- [ ] Repository/port ownership decision from `beebox-project` is resolved consistently with the current Clean Architecture rule

## Phase 1 — Bootstrap

- [x] Create `services/beebox-identity`
- [x] Initialize Go module with Go 1.26.5
- [x] Create Clean Architecture package structure
- [x] Add service entrypoint and composition root
- [x] Add local `apperror`
- [x] Add configuration loading and startup validation
- [x] Add initial README

Acceptance:

- [x] Service builds
- [x] No business logic in transport
- [x] No infrastructure dependency in domain
- [x] CI can discover and test the service

## Phase 2 — Domain foundation

- [x] Define `User`
- [x] Define identity identifier/value objects needed by current use cases
- [x] Define credential model
- [x] Define password credential state
- [x] Define session model
- [x] Define domain errors
- [x] Add domain invariant tests

Acceptance:

- Domain package has no infrastructure imports
- Domain tests run without PostgreSQL
- No premature organization/project authorization model

## Phase 3 — Application boundaries

- [x] Define user repository port
- [x] Define credential repository port
- [x] Define session repository port
- [x] Define password hashing port
- [x] Define required clock/token/security ports only when a use case needs them
- [x] Implement application use cases
- [x] Add application error translation
- [x] Add application tests

Initial use cases:

- [x] Sign up
- [x] Sign in
- [x] Sign out
- [x] Revoke session

Verification foundation:

- [x] Define verification domain model and types (email, phone)
- [x] Define verification repository port
- [x] Implement RequestVerification use case
- [x] Implement Verify use case
- [x] Secure one-time code generation and hash storage
- [x] Application and domain tests for verification lifecycle

Acceptance:

- Use cases are independently testable
- Domain rules stay in domain
- Infrastructure is not imported by application/domain

## Phase 4 — PostgreSQL infrastructure

- [ ] Add PostgreSQL connection configuration
- [ ] Create identity-owned migrations
- [ ] Implement user repository
- [ ] Implement credential repository
- [ ] Implement session repository
- [ ] Add transaction handling where a use case requires atomic writes
- [ ] Add integration tests where repository behavior cannot be validated with unit tests

Acceptance:

- `beebox-identity` owns its own schema
- No other service accesses identity tables
- Database errors do not leak through the API

## Phase 5 — Password security

- [ ] Select a maintained password hashing implementation
- [ ] Implement password hash/verify adapter
- [ ] Enforce password policy at the appropriate application/domain boundary
- [ ] Ensure plaintext passwords never enter logs
- [ ] Test valid password verification
- [ ] Test invalid password verification
- [ ] Test malformed/invalid stored hashes safely

Acceptance:

- Passwords are never stored plaintext
- Password hashing is not custom cryptography
- Authentication failures do not expose sensitive details

## Phase 6 — HTTP API

- [x] Define public request/response DTOs
- [x] Implement signup endpoint
- [x] Implement signin endpoint
- [x] Implement signout endpoint
- [x] Implement session/revocation endpoint only when required
- [x] Add request validation
- [x] Map `apperror.Code` to HTTP status
- [x] Standardize public error JSON
- [x] Implement verification request/verify endpoints
- [x] Implement password-reset request/reset endpoints
- [x] Enumeration protection for password-reset request

Initial public shape:

```text
GET  /healthz
POST /auth/signup
POST /auth/signin
POST /auth/signout
POST /auth/verification/request
POST /auth/verification/verify
POST /auth/password-reset/request
POST /auth/password-reset/reset
```

Do not add endpoints merely for future compatibility.

Acceptance:

- Transport contains HTTP concerns only
- Public responses never expose internal errors
- Authentication failure behavior is intentionally defined

## Phase 7 — Session and token lifecycle

- [ ] Define access/session model based on the actual client contract
- [ ] Define expiration
- [ ] Define revocation
- [ ] Define refresh behavior if required
- [ ] Implement secure session/token storage
- [ ] Add tests for expiration
- [ ] Add tests for revocation
- [ ] Add tests for refresh/reuse behavior if refresh tokens are introduced

Decision rule:

Prefer the simplest secure session model that satisfies the frontend contract.

Do not introduce JWT unless there is a concrete requirement.

Do not introduce Redis for sessions until PostgreSQL is proven insufficient.

## Phase 8 — Email and phone verification

- [ ] Define email verification product flow
- [ ] Define phone verification product flow
- [ ] Define verification request endpoint
- [ ] Define verification confirmation endpoint
- [ ] Define how successful verification updates identity state
- [ ] Add email delivery port
- [ ] Add phone/SMS delivery port
- [ ] Implement email verification flow
- [ ] Implement phone verification flow
- [ ] Add tests for verification success/failure
- [ ] Add tests for expiry and reuse
- [ ] Add account-enumeration protection
- [ ] Add delivery failure handling

Delivery rules:

- Do not couple identity domain to SMTP/SMS providers.
- Do not introduce RabbitMQ merely because delivery is asynchronous.
- Introduce a queue only when reliable asynchronous delivery becomes a concrete requirement.

## Phase 9 — Password recovery

- [x] Define recovery state
- [x] Define recovery token lifecycle
- [x] Define expiration
- [x] Define single-use behavior
- [ ] Add request-recovery endpoint
- [ ] Add reset-password endpoint
- [ ] Prevent account enumeration through public responses
- [x] Add tests for expiry and reuse

Password recovery foundation:

- [x] Define recovery domain model
- [x] Define recovery repository port
- [x] Implement RequestPasswordReset use case
- [x] Add recovery lifecycle tests

## Phase 10 — Authentication context

- [ ] Define the authenticated principal contract
- [ ] Define how downstream BeeBox services validate authentication
- [ ] Define stable identity/user identifiers
- [ ] Document trust boundaries
- [ ] Document token/session verification contract

Do not build `beebox-gateway` as part of this phase.

The goal is to establish a clean identity contract that a later gateway or service can consume.

## Phase 11 — Integration contract

- [ ] Document API endpoints
- [ ] Document authentication headers/tokens
- [ ] Document error codes
- [ ] Document session lifecycle
- [ ] Document downstream authentication contract
- [ ] Document what `beebox-project` may consume from identity
- [ ] Document what `modules/beebox-auth` exposes versus what identity implements

Acceptance:

- Another BeeBox service can integrate without reading identity source code.

## Phase 12 — CI and final review

- [ ] `gofmt -l .` returns no files
- [ ] `go mod verify` passes
- [ ] `go vet ./...` passes
- [ ] `go test -race -count=1 ./...` passes
- [ ] No comments in source code
- [ ] No TODO/FIXME in source code
- [ ] No real credentials/secrets committed
- [ ] Dependency direction reviewed
- [ ] README reflects the real architecture
- [ ] API contract matches implementation

---

## PR slicing

### PR 1 — bootstrap

- Phase 1
- CI/discovery compatibility
- README

### PR 2 — domain

- Phase 2
- Domain tests

### PR 3 — application auth foundation

- Phase 3
- Signup/signin application use cases

### PR 4 — PostgreSQL

- Phase 4
- Repository integration tests

### PR 5 — HTTP auth

- Phase 5
- Phase 6

### PR 6 — sessions

- Phase 7

### PR 7 — Verification and recovery

- Phase 8
- Phase 9

### PR 8 — downstream contract

- Phase 10
- Phase 11
- Final CI/review

---

## Progress tracker

- [ ] Phase 0 — Preconditions
- [x] Phase 1 — Bootstrap
- [x] Phase 2 — Domain foundation
- [x] Phase 3 — Application boundaries
- [ ] Phase 4 — PostgreSQL infrastructure
- [ ] Phase 5 — Password security
- [x] Phase 6 — HTTP API
- [ ] Phase 7 — Session/token lifecycle
- [ ] Phase 8 — Email and phone verification
- [ ] Phase 9 — Password recovery
- [ ] Phase 10 — Authentication context
- [ ] Phase 11 — Integration contract
- [ ] Phase 12 — CI and final review
