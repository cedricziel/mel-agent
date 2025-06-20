## @mel-agent/api-client@1.0.0

This generator creates TypeScript/JavaScript client that utilizes [axios](https://github.com/axios/axios). The generated Node module can be used in the following environments:

Environment
* Node.js
* Webpack
* Browserify

Language level
* ES5 - you must have a Promises/A+ library installed
* ES6

Module system
* CommonJS
* ES6 module system

It can be used in both TypeScript and JavaScript. In TypeScript, the definition will be automatically resolved via `package.json`. ([Reference](https://www.typescriptlang.org/docs/handbook/declaration-files/consumption.html))

### Building

To build and compile the typescript sources to javascript use:
```
npm install
npm run build
```

### Publishing

First build the package then run `npm publish`

### Consuming

navigate to the folder of your consuming project and run one of the following commands.

_published:_

```
npm install @mel-agent/api-client@1.0.0 --save
```

_unPublished (not recommended):_

```
npm install PATH_TO_GENERATED_PACKAGE --save
```

### Documentation for API Endpoints

All URIs are relative to *http://localhost:8080*

Class | Method | HTTP request | Description
------------ | ------------- | ------------- | -------------
*AgentsApi* | [**createAgent**](docs/AgentsApi.md#createagent) | **POST** /api/agents | Create a new agent
*AgentsApi* | [**createAgentVersion**](docs/AgentsApi.md#createagentversion) | **POST** /api/agents/{agentId}/versions | Create a new agent version
*AgentsApi* | [**deleteAgent**](docs/AgentsApi.md#deleteagent) | **DELETE** /api/agents/{id} | Delete agent
*AgentsApi* | [**deployAgentVersion**](docs/AgentsApi.md#deployagentversion) | **POST** /api/agents/{agentId}/deploy | Deploy a specific agent version
*AgentsApi* | [**executeAgentNode**](docs/AgentsApi.md#executeagentnode) | **POST** /api/agents/{agentId}/nodes/{nodeId}/execute | Execute a single node with provided input
*AgentsApi* | [**getAgent**](docs/AgentsApi.md#getagent) | **GET** /api/agents/{id} | Get agent by ID
*AgentsApi* | [**getAgentDraft**](docs/AgentsApi.md#getagentdraft) | **GET** /api/agents/{agentId}/draft | Get current draft for an agent
*AgentsApi* | [**getLatestAgentVersion**](docs/AgentsApi.md#getlatestagentversion) | **GET** /api/agents/{agentId}/versions/latest | Get latest agent version
*AgentsApi* | [**listAgents**](docs/AgentsApi.md#listagents) | **GET** /api/agents | List all agents
*AgentsApi* | [**testDraftNode**](docs/AgentsApi.md#testdraftnode) | **POST** /api/agents/{agentId}/draft/nodes/{nodeId}/test | Test a single node in draft context
*AgentsApi* | [**updateAgent**](docs/AgentsApi.md#updateagent) | **PUT** /api/agents/{id} | Update agent
*AgentsApi* | [**updateAgentDraft**](docs/AgentsApi.md#updateagentdraft) | **PUT** /api/agents/{agentId}/draft | Update agent draft with auto-persistence
*AssistantApi* | [**assistantChat**](docs/AssistantApi.md#assistantchat) | **POST** /api/agents/{agentId}/assistant/chat | Chat with AI assistant for workflow building
*ConnectionsApi* | [**createConnection**](docs/ConnectionsApi.md#createconnection) | **POST** /api/connections | Create a connection
*ConnectionsApi* | [**deleteConnection**](docs/ConnectionsApi.md#deleteconnection) | **DELETE** /api/connections/{id} | Delete connection
*ConnectionsApi* | [**getConnection**](docs/ConnectionsApi.md#getconnection) | **GET** /api/connections/{id} | Get connection by ID
*ConnectionsApi* | [**listConnections**](docs/ConnectionsApi.md#listconnections) | **GET** /api/connections | List connections
*ConnectionsApi* | [**updateConnection**](docs/ConnectionsApi.md#updateconnection) | **PUT** /api/connections/{id} | Update connection
*CredentialTypesApi* | [**getCredentialTypeSchema**](docs/CredentialTypesApi.md#getcredentialtypeschema) | **GET** /api/credential-types/{type}/schema | Get JSON schema for credential type
*CredentialTypesApi* | [**listCredentialTypes**](docs/CredentialTypesApi.md#listcredentialtypes) | **GET** /api/credential-types | List credential type definitions
*CredentialTypesApi* | [**testCredentials**](docs/CredentialTypesApi.md#testcredentials) | **POST** /api/credential-types/{type}/test | Test credentials for a specific type
*CredentialsApi* | [**listCredentials**](docs/CredentialsApi.md#listcredentials) | **GET** /api/credentials | List credentials for selection in nodes
*IntegrationsApi* | [**listIntegrations**](docs/IntegrationsApi.md#listintegrations) | **GET** /api/integrations | List available integrations
*NodeTypesApi* | [**getNodeParameterOptions**](docs/NodeTypesApi.md#getnodeparameteroptions) | **GET** /api/node-types/{type}/parameters/{parameter}/options | Get dynamic options for node parameters
*NodeTypesApi* | [**listNodeTypes**](docs/NodeTypesApi.md#listnodetypes) | **GET** /api/node-types | List available node types
*SystemApi* | [**getHealth**](docs/SystemApi.md#gethealth) | **GET** /api/health | Health check endpoint
*SystemApi* | [**listExtensions**](docs/SystemApi.md#listextensions) | **GET** /api/extensions | List available extensions and plugins
*TriggersApi* | [**createTrigger**](docs/TriggersApi.md#createtrigger) | **POST** /api/triggers | Create a trigger
*TriggersApi* | [**deleteTrigger**](docs/TriggersApi.md#deletetrigger) | **DELETE** /api/triggers/{id} | Delete trigger
*TriggersApi* | [**getTrigger**](docs/TriggersApi.md#gettrigger) | **GET** /api/triggers/{id} | Get trigger by ID
*TriggersApi* | [**listTriggers**](docs/TriggersApi.md#listtriggers) | **GET** /api/triggers | List triggers
*TriggersApi* | [**updateTrigger**](docs/TriggersApi.md#updatetrigger) | **PUT** /api/triggers/{id} | Update trigger
*WebhooksApi* | [**handleWebhook**](docs/WebhooksApi.md#handlewebhook) | **POST** /webhooks/{token} | Webhook endpoint
*WorkersApi* | [**claimWork**](docs/WorkersApi.md#claimwork) | **POST** /api/workers/{id}/claim-work | Claim work items
*WorkersApi* | [**completeWork**](docs/WorkersApi.md#completework) | **POST** /api/workers/{id}/complete-work/{itemId} | Complete a work item
*WorkersApi* | [**listWorkers**](docs/WorkersApi.md#listworkers) | **GET** /api/workers | List all workers
*WorkersApi* | [**registerWorker**](docs/WorkersApi.md#registerworker) | **POST** /api/workers | Register a new worker
*WorkersApi* | [**unregisterWorker**](docs/WorkersApi.md#unregisterworker) | **DELETE** /api/workers/{id} | Unregister a worker
*WorkersApi* | [**updateWorkerHeartbeat**](docs/WorkersApi.md#updateworkerheartbeat) | **PUT** /api/workers/{id}/heartbeat | Update worker heartbeat
*WorkflowRunsApi* | [**getWorkflowRun**](docs/WorkflowRunsApi.md#getworkflowrun) | **GET** /api/workflow-runs/{id} | Get workflow run details
*WorkflowRunsApi* | [**getWorkflowRunSteps**](docs/WorkflowRunsApi.md#getworkflowrunsteps) | **GET** /api/workflow-runs/{id}/steps | Get workflow run steps
*WorkflowRunsApi* | [**listWorkflowRuns**](docs/WorkflowRunsApi.md#listworkflowruns) | **GET** /api/workflow-runs | List workflow runs
*WorkflowsApi* | [**autoLayoutWorkflow**](docs/WorkflowsApi.md#autolayoutworkflow) | **POST** /api/workflows/{workflowId}/layout | Auto-layout workflow nodes
*WorkflowsApi* | [**createWorkflow**](docs/WorkflowsApi.md#createworkflow) | **POST** /api/workflows | Create a new workflow
*WorkflowsApi* | [**createWorkflowEdge**](docs/WorkflowsApi.md#createworkflowedge) | **POST** /api/workflows/{workflowId}/edges | Create a new edge in workflow
*WorkflowsApi* | [**createWorkflowNode**](docs/WorkflowsApi.md#createworkflownode) | **POST** /api/workflows/{workflowId}/nodes | Create a new node in workflow
*WorkflowsApi* | [**deleteWorkflow**](docs/WorkflowsApi.md#deleteworkflow) | **DELETE** /api/workflows/{id} | Delete workflow
*WorkflowsApi* | [**deleteWorkflowEdge**](docs/WorkflowsApi.md#deleteworkflowedge) | **DELETE** /api/workflows/{workflowId}/edges/{edgeId} | Delete workflow edge
*WorkflowsApi* | [**deleteWorkflowNode**](docs/WorkflowsApi.md#deleteworkflownode) | **DELETE** /api/workflows/{workflowId}/nodes/{nodeId} | Delete workflow node
*WorkflowsApi* | [**executeWorkflow**](docs/WorkflowsApi.md#executeworkflow) | **POST** /api/workflows/{id}/execute | Execute a workflow
*WorkflowsApi* | [**getWorkflow**](docs/WorkflowsApi.md#getworkflow) | **GET** /api/workflows/{id} | Get workflow by ID
*WorkflowsApi* | [**getWorkflowNode**](docs/WorkflowsApi.md#getworkflownode) | **GET** /api/workflows/{workflowId}/nodes/{nodeId} | Get specific workflow node
*WorkflowsApi* | [**listWorkflowEdges**](docs/WorkflowsApi.md#listworkflowedges) | **GET** /api/workflows/{workflowId}/edges | List all edges in a workflow
*WorkflowsApi* | [**listWorkflowNodes**](docs/WorkflowsApi.md#listworkflownodes) | **GET** /api/workflows/{workflowId}/nodes | List all nodes in a workflow
*WorkflowsApi* | [**listWorkflows**](docs/WorkflowsApi.md#listworkflows) | **GET** /api/workflows | List all workflows
*WorkflowsApi* | [**updateWorkflow**](docs/WorkflowsApi.md#updateworkflow) | **PUT** /api/workflows/{id} | Update workflow
*WorkflowsApi* | [**updateWorkflowNode**](docs/WorkflowsApi.md#updateworkflownode) | **PUT** /api/workflows/{workflowId}/nodes/{nodeId} | Update workflow node


### Documentation For Models

 - [Agent](docs/Agent.md)
 - [AgentDeployment](docs/AgentDeployment.md)
 - [AgentDraft](docs/AgentDraft.md)
 - [AgentList](docs/AgentList.md)
 - [AgentVersion](docs/AgentVersion.md)
 - [AssistantChatRequest](docs/AssistantChatRequest.md)
 - [AssistantChatResponse](docs/AssistantChatResponse.md)
 - [ChatChoice](docs/ChatChoice.md)
 - [ChatMessage](docs/ChatMessage.md)
 - [ChatUsage](docs/ChatUsage.md)
 - [ClaimWorkRequest](docs/ClaimWorkRequest.md)
 - [CompleteWorkRequest](docs/CompleteWorkRequest.md)
 - [Connection](docs/Connection.md)
 - [CreateAgentRequest](docs/CreateAgentRequest.md)
 - [CreateAgentVersionRequest](docs/CreateAgentVersionRequest.md)
 - [CreateConnectionRequest](docs/CreateConnectionRequest.md)
 - [CreateTriggerRequest](docs/CreateTriggerRequest.md)
 - [CreateWorkflowEdgeRequest](docs/CreateWorkflowEdgeRequest.md)
 - [CreateWorkflowNodeRequest](docs/CreateWorkflowNodeRequest.md)
 - [CreateWorkflowNodeRequestPosition](docs/CreateWorkflowNodeRequestPosition.md)
 - [CreateWorkflowRequest](docs/CreateWorkflowRequest.md)
 - [Credential](docs/Credential.md)
 - [CredentialTestResult](docs/CredentialTestResult.md)
 - [CredentialType](docs/CredentialType.md)
 - [CredentialTypeSchema](docs/CredentialTypeSchema.md)
 - [DeployAgentVersionRequest](docs/DeployAgentVersionRequest.md)
 - [ExecuteAgentNodeRequest](docs/ExecuteAgentNodeRequest.md)
 - [ExecuteWorkflowRequest](docs/ExecuteWorkflowRequest.md)
 - [Extension](docs/Extension.md)
 - [FunctionCall](docs/FunctionCall.md)
 - [GetHealth200Response](docs/GetHealth200Response.md)
 - [Integration](docs/Integration.md)
 - [ModelError](docs/ModelError.md)
 - [NodeExecutionResult](docs/NodeExecutionResult.md)
 - [NodeInput](docs/NodeInput.md)
 - [NodeOutput](docs/NodeOutput.md)
 - [NodeParameterOptions](docs/NodeParameterOptions.md)
 - [NodeParameterOptionsOptionsInner](docs/NodeParameterOptionsOptionsInner.md)
 - [NodeTestResult](docs/NodeTestResult.md)
 - [NodeType](docs/NodeType.md)
 - [ParamSpec](docs/ParamSpec.md)
 - [RegisterWorkerRequest](docs/RegisterWorkerRequest.md)
 - [TestCredentialsRequest](docs/TestCredentialsRequest.md)
 - [Trigger](docs/Trigger.md)
 - [UpdateAgentDraftRequest](docs/UpdateAgentDraftRequest.md)
 - [UpdateAgentRequest](docs/UpdateAgentRequest.md)
 - [UpdateConnectionRequest](docs/UpdateConnectionRequest.md)
 - [UpdateTriggerRequest](docs/UpdateTriggerRequest.md)
 - [UpdateWorkflowNodeRequest](docs/UpdateWorkflowNodeRequest.md)
 - [UpdateWorkflowRequest](docs/UpdateWorkflowRequest.md)
 - [ValidatorSpec](docs/ValidatorSpec.md)
 - [WorkItem](docs/WorkItem.md)
 - [Worker](docs/Worker.md)
 - [Workflow](docs/Workflow.md)
 - [WorkflowDefinition](docs/WorkflowDefinition.md)
 - [WorkflowEdge](docs/WorkflowEdge.md)
 - [WorkflowExecution](docs/WorkflowExecution.md)
 - [WorkflowLayoutResult](docs/WorkflowLayoutResult.md)
 - [WorkflowLayoutResultNodesInner](docs/WorkflowLayoutResultNodesInner.md)
 - [WorkflowList](docs/WorkflowList.md)
 - [WorkflowNode](docs/WorkflowNode.md)
 - [WorkflowRun](docs/WorkflowRun.md)
 - [WorkflowRunList](docs/WorkflowRunList.md)
 - [WorkflowStep](docs/WorkflowStep.md)


<a id="documentation-for-authorization"></a>
## Documentation For Authorization


Authentication schemes defined for the API:
<a id="BearerAuth"></a>
### BearerAuth

- **Type**: Bearer authentication (JWT)

<a id="ApiKeyAuth"></a>
### ApiKeyAuth

- **Type**: API key
- **API key parameter name**: X-API-Key
- **Location**: HTTP header

