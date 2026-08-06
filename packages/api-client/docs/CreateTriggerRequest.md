# CreateTriggerRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**name** | **string** |  | [default to undefined]
**type** | [**TriggerType**](TriggerType.md) |  | [default to undefined]
**workflow_id** | **string** |  | [default to undefined]
**config** | **{ [key: string]: any; }** | Trigger configuration containing trigger-specific parameters and settings | [optional] [default to undefined]
**enabled** | **boolean** |  | [optional] [default to true]

## Example

```typescript
import { CreateTriggerRequest } from '@mel-agent/api-client';

const instance: CreateTriggerRequest = {
    name,
    type,
    workflow_id,
    config,
    enabled,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
