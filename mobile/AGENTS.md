# Opero Mobile — Agent Instructions

Field-staff mobile client (Flutter/Dart) in `mobile/`. Read the repo-root
`CLAUDE.md` for product/architecture; this file governs work inside `mobile/`.
See `SHIPPING.md` for how to build/run and the current verification status.

## The contract boundary (most important rule)

The backend API is defined by **`../api/openapi.yaml`** — the single source of
truth. This app must match it exactly (snake_case JSON, ISO-8601 timestamps,
UUID strings).

- If you need an endpoint that does not exist, or one whose shape differs:
  **STOP and report it.** Do not invent client-side behavior that diverges from
  the spec. The spec changes on the backend side first, then this app follows.
- **Deliberate documented deviation:** the API layer (`lib/api/`) is a
  hand-written thin client, NOT generated, because the mobile surface is tiny
  and a Dart generator could not be run in the authoring environment. This is
  the §6 guardrail exception, recorded here and in SHIPPING.md §7. If you change
  the data layer to a generated client, keep `lib/attendance/` + `lib/offline/`
  intact and update SHIPPING.md.

## Scope — what you may and may not touch

- **You own `mobile/` only.** Do NOT edit `../backend/`, `../api/openapi.yaml`,
  `../backend/gen/`, or `../web/`.
- The app is field-staff-facing: sign in, see my shifts, check in/out. Manager/
  admin features belong in the web client.

## Architecture (keep these boundaries)

- `lib/api/` — thin typed HTTP client + DTOs + auth token store. The only place
  that talks to the network.
- `lib/offline/` — durable FIFO queue of pending attendance actions.
- `lib/attendance/` — `AttendanceService` (client_id lifecycle + sync) and
  capture (geo/photo). The offline-tolerance core lives here.
- `lib/screens/` — UI only; calls services, holds no business logic.

## Offline-tolerance invariants (do not break)

- Every check-in generates exactly one `client_id` (UUID v4), persisted BEFORE
  any network attempt. A check-out reuses the SAME `client_id` as its check-in.
- Actions are enqueued durably first, then synced. The server is idempotent on
  `client_id`, so the sync loop may replay any action any number of times.
- Sync is FIFO and stops on the first retryable (network/5xx) failure to
  preserve ordering; it drops terminal (4xx) poison messages so the queue
  progresses. Don't reorder or parallelize the queue.

## Definition of done

- `flutter analyze` clean, `flutter build` succeeds for at least one platform.
- No hand-written drift from `../api/openapi.yaml`.
- New logic (especially the offline queue / sync) has tests.
- Permissions configured per SHIPPING.md §4.
