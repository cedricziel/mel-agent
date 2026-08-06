# AssistantChatResponse


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**id** | **string** | Response ID | [optional] [default to undefined]
**object** | **string** | Object type (e.g., \&quot;chat.completion\&quot;) | [optional] [default to undefined]
**created** | **number** | Unix timestamp | [optional] [default to undefined]
**model** | **string** | Model used | [optional] [default to undefined]
**choices** | [**Array&lt;ChatChoice&gt;**](ChatChoice.md) | Array of response choices | [optional] [default to undefined]
**usage** | [**ChatUsage**](ChatUsage.md) |  | [optional] [default to undefined]

## Example

```typescript
import { AssistantChatResponse } from '@mel-agent/api-client';

const instance: AssistantChatResponse = {
    id,
    object,
    created,
    model,
    choices,
    usage,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
