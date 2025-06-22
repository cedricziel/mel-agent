# AssistantChatRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**messages** | [**Array&lt;ChatMessage&gt;**](ChatMessage.md) | Array of chat messages | [default to undefined]
**workflow_id** | **string** | Optional workflow ID for context-aware assistance | [optional] [default to undefined]

## Example

```typescript
import { AssistantChatRequest } from '@mel-agent/api-client';

const instance: AssistantChatRequest = {
    messages,
    workflow_id,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
