# UpdateTriggerRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**name** | **string** |  | [optional] [default to undefined]
**config** | **{ [key: string]: any; }** | Trigger configuration containing trigger-specific parameters and settings | [optional] [default to undefined]
**enabled** | **boolean** |  | [optional] [default to undefined]

## Example

```typescript
import { UpdateTriggerRequest } from '@mel-agent/api-client';

const instance: UpdateTriggerRequest = {
    name,
    config,
    enabled,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
