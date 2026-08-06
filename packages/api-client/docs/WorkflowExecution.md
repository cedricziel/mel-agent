# WorkflowExecution


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**id** | **string** |  | [optional] [default to undefined]
**workflow_id** | **string** |  | [optional] [default to undefined]
**status** | [**WorkflowRunStatus**](WorkflowRunStatus.md) |  | [optional] [default to undefined]
**started_at** | **string** |  | [optional] [default to undefined]
**completed_at** | **string** |  | [optional] [default to undefined]
**result** | **{ [key: string]: any; }** |  | [optional] [default to undefined]
**error** | **string** |  | [optional] [default to undefined]

## Example

```typescript
import { WorkflowExecution } from '@mel-agent/api-client';

const instance: WorkflowExecution = {
    id,
    workflow_id,
    status,
    started_at,
    completed_at,
    result,
    error,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
