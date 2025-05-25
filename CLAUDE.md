# MEL Agent Project Context

## Project Overview
AI Agents SaaS platform with Go backend and React frontend. This is a monorepo that provides a visual workflow builder for AI agents with support for various node types, triggers, and integrations.

## Architecture
- **Backend**: Go with Chi router, PostgreSQL database
- **Frontend**: React + Vite + Tailwind CSS
- **Database**: PostgreSQL with migrations
- **Containerization**: Docker Compose for local development

## Development Commands

### Backend (Go)
- **Run server**: `go run ./cmd/server`
- **Run with Docker**: `docker compose up --build`
- **Test**: `go test ./...`
- **Build**: `go build ./cmd/server`
- **Lint**: `go vet ./...`

### Frontend (React)
- **Install dependencies**: `cd web && pnpm install`
- **Start dev server**: `cd web && pnpm dev`
- **Build**: `cd web && pnpm build`
- **Lint**: `cd web && pnpm lint`

### Database
- **Connection**: `postgres://postgres:postgres@localhost:5432/agentsaas?sslmode=disable`
- **Migrations**: Located in `migrations/` directory
- **Docker setup**: Postgres runs on localhost:5432

## Key Components

### Backend Structure
- `cmd/server/` - Main application entry point
- `internal/api/` - HTTP handlers and routers
- `internal/models/` - Domain models
- `internal/db/` - Database layer
- `internal/runs/` - Agent execution engine
- `internal/triggers/` - Trigger management
- `pkg/nodes/` - Node type definitions
- `pkg/plugin/` - Plugin system

### Frontend Structure
- `web/src/components/` - React components
- `web/src/pages/` - Page components
- Key components: ChatAssistant, NodeDetailsPanel, RunDetailsPanel

### Node Types
Available node types include:
- Agent nodes (LLM interactions)
- HTTP request/response
- Database queries
- Email sending
- File I/O operations
- Control flow (if/else, for_each, switch)
- Integrations (Slack, webhooks)
- Utility nodes (log, delay, transform)

## API Endpoints
- `GET /api/agents` - List agents
- `POST /api/agents` - Create agent
- `GET /api/connections` - List connections
- `POST /api/connections` - Create connection
- WebSocket endpoint for real-time updates

## Environment Setup
1. Start services: `docker compose up --build`
2. API available at: http://localhost:8080
3. Frontend dev server: `cd web && pnpm dev` (http://localhost:5173)
4. Health check: http://localhost:8080/health

## Testing
- Go tests: `go test ./...`
- Frontend tests: Not currently configured (placeholder in package.json)

## Public API for Node Development

The `pkg/api` package provides a complete public API for building node types:

### Core Types
- `NodeDefinition` - Interface for implementing node types
- `ExecutionContext` - Execution context with agent ID and variables
- `Node` - Represents a workflow node with configuration
- `NodeType` - Metadata for node type (label, category, parameters)
- `ParameterDefinition` - Defines node configuration parameters

### Registration and Discovery
```go
import "github.com/cedricziel/mel-agent/pkg/api"

// Register a node definition globally
api.RegisterNodeDefinition(myNodeDef)

// List all registered definitions
definitions := api.ListNodeDefinitions()

// Find specific definition by type
def := api.FindDefinition("my_node_type")
```

### Creating Node Types
```go
type MyNodeDefinition struct{}

func (d MyNodeDefinition) Meta() api.NodeType {
    return api.NodeType{
        Type:     "my_node",
        Label:    "My Node",
        Category: "Custom",
        Parameters: []api.ParameterDefinition{
            {Name: "config", Label: "Configuration", Type: "string", Required: true},
        },
    }
}

func (d MyNodeDefinition) Execute(ctx api.ExecutionContext, node api.Node, input interface{}) (interface{}, error) {
    // Implementation here
    return input, nil
}

func (d MyNodeDefinition) Initialize(mel api.Mel) error {
    return nil
}
```

## Common Workflows
1. **Adding new node types**: Create in `pkg/nodes/[type]/` implementing `api.NodeDefinition` interface
2. **API changes**: Update handlers in `internal/api/` and types in `pkg/api/`
3. **Frontend updates**: Components in `web/src/components/`, pages in `web/src/pages/`
4. **Database changes**: Add migrations to `migrations/` directory