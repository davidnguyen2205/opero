# Opero Web — Agent Instructions (Codex / any coding agent)

This is the **manager-facing web client** for Opero (React + TypeScript + Vite).
Read the repo-root `CLAUDE.md` for the overall product and architecture; this
file is the authoritative guide for work inside `web/`.

## The contract boundary (most important rule)

The backend API is defined by **`../api/openapi.yaml`** — the single source of
truth. This client consumes a **generated** TypeScript client from that spec.

- **Never hand-write API request/response types.** They are generated into
  `src/api/schema.d.ts` by `bun run gen:api`. Regenerate; never edit that file.
- **Never call the API with ad-hoc `fetch` and inline types.** Use the typed
  client in `src/api/client.ts` (wraps `openapi-fetch` over the generated types).
- If you need an endpoint that does not exist in `../api/openapi.yaml`, or one
  whose shape differs from what you need: **STOP and report it.** Do not invent
  it client-side. The spec changes on the backend side first, then we regenerate.

## Scope — what you may and may not touch

- **You own `web/` only.** Do **not** edit `../backend/`, `../api/openapi.yaml`,
  or anything under `../backend/gen/`. The spec is a frozen input here.
- If the API contract needs to change, surface it; it is handled outside `web/`.

## Tech & conventions

- React + TypeScript, built with Vite. Package manager: **bun** (use `bun install`,
  `bun add <pkg>`, `bun run <script>` — not npm).
- API types via **`openapi-typescript`** (generates `src/api/schema.d.ts`).
- API calls via **`openapi-fetch`** (typed client in `src/api/client.ts`).
  - This was chosen for being lightweight and framework-agnostic. If the team
    later wants React Query hooks, `orval` is the likely swap — but discuss
    before changing the data layer.
- JSON fields are **snake_case**; timestamps are **ISO-8601**; ids are UUID
  strings. Match the generated types exactly — do not rename fields.
- Keep all API access in `src/api/`, not scattered through components.
- Auth: the API uses a bearer JWT obtained from `POST /auth/login` (or signup).
  Store it in memory/an auth context and set it via `setAuthToken()` in
  `src/api/client.ts`. Do not log tokens.

## Workflow

1. Ensure the spec is current, then `bun run gen:api` to (re)generate types.
2. Build features against the generated client. Type-check with `bun run build`
   (or `bun run typecheck`) before considering work done.
3. The backend must be running and CORS must allow this origin
   (`CORS_ALLOWED_ORIGINS` in the backend `.env`, default `http://localhost:5173`).

## Milestones (build only what the spec exposes)

- **M3 — Roster** (when shifts endpoints exist): create/edit/publish shifts,
  list by employee and date range.
- **M5 — Manager live view**: today's published shifts joined with current
  attendance ("who's working now").
- Already in the spec now: auth (signup/login/me) and identity
  (departments, employees, roles).

## Definition of done

- `bun run build` / `tsc` type-checks clean; lint passes; no hand-written API
  types; `src/api/schema.d.ts` is generated, not edited.
