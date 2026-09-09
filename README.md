# BeeBox

> **Configurable Backend & Cloud Platform for Developers**

BeeBox is a platform that provides developers with a configurable backend, infrastructure, authentication, authorization, storage, and application capabilities without requiring them to design and operate a traditional backend from scratch.

The core idea is simple:

**Developers build the frontend. BeeBox provides the backend capabilities and infrastructure.**

Instead of starting every application by implementing authentication, CRUD APIs, database schemas, authorization, storage, and deployment infrastructure, developers create a BeeBox project, select the capabilities they need, configure supported options, and receive a ready-to-use backend API.

---

## Core Concept

BeeBox turns backend development into a **controlled composition model**.

A developer creates a project and configures:

1. **Infrastructure**

   * Compute
   * Storage

2. **Modules**

   * Authentication
   * Products
   * Cart
   * Payments
   * Users
   * Other supported application capabilities

3. **Capabilities**

   * Individual operations provided by each module

4. **Data Fields**

   * Supported fields exposed by each capability

5. **Security Policies**

   * Authentication requirements
   * RBAC
   * Permissions
   * Tenant isolation
   * Rate limits and other platform-level controls

BeeBox then provisions and exposes the resulting backend through a versioned API.

The developer does not receive an arbitrary generated backend codebase. They receive a **managed backend capability surface** operated by BeeBox.

---

# Product Model

A BeeBox deployment is organized around the following hierarchy:

```text
Organization
└── Project
    ├── Infrastructure
    │   ├── Compute
    │   └── Storage
    │
    ├── Modules
    │   ├── Capabilities
    │   └── Data Field Configuration
    │
    ├── Authentication
    ├── Authorization / RBAC
    └── API
```

## Project

A **Project** represents one application backend.

A project owns its:

* API namespace
* configuration
* enabled modules
* capability configuration
* data schema configuration
* authentication configuration
* authorization policies
* compute allocation
* storage allocation
* runtime resources
* project credentials

Projects are isolated logical tenants within the BeeBox platform.

---

# Infrastructure

Infrastructure is a first-class part of the BeeBox product.

Each project receives a controlled allocation of:

* Compute
* Storage

Compute and storage are intentionally modeled together as the project's infrastructure capacity.

## Compute

Compute represents the runtime resources available to the project's backend workloads.

The initial product model supports predefined resource plans such as:

| Plan       |     CPU |    RAM |       Price |
| ---------- | ------: | -----: | ----------: |
| Free       | 0.1 CPU | 512 MB |  $0 / month |
| 0.5c-512mb | 0.5 CPU | 512 MB |  $7 / month |
| 1c-2g      |   1 CPU |   2 GB | $25 / month |
| 2c-4g      |   2 CPU |   4 GB | $85 / month |

Pricing and resource quotas are product configuration, not hard-coded application behavior.

## Storage

Storage is provisioned as part of the project's infrastructure plan.

Storage may be used for:

* Application data
* Database-backed resources
* Uploaded files
* Images
* User assets
* Documents
* Other supported persistent data

Storage quotas are controlled by BeeBox and associated with the selected infrastructure plan.

The platform must prevent one project from consuming storage allocated to another project.

---

# Managed Infrastructure

BeeBox separates application configuration from infrastructure implementation.

A developer specifies:

```text
Project
    ↓
Infrastructure Plan
    ↓
Modules
    ↓
Capabilities
    ↓
Configuration
```

The platform determines how those requirements are deployed and operated.

The developer should not need to directly manage:

* backend server provisioning
* runtime process management
* database server administration
* cache infrastructure
* storage allocation
* service discovery
* internal networking
* internal authentication between BeeBox services

Infrastructure management remains an internal platform responsibility.

---

# Technology Foundation

## Backend

* **Language:** Go
* **Go version:** `1.26.5`

Go services should follow clean dependency direction and explicit boundaries.

## Architecture

BeeBox uses:

* **Microservice Architecture**
* **Clean Architecture**

Each service represents a clearly defined bounded context and owns its responsibilities independently.

Clean Architecture defines dependency direction:

```text
Delivery / Transport
        ↓
Application
        ↓
Domain
        ↓
Infrastructure
```

Dependencies should point inward toward business rules.

Infrastructure implementations must not define or leak domain behavior.

Microservices should be separated by business and ownership boundaries rather than by arbitrary technical layers.

## Database

* **PostgreSQL**
* Hosted through **Supabase**

PostgreSQL is the primary persistent data store for BeeBox platform and service data where relational persistence is required.

Database ownership must remain explicit at the service boundary.

Services should not depend on another service's private tables as an implicit API.

## Cache

* **Redis**
* Hosted through **Upstash**

Redis is used for:

