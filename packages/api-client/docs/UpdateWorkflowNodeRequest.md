# UpdateWorkflowNodeRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**name** | **string** |  | [optional] [default to undefined]
**config** | **{ [key: string]: any; }** | Node configuration containing node-specific parameters and settings | [optional] [default to undefined]
**position** | [**NodePosition**](NodePosition.md) |  | [optional] [default to undefined]

## Example

```typescript
import { UpdateWorkflowNodeRequest } from '@mel-agent/api-client';

const instance: UpdateWorkflowNodeRequest = {
    name,
    config,
    position,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
