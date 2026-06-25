# Opero — Project Setup Guide for Claude Code

> **How to use this file.** Drop it at the root of the Opero repo and use it as your `CLAUDE.md` (project context). It encodes every architectural decision already made. Claude Code should treat the **Conventions & Guardrails** section as hard rules, and build features in the order given by **Build Sequence**. When a decision here conflicts with a default instinct, this file wins. If something is genuinely ambiguous, stop and ask rather than guessing.

---

## 1. What Opero is

Opero is a **people-operations (HR) SaaS** for service companies under ~100 employees, starting with **travel/tourism agencies** (tour guides, drivers, ops staff + office staff). The wedge is the **daily field-ops loop**: a manager builds a roster, field staff see their shifts and check in/out from their phones, and the manager sees who is working in real time.

It is sold as multi-tenant SaaS, with a "dedicated" single-tenant deployment for larger corporates being *the same codebase deployed in isolation* — never a separate product.

**v1 scope (build only this):**
1. **Org / people core** — employees, departments, roles, employment types.
2. **Roster scheduling** — managers create and publish shifts (web).
3. **Mobile attendance** — field staff check in/out with geolocation + photo, offline-tolerant (mobile).
4. **Manager live view** — who is working right now.

**Explicitly out of scope for v1** (do not build, but do not architect against): leave management & document/cert tracking (v1.1); payroll, cross-tenant analytics/insights, and an in-product AI assistant (phase 2). The AI assistant will be built later as tool-calling over the REST API — so keep endpoints clean and well-described, but write no AI code now.

**Next control-plane scope:** add a separate **Super Admin** surface for Opero staff to manage the SaaS platform itself: tenants, platform users, subscription state, provisioning/migration health, system health, and audited support access. Super Admin is not a tenant admin role and must not live inside any tenant database.

---

## 2. Tech stack (locked — do not substitute)

| Layer | Choice |
|---|---|
| Backend | **Go** (modular monolith) |
| HTTP router | **chi** (stdlib-compatible) |
| DB access | **sqlc** + **pgx** — type-safe SQL, **no ORM** |
| Database | **PostgreSQL**, one **logical database per tenant** + one shared **control-plane** database |
| Migrations | **goose** (or golang-migrate) |
| API style | **REST**, **spec-first** with **OpenAPI** + **oapi-codegen** |
| Validation | go-playground/validator |
| Logging | stdlib **slog** (structured) |
| Web client | **React + TypeScript** (Vite) |
| Mobile client | **Flutter** (Dart) — field-facing only |
| Local infra | Docker Compose (Postgres) |

**Hard rules:** no ORM (GORM/ent), no GraphQL, no gRPC, no microservices, no message brokers (Kafka/Redis) in v1.

---

## 3. Architecture

### 3.1 Modular monolith
One deployable Go binary. Code is organised into **domain modules** with hard internal boundaries. A module owns its data and exposes a clean Go interface to others. Modules never reach into another module's store directly.

v1 domain modules: `identity` (people/org), `roster`, `attendance`. Plus `controlplane` for tenant-level and platform-level concerns, and a `platform` package for cross-cutting infrastructure.

### 3.2 Two kinds of database
- **Control-plane DB (one, shared):** the tenant registry, authentication/identity, billing, and platform-level metadata. It is the routing source of truth — it knows which tenant maps to which database.
- **Tenant DB (one per tenant, identical schema):** all of a tenant's operational data (employees, shifts, attendance). Physically isolated from other tenants.

"Database per tenant" means **separate logical databases on a shared Postgres server**, not a server per tenant. ~100 tenant DBs live comfortably on one instance.

### 3.3 Request lifecycle (the core flow)
```
Request
  → AuthMiddleware:    validate JWT, extract user_id + tenant_id
  → TenantMiddleware:  resolve tenant's DB connection from the registry (cached),
                       attach a tenant-scoped *pgxpool handle to request context
  → Handler:           thin; decode/validate input, call Service
  → Service:           business logic; uses the tenant DB handle from context
  → Store (sqlc):      executes generated, type-safe queries
  → Response
```
Control-plane operations (tenant signup/login, platform login, tenant administration, subscription administration, provisioning/migration status) use the control-plane pool directly and skip TenantMiddleware.

### 3.3.1 Tenant auth vs platform auth
Use the same low-level auth primitives for password hashing, password verification, JWT signing, and JWT verification. Do **not** use the same business auth service for tenant users and Super Admin users.

Tenant auth:
```
POST /auth/login
  → look up tenants by slug in the control-plane DB
  → look up users by tenant_id + email in the control-plane DB
  → issue a tenant JWT with kind=tenant, tenant_id, user_id, role
  → TenantAuthMiddleware validates kind=tenant
  → TenantMiddleware resolves the tenant DB from tenant_id
```