* caching
* short-lived state
* rate limiting
* distributed coordination where necessary
* session-related ephemeral data where appropriate

Redis must not become the authoritative source of durable business data unless explicitly defined by the domain.

---

# Microservice Model

BeeBox consists of independently deployable services aligned with bounded contexts.

The exact service decomposition may evolve, but each service must have:

* a clear responsibility
* explicit ownership
* explicit API contracts
* controlled dependencies
* independent failure boundaries
* observable runtime behavior

A service should not exist merely because a package or database table exists.

The architecture should avoid:

* function-per-service decomposition
* arbitrary service fragmentation
* shared business logic duplicated across services
* hidden database coupling
* synchronous dependency chains without failure handling

A service boundary should exist because the boundary represents meaningful ownership, isolation, scalability, security, or operational requirements.

---

# Core Platform Services

The initial system can be organized around several major bounded contexts.

## Identity and Authentication

Responsible for:

* sign up
* sign in
* sign out
* password reset
* email verification
* session management
* token issuance
* JWT-related functionality
* authentication policies
* identity lifecycle

Authentication is a platform capability rather than application-specific frontend logic.

## Authorization

Responsible for:

* roles
* permissions
* RBAC policies
* resource authorization
* capability-level access control
* tenant-aware authorization

Authentication answers:

> Who is this?

Authorization answers:

> What is this identity allowed to do?

These concerns must remain explicitly separated.

## Project Management

Responsible for:

* project creation
* project lifecycle
* project metadata
* project credentials
* enabled services/modules
* configuration state

## Module Configuration

Responsible for:

* module activation
* capability activation
* data-field configuration
* configuration validation
* configuration versioning
* configuration publication

## Infrastructure Management

Responsible for:

* compute plans
* storage quotas
* provisioning state
* resource allocation
* lifecycle operations
* infrastructure configuration

## API Gateway / Edge

Responsible for:

* API routing
* authentication enforcement
* project identification
* request validation
* rate limiting
* version routing
* request correlation
* edge-level security controls

The gateway must not become a business-logic monolith.

---

# Modules

A **Module** is a reusable backend capability group.

Modules represent product-level functionality, not technical infrastructure.

Examples include:

```text
Auth
Products
Cart
Payment
Users
Orders
Files
Notifications
```

A module contains multiple **capabilities**.

For example:

```text
Auth
├── Sign Up
├── Sign In
├── Forgot Password
├── Reset Password
├── Email Verification
├── Session Management
└── Token Management
```

A capability represents an independently configurable operation or workflow.

---

# Capabilities

Capabilities are the fundamental building blocks of a BeeBox backend.

Each capability defines a controlled contract containing:

* input schema
* output schema
* supported data fields
* validation rules
* authentication requirements
* authorization requirements
* business rules
* error contract
* API endpoint
* version
* observability requirements

A capability is not arbitrary user-generated backend code.

It is a platform-defined behavior that can be safely composed into projects.

---

# Configuration Model

BeeBox configuration is intentionally constrained.

The platform does not allow every developer to freely redesign the backend schema or behavior.

Instead, developers configure a controlled set of options exposed by BeeBox.

The configuration process is:

```text
Select Module
      ↓
Select Capability
      ↓
Select Supported Fields
      ↓
Configure Supported Policies
      ↓
Validate Configuration
      ↓
Publish Configuration
```

The configuration engine is responsible for ensuring that combinations of options are valid.

---

# Controlled Data Fields

Each capability exposes a predefined catalog of supported data fields.

For example, an authentication capability may define a controlled field set including:

```text
id
email
phoneNumber
password
firstName
lastName
fullName
avatar
...
```

Only fields registered by BeeBox may be enabled.

A project configuration therefore defines:

```text
Capability
    +
Enabled Fields
    +
Policies
```

rather than an unrestricted user-defined schema.

## Why fields are controlled

Controlled schemas provide:

* predictable API contracts
* consistent validation
* safer migrations
* stronger security guarantees
* better compatibility
* easier versioning
* centralized platform governance
* better operational control

BeeBox intentionally favors **controlled extensibility** over unrestricted customization.

---

# Configuration as a Product Boundary

The configuration system is a major product boundary.

A project configuration should be treated as declarative state describing the desired backend.

Conceptually:

```text
Project Configuration
        ↓
Validation
        ↓
Resolved Configuration
        ↓
Provisioning / Runtime
        ↓
Public API
```

The runtime should only expose capabilities that are enabled and valid for the project.

Configuration changes must therefore be:

* validated
* versioned
* auditable
* safely applied
* compatible with existing data
* protected against destructive changes

Breaking changes should never be silently introduced by a configuration switch.

---

# API Exposure

