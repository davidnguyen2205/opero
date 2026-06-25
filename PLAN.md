# Opero — Implementation Plan & Status

> Living status doc for Opero v1. The authoritative architecture and rules live in
> [`CLAUDE.md`](./CLAUDE.md); this file tracks **what is built vs. what remains**.
> Last verified: **2026-06-23**.

## How status was determined

- **Code presence**: read of `backend/`, `web/`, `mobile/`, and `api/openapi.yaml`.
- **Quality gates** (run on `backend/`, 2026-06-23): `gofmt -l` clean · `go vet ./...` clean ·
  `golangci-lint run` → 0 issues · `go test ./...` → all packages `ok`.
- Test results above were partly served from Go's build cache (no source changed since last run),
  but every test package reports `ok`. Integration tests run against the local Dockerized
  Postgres (`opero-postgres`, healthy).

## Status at a glance

| Milestone | Scope | Status |
|---|---|---|
| **M0** | Scaffold (server, config, middleware, db pools, codegen, CI, `/health`) | ✅ Done |
| **M1** | Control plane & tenancy (tenants/users, auth, signup/login, middleware, provision/migrate) | ✅ Done |
| **M2** | People core / identity (employees, departments, roles, hierarchy) | ✅ Done |
| **M3** | Roster (locations, shifts CRUD + publish, list by employee/date) | ✅ Done |
| **M4** | Attendance (check-in/out, geo + photo, offline-idempotent sync) | ✅ Done |
| **M5** | Manager live view ("who's working now") | ✅ Done |
| **M6** | Super Admin / platform console (separate platform auth, tenants, subscriptions, audit, health) | 🚧 Backend done, web pending |

**All v1 milestones (M0–M5) are implemented and the quality gates pass.** Remaining work is
hardening plus the next control-plane Super Admin scope.

---

## M0 — Scaffold ✅

- `cmd/api/main.go` wires chi via `platform/httpserver`; `slog` JSON logging.
- `platform/config` (env), `platform/db` (control-plane `*pgxpool.Pool` + cached `TenantResolver`).
- `platform/middleware`: recovery, request logging, CORS.
- `sqlc.yaml` (pgx/v5), `oapi-codegen.yaml` (chi-server + types), goose migration dirs, Docker Compose Postgres.
- `.github/workflows/ci.yml` runs gofmt / vet / golangci-lint (v2.12.2) / test.
- `/health` endpoint live.

## M1 — Control plane & tenancy ✅

- `db/migrations/controlplane/00001_init.sql`: `tenants`, `users` (unique `(tenant_id, lower(email))`), `subscriptions` stub, `set_updated_at()` trigger.
- `platform/auth`: password hashing + JWT issue/verify (`auth_test.go`).
- Endpoints: `signup` (creates tenant + admin + provisions tenant DB), `login`, `getCurrentUser` (`GET /auth/me`).
- `middleware/auth.go` (JWT → user_id + tenant_id) and `middleware/tenant.go` (registry lookup → scoped pool in context).
- `cmd/provision` (create + migrate + seed one tenant), `cmd/migrate` (fan-out across tenant DBs).
- `controlplane/integration_test.go` covers signup → login.

## M2 — People core (identity) ✅

- Migrations: `00001_identity.sql` (`departments` w/ `parent_id` hierarchy, `employees`), `00002_roles.sql` (`roles` + `role_id` FK).
- CRUD operationIds for **departments**, **employees**, **roles** (5 each) + `createEmployeeLogin` (`POST /employees/{id}/login`).
- `identity/{handler,service,store}.go`; unit (`service_test.go`) + integration tests.

## M3 — Roster ✅

- Migrations: `00003_roster.sql` (`locations`, `shifts`), `00004_shifts_employee_restrict.sql` (FK `ON DELETE RESTRICT`).
- **Locations** CRUD; **Shifts** CRUD with filters (employee_id, status, from/to); `publishShift`; `listMyShifts` (`GET /me/shifts`).
- `roster/{handler,service,store}.go`; unit + integration tests.

## M4 — Attendance ✅

- Migration `00005_attendance.sql`: `attendance_records` with `client_id` **unique** (offline idempotency key), geo lat/lng + photo URLs for check-in and check-out.
- Endpoints: `checkIn`, `checkOut` (idempotent on `client_id` — replay returns existing record), `listAttendance` (filters).
- `attendance/{handler,service,store}.go`; idempotency covered in `service_test.go`.

## M5 — Manager live view ✅

- `GET /live` → `LiveViewEntry[]` (employee, shift, attendance_status: not_checked_in | checked_in | checked_out, timestamps).
- `liveview/{handler,service}.go` owns no tables — composes `roster`, `attendance`, `identity` via narrow Go interfaces (respects module isolation).
- `service_test.go` covers the join logic.

---

## M6 — Super Admin / platform console 🚧

Goal: build an internal Opero control-plane surface for managing the SaaS platform itself. Super Admin is separate from tenant admins and must not live inside any tenant database.

### Backend status

- ✅ `platform_users` and `super_admin_audit_events` control-plane migration added.
- ✅ Platform JWTs use `kind=platform`; tenant JWTs use `kind=tenant`.
- ✅ `PlatformAuthMiddleware` rejects tenant tokens on `/platform/...` routes.
- ✅ Platform auth endpoints implemented: `POST /platform/auth/login`, `GET /platform/auth/me`.
- ✅ Bootstrap CLI added: `go run ./cmd/platform-user -email ... -password ... -role super_admin`.
- ✅ Platform admin endpoints implemented for tenants, tenant login users, subscriptions, system health, and audit events.
- ✅ Mutating platform operations write `super_admin_audit_events`.
- ✅ OpenAPI, oapi-codegen, and sqlc generated code updated.
- ✅ Backend tests, vet, gofmt, and golangci-lint pass.
- ⏳ Web `/super-admin` UI is not implemented yet.

### Architecture decisions

- Add `platform_users` in the control-plane DB for Opero staff only.
- Keep tenant users in existing control-plane `users`; do **not** add `super_admin` to tenant `users.role`.
- Share low-level `platform/auth` primitives for password hashing and JWT signing/verification.
- Split business auth:
  - tenant auth: `/auth/login`, `/auth/me`, `users`, token `kind=tenant`, includes `tenant_id`
  - platform auth: `/platform/auth/login`, `/platform/auth/me`, `platform_users`, token `kind=platform`, no `tenant_id`
- Add `PlatformAuthMiddleware`; keep it separate from tenant `AuthMiddleware` + `TenantMiddleware`.
- Platform routes never resolve a tenant DB implicitly. Any tenant-specific platform access must name the tenant explicitly and write an audit event.

### Control-plane schema

- `platform_users`
  - `id`
  - `email`
  - `password_hash`
  - `role`: `super_admin` | `support` | `ops`
  - `status`: `active` | `disabled`
  - timestamps
- `super_admin_audit_events`
  - `id`
  - `actor_platform_user_id`
  - `action`
  - `target_type`
  - `target_id`
  - `tenant_id`
  - `metadata jsonb`
  - `created_at`

### Initial API surface

- `POST /platform/auth/login`
- `GET /platform/auth/me`
- `GET /platform/tenants`
- `GET /platform/tenants/{id}`
- `PATCH /platform/tenants/{id}` — status/name/plan changes
- `GET /platform/users`
- `PATCH /platform/users/{id}` — disable/enable tenant login users
- `GET /platform/subscriptions`
- `PATCH /platform/subscriptions/{id}`
- `GET /platform/system/health`
- `GET /platform/audit-events`

### Web surface

- Add a separate `/super-admin` route group.
- First views:
  - platform login
  - tenant directory
  - tenant detail
  - platform users / tenant users
  - subscriptions
  - system health
  - audit events

### Implementation order

1. Update `api/openapi.yaml` with platform auth, tenants, users, subscriptions, health, and audit endpoints.
2. Add control-plane migrations for `platform_users` and `super_admin_audit_events`.
3. Add sqlc queries for platform auth, tenant listing/detail/update, user status updates, subscriptions, health inputs, and audit event creation/listing.
4. Generate backend code.
5. Implement `controlplane/platformauth` and `PlatformAuthMiddleware`.
6. Implement platform tenant/user/subscription/audit handlers and services.
7. Add unit and integration tests for platform login, middleware rejection of tenant tokens, tenant listing, status updates, and audit writes.
8. Regenerate web client types and build the `/super-admin` UI.
9. Run backend quality gates and web build.

### Future additions

- Read-only support mode with mandatory reason and audit event.
- Provisioning step/event table for detailed retry/status visibility.
- Cross-tenant aggregate usage metrics.
- Stripe/payment integration when billing is no longer manual.

---

## Clients

- **Web** (`web/`, React + TS + Vite): full manager UI in `src/App.tsx` — Live, Roster, People, Departments, Roles, Locations; signup/login; CRUD + publish. Typed client generated from `openapi.yaml`.
  - **Build verified 2026-06-23** via `bun run build` (`tsc && vite build`): typecheck clean, 34 modules transformed, production bundle emitted to `dist/` (~178 kB JS / 56 kB gzip). No errors.
- **Mobile** (`mobile/`, Flutter): login + shifts screens; attendance capture (geo + photo); offline queue with client-id idempotency; generated Dart API client.
  - **Build NOT run** — no Flutter/Dart SDK is installed on this machine (checked PATH, fvm, asdf, common dirs, `mdfind`). Cannot be verified here.
  - `mobile/pubspec.yaml` carries an author note that the package versions are best-effort from training knowledge and may need `flutter pub upgrade --major-versions` reconciliation (esp. `geolocator`, `image_picker`). So the mobile build is **unverified and at some risk** until run on a machine with Flutter.

---

## Verification gaps / suggested next checks

These are not known failures — they are things **not yet verified in this pass**:

1. **Web build & typecheck** — run `npm run build` / `tsc` in `web/`.
2. **Mobile build** — `flutter analyze` / build in `mobile/`.
3. **Client regeneration drift** — confirm generated TS/Dart clients match the current `openapi.yaml`.
4. **`cmd/migrate` partial-failure reporting** — CLAUDE.md §3.5 requires clear per-tenant success/failure output; confirm behavior against multiple tenant DBs.
5. **End-to-end smoke** — signup → login → create employee → publish shift → mobile check-in → live view, against Dockerized Postgres.
6. **Untracked artifact** — `backend/design/Opero (standalone).html` (~57KB Figma export) is uncommitted; decide whether it belongs in the repo.

---

## Out of scope for v1 (do not build now)

- **v1.1**: leave management (`leave_requests`, `leave_balances`), document/cert tracking (`documents`). Schema leaves room; not implemented.
- **Phase 2**: payroll, advanced cross-tenant analytics/insights, in-product AI assistant (tool-calling over the existing REST API — keep operationIds/descriptions clean for this).
