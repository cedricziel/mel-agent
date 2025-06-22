# API Client

This directory contains the generated TypeScript API client for MEL Agent.

## Usage

```typescript
import { workflowsApi, Workflow, CreateWorkflowRequest } from './client';

// List all workflows
const workflows = await workflowsApi.listWorkflows({ page: 1, limit: 20 });

// Create a new workflow
const newWorkflow: CreateWorkflowRequest = {
  name: 'My Workflow',
  description: 'Workflow description',
};
const workflow = await workflowsApi.createWorkflow(newWorkflow);

// Get a specific workflow
const workflow = await workflowsApi.getWorkflow('workflow-id');
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