Platform auth:
```
POST /platform/auth/login
  → look up platform_users by email in the control-plane DB
  → issue a platform JWT with kind=platform, platform_user_id, role
  → PlatformAuthMiddleware validates kind=platform
  → no tenant DB is selected automatically
```

Do not model Super Admin as a nullable-tenant user in `users`. Platform users live in `platform_users`, have no `tenant_id`, and are Opero staff, not customer employees.

### 3.4 Tenant isolation — the most important guarantee
- A service **only** ever uses the tenant DB handle placed in the request context. It must never open a tenant connection ad hoc or accept a `tenant_id` to pick a database itself.
- Control-plane data and tenant data are **never** mixed in one query — they live in different databases.
- Because each tenant is a physically separate database reached only through the scoped handle, cross-tenant leakage is structurally impossible. Preserve that property; do not add code paths that bypass the middleware-provided handle.

### 3.5 Provisioning & migrations
- **Provisioning** (on tenant creation): create the logical DB → run all tenant migrations on it → seed defaults (e.g., an admin user, default roles). Start as a callable function/script; automate on signup later.
- **Migration orchestrator** (`cmd/migrate`): iterate every tenant in the control-plane registry, apply `db/migrations/tenant/` to each, and **report partial failures clearly** (which tenant DBs succeeded/failed). Control-plane migrations (`db/migrations/controlplane/`) run separately.

### 3.6 Super Admin / platform administration
Super Admin is the internal Opero control-plane console. It manages the SaaS platform, not a single tenant's HR operations.

Required boundaries:
- Super Admin users live in the control-plane `platform_users` table, not tenant DBs and not tenant-scoped `users`.
- Super Admin routes use a separate `/platform/...` API namespace and `PlatformAuthMiddleware`.
- Super Admin tokens must use `kind=platform` and must not include `tenant_id`.
- Tenant-user tokens must use `kind=tenant` and must include `tenant_id`.
- A platform request must never resolve a tenant DB implicitly. Tenant access from the platform side must name a tenant explicitly and create an audit event.
- Support-mode access must be audited with actor, target tenant, reason, action, and timestamp. Prefer read-only support-mode first.
- Cross-tenant dashboards may aggregate usage/health metrics, but raw tenant operational data should not be casually exposed across tenants.

Initial platform API namespace:
```
/platform/auth/login
/platform/auth/me
/platform/tenants
/platform/tenants/{id}
/platform/users
/platform/users/{id}
/platform/subscriptions
/platform/system/health
/platform/audit-events
```

Super Admin UI should be a separate web route group such as `/super-admin`, visually and logically separate from the tenant manager UI.

---

## 4. Repository structure

Use a **monorepo** so the OpenAPI spec is a single shared contract all clients build against.

```
opero/
├── CLAUDE.md                  # this file
├── api/
│   └── openapi.yaml           # API contract — SINGLE SOURCE OF TRUTH
├── backend/
│   ├── cmd/
│   │   ├── api/main.go        # server entrypoint
│   │   ├── migrate/main.go    # migration orchestrator (fan-out across tenant DBs)
│   │   └── provision/main.go  # create + migrate + seed a new tenant DB
│   ├── internal/
│   │   ├── platform/          # cross-cutting infra (no business logic)
│   │   │   ├── config/        # env-based config
│   │   │   ├── db/            # control-plane pool + tenant pool resolver/cache
│   │   │   ├── httpserver/    # chi router, server wiring
│   │   │   ├── middleware/    # auth, tenant resolution, request logging, recovery
│   │   │   └── auth/          # password hashing, JWT issue/verify
│   │   ├── controlplane/      # tenants, platform users, billing (control-plane DB)
│   │   ├── identity/          # employees, departments, roles (tenant DB)
│   │   ├── roster/            # shifts / scheduling (tenant DB)
│   │   └── attendance/        # check-in/out (tenant DB)
│   ├── gen/                   # GENERATED — do not edit by hand
│   │   ├── oapi/              # oapi-codegen server types + interfaces
│   │   └── sqlc/              # sqlc query code, per module
│   ├── db/
│   │   ├── migrations/
│   │   │   ├── controlplane/  # migrations for the control-plane DB
│   │   │   └── tenant/        # migrations applied to EVERY tenant DB
│   │   └── queries/           # raw .sql files consumed by sqlc, per module
│   ├── sqlc.yaml
│   ├── oapi-codegen.yaml
│   ├── go.mod
│   └── docker-compose.yml     # local Postgres
├── web/                       # React + TS (Vite); consumes generated TS client
└── mobile/                    # Flutter; consumes generated Dart client
```

Each domain module follows the same internal layering:
```
internal/<module>/
├── handler.go     # thin HTTP layer; implements the oapi-generated interface
├── service.go     # business logic; the only place rules live
├── store.go       # thin wrapper over sqlc-generated queries
└── types.go       # domain types (mapped to/from generated API + sqlc types)
```
**Handlers thin, services own the logic, stores only touch the DB.** Keep these boundaries.

