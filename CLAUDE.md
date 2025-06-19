# MEL Agent Project Context

## Project Overview
AI Agents SaaS platform with Go backend and React frontend. This is a monorepo that provides a visual workflow builder for AI agents with support for various node types, triggers, and integrations.

## Architecture
- **Backend**: Go with Chi router, PostgreSQL database, WebSockets
- **Frontend**: React + Vite + Tailwind CSS
- **Workers**: Dual worker model - local embedded workers + distributed remote workers
- **Database**: PostgreSQL with migrations
- **Testing**: Testcontainers for database integration tests
- **Containerization**: Docker Compose for local development

## Development Commands

### Backend (Go)
- **Run server**: `go run ./cmd/server server` (includes local workers)
- **Run API server only**: `go run ./cmd/server api-server` (no embedded workers)
- **Run remote worker**: `go run ./cmd/server worker --token <token>`
- **Run with Docker**: `docker compose up --build`
- **Test**: `go test ./...` (includes testcontainer integration tests)
- **Build**: `go build ./cmd/server`
- **Lint**: `go vet ./...`
- **Format**: `go fmt ./...`

### CLI Help & Configuration
- **Show help**: `go run ./cmd/server --help`
- **Server help**: `go run ./cmd/server server --help`
- **Worker help**: `go run ./cmd/server worker --help`
- **Shell completion**: `go run ./cmd/server completion bash`

### Advanced CLI Usage
- **Custom port**: `go run ./cmd/server server --port 9090`
- **Worker with options**: `go run ./cmd/server worker --token <token> --concurrency 10 --id worker-custom`
- **Config file**: Create `config.yaml` (see config.yaml.example)
- **Environment vars**: `PORT=8080 MEL_WORKER_TOKEN=abc123 go run ./cmd/server server`

### Frontend (React)
- **Install dependencies**: `cd web && pnpm install`
- **Start dev server**: `cd web && pnpm dev`
- **Build**: `cd web && pnpm build`
- **Lint**: `cd web && pnpm lint`
- **Reformat the javascript using pnpm format from time to time**

### Database
- **Connection**: `postgres://postgres:postgres@localhost:5432/agentsaas?sslmode=disable`
- **Migrations**: Located in `migrations/` directory
- **Docker setup**: Postgres runs on localhost:5432

## Development Principles
- Our goal is to create a well-tested project. Always be mindful of creating reasonably well-sized components and testing them. Create e2e tests with Cypress where applicable
- Remember to keep components at a reasonable size and add new tests for new components and functionality
- Use and contribute to testutils we already have for db migrations etc

## API Endpoints

### Node Types API
- **Get all node types**: `GET /api/node-types`
- **Filter by kind**: `GET /api/node-types?kind=model,memory,action,tool,trigger`
- **Single kind**: `GET /api/node-types?kind=model`

#### Node Kinds
- `action` - Can execute as workflow steps (default for all nodes)
- `model` - Provides AI model interaction capabilities
- `memory` - Offers memory storage and retrieval
- `tool` - Enables tool execution capabilities  
- `trigger` - Can initiate workflow execution

Nodes can implement multiple kinds (e.g., OpenAI model has kinds: `["action", "model"]`).

## Worker System

### Local Workers
- Embedded within the API server process
- Zero configuration required
- Automatic startup with server

### Remote Workers
- Standalone processes that connect to API server
- Require authentication token via `MEL_WORKER_TOKEN`
- Support horizontal scaling and geographic distribution
- Auto-registration with heartbeat monitoring

### Worker Commands
- `go run ./cmd/server worker -token <token>` - Start remote worker
- `go run ./cmd/server worker -id <id> -token <token> -concurrency <n>` - Start with custom settings

### Worker API Endpoints
- `POST /api/workers` - Register worker
- `PUT /api/workers/{id}/heartbeat` - Update worker heartbeat
- `POST /api/workers/{id}/claim-work` - Claim work items
- `POST /api/workers/{id}/complete-work/{itemID}` - Complete work
- `DELETE /api/workers/{id}` - Unregister worker

## Development Guidelines
- When committing code via git, never mention claude

## Testing
- We are using testcontainers to test database dependencies
- Database tests rely on our migration system via `testutil.SetupPostgresWithMigrations()`
- Integration tests cover worker registration, work claiming, lifecycle management
- Router integration tests verify all API endpoints are accessible
- Use existing testutil functions for database setup in tests

## Router Architecture
- Server uses a merged API handler approach to combine main API and workflow engine routes
- Fixed Chi router conflict where multiple r.Mount("/api", ...) calls would override each other
- `createMergedAPIHandler()` function implements fallback routing: tries main API first, then workflow engine
- Comprehensive router integration tests ensure all endpoints remain accessible