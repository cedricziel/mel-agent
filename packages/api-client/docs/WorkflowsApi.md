# WorkflowsApi

All URIs are relative to *http://localhost:8080*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**autoLayoutWorkflow**](#autolayoutworkflow) | **POST** /api/workflows/{workflowId}/layout | Auto-layout workflow nodes|
|[**createWorkflow**](#createworkflow) | **POST** /api/workflows | Create a new workflow|
|[**createWorkflowEdge**](#createworkflowedge) | **POST** /api/workflows/{workflowId}/edges | Create a new edge in workflow|
|[**createWorkflowNode**](#createworkflownode) | **POST** /api/workflows/{workflowId}/nodes | Create a new node in workflow|
|[**createWorkflowVersion**](#createworkflowversion) | **POST** /api/workflows/{workflowId}/versions | Create a new workflow version|
|[**deleteWorkflow**](#deleteworkflow) | **DELETE** /api/workflows/{id} | Delete workflow|
|[**deleteWorkflowEdge**](#deleteworkflowedge) | **DELETE** /api/workflows/{workflowId}/edges/{edgeId} | Delete workflow edge|
|[**deleteWorkflowNode**](#deleteworkflownode) | **DELETE** /api/workflows/{workflowId}/nodes/{nodeId} | Delete workflow node|
|[**deployWorkflowVersion**](#deployworkflowversion) | **POST** /api/workflows/{workflowId}/versions/{versionNumber}/deploy | Deploy a specific workflow version|
|[**executeWorkflow**](#executeworkflow) | **POST** /api/workflows/{id}/execute | Execute a workflow|
|[**getLatestWorkflowVersion**](#getlatestworkflowversion) | **GET** /api/workflows/{workflowId}/versions/latest | Get latest version of a workflow|
|[**getWorkflow**](#getworkflow) | **GET** /api/workflows/{id} | Get workflow by ID|
|[**getWorkflowDraft**](#getworkflowdraft) | **GET** /api/workflows/{workflowId}/draft | Get current draft for a workflow|
|[**getWorkflowNode**](#getworkflownode) | **GET** /api/workflows/{workflowId}/nodes/{nodeId} | Get specific workflow node|
|[**listWorkflowEdges**](#listworkflowedges) | **GET** /api/workflows/{workflowId}/edges | List all edges in a workflow|
|[**listWorkflowNodes**](#listworkflownodes) | **GET** /api/workflows/{workflowId}/nodes | List all nodes in a workflow|
|[**listWorkflowVersions**](#listworkflowversions) | **GET** /api/workflows/{workflowId}/versions | List all versions of a workflow|
|[**listWorkflows**](#listworkflows) | **GET** /api/workflows | List all workflows|
|[**testWorkflowDraftNode**](#testworkflowdraftnode) | **POST** /api/workflows/{workflowId}/draft/nodes/{nodeId}/test | Test a single node in draft context|
|[**updateWorkflow**](#updateworkflow) | **PUT** /api/workflows/{id} | Update workflow|
|[**updateWorkflowDraft**](#updateworkflowdraft) | **PUT** /api/workflows/{workflowId}/draft | Update workflow draft with auto-persistence|
|[**updateWorkflowNode**](#updateworkflownode) | **PUT** /api/workflows/{workflowId}/nodes/{nodeId} | Update workflow node|

# **autoLayoutWorkflow**
> WorkflowLayoutResult autoLayoutWorkflow()


### Example

```typescript
import {
    WorkflowsApi,
    Configuration
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new WorkflowsApi(configuration);

let workflowId: string; // (default to undefined)

const { status, data } = await apiInstance.autoLayoutWorkflow(
    workflowId
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **workflowId** | [**string**] |  | defaults to undefined|


### Return type

**WorkflowLayoutResult**

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth), [BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Layout updated successfully |  -  |
|**404** | Workflow not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **createWorkflow**
> Workflow createWorkflow(createWorkflowRequest)


### Example

```typescript
import {
    WorkflowsApi,
    Configuration,
    CreateWorkflowRequest
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new WorkflowsApi(configuration);

let createWorkflowRequest: CreateWorkflowRequest; //

const { status, data } = await apiInstance.createWorkflow(
    createWorkflowRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **createWorkflowRequest** | **CreateWorkflowRequest**|  | |


### Return type

**Workflow**

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth), [BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**201** | Workflow created successfully |  -  |
|**400** | Bad request |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **createWorkflowEdge**
> WorkflowEdge createWorkflowEdge(createWorkflowEdgeRequest)


### Example

```typescript
import {
    WorkflowsApi,
    Configuration,
    CreateWorkflowEdgeRequest
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new WorkflowsApi(configuration);

let workflowId: string; // (default to undefined)
let createWorkflowEdgeRequest: CreateWorkflowEdgeRequest; //

const { status, data } = await apiInstance.createWorkflowEdge(
    workflowId,
    createWorkflowEdgeRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **createWorkflowEdgeRequest** | **CreateWorkflowEdgeRequest**|  | |
| **workflowId** | [**string**] |  | defaults to undefined|


### Return type

**WorkflowEdge**

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth), [BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**201** | Workflow edge created successfully |  -  |
|**400** | Bad request |  -  |
|**404** | Workflow not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **createWorkflowNode**
> WorkflowNode createWorkflowNode(createWorkflowNodeRequest)


### Example

```typescript
import {
    WorkflowsApi,
    Configuration,
    CreateWorkflowNodeRequest
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new WorkflowsApi(configuration);

let workflowId: string; // (default to undefined)
let createWorkflowNodeRequest: CreateWorkflowNodeRequest; //

const { status, data } = await apiInstance.createWorkflowNode(
    workflowId,
    createWorkflowNodeRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **createWorkflowNodeRequest** | **CreateWorkflowNodeRequest**|  | |
| **workflowId** | [**string**] |  | defaults to undefined|


### Return type

**WorkflowNode**

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth), [BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**201** | Workflow node created successfully |  -  |
|**400** | Bad request |  -  |
|**404** | Workflow not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **createWorkflowVersion**
> WorkflowVersion createWorkflowVersion(createWorkflowVersionRequest)


### Example

```typescript
import {
    WorkflowsApi,
    Configuration,
    CreateWorkflowVersionRequest
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new WorkflowsApi(configuration);

let workflowId: string; // (default to undefined)
let createWorkflowVersionRequest: CreateWorkflowVersionRequest; //

const { status, data } = await apiInstance.createWorkflowVersion(
    workflowId,
    createWorkflowVersionRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **createWorkflowVersionRequest** | **CreateWorkflowVersionRequest**|  | |
| **workflowId** | [**string**] |  | defaults to undefined|


### Return type

**WorkflowVersion**

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth), [BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**201** | Version created successfully |  -  |
|**400** | Invalid request body |  -  |
|**404** | Workflow not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **deleteWorkflow**
> deleteWorkflow()


### Example

```typescript
import {
    WorkflowsApi,
    Configuration
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new WorkflowsApi(configuration);

let id: string; // (default to undefined)

const { status, data } = await apiInstance.deleteWorkflow(
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
|**204** | Workflow deleted successfully |  -  |
|**404** | Workflow not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **deleteWorkflowEdge**
> deleteWorkflowEdge()


### Example

```typescript
import {
    WorkflowsApi,
    Configuration
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new WorkflowsApi(configuration);

let workflowId: string; // (default to undefined)
let edgeId: string; // (default to undefined)

const { status, data } = await apiInstance.deleteWorkflowEdge(
    workflowId,
    edgeId
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **workflowId** | [**string**] |  | defaults to undefined|
| **edgeId** | [**string**] |  | defaults to undefined|


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
|**204** | Workflow edge deleted successfully |  -  |
|**404** | Workflow or edge not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **deleteWorkflowNode**
> deleteWorkflowNode()


### Example

```typescript
import {
    WorkflowsApi,
    Configuration
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new WorkflowsApi(configuration);

let workflowId: string; // (default to undefined)
let nodeId: string; // (default to undefined)

const { status, data } = await apiInstance.deleteWorkflowNode(
    workflowId,
    nodeId
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **workflowId** | [**string**] |  | defaults to undefined|
| **nodeId** | [**string**] |  | defaults to undefined|


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
|**204** | Workflow node deleted successfully |  -  |
|**404** | Workflow or node not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **deployWorkflowVersion**
> WorkflowVersion deployWorkflowVersion()


### Example

```typescript
import {
    WorkflowsApi,
    Configuration
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new WorkflowsApi(configuration);

let workflowId: string; // (default to undefined)
let versionNumber: number; // (default to undefined)

const { status, data } = await apiInstance.deployWorkflowVersion(
    workflowId,
    versionNumber
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **workflowId** | [**string**] |  | defaults to undefined|
| **versionNumber** | [**number**] |  | defaults to undefined|


### Return type

**WorkflowVersion**

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth), [BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Version deployed successfully |  -  |
|**404** | Workflow or version not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **executeWorkflow**
> WorkflowExecution executeWorkflow()


### Example

```typescript
import {
    WorkflowsApi,
    Configuration,
    ExecuteWorkflowRequest
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new WorkflowsApi(configuration);

let id: string; // (default to undefined)
let executeWorkflowRequest: ExecuteWorkflowRequest; // (optional)

const { status, data } = await apiInstance.executeWorkflow(
    id,
    executeWorkflowRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **executeWorkflowRequest** | **ExecuteWorkflowRequest**|  | |
| **id** | [**string**] |  | defaults to undefined|


### Return type

**WorkflowExecution**

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth), [BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Workflow execution started |  -  |
|**400** | Bad request |  -  |
|**404** | Workflow not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getLatestWorkflowVersion**
> WorkflowVersion getLatestWorkflowVersion()


### Example

```typescript
import {
    WorkflowsApi,
    Configuration
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new WorkflowsApi(configuration);

let workflowId: string; // (default to undefined)

const { status, data } = await apiInstance.getLatestWorkflowVersion(
    workflowId
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **workflowId** | [**string**] |  | defaults to undefined|


### Return type

**WorkflowVersion**

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth), [BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Latest workflow version |  -  |
|**404** | Workflow not found or no versions exist |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getWorkflow**
> Workflow getWorkflow()


### Example

```typescript
import {
    WorkflowsApi,
    Configuration
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new WorkflowsApi(configuration);

let id: string; // (default to undefined)

const { status, data } = await apiInstance.getWorkflow(
    id
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **id** | [**string**] |  | defaults to undefined|


### Return type

**Workflow**

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth), [BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Workflow details |  -  |
|**404** | Workflow not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getWorkflowDraft**
> WorkflowDraft getWorkflowDraft()


### Example

```typescript
import {
    WorkflowsApi,
    Configuration
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new WorkflowsApi(configuration);

let workflowId: string; // (default to undefined)

const { status, data } = await apiInstance.getWorkflowDraft(
    workflowId
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **workflowId** | [**string**] |  | defaults to undefined|


### Return type

**WorkflowDraft**

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth), [BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Workflow draft |  -  |
|**404** | Workflow not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getWorkflowNode**
> WorkflowNode getWorkflowNode()


### Example

```typescript
import {
    WorkflowsApi,
    Configuration
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new WorkflowsApi(configuration);

let workflowId: string; // (default to undefined)
let nodeId: string; // (default to undefined)

const { status, data } = await apiInstance.getWorkflowNode(
    workflowId,
    nodeId
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **workflowId** | [**string**] |  | defaults to undefined|
| **nodeId** | [**string**] |  | defaults to undefined|


### Return type

**WorkflowNode**

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth), [BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Workflow node details |  -  |
|**404** | Workflow or node not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listWorkflowEdges**
> Array<WorkflowEdge> listWorkflowEdges()


### Example

```typescript
import {
    WorkflowsApi,
    Configuration
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new WorkflowsApi(configuration);

let workflowId: string; // (default to undefined)

const { status, data } = await apiInstance.listWorkflowEdges(
    workflowId
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **workflowId** | [**string**] |  | defaults to undefined|


### Return type

**Array<WorkflowEdge>**

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth), [BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | List of workflow edges |  -  |
|**404** | Workflow not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listWorkflowNodes**
> Array<WorkflowNode> listWorkflowNodes()


### Example

```typescript
import {
    WorkflowsApi,
    Configuration
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new WorkflowsApi(configuration);

let workflowId: string; // (default to undefined)

const { status, data } = await apiInstance.listWorkflowNodes(
    workflowId
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **workflowId** | [**string**] |  | defaults to undefined|


### Return type

**Array<WorkflowNode>**

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth), [BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | List of workflow nodes |  -  |
|**404** | Workflow not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listWorkflowVersions**
> WorkflowVersionList listWorkflowVersions()


### Example

```typescript
import {
    WorkflowsApi,
    Configuration
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new WorkflowsApi(configuration);

let workflowId: string; // (default to undefined)
let page: number; // (optional) (default to 1)
let limit: number; // (optional) (default to 20)

const { status, data } = await apiInstance.listWorkflowVersions(
    workflowId,
    page,
    limit
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **workflowId** | [**string**] |  | defaults to undefined|
| **page** | [**number**] |  | (optional) defaults to 1|
| **limit** | [**number**] |  | (optional) defaults to 20|


### Return type

**WorkflowVersionList**

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth), [BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | List of workflow versions |  -  |
|**404** | Workflow not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listWorkflows**
> WorkflowList listWorkflows()


### Example

```typescript
import {
    WorkflowsApi,
    Configuration
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new WorkflowsApi(configuration);

let page: number; // (optional) (default to 1)
let limit: number; // (optional) (default to 20)

const { status, data } = await apiInstance.listWorkflows(
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

**WorkflowList**

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth), [BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | List of workflows |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **testWorkflowDraftNode**
> NodeTestResult testWorkflowDraftNode(executeWorkflowRequest)


### Example

```typescript
import {
    WorkflowsApi,
    Configuration,
    ExecuteWorkflowRequest
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new WorkflowsApi(configuration);

let workflowId: string; // (default to undefined)
let nodeId: string; // (default to undefined)
let executeWorkflowRequest: ExecuteWorkflowRequest; //

const { status, data } = await apiInstance.testWorkflowDraftNode(
    workflowId,
    nodeId,
    executeWorkflowRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **executeWorkflowRequest** | **ExecuteWorkflowRequest**|  | |
| **workflowId** | [**string**] |  | defaults to undefined|
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
|**400** | Invalid request body |  -  |
|**404** | Workflow or node not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updateWorkflow**
> Workflow updateWorkflow(updateWorkflowRequest)


### Example

```typescript
import {
    WorkflowsApi,
    Configuration,
    UpdateWorkflowRequest
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new WorkflowsApi(configuration);

let id: string; // (default to undefined)
let updateWorkflowRequest: UpdateWorkflowRequest; //

const { status, data } = await apiInstance.updateWorkflow(
    id,
    updateWorkflowRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **updateWorkflowRequest** | **UpdateWorkflowRequest**|  | |
| **id** | [**string**] |  | defaults to undefined|


### Return type

**Workflow**

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth), [BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Workflow updated successfully |  -  |
|**400** | Bad request |  -  |
|**404** | Workflow not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updateWorkflowDraft**
> WorkflowDraft updateWorkflowDraft(updateWorkflowDraftRequest)


### Example

```typescript
import {
    WorkflowsApi,
    Configuration,
    UpdateWorkflowDraftRequest
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new WorkflowsApi(configuration);

let workflowId: string; // (default to undefined)
let updateWorkflowDraftRequest: UpdateWorkflowDraftRequest; //

const { status, data } = await apiInstance.updateWorkflowDraft(
    workflowId,
    updateWorkflowDraftRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **updateWorkflowDraftRequest** | **UpdateWorkflowDraftRequest**|  | |
| **workflowId** | [**string**] |  | defaults to undefined|


### Return type

**WorkflowDraft**

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth), [BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Draft updated successfully |  -  |
|**400** | Invalid request body |  -  |
|**404** | Workflow not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updateWorkflowNode**
> WorkflowNode updateWorkflowNode(updateWorkflowNodeRequest)


### Example

```typescript
import {
    WorkflowsApi,
    Configuration,
    UpdateWorkflowNodeRequest
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new WorkflowsApi(configuration);

let workflowId: string; // (default to undefined)
let nodeId: string; // (default to undefined)
let updateWorkflowNodeRequest: UpdateWorkflowNodeRequest; //

const { status, data } = await apiInstance.updateWorkflowNode(
    workflowId,
    nodeId,
    updateWorkflowNodeRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **updateWorkflowNodeRequest** | **UpdateWorkflowNodeRequest**|  | |
| **workflowId** | [**string**] |  | defaults to undefined|
| **nodeId** | [**string**] |  | defaults to undefined|


### Return type

**WorkflowNode**

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth), [BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Workflow node updated successfully |  -  |
|**400** | Bad request |  -  |
|**404** | Workflow or node not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

