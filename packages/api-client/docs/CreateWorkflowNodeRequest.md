# CreateWorkflowNodeRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**id** | **string** |  | [default to undefined]
**name** | **string** |  | [default to undefined]
**type** | **string** |  | [default to undefined]
**config** | **{ [key: string]: any; }** | Node configuration containing node-specific parameters and settings | [default to undefined]
**position** | [**NodePosition**](NodePosition.md) |  | [optional] [default to undefined]

## Example

```typescript
import { CreateWorkflowNodeRequest } from '@mel-agent/api-client';

const instance: CreateWorkflowNodeRequest = {
    id,
    name,
    type,
    config,
    position,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
