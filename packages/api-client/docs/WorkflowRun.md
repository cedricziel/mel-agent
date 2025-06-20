# WorkflowRun


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**id** | **string** |  | [optional] [default to undefined]
**workflow_id** | **string** |  | [optional] [default to undefined]
**status** | **string** |  | [optional] [default to undefined]
**started_at** | **string** |  | [optional] [default to undefined]
**completed_at** | **string** |  | [optional] [default to undefined]
**context** | **{ [key: string]: any; }** |  | [optional] [default to undefined]
**error** | **string** |  | [optional] [default to undefined]

## Example

```typescript
import { WorkflowRun } from '@mel-agent/api-client';

const instance: WorkflowRun = {
    id,
    workflow_id,
    status,
    started_at,
    completed_at,
    context,
    error,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
