# Opero backend

Go modular monolith. See the repo-root `CLAUDE.md` for architecture and rules.

## Prerequisites

Install: Go (latest stable), Docker + Docker Compose, and for codegen/migrations:
`sqlc`, `oapi-codegen`, `goose`, `golangci-lint`.

## First run (M0)

```bash
cd backend
cp .env.example .env        # adjust if needed
make tidy                   # resolves chi + pgx, writes go.sum
make up                     # starts Postgres with the control-plane DB
make run                    # serves the API
# in another shell:
curl -s localhost:8080/health   # -> {"status":"ok"}
```

## Common targets

`make help` lists them. Key ones: `generate` (oapi + sqlc), `migrate`,
`lint`, `vet`, `fmt`, `test`, `run`, `up`, `down`.

## Notes

- `gen/` is committed and generated — never hand-edit it; run `make generate`.
- `make generate` needs migrations + queries to exist (from M1); it is a no-op
  contract-wise until then.
- Config is environment-only; see `.env.example`.