After a project has been configured, BeeBox exposes a public API for the enabled capabilities.

The public API is the primary interface between the developer's application and BeeBox.

Conceptually:

```text
Developer Frontend
        ↓
BeeBox Public API
        ↓
Authentication / Authorization
        ↓
Enabled Capability
        ↓
Domain Logic
        ↓
Persistence / Infrastructure
```

APIs must be:

* versioned
* documented
* language-neutral
* stable
* explicit about authentication
* explicit about authorization
* explicit about errors
* tenant-aware

The frontend should not need to know how the internal BeeBox microservices are deployed.

Internal service topology is an implementation detail.

---

# API Versioning

Public APIs are contracts.

Changes to public behavior must be managed through explicit versioning and compatibility rules.

BeeBox must avoid silently changing:

* request fields
* response fields
* authentication behavior
* authorization semantics
* error codes
* resource semantics
* data constraints

Configuration changes must not unexpectedly invalidate an existing client.

---

# Project Identity and Credentials

Every project has a platform identity.

Project credentials allow BeeBox to identify the consuming application/project.

Project identity and end-user identity are different concepts.

```text
Project Credential
    ↓
Identifies the application/project

User Token / Session
    ↓
Identifies an authenticated user
```

Project credentials must not be treated as a replacement for user authentication.

Credentials must also be classified by trust level.

A credential that is safe to expose to a browser must not provide privileged server-side capabilities.

Server-only credentials must never be exposed to client-side applications.

---

# Security and Control Principles

Security is a platform-level responsibility.

BeeBox must enforce security consistently instead of relying on every application developer to implement it correctly.

## Tenant Isolation

Projects are isolated tenants.

A request associated with Project A must never access:

* Project B's data
* Project B's configuration
* Project B's storage
* Project B's credentials
* Project B's internal resources

Tenant identity must be established and validated at the platform boundary.

Tenant IDs must not be trusted solely because they are supplied by the client.

## Authentication

Protected capabilities must require explicit authentication policies.

Authentication must support the defined identity flows of the selected module.

Token validation must include appropriate:

* signature validation
* expiration checks
* issuer validation
* audience validation
* token type validation
* project/tenant context validation

## Authorization

Every protected operation must have an authorization decision.

Authorization should be based on:

```text
Identity
+
Project / Tenant
+
Role
+
Permission
+
Resource Context
```

RBAC is provided as a platform capability, but resource ownership and tenant boundaries must still be enforced.

## Least Privilege

Services, users, credentials, and internal components should receive only the permissions required for their responsibilities.

## Secret Management

Secrets must never be embedded in:

* source code
* public API responses
* frontend bundles
* logs
* version-controlled configuration

## Rate Limiting

Public endpoints must support platform-controlled rate limiting.

Rate limits may be applied using combinations of:

* project
* user
* credential
* IP
* endpoint
* authentication state

Sensitive operations such as authentication and password recovery require stronger protection.

## Auditability

Security-sensitive operations should produce auditable events where appropriate.

Examples include:

* authentication changes
* credential changes
* role changes
* permission changes
* configuration changes
* destructive operations
* sensitive account operations

---

# Data Ownership

Every service must have explicit ownership over the data it is responsible for.

Services must not use another service's internal database tables as an implicit communication mechanism.

Cross-service communication should occur through:

* public service contracts
* internal APIs
* messaging/events where justified

Database sharing must not become the mechanism that destroys service boundaries.

PostgreSQL may be physically hosted as a common platform database environment, but logical ownership must remain explicit.

---

# Failure Boundaries

BeeBox is a distributed system.

A failure in one service must not automatically imply total platform failure.

The platform should explicitly model:

* timeouts
* retries
* idempotency
* partial failure
* dependency failure
* unavailable infrastructure
* stale cache
* provisioning failure
* configuration failure

Retries must be bounded and safe.

Operations that can be repeated must have explicit idempotency semantics.

---

# Storage and Persistence Boundaries

Persistent application data belongs to the appropriate data-owning component.

Storage should be treated as two separate concerns:

```text
Structured Data
    ↓
PostgreSQL

Ephemeral / Cache / Coordination State
    ↓
Redis

Persistent File/Object Data
    ↓
BeeBox Storage Layer
```

Redis must not silently become a second primary database.

Cache invalidation and consistency rules must be explicitly defined for every cached domain.

---

# Infrastructure Provisioning

Provisioning is driven by project configuration.

At a conceptual level:

```text
Project Configuration
        ↓
Infrastructure Validation
        ↓
Resource Allocation
        ↓
Service Configuration
        ↓
Runtime Activation
        ↓
Health Verification
        ↓
API Availability
```

Provisioning must be:

