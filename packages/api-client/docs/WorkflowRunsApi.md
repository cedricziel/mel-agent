# WorkflowRunsApi

All URIs are relative to *http://localhost:8080*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**getWorkflowRun**](#getworkflowrun) | **GET** /api/workflow-runs/{id} | Get workflow run details|
|[**getWorkflowRunSteps**](#getworkflowrunsteps) | **GET** /api/workflow-runs/{id}/steps | Get workflow run steps|
|[**listWorkflowRuns**](#listworkflowruns) | **GET** /api/workflow-runs | List workflow runs|

# **getWorkflowRun**
> WorkflowRun getWorkflowRun()


### Example

```typescript
import {
    WorkflowRunsApi,
    Configuration
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new WorkflowRunsApi(configuration);

let id: string; // (default to undefined)

const { status, data } = await apiInstance.getWorkflowRun(
    id
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **id** | [**string**] |  | defaults to undefined|


### Return type

**WorkflowRun**

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth), [BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Workflow run details |  -  |
|**404** | Workflow run not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getWorkflowRunSteps**
> Array<WorkflowStep> getWorkflowRunSteps()


### Example

```typescript
import {
    WorkflowRunsApi,
    Configuration
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new WorkflowRunsApi(configuration);

let id: string; // (default to undefined)

const { status, data } = await apiInstance.getWorkflowRunSteps(
    id
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **id** | [**string**] |  | defaults to undefined|


### Return type

**Array<WorkflowStep>**

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth), [BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | List of workflow steps |  -  |
|**404** | Workflow run not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listWorkflowRuns**
> WorkflowRunList listWorkflowRuns()


### Example

```typescript
import {
    WorkflowRunsApi,
    Configuration
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new WorkflowRunsApi(configuration);

let workflowId: string; // (optional) (default to undefined)
let status: 'pending' | 'running' | 'completed' | 'failed'; // (optional) (default to undefined)
let page: number; // (optional) (default to 1)
let limit: number; // (optional) (default to 20)

const { status, data } = await apiInstance.listWorkflowRuns(
    workflowId,
    status,
    page,
    limit
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **workflowId** | [**string**] |  | (optional) defaults to undefined|
| **status** | [**&#39;pending&#39; | &#39;running&#39; | &#39;completed&#39; | &#39;failed&#39;**]**Array<&#39;pending&#39; &#124; &#39;running&#39; &#124; &#39;completed&#39; &#124; &#39;failed&#39;>** |  | (optional) defaults to undefined|
| **page** | [**number**] |  | (optional) defaults to 1|
| **limit** | [**number**] |  | (optional) defaults to 20|


### Return type

**WorkflowRunList**

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth), [BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | List of workflow runs |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

