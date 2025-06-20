# AssistantApi

All URIs are relative to *http://localhost:8080*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**assistantChat**](#assistantchat) | **POST** /api/assistant/chat | Chat with AI assistant|

# **assistantChat**
> AssistantChatResponse assistantChat(assistantChatRequest)


### Example

```typescript
import {
    AssistantApi,
    Configuration,
    AssistantChatRequest
} from '@mel-agent/api-client';

const configuration = new Configuration();
const apiInstance = new AssistantApi(configuration);

let assistantChatRequest: AssistantChatRequest; //

const { status, data } = await apiInstance.assistantChat(
    assistantChatRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **assistantChatRequest** | **AssistantChatRequest**|  | |


### Return type

**AssistantChatResponse**

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth), [BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Assistant response |  -  |
|**400** | Bad request |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

