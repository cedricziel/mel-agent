# WorkflowStep


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**id** | **string** |  | [optional] [default to undefined]
**run_id** | **string** |  | [optional] [default to undefined]
**node_id** | **string** |  | [optional] [default to undefined]
**status** | **string** |  | [optional] [default to undefined]
**started_at** | **string** |  | [optional] [default to undefined]
**completed_at** | **string** |  | [optional] [default to undefined]
**input** | **{ [key: string]: any; }** |  | [optional] [default to undefined]
**output** | **{ [key: string]: any; }** |  | [optional] [default to undefined]
**error** | **string** |  | [optional] [default to undefined]

## Example

```typescript
import { WorkflowStep } from '@mel-agent/api-client';

const instance: WorkflowStep = {
    id,
    run_id,
    node_id,
    status,
    started_at,
    completed_at,
    input,
    output,
    error,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
