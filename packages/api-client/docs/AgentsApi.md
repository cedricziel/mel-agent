# AgentsApi

All URIs are relative to *http://localhost:8080*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**createAgent**](#createagent) | **POST** /api/agents | Create a new agent|
|[**createAgentVersion**](#createagentversion) | **POST** /api/agents/{agentId}/versions | Create a new agent version|
|[**deleteAgent**](#deleteagent) | **DELETE** /api/agents/{id} | Delete agent|
|[**deployAgentVersion**](#deployagentversion) | **POST** /api/agents/{agentId}/deploy | Deploy a specific agent version|
|[**executeAgentNode**](#executeagentnode) | **POST** /api/agents/{agentId}/nodes/{nodeId}/execute | Execute a single node with provided input|
|[**getAgent**](#getagent) | **GET** /api/agents/{id} | Get agent by ID|
|[**getAgentDraft**](#getagentdraft) | **GET** /api/agents/{agentId}/draft | Get current draft for an agent|
|[**getLatestAgentVersion**](#getlatestagentversion) | **GET** /api/agents/{agentId}/versions/latest | Get latest agent version|
|[**listAgents**](#listagents) | **GET** /api/agents | List all agents|
|[**testDraftNode**](#testdraftnode) | **POST** /api/agents/{agentId}/draft/nodes/{nodeId}/test | Test a single node in draft context|
|[**updateAgent**](#updateagent) | **PUT** /api/agents/{id} | Update agent|
|[**updateAgentDraft**](#updateagentdraft) | **PUT** /api/agents/{agentId}/draft | Update agent draft with auto-persistence|

# **createAgent**
> Agent createAgent(createAgentRequest)


### Example

```typescript
import {
    AgentsApi,
    Configuration,
    CreateAgentRequest
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new AgentsApi(configuration);

let createAgentRequest: CreateAgentRequest; //

const { status, data } = await apiInstance.createAgent(
    createAgentRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **createAgentRequest** | **CreateAgentRequest**|  | |


### Return type

**Agent**

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth), [BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**201** | Agent created successfully |  -  |
|**400** | Bad request |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **createAgentVersion**
> AgentVersion createAgentVersion(createAgentVersionRequest)


### Example

```typescript
import {
    AgentsApi,
    Configuration,
    CreateAgentVersionRequest
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new AgentsApi(configuration);

let agentId: string; // (default to undefined)
let createAgentVersionRequest: CreateAgentVersionRequest; //

const { status, data } = await apiInstance.createAgentVersion(
    agentId,
    createAgentVersionRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **createAgentVersionRequest** | **CreateAgentVersionRequest**|  | |
| **agentId** | [**string**] |  | defaults to undefined|


### Return type

**AgentVersion**

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth), [BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**201** | Agent version created successfully |  -  |
|**400** | Bad request |  -  |
|**404** | Agent not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **deleteAgent**
> deleteAgent()


### Example

```typescript
import {
    AgentsApi,
    Configuration
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new AgentsApi(configuration);

let id: string; // (default to undefined)

const { status, data } = await apiInstance.deleteAgent(
    id
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **id** | [**string**] |  | defaults to undefined|


### Return type

void (empty response body)

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth), [BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**204** | Agent deleted successfully |  -  |
|**404** | Agent not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **deployAgentVersion**
> AgentDeployment deployAgentVersion(deployAgentVersionRequest)


### Example

```typescript
import {
    AgentsApi,
    Configuration,
    DeployAgentVersionRequest
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new AgentsApi(configuration);

let agentId: string; // (default to undefined)
let deployAgentVersionRequest: DeployAgentVersionRequest; //

const { status, data } = await apiInstance.deployAgentVersion(
    agentId,
    deployAgentVersionRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **deployAgentVersionRequest** | **DeployAgentVersionRequest**|  | |
| **agentId** | [**string**] |  | defaults to undefined|


### Return type

**AgentDeployment**

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth), [BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Agent version deployed successfully |  -  |
|**400** | Bad request |  -  |
|**404** | Agent not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **executeAgentNode**
> NodeExecutionResult executeAgentNode(executeAgentNodeRequest)


### Example

```typescript
import {
    AgentsApi,
    Configuration,
    ExecuteAgentNodeRequest
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new AgentsApi(configuration);

let agentId: string; // (default to undefined)
let nodeId: string; // (default to undefined)
let executeAgentNodeRequest: ExecuteAgentNodeRequest; //

const { status, data } = await apiInstance.executeAgentNode(
    agentId,
    nodeId,
    executeAgentNodeRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **executeAgentNodeRequest** | **ExecuteAgentNodeRequest**|  | |
| **agentId** | [**string**] |  | defaults to undefined|
| **nodeId** | [**string**] |  | defaults to undefined|


### Return type

**NodeExecutionResult**

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth), [BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Node execution result |  -  |
|**400** | Invalid input or execution failed |  -  |
|**404** | Agent or node not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getAgent**
> Agent getAgent()


### Example

```typescript
import {
    AgentsApi,
    Configuration
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new AgentsApi(configuration);

let id: string; // (default to undefined)

const { status, data } = await apiInstance.getAgent(
    id
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **id** | [**string**] |  | defaults to undefined|


### Return type

**Agent**

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth), [BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Agent details |  -  |
|**404** | Agent not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getAgentDraft**
> AgentDraft getAgentDraft()


### Example

```typescript
import {
    AgentsApi,
    Configuration
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new AgentsApi(configuration);

let agentId: string; // (default to undefined)

const { status, data } = await apiInstance.getAgentDraft(
    agentId
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **agentId** | [**string**] |  | defaults to undefined|


### Return type

**AgentDraft**

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth), [BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Agent draft |  -  |
|**404** | Agent not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getLatestAgentVersion**
> AgentVersion getLatestAgentVersion()


### Example

```typescript
import {
    AgentsApi,
    Configuration
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new AgentsApi(configuration);

let agentId: string; // (default to undefined)

const { status, data } = await apiInstance.getLatestAgentVersion(
    agentId
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **agentId** | [**string**] |  | defaults to undefined|


### Return type

**AgentVersion**

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth), [BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Latest agent version |  -  |
|**404** | Agent not found or no versions exist |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listAgents**
> AgentList listAgents()


### Example

```typescript
import {
    AgentsApi,
    Configuration
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new AgentsApi(configuration);

let page: number; // (optional) (default to 1)
let limit: number; // (optional) (default to 20)

const { status, data } = await apiInstance.listAgents(
    page,
    limit
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **page** | [**number**] |  | (optional) defaults to 1|
| **limit** | [**number**] |  | (optional) defaults to 20|


### Return type

**AgentList**

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth), [BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | List of agents |  -  |
|**400** | Invalid query parameters |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **testDraftNode**
> NodeTestResult testDraftNode(executeWorkflowRequest)


### Example

```typescript
import {
    AgentsApi,
    Configuration,
    ExecuteWorkflowRequest
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new AgentsApi(configuration);

let agentId: string; // (default to undefined)
let nodeId: string; // (default to undefined)
let executeWorkflowRequest: ExecuteWorkflowRequest; //

const { status, data } = await apiInstance.testDraftNode(
    agentId,
    nodeId,
    executeWorkflowRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **executeWorkflowRequest** | **ExecuteWorkflowRequest**|  | |
| **agentId** | [**string**] |  | defaults to undefined|
| **nodeId** | [**string**] |  | defaults to undefined|


### Return type

**NodeTestResult**

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth), [BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Node test result |  -  |
|**400** | Bad request |  -  |
|**404** | Agent or node not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updateAgent**
> Agent updateAgent(updateAgentRequest)


### Example

```typescript
import {
    AgentsApi,
    Configuration,
    UpdateAgentRequest
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new AgentsApi(configuration);

let id: string; // (default to undefined)
let updateAgentRequest: UpdateAgentRequest; //

const { status, data } = await apiInstance.updateAgent(
    id,
    updateAgentRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **updateAgentRequest** | **UpdateAgentRequest**|  | |
| **id** | [**string**] |  | defaults to undefined|


### Return type

**Agent**

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth), [BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Agent updated successfully |  -  |
|**400** | Bad request |  -  |
|**404** | Agent not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updateAgentDraft**
> AgentDraft updateAgentDraft(updateAgentDraftRequest)


### Example

```typescript
import {
    AgentsApi,
    Configuration,
    UpdateAgentDraftRequest
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new AgentsApi(configuration);

let agentId: string; // (default to undefined)
let updateAgentDraftRequest: UpdateAgentDraftRequest; //

const { status, data } = await apiInstance.updateAgentDraft(
    agentId,
    updateAgentDraftRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **updateAgentDraftRequest** | **UpdateAgentDraftRequest**|  | |
| **agentId** | [**string**] |  | defaults to undefined|


### Return type

**AgentDraft**

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth), [BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Draft updated successfully |  -  |
|**400** | Bad request |  -  |
|**404** | Agent not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

