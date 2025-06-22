# Worker


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**id** | **string** |  | [optional] [default to undefined]
**name** | **string** |  | [optional] [default to undefined]
**status** | [**WorkerStatus**](WorkerStatus.md) |  | [optional] [default to undefined]
**last_heartbeat** | **string** |  | [optional] [default to undefined]
**concurrency** | **number** |  | [optional] [default to undefined]
**registered_at** | **string** |  | [optional] [default to undefined]

## Example

```typescript
import { Worker } from '@mel-agent/api-client';

const instance: Worker = {
    id,
    name,
    status,
    last_heartbeat,
    concurrency,
    registered_at,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
