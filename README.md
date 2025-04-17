# AI Agents SaaS – Monorepo

This repository boot‑straps a **Go backend** and a **React + Tailwind** frontend in a single mono‑repo, based on the design captured in `docs/design/0-agents.md`.

## Structure

```
.
├── cmd/            # Go entry‑points (binaries)
│   └── server/     # HTTP API server
├── internal/       # Private Go packages (not exported)
│   ├── api/        # HTTP handlers & routers
│   └── models/     # Domain structs – mirrors DB schema
├── web/            # React + Tailwind app (Vite)
└── docs/           # Product & technical docs
```

## Running the backend

```
go run ./cmd/server
```

The server listens on `:8080` by default and exposes:

* `GET /health` – health‑check
* `GET /api/agents` – list stub agents (in‑memory data)

## Running the frontend

The frontend is a standard Vite application. Inside `web/` run:

```
pnpm install   # or npm / yarn
pnpm dev
```

It proxies API requests to `localhost:8080` via the `vite.config.js` proxy setting.

## Next steps

1. Plug the Go server into a real Postgres instance using the schema from the design doc.
2. Add OpenAPI / Swagger generation from Go types so that the frontend SDK stays type‑safe.
3. Expand CI pipeline (Go test, `go vet`, frontend `eslint` + `cypress`).