* repeatable
* observable
* idempotent where possible
* safely reversible
* stateful
* auditable

A project must not be considered ready until the platform has verified that its required resources and capabilities are available.

---

# Configuration Lifecycle

BeeBox configuration should have an explicit lifecycle:

```text
Draft
  ↓
Validated
  ↓
Published
  ↓
Provisioning
  ↓
Active
```

A failed configuration must not partially activate an invalid backend state.

Configuration changes must distinguish between:

* additive changes
* compatible changes
* migration-required changes
* destructive changes
* breaking changes

Destructive and breaking operations require stronger controls than ordinary configuration changes.

---

# Scope Boundaries

BeeBox is intentionally not an unrestricted application platform.

The initial product should **not** attempt to become:

* a fully arbitrary backend generator
* a general-purpose code generation platform
* a replacement for every cloud provider
* a serverless function marketplace
* a Kubernetes abstraction layer
* a general-purpose workflow engine
* an unrestricted database designer
* a platform where every customer defines arbitrary backend logic

The platform's value depends on keeping the capability surface controlled.

---

# What Developers Own

Developers are responsible for:

* frontend applications
* user experience
* application presentation
* client-side state
* integration with BeeBox APIs
* product-specific UI behavior
* application-specific workflows that exist outside supported BeeBox capabilities

---

# What BeeBox Owns

BeeBox is responsible for:

* backend capability implementation
* public API contracts
* authentication
* authorization
* RBAC
* supported data schemas
* database persistence
* storage
* compute allocation
* cache infrastructure
* project isolation
* service orchestration
* configuration validation
* provisioning
* platform security controls
* backend observability
* API lifecycle and compatibility

---

# Architectural Principles

## 1. Capability Over Code Generation

The primary abstraction is a capability, not generated source code.

## 2. Configuration Over Arbitrary Customization

Developers configure supported behavior instead of modifying platform internals.

## 3. Explicit Boundaries

Modules, services, domains, APIs, and data ownership must have clear boundaries.

## 4. Security by Platform

Security-critical behavior must be enforced centrally and consistently.

## 5. Tenant Isolation by Default

Cross-project access must be impossible unless explicitly authorized by the platform.

## 6. Public Contracts Are Stable

Public APIs and schemas must be treated as long-lived contracts.

## 7. Infrastructure Is Managed

Developers consume compute and storage; BeeBox operates the underlying platform.

## 8. Failure Is Explicit

Distributed-system failure modes must be designed rather than assumed away.

## 9. Controlled Extensibility

BeeBox should expand through new modules, capabilities, and approved configuration options rather than unrestricted backend customization.

## 10. Domain-Driven Service Boundaries

Microservices should represent meaningful bounded contexts and operational boundaries.

---

# Initial Scope

The first version of BeeBox should focus on the smallest complete platform capable of proving the core product model:

```text
Project Management
        +
Compute / Storage Plans
        +
Module Selection
        +
Capability Selection
        +
Controlled Data Fields
        +
Authentication
        +
Authorization / RBAC
        +
API Exposure
        +
Provisioning
```

Everything outside this core should be treated as future scope unless it is required to make the core product functional and secure.

---

# Long-Term Product Direction

BeeBox is intended to evolve into a platform where application backends can be assembled from standardized capabilities.

The long-term model is:

```text
Infrastructure
      +
Capabilities
      +
Data Schema
      +
Security Policies
      +
API Contracts
      =
Managed Application Backend
```

The ultimate goal is to make backend infrastructure a **programmable product surface** without requiring every developer to independently design, implement, secure, deploy, and operate the underlying backend.

---

# Development Principles

All implementation should preserve the following boundaries:

```text
Product Configuration
        ↓
Domain Model
        ↓
Application Use Cases
        ↓
Service Contracts
        ↓
Infrastructure Implementations
```

The system should favor:

* explicit contracts
* small bounded contexts
* strong domain ownership
* versioned public APIs
* deterministic configuration
* tenant-aware data access
* observable infrastructure
* secure defaults
* backward-compatible evolution

The system should avoid premature complexity that does not directly support the core BeeBox product model.

---

# Summary

BeeBox is a **configurable backend and cloud platform**.

Developers do not build the traditional backend themselves. They define what backend capabilities their project needs, configure supported data and security options, select infrastructure resources, and consume the resulting API.

The platform's core abstractions are:

```text
Project
Infrastructure
Module
Capability
Data Field
Configuration
Authentication
Authorization
API
```

The platform's core responsibilities are:

```text
Provision
Configure
Secure
Expose
Operate
```

The developer's primary responsibility is:

```text
Build the application experience.
```

BeeBox's primary responsibility is:

```text
Provide and operate the backend capability layer.
```
