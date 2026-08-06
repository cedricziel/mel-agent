# WorkersApi

All URIs are relative to *http://localhost:8080*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**claimWork**](#claimwork) | **POST** /api/workers/{id}/claim-work | Claim work items|
|[**completeWork**](#completework) | **POST** /api/workers/{id}/complete-work/{itemId} | Complete a work item|
|[**listWorkers**](#listworkers) | **GET** /api/workers | List all workers|
|[**registerWorker**](#registerworker) | **POST** /api/workers | Register a new worker|
|[**unregisterWorker**](#unregisterworker) | **DELETE** /api/workers/{id} | Unregister a worker|
|[**updateWorkerHeartbeat**](#updateworkerheartbeat) | **PUT** /api/workers/{id}/heartbeat | Update worker heartbeat|

# **claimWork**
> Array<WorkItem> claimWork(claimWorkRequest)


### Example

```typescript
import {
    WorkersApi,
    Configuration,
    ClaimWorkRequest
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new WorkersApi(configuration);

let id: string; // (default to undefined)
let claimWorkRequest: ClaimWorkRequest; //

const { status, data } = await apiInstance.claimWork(
    id,
    claimWorkRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **claimWorkRequest** | **ClaimWorkRequest**|  | |
| **id** | [**string**] |  | defaults to undefined|


### Return type

**Array<WorkItem>**

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth), [BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Work items claimed |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **completeWork**
> completeWork(completeWorkRequest)


### Example

```typescript
import {
    WorkersApi,
    Configuration,
    CompleteWorkRequest
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new WorkersApi(configuration);

let id: string; // (default to undefined)
let itemId: string; // (default to undefined)
let completeWorkRequest: CompleteWorkRequest; //

const { status, data } = await apiInstance.completeWork(
    id,
    itemId,
    completeWorkRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **completeWorkRequest** | **CompleteWorkRequest**|  | |
| **id** | [**string**] |  | defaults to undefined|
| **itemId** | [**string**] |  | defaults to undefined|


### Return type

void (empty response body)

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth), [BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Work item completed |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listWorkers**
> Array<Worker> listWorkers()


### Example

```typescript
import {
    WorkersApi,
    Configuration
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new WorkersApi(configuration);

const { status, data } = await apiInstance.listWorkers();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**Array<Worker>**

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth), [BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | List of workers |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **registerWorker**
> Worker registerWorker(registerWorkerRequest)


### Example

```typescript
import {
    WorkersApi,
    Configuration,
    RegisterWorkerRequest
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new WorkersApi(configuration);

let registerWorkerRequest: RegisterWorkerRequest; //

const { status, data } = await apiInstance.registerWorker(
    registerWorkerRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **registerWorkerRequest** | **RegisterWorkerRequest**|  | |


### Return type

**Worker**

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth), [BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**201** | Worker registered successfully |  -  |
|**400** | Bad request |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **unregisterWorker**
> unregisterWorker()


### Example

```typescript
import {
    WorkersApi,
    Configuration
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new WorkersApi(configuration);

let id: string; // (default to undefined)

const { status, data } = await apiInstance.unregisterWorker(
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
|**204** | Worker unregistered successfully |  -  |
|**404** | Worker not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updateWorkerHeartbeat**
> updateWorkerHeartbeat()


### Example

```typescript
import {
    WorkersApi,
    Configuration
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new WorkersApi(configuration);

let id: string; // (default to undefined)

const { status, data } = await apiInstance.updateWorkerHeartbeat(
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
|**200** | Heartbeat updated successfully |  -  |
|**404** | Worker not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

