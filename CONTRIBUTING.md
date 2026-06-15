# Contributing to VortexUI

Thanks for your interest in improving VortexUI! This guide covers how to set up a
dev environment, the project layout, and the conventions we follow.

## Development setup

```bash
git clone https://github.com/iPmartNetwork/VortexUI && cd VortexUI

# Backend deps (PostgreSQL/TimescaleDB + Redis)
docker compose up -d

# Configure
cp .env.example .env
# set VORTEX_JWT_SECRET: openssl rand -hex 32

# Run the control plane
make certs        # dev mTLS chain
make run-panel

# Frontend (separate terminal)
cd web
npm install
npm run dev -- --host 127.0.0.1
```

For a UI-only demo without a database, run the in-memory mock API:

```bash
go run ./tools/mockapi      # serves :8080
cd web && npm run dev -- --host 127.0.0.1
```

## Project layout

| Path | Purpose |
|------|---------|
| `cmd/panel`, `cmd/node` | Binaries: control plane and node agent |
| `internal/core/{xray,singbox}` | Engine drivers behind a common interface |
| `internal/panel/{api,service,hub,port}` | HTTP API, services, node hub, ports |
| `internal/platform/postgres` | sqlc-generated DB access (`queries/*.sql`) |
| `internal/transport` | gRPC server/client + generated protobufs |
| `proto/` | gRPC definitions (regenerate with `make proto`) |
| `migrations/` | goose migrations (schema mirror in `schema.sql`) |
| `web/` | React + Vite frontend |
| `deploy/`, `install.sh`, `scripts/` | Deployment and installer |

## Conventions

- **Commits** follow Conventional Commits: `feat:`, `fix:`, `docs:`, `refactor:`,
  `chore:`, etc.
- **Go**: keep it idiomatic; run `go build ./...`, `go vet ./...`, and
  `go test ./...` before pushing. Match the surrounding style.
- **Codegen**: after editing `proto/` run `make proto`; after editing
  `queries/*.sql` run `make sqlc`. Commit the generated files.
- **Migrations**: add a new `migrations/000N_*.sql` *and* mirror the change in
  `internal/platform/postgres/schema.sql` (sqlc reads the latter).
- **Frontend**: `npm run build` must pass (it type-checks). Add any new i18n key
  to all 8 languages in `web/src/i18n/dict.ts`.

## Pull requests

1. Fork and branch from `master` (`git checkout -b feat/amazing`).
2. Keep changes focused; update docs/tests alongside code.
3. Ensure CI is green (build, vet, test, lint, web build).
4. Open the PR with a clear description of what and why.

## Reporting issues

Include the version/commit, your platform, repro steps, and relevant logs
(`vortexui logs`). For security issues, please disclose privately rather than via a
public issue.

By contributing, you agree that your contributions are licensed under the project's
**GPL-3.0** license.
