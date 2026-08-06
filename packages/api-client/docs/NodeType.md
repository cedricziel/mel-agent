# NodeType


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**id** | **string** |  | [optional] [default to undefined]
**name** | **string** |  | [optional] [default to undefined]
**description** | **string** |  | [optional] [default to undefined]
**kinds** | [**Array&lt;NodeKind&gt;**](NodeKind.md) |  | [optional] [default to undefined]
**inputs** | [**Array&lt;NodeInput&gt;**](NodeInput.md) |  | [optional] [default to undefined]
**outputs** | [**Array&lt;NodeOutput&gt;**](NodeOutput.md) |  | [optional] [default to undefined]

## Example

```typescript
import { NodeType } from '@mel-agent/api-client';

const instance: NodeType = {
    id,
    name,
    description,
    kinds,
    inputs,
    outputs,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