---

## 5. Data model (v1)

Express schemas as goose migrations. All timestamps `timestamptz`, default `now()`. All IDs `uuid` (default `gen_random_uuid()`). Add `created_at` / `updated_at` to every table.

### 5.1 Control-plane DB

**`tenants`**
- `id` uuid PK
- `name` text — display name (e.g. "Saigon Tours Co.")
- `slug` text unique — used in URLs/login
- `db_name` text — the tenant's logical database name
- `status` text — `active` | `suspended` | `provisioning`
- `plan` text — billing plan key (stub for now)

**`users`** (authentication + tenant routing — the login record)
- `id` uuid PK
- `tenant_id` uuid FK → tenants
- `email` text — unique per tenant (unique index on `(tenant_id, email)`)
- `password_hash` text
- `role` text — `admin` | `manager` | `employee`
- `status` text — `active` | `disabled`
- *Note:* a `users` row authenticates a person and routes them to their tenant DB. Their richer HR profile lives in the tenant DB `employees` table, linked by `user_id`.

**`platform_users`** (Opero staff authentication — Super Admin/support/ops)
- `id` uuid PK
- `email` text unique
- `password_hash` text
- `role` text — `super_admin` | `support` | `ops`
- `status` text — `active` | `disabled`
- *Note:* a `platform_users` row has no `tenant_id`. These users authenticate only to `/platform/...` routes.

**`billing` / `subscriptions`** — stub a minimal table now (tenant_id, plan, status); flesh out in phase 2.

**`super_admin_audit_events`** (platform action audit log)
- `id` uuid PK
- `actor_platform_user_id` uuid FK → platform_users
- `action` text — e.g. `tenant.suspended`, `tenant.reactivated`, `user.disabled`, `subscription.updated`, `support_mode.entered`
- `target_type` text
- `target_id` uuid null
- `tenant_id` uuid null FK → tenants
- `metadata` jsonb default `{}`
- `created_at` timestamptz default `now()`

### 5.2 Tenant DB (identical schema in every tenant database)

**`departments`**
- `id` uuid PK
- `name` text
- `parent_id` uuid null FK → departments (org hierarchy)

**`employees`**
- `id` uuid PK
- `user_id` uuid null — links to control-plane `users.id` (null for staff without login, e.g. some seasonal guides)
- `full_name` text
- `email` text null
- `phone` text null
- `employment_type` text — `full_time` | `part_time` | `freelance` | `seasonal`
- `department_id` uuid null FK → departments
- `title` text null
- `status` text — `active` | `inactive`
- `hired_at` date null

**`locations`** (where shifts/tours happen; used for attendance geofencing)
- `id` uuid PK
- `name` text
- `address` text null
- `lat` double precision null
- `lng` double precision null

**`shifts`** (a rostered assignment)
- `id` uuid PK
- `employee_id` uuid FK → employees
- `location_id` uuid null FK → locations
- `starts_at` timestamptz
- `ends_at` timestamptz
- `notes` text null
- `status` text — `draft` | `published`

**`attendance_records`**
- `id` uuid PK
- `employee_id` uuid FK → employees
- `shift_id` uuid null FK → shifts (attendance may be unscheduled)
- `check_in_at` timestamptz null
- `check_in_lat` / `check_in_lng` double precision null
- `check_in_photo_url` text null
- `check_out_at` timestamptz null
- `check_out_lat` / `check_out_lng` double precision null
- `status` text — `checked_in` | `checked_out` | `missed`

> v1.1 will add `leave_requests`, `leave_balances`, `documents` to the tenant DB. Leave room for them; don't build them.

---

## 6. API workflow (spec-first — follow this order every time)

1. **Edit `api/openapi.yaml` first.** The spec is the contract; nothing is built before it's described there. Give every operation a clear `operationId`, summary, and description (these double as future AI-assistant tool definitions — keep them clean).
2. **Generate server code:** `oapi-codegen` → `backend/gen/oapi/` (request/response types + a server interface per tag).
3. **Write SQL** in `backend/db/queries/<module>.sql`, then **`sqlc generate`** → `backend/gen/sqlc/`.
4. **Implement** the generated server interface in the module's `handler.go`, delegating to `service.go` → `store.go`.
5. **Regenerate the clients** for web (TypeScript) and mobile (Dart) from the same spec.

Never hand-edit anything in `gen/`. Never let the API contract and the implementation drift — the spec leads.

REST conventions: resource-based paths (`/employees`, `/shifts`, `/attendance`), plural nouns, standard verbs/status codes, cursor or page/limit pagination, ISO-8601 timestamps, snake_case JSON fields.

---

## 7. Conventions & guardrails (hard rules)

