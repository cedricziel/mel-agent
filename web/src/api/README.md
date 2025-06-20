# API Client

This directory contains the generated TypeScript API client for MEL Agent.

## Usage

```typescript
import { agentsApi, workflowsApi, Agent, CreateAgentRequest } from './client';

// List all agents
const agents = await agentsApi.listAgents({ page: 1, limit: 20 });

// Create a new agent
const newAgent: CreateAgentRequest = {
  name: 'My Agent',
  description: 'Agent description'
};
const agent = await agentsApi.createAgent(newAgent);

// Get a specific agent
const agent = await agentsApi.getAgent('agent-id');
```

## Generated Code

The client is automatically generated from the OpenAPI specification at `../api/openapi.yaml`.

To regenerate the client:
```bash
pnpm generate:api:clean
```

## Configuration

The client is configured to use the backend URL from environment variables:
- `VITE_API_BASE_URL` (defaults to `http://localhost:8080`)

## License

This generated code is licensed under AGPL-3.0.