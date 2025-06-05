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
- Workflow communication (workflow_call, workflow_return, workflow_trigger)
- Triggers (manual_trigger, schedule, webhook)
- Utility nodes (log, delay, transform, variable_get, variable_set, variable_list)

## API Endpoints

### Agents/Workflows
- `GET /api/agents` - List agents
- `POST /api/agents` - Create agent
- `POST /api/agents/{agentID}/versions` - Create agent version
- `GET /api/agents/{agentID}/versions/latest` - Get latest agent version
- `POST /api/agents/{agentID}/runs` - Create run
- `POST /api/agents/{agentID}/runs/test` - Test run agent
- `GET /api/agents/{agentID}/runs` - List runs
- `GET /api/agents/{agentID}/runs/{runID}` - Get specific run

### Connections & Credentials
- `GET /api/connections` - List connections
- `POST /api/connections` - Create connection
- `GET /api/connections/{connectionID}` - Get connection
- `PUT /api/connections/{connectionID}` - Update connection
- `DELETE /api/connections/{connectionID}` - Delete connection
- `GET /api/credentials` - List credentials for selection

### Workflow Builder
- `GET /api/workflows/{workflowID}/nodes` - List workflow nodes
- `POST /api/workflows/{workflowID}/nodes` - Create workflow node
- `GET /api/workflows/{workflowID}/edges` - List workflow edges
- `POST /api/workflows/{workflowID}/edges` - Create workflow edge
- `POST /api/workflows/{workflowID}/layout` - Auto-layout workflow

### Meta & Schema
- `GET /api/node-types` - List node type definitions
- `GET /api/node-types/schema/{type}` - Get node schema
- `GET /api/credential-types` - List credential types
- `GET /api/integrations` - List integrations

### Real-time
- `GET /api/ws/agents/{agentID}` - WebSocket for collaborative updates

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
- `ExecutionContext` - Execution context with agent ID, variables, and platform utilities (Mel)
- `Node` - Represents a workflow node with configuration
- `NodeType` - Metadata for node type (label, category, parameters)
- `ParameterDefinition` - Defines node configuration parameters
- `Mel` - Platform utilities interface providing HTTP client, workflow calling, and data storage

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

func (d MyNodeDefinition) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[interface{}]) (*api.Envelope[interface{}], error) {
    // Implementation here - process envelope and return result
    result := envelope.Clone()
    result.Trace = envelope.Trace.Next(node.ID)
    result.Data = map[string]interface{}{"processed": true}
    return result, nil
}

func (d MyNodeDefinition) Initialize(mel api.Mel) error {
    return nil
}
```

## Platform Utilities (Mel Interface)

Nodes have access to platform utilities through the `ctx.Mel` interface:

### HTTP Client
```go
// Make HTTP requests from nodes
httpReq := api.HTTPRequest{
    Method:  "POST",
    URL:     "https://api.example.com/data",
    Headers: map[string]string{"Authorization": "Bearer token"},
    Body:    strings.NewReader(`{"data": "value"}`), // requires import "strings"
    Timeout: 30 * time.Second,
}

response, err := ctx.Mel.HTTPRequest(context.Background(), httpReq)
```

### Workflow Communication
```go
// Call another workflow synchronously
req := api.WorkflowCallRequest{
    TargetWorkflowID: "target-workflow-id",
    CallData:         map[string]interface{}{"input": "data"},
    CallMode:         "sync", // or "async"
    TimeoutSeconds:   30,
    SourceContext:    ctx,
}

response, err := ctx.Mel.CallWorkflow(context.Background(), req)

// Return data to calling workflow (from workflow_return node)
err := ctx.Mel.ReturnToWorkflow(context.Background(), callID, returnData, "success")
```

### Data Storage
```go
// Store data for cross-workflow communication
err := ctx.Mel.StoreData(context.Background(), "my-key", data, 1*time.Hour)

// Retrieve data
data, err := ctx.Mel.RetrieveData(context.Background(), "my-key")

// Delete data
err := ctx.Mel.DeleteData(context.Background(), "my-key")
```

## Envelope-Based Architecture

All node execution uses an envelope-based data flow system:

```go
type Envelope[T any] struct {
    ID        string                 `json:"id"`
    IssuedAt  time.Time             `json:"issuedAt"`
    Version   int                   `json:"version"`
    DataType  string                `json:"dataType"`
    Data      T                     `json:"data"`
    Trace     Trace                 `json:"trace"`
    Variables map[string]interface{} `json:"variables,omitempty"`
}
```

### Key Concepts:
- **Immutable Flow**: Envelopes are cloned and modified, never mutated
- **Tracing**: Each envelope carries execution trace information
- **Type Safety**: Generic envelope supports any data type
- **Context Preservation**: Variables and trace information flow through nodes

## Common Workflows
1. **Adding new node types**: Create in `pkg/nodes/[type]/` implementing `api.NodeDefinition` interface with `ExecuteEnvelope` method
2. **API changes**: Update handlers in `internal/api/` and types in `pkg/api/`
3. **Frontend updates**: Components in `web/src/components/`, pages in `web/src/pages/`
4. **Database changes**: Add migrations to `migrations/` directory
5. **Workflow communication**: Use `workflow_call`, `workflow_return`, and `workflow_trigger` nodes for inter-workflow communication