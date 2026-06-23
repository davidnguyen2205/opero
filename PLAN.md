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

**All v1 milestones (M0–M5) are implemented and the quality gates pass.** Remaining work is
hardening and the deferred v1.1 / phase-2 scope (see end).

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
- **Phase 2**: payroll, cross-tenant analytics/insights, in-product AI assistant (tool-calling over the existing REST API — keep operationIds/descriptions clean for this).
