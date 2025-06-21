# Trigger


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**id** | **string** |  | [optional] [default to undefined]
**name** | **string** |  | [optional] [default to undefined]
**type** | [**TriggerType**](TriggerType.md) |  | [optional] [default to undefined]
**workflow_id** | **string** |  | [optional] [default to undefined]
**config** | **{ [key: string]: any; }** | Trigger configuration containing trigger-specific parameters and settings | [optional] [default to undefined]
**enabled** | **boolean** |  | [optional] [default to undefined]
**created_at** | **string** |  | [optional] [default to undefined]
**updated_at** | **string** |  | [optional] [default to undefined]

## Example

```typescript
import { Trigger } from '@mel-agent/api-client';

const instance: Trigger = {
    id,
    name,
    type,
    workflow_id,
    config,
    enabled,
    created_at,
    updated_at,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
