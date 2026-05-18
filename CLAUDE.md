# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

**Evolution Go** is a high-performance WhatsApp API built in Go, using a forked [whatsmeow](https://github.com/tulir/whatsmeow) library vendored locally at `./whatsmeow-lib/`. The fork is declared via `go.mod` replace directive: `replace go.mau.fi/whatsmeow => ./whatsmeow-lib`. Changes to WhatsApp protocol behavior go in `whatsmeow-lib/`, not in `pkg/`.

## Build & Development Commands

```bash
# Development (loads .env automatically)
make dev           # go run with -dev flag (enables godotenv)
make run           # production mode (no .env loading)
make watch         # hot reload via air

# Build
make build         # outputs to build/evolution-go
make build-linux   # GOOS=linux GOARCH=amd64

# Testing
make test          # go test -v ./...
make test-race     # race detector enabled
go test -v ./pkg/instance/...   # single package
go test -run TestFoo ./pkg/...  # single test

# Code quality
make vet           # go vet ./...
make fmt           # go fmt ./...
make lint          # golangci-lint (must be installed separately)

# Swagger docs
make swagger       # requires: go install github.com/swaggo/swag/cmd/swag@latest

# Dependency management
make deps          # download + verify
make deps-reset    # clean cache + re-download (fixes module issues)
```

**Note:** `github.com/chai2010/webp` requires CGO. On Windows without a C toolchain, `go build ./...` will fail on that package. Use `go vet ./pkg/...` to validate specific packages instead.

## Configuration (.env)

Required:
- `GLOBAL_API_KEY` — admin-level API key (used by `AuthAdmin` middleware)
- `SERVER_PORT` — HTTP listen port

Key optional variables (see `pkg/config/env/env.go` for full list):
- `POSTGRES_AUTH_DB` / `POSTGRES_USERS_DB` — if unset, falls back to SQLite at `./dbdata/`
- `WEBHOOK_URL` — global webhook URL for all instances
- `AMQP_URL`, `NATS_URL` — message queue integration
- `CONNECT_ON_STARTUP` — reconnect all previously-connected instances on boot
- `QRCODE_MAX_COUNT` — max QR regenerations before forcing logout (0 = unlimited)
- `CLIENT_NAME` — used for multi-server deployments (filters which instances to reconnect on startup)
- `MINIO_ENABLED` / `MINIO_*` — object storage for media

## Architecture

### Request Flow

```
HTTP Request
  → GateMiddleware (core license check — blocks with 503 if unlicensed)
  → Auth middleware (two variants):
      AuthAdmin: validates apikey == GLOBAL_API_KEY (admin routes)
      Auth: looks up instance by apikey token → sets "instance" in gin context
  → Handler → Service → Repository / WhatsmeowService
```

### Shared Mutable State

Two maps are created in `main.go` and **passed by reference** to all services:

```go
killChannel   map[string]chan bool          // one channel per instance ID
clientPointer map[string]*whatsmeow.Client  // live WA client per instance ID
```

Both `instanceService` and `whatsmeowService` (and every other service) share these exact maps. Mutations to `clientPointer` or `killChannel` in any service are immediately visible everywhere. This is the concurrency model — there is no locking around these maps.

### Instance Lifecycle

1. **Create** (`POST /instance/create`) — persists to DB; optionally sets `webhook` field; no WA connection yet
2. **Connect** (`POST /instance/connect`, auth: instance token) — updates DB (events, webhook, etc.), starts `StartClient` in a goroutine
3. **GetQr** (`GET /instance/qr`, auth: instance token) — checks `clientPointer` and `killChannel` to avoid double-starting; polls DB for QR code with retries
4. **Pair** (`POST /instance/pair`) — pairing code alternative to QR scan
5. **Disconnect/Logout/Delete** — sends to `killChannel`, cleans up maps

`StartClient` in `pkg/whatsmeow/service/whatsmeow.go` is the core goroutine. It:
- Creates/reuses a whatsmeow device store (SQLite or PostgreSQL)
- Opens a QR channel (new device) or directly connects (existing session)
- Saves QR codes to DB via `instanceRepository.UpdateQrcode`
- Runs an event loop (`myEventHandler`) that produces events to all configured producers
- Blocks on `killChannel[instanceId]` after the QR/session loop ends

### Event System

All WA events are dispatched through the `Producer` interface (`pkg/events/interfaces/producer.go`):

```go
Produce(queueName string, payload []byte, webhookUrl string, userID string) error
```

Four implementations:
- `pkg/events/webhook/` — HTTP POST with retry (5 attempts, 30s interval)
- `pkg/events/rabbitmq/` — AMQP with global queue support
- `pkg/events/nats/` — NATS pub/sub
- `pkg/events/websocket/` — WebSocket via `gorilla/websocket`

Per-instance webhook URL (`instance.Webhook`) is passed alongside the global webhook. Both are fired concurrently.

### Authentication

- **Admin routes** (`/instance/create`, `/instance/all`, `/instance/delete/:id`, etc.) — require `apikey: GLOBAL_API_KEY`
- **Instance routes** (`/instance/connect`, `/instance/qr`, `/send/*`, etc.) — require `apikey: <instance.Token>`; the instance object is injected into gin context under `"instance"` key

### License / Core Gate

`pkg/core/c0.go` implements license management. `GateMiddleware` is applied globally and returns `503` for all routes except `/server/ok` and `/license/*` if the license is not active. License state is stored in the `runtime_configs` table. On first run, activate via the Manager UI at `/manager/`.

### Database

- **GORM (PostgreSQL or SQLite)** — `Instance`, `Message`, `Label` models auto-migrated on startup
- **raw `*sql.DB` (PostgreSQL or SQLite)** — used by whatsmeow for device/session storage (WA protocol state)
- If `POSTGRES_AUTH_DB` is unset, whatsmeow uses `./dbdata/main.db` (SQLite)
- If `POSTGRES_USERS_DB` is unset, GORM uses a separate SQLite at `./dbdata/users.db`

### Manager UI

A pre-built React SPA served from `manager/dist/`. Routes under `/manager/*` serve `index.html` for client-side routing. The JS bundle communicates with the same backend via the REST API using instance tokens.

## Key Files for Common Tasks

| Task | File |
|---|---|
| Add/change instance fields | `pkg/instance/model/instance_model.go` |
| Instance create/connect/QR logic | `pkg/instance/service/instance_service.go` |
| WA event handling & QR generation | `pkg/whatsmeow/service/whatsmeow.go` |
| HTTP routes | `pkg/routes/routes.go` |
| Environment variables | `pkg/config/env/env.go`, `pkg/config/config.go` |
| Event fanout | `pkg/events/*/` |
| Bootstrap / wiring | `cmd/evolution-go/main.go` |
