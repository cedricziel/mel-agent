# ConnectionsApi

All URIs are relative to *http://localhost:8080*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**createConnection**](#createconnection) | **POST** /api/connections | Create a connection|
|[**deleteConnection**](#deleteconnection) | **DELETE** /api/connections/{id} | Delete connection|
|[**getConnection**](#getconnection) | **GET** /api/connections/{id} | Get connection by ID|
|[**listConnections**](#listconnections) | **GET** /api/connections | List connections|
|[**updateConnection**](#updateconnection) | **PUT** /api/connections/{id} | Update connection|

# **createConnection**
> Connection createConnection(createConnectionRequest)


### Example

```typescript
import {
    ConnectionsApi,
    Configuration,
    CreateConnectionRequest
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new ConnectionsApi(configuration);

let createConnectionRequest: CreateConnectionRequest; //

const { status, data } = await apiInstance.createConnection(
    createConnectionRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **createConnectionRequest** | **CreateConnectionRequest**|  | |


### Return type

**Connection**

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth), [BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**201** | Connection created |  -  |
|**400** | Bad request |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **deleteConnection**
> deleteConnection()


### Example

```typescript
import {
    ConnectionsApi,
    Configuration
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new ConnectionsApi(configuration);

let id: string; // (default to undefined)

const { status, data } = await apiInstance.deleteConnection(
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
|**204** | Connection deleted |  -  |
|**404** | Connection not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getConnection**
> Connection getConnection()


### Example

```typescript
import {
    ConnectionsApi,
    Configuration
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new ConnectionsApi(configuration);

let id: string; // (default to undefined)

const { status, data } = await apiInstance.getConnection(
    id
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **id** | [**string**] |  | defaults to undefined|


### Return type

**Connection**

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth), [BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Connection details |  -  |
|**404** | Connection not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listConnections**
> Array<Connection> listConnections()


### Example

```typescript
import {
    ConnectionsApi,
    Configuration
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new ConnectionsApi(configuration);

const { status, data } = await apiInstance.listConnections();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**Array<Connection>**

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth), [BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | List of connections |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updateConnection**
> Connection updateConnection(updateConnectionRequest)


### Example

```typescript
import {
    ConnectionsApi,
    Configuration,
    UpdateConnectionRequest
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new ConnectionsApi(configuration);

let id: string; // (default to undefined)
let updateConnectionRequest: UpdateConnectionRequest; //

const { status, data } = await apiInstance.updateConnection(
    id,
    updateConnectionRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **updateConnectionRequest** | **UpdateConnectionRequest**|  | |
| **id** | [**string**] |  | defaults to undefined|


### Return type

**Connection**

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth), [BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Connection updated |  -  |
|**400** | Bad request |  -  |
|**404** | Connection not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

