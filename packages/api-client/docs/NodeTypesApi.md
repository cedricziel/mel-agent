# NodeTypesApi

All URIs are relative to *http://localhost:8080*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**getNodeParameterOptions**](#getnodeparameteroptions) | **GET** /api/node-types/{type}/parameters/{parameter}/options | Get dynamic options for node parameters|
|[**listNodeTypes**](#listnodetypes) | **GET** /api/node-types | List available node types|

# **getNodeParameterOptions**
> NodeParameterOptions getNodeParameterOptions()


### Example

```typescript
import {
    NodeTypesApi,
    Configuration
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new NodeTypesApi(configuration);

let type: string; // (default to undefined)
let parameter: string; // (default to undefined)
let context: string; //Context for dynamic option generation (optional) (default to undefined)

const { status, data } = await apiInstance.getNodeParameterOptions(
    type,
    parameter,
    context
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **type** | [**string**] |  | defaults to undefined|
| **parameter** | [**string**] |  | defaults to undefined|
| **context** | [**string**] | Context for dynamic option generation | (optional) defaults to undefined|


### Return type

**NodeParameterOptions**

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth), [BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Dynamic parameter options |  -  |
|**404** | Node type or parameter not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listNodeTypes**
> Array<NodeType> listNodeTypes()


### Example

```typescript
import {
    NodeTypesApi,
    Configuration
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new NodeTypesApi(configuration);

let kind: string; //Filter by node kind (can be comma-separated) (optional) (default to undefined)

const { status, data } = await apiInstance.listNodeTypes(
    kind
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **kind** | [**string**] | Filter by node kind (can be comma-separated) | (optional) defaults to undefined|


### Return type

**Array<NodeType>**

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth), [BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | List of node types |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

