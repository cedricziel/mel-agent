# ChatMessage


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**role** | [**ChatMessageRole**](ChatMessageRole.md) |  | [default to undefined]
**content** | **string** | Message content | [default to undefined]
**name** | **string** | Function name (for function role) | [optional] [default to undefined]
**function_call** | [**FunctionCall**](FunctionCall.md) |  | [optional] [default to undefined]

## Example

```typescript
import { ChatMessage } from '@mel-agent/api-client';

const instance: ChatMessage = {
    role,
    content,
    name,
    function_call,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
