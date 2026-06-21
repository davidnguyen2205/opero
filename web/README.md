# Opero web

Manager-facing web client (React + TypeScript + Vite). See `AGENTS.md` for the
rules any coding agent (Codex/Claude) must follow here.

This project uses **bun** as its package manager and runner.

## Setup

```bash
cd web
bun install
bun run gen:api      # generates src/api/schema.d.ts from ../api/openapi.yaml
bun run dev          # http://localhost:5173
```

The backend must be running (`cd ../backend && make up && make run`) and its
`CORS_ALLOWED_ORIGINS` must include `http://localhost:5173` (it does by default
in `backend/.env.example`).

## How the contract works

`../api/openapi.yaml` is the single source of truth. `bun run gen:api` regenerates
`src/api/schema.d.ts`; the typed API layer under `src/api/` wraps `openapi-fetch`
over those types. Never hand-write API types or call the API with untyped `fetch`.
If you need an endpoint that isn't in the spec, stop — the spec changes on the
backend first.

## Status / caveats

- This scaffold was authored without a local JS toolchain, so **dependency
  versions in `package.json` are best-effort and unverified**. Run `bun install`;
  if a version is unavailable, `bun add <pkg>@latest` and adjust.
- The exact `openapi-fetch` middleware API may vary by version — align the
  `src/api/` client with the installed version if `bun run build` flags it.
- `src/api/schema.d.ts` does not exist until you run `bun run gen:api`; type
  errors importing the generated schema before that are expected.
