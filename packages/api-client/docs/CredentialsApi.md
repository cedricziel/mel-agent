# CredentialsApi

All URIs are relative to *http://localhost:8080*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**listCredentials**](#listcredentials) | **GET** /api/credentials | List credentials for selection in nodes|

# **listCredentials**
> Array<Credential> listCredentials()


### Example

```typescript
import {
    CredentialsApi,
    Configuration
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new CredentialsApi(configuration);

let credentialType: string; //Filter by credential type (optional) (default to undefined)

const { status, data } = await apiInstance.listCredentials(
    credentialType
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **credentialType** | [**string**] | Filter by credential type | (optional) defaults to undefined|


### Return type

**Array<Credential>**

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth), [BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | List of credentials |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

