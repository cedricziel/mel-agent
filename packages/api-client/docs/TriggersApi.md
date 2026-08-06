# TriggersApi

All URIs are relative to *http://localhost:8080*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**createTrigger**](#createtrigger) | **POST** /api/triggers | Create a trigger|
|[**deleteTrigger**](#deletetrigger) | **DELETE** /api/triggers/{id} | Delete trigger|
|[**getTrigger**](#gettrigger) | **GET** /api/triggers/{id} | Get trigger by ID|
|[**listTriggers**](#listtriggers) | **GET** /api/triggers | List triggers|
|[**updateTrigger**](#updatetrigger) | **PUT** /api/triggers/{id} | Update trigger|

# **createTrigger**
> Trigger createTrigger(createTriggerRequest)


### Example

```typescript
import {
    TriggersApi,
    Configuration,
    CreateTriggerRequest
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new TriggersApi(configuration);

let createTriggerRequest: CreateTriggerRequest; //

const { status, data } = await apiInstance.createTrigger(
    createTriggerRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **createTriggerRequest** | **CreateTriggerRequest**|  | |


### Return type

**Trigger**

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth), [BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**201** | Trigger created |  -  |
|**400** | Bad request |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **deleteTrigger**
> deleteTrigger()


### Example

```typescript
import {
    TriggersApi,
    Configuration
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new TriggersApi(configuration);

let id: string; // (default to undefined)

const { status, data } = await apiInstance.deleteTrigger(
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
|**204** | Trigger deleted |  -  |
|**404** | Trigger not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getTrigger**
> Trigger getTrigger()


### Example

```typescript
import {
    TriggersApi,
    Configuration
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new TriggersApi(configuration);

let id: string; // (default to undefined)

const { status, data } = await apiInstance.getTrigger(
    id
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **id** | [**string**] |  | defaults to undefined|


### Return type

**Trigger**

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth), [BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Trigger details |  -  |
|**404** | Trigger not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listTriggers**
> Array<Trigger> listTriggers()


### Example

```typescript
import {
    TriggersApi,
    Configuration
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new TriggersApi(configuration);

const { status, data } = await apiInstance.listTriggers();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**Array<Trigger>**

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth), [BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | List of triggers |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updateTrigger**
> Trigger updateTrigger(updateTriggerRequest)


### Example

```typescript
import {
    TriggersApi,
    Configuration,
    UpdateTriggerRequest
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new TriggersApi(configuration);

let id: string; // (default to undefined)
let updateTriggerRequest: UpdateTriggerRequest; //

const { status, data } = await apiInstance.updateTrigger(
    id,
    updateTriggerRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **updateTriggerRequest** | **UpdateTriggerRequest**|  | |
| **id** | [**string**] |  | defaults to undefined|


### Return type

**Trigger**

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth), [BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Trigger updated |  -  |
|**400** | Bad request |  -  |
|**404** | Trigger not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

