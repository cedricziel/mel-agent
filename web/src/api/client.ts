import { Configuration, AgentsApi, WorkflowsApi, NodeTypesApi, TriggersApi, WorkersApi, SystemApi } from '../generated/api';

const baseURL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';

const configuration = new Configuration({
  basePath: baseURL,
});

export const agentsApi = new AgentsApi(configuration);
export const workflowsApi = new WorkflowsApi(configuration);
export const nodeTypesApi = new NodeTypesApi(configuration);
export const triggersApi = new TriggersApi(configuration);
export const workersApi = new WorkersApi(configuration);
export const systemApi = new SystemApi(configuration);

export * from '../generated/api/models';
export * from '../generated/api/api';