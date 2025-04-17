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

1. Start Postgres + server via Docker Compose (recommended):

```bash
docker compose up --build
```

This automatically exposes:

* Go API on http://localhost:8080  (health at `/health`)
* Postgres on localhost:5432 (`postgres / postgres`)

2. Alternatively run Postgres yourself, apply migrations in `migrations/`, then:

```bash
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/agentsaas?sslmode=disable"
go run ./cmd/server
```

Initial endpoints:

* `GET /api/agents` – list agents (persisted)
* `POST /api/agents` – create agent `{name, description, user_id?}`
* `GET /api/connections` – list connections
* `POST /api/connections` – create connection

## Running the frontend

The frontend is a standard Vite application. Run:

```bash
cd web
pnpm install   # or npm / yarn
pnpm dev       # starts Vite on :5173 and proxies to localhost:8080
```

It proxies API requests to `localhost:8080` via the `vite.config.js` proxy setting.

## Next steps

1. Plug the Go server into a real Postgres instance using the schema from the design doc.
2. Add OpenAPI / Swagger generation from Go types so that the frontend SDK stays type‑safe.
3. Expand CI pipeline (Go test, `go vet`, frontend `eslint` + `cypress`).
