// TypeScript types for progressive migration
// You can gradually move types here and import them in JS files

// Re-export generated types for convenience
export type {
  Agent,
  AgentList,
  CreateAgentRequest,
  UpdateAgentRequest,
  Workflow,
  WorkflowList,
  CreateWorkflowRequest,
  UpdateWorkflowRequest,
  NodeType,
  Trigger,
  Connection,
  Credential,
  WorkflowNode,
  WorkflowEdge,
  WorkflowDefinition,
} from '@mel-agent/api-client';

// Additional types for the application
export interface APIResponse<T> {
  data: T;
  status: number;
  statusText: string;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  limit: number;
}

export interface APIError {
  message: string;
  code?: number;
  details?: unknown;
}