- **Spec-first, always.** Change `openapi.yaml` before implementing. Spec is source of truth.
- **All SQL through sqlc.** No raw query strings scattered in services. No ORM.
- **Tenant DB access only via the request-context handle.** Never open a tenant connection by hand; never mix control-plane and tenant queries.
- **Platform auth and tenant auth stay separate.** Share password/JWT primitives only; keep tenant auth services, platform auth services, middleware, token claims, and route namespaces separate.
- **Super Admin is control-plane only.** Do not add `super_admin` to tenant `users.role`; do not put platform users in tenant DBs; do not make `tenant_id` nullable to support platform users.
- **Audited platform access.** Any platform operation that changes a tenant, changes a user/subscription, or explicitly inspects tenant data must write a `super_admin_audit_events` row.
- **Layering:** handler (thin) → service (logic) → store (sqlc). No business logic in handlers or stores.
- **Modules are isolated.** A module talks to another only through its exported Go interface, never its store or tables.
- **Errors are explicit and wrapped** (`fmt.Errorf("...: %w", err)`). No panics in request paths; a recovery middleware is the only catch-all.
- **Structured logging via slog.** Never log secrets, passwords, tokens, or PII (names/emails/phones).
- **Config from environment only.** No secrets in code or committed files. Provide a `.env.example`.
- **Quality gates must pass before any change is "done":** `gofmt`, `go vet`, `golangci-lint`, `go test ./...` all green.
- **Tests:** every service method gets unit tests; the request lifecycle (auth → tenant resolution → handler) gets at least one integration test against a real Postgres (Docker).
- **Keep endpoint names and descriptions clean** — they are the future AI-assistant tool surface.

---

## 8. Build sequence (milestones — build in this order)

**M0 — Scaffold.** Repo + `go.mod`; Docker Compose Postgres; chi server with config, slog, recovery + request-logging middleware; control-plane pool + tenant pool resolver/cache in `platform/db`; `sqlc.yaml`, `oapi-codegen.yaml`, goose wired; empty `openapi.yaml`; CI running the quality gates. A `/health` endpoint proves the wiring.

**M1 — Control plane & tenancy.** `tenants` + `users` tables (control-plane migrations); password hashing + JWT; `signup` (creates tenant + admin user + provisions the tenant DB), `login`; AuthMiddleware + TenantMiddleware; `cmd/provision` and `cmd/migrate`. This is the spine — get tenant isolation right here before anything else.

**M2 — People core (identity).** Employees, departments, roles CRUD; org hierarchy. (Tenant DB.)

**M3 — Roster.** Shifts CRUD + publish; list by employee/date range. (Web-facing.)

**M4 — Attendance.** Mobile check-in/out with geolocation + photo upload; offline-tolerant sync (client queues, server is idempotent on a client-supplied id). (Mobile-facing.)

**M5 — Manager live view.** "Who's working now": today's published shifts joined with current attendance state.

**M6 — Super Admin / platform console.** Separate `platform_users` table; platform login/me; `PlatformAuthMiddleware`; Super Admin audit log; tenant directory/detail; tenant suspend/reactivate/change plan; platform user disable/enable; subscription management; provisioning/migration/system health; optional read-only support-mode with mandatory reason and audit event.

Ship M1–M4 to design partners before polishing. M2 can't start before M1's tenancy works; M5 needs M3 + M4.

---

## 9. Project setup steps (M0 concretely)

Tooling to install: Go (latest stable), Docker + Docker Compose, `sqlc`, `oapi-codegen`, `goose`, `golangci-lint`.

Suggested first actions for Claude Code:
1. `go mod init github.com/<you>/opero/backend` and create the directory skeleton from §4.
2. Add `docker-compose.yml` with a Postgres service; create the control-plane DB on startup.
3. Add `platform/config` (env loading), `platform/db` (build the control-plane `*pgxpool.Pool`; implement a `TenantResolver` that looks up `db_name` in the registry and returns a cached pool per tenant).
4. Add `platform/httpserver` (chi router) + `platform/middleware` (recovery, request-logging; auth/tenant come in M1) and a `/health` route.
5. Create `sqlc.yaml` (engine: postgresql, pgx/v5), `oapi-codegen.yaml` (chi-server + types), and a `Makefile` with `generate`, `migrate`, `lint`, `test`, `run`.
6. Wire CI to run `gofmt -l`, `go vet ./...`, `golangci-lint run`, `go test ./...`.
7. Confirm `make run` serves `/health` against the Dockerized Postgres.

Then proceed to M1.

---

## 10. Definition of done (every change)

A change is complete only when: the spec is updated (if the API changed) and clients regenerated; code follows the layering and tenancy rules above; all four quality gates pass; new logic has tests; and nothing in `gen/` was hand-edited. If a requirement here can't be met, stop and surface the conflict rather than working around a guardrail.
