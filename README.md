# VortexUI

A high-performance, **core-agnostic** proxy management panel. VortexUI runs
Xray-core **and** sing-box behind one abstraction, uses a **user-centric** data
model (one identity, many protocols/nodes), and collects traffic via
**push-based delta streaming** for accurate, restart-safe accounting.

> Status: **early scaffolding.** The architecture and domain layer are in place;
> infrastructure adapters and transport are being wired next.

## Why VortexUI

| Capability | VortexUI | Typical panels |
|---|---|---|
| Proxy engine | Xray **and** sing-box per node/inbound | locked to one |
| Data model | user-centric (user → many inbounds) | mostly inbound-centric |
| Traffic stats | gRPC push, delta-based, TimescaleDB | REST polling, absolute counters |
| Node failover | health-gated, automatic | manual |
| Backend | Go (low RAM, high concurrency) | often Python |

## Architecture

```
cmd/panel   → control plane: REST API + web dashboard + gRPC hub
cmd/node    → agent on each server: drives local core, streams deltas

internal/
  domain/        pure entities & business rules (no I/O, fully unit-tested)
  core/          CoreDriver abstraction over Xray / sing-box
  stats/         delta aggregator (single-consumer, idempotent fold)
  config/        env-based, fail-fast configuration
  panel/port/    repository interfaces (hexagonal ports)
  platform/      adapters: postgres, redis, logger        (pending)
proto/           gRPC service definitions                 (pending)
migrations/      SQL migrations (goose/atlas)             (pending)
web/             frontend (React + TanStack + shadcn/ui)  (pending)
deploy/          docker, systemd, certs                   (pending)
```

The single most important seam is [`internal/core.CoreDriver`](internal/core/driver.go):
the whole system talks to a proxy engine only through it, so adding an engine is
implementing one interface.

## Develop

```bash
cp .env.example .env        # then set VORTEX_JWT_SECRET (openssl rand -hex 32)
make docker-up              # postgres+timescale, redis
make migrate-up             # apply schema
make build                  # bin/panel, bin/node
make test                   # unit tests (race in CI; this box lacks cgo)

# bootstrap the first admin (nothing can log in until you do)
./bin/panel admin create --username root --password 'change-me' --sudo
```

Run a node agent on each proxy server (it serves the NodeService over mTLS; the
panel dials in):

```bash
VORTEX_NODE_LISTEN=:50051 \
VORTEX_CORE=xray \                       # or: singbox
VORTEX_CORE_BIN=/usr/local/bin/xray \
VORTEX_CORE_CONFIG=/etc/vortex/core.json \
VORTEX_TLS_CERT=node.crt VORTEX_TLS_KEY=node.key VORTEX_TLS_CA=ca.crt \
./bin/node
```

Each node independently runs **Xray-core or sing-box** behind the same driver
interface — set `VORTEX_CORE` per node.

Clients fetch their config from `GET /sub/<token>`; the response format adapts to
the client (Clash YAML, sing-box JSON, or base64) and can be forced with
`?format=clash|singbox|base64`.

The full REST API is documented in [docs/openapi.yaml](docs/openapi.yaml).

## Engineering guardrails

- `go test -race` on every package; domain logic stays at high coverage.
- Type-safe DB access (sqlc) — no dynamic ORM string-building.
- `context.Context` end-to-end; no unbounded goroutines.
- Structured logging (slog) + OpenTelemetry (planned).
- `golangci-lint` + `gosec` gate merges.

## License

AGPL-3.0 (planned, to match the copyleft norm of the ecosystem).
