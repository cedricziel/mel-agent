import {
  Configuration,
  WorkflowsApi,
  NodeTypesApi,
  TriggersApi,
  WorkersApi,
  SystemApi,
  CredentialsApi,
  ConnectionsApi,
  IntegrationsApi,
  CredentialTypesApi,
  WorkflowRunsApi,
  WebhooksApi,
  AssistantApi,
} from '@mel-agent/api-client';

const baseURL = import.meta.env.VITE_API_BASE_URL || '';

const configuration = new Configuration({
  basePath: baseURL,
});

export const workflowsApi = new WorkflowsApi(configuration);
export const nodeTypesApi = new NodeTypesApi(configuration);
export const triggersApi = new TriggersApi(configuration);
export const workersApi = new WorkersApi(configuration);
export const systemApi = new SystemApi(configuration);
export const credentialsApi = new CredentialsApi(configuration);
export const connectionsApi = new ConnectionsApi(configuration);
export const integrationsApi = new IntegrationsApi(configuration);
export const credentialTypesApi = new CredentialTypesApi(configuration);
export const workflowRunsApi = new WorkflowRunsApi(configuration);
export const assistantApi = new AssistantApi(configuration);
export const webhooksApi = new WebhooksApi(configuration);

export * from '@mel-agent/api-client';
