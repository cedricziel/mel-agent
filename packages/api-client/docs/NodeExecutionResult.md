# NodeExecutionResult


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**success** | **boolean** |  | [optional] [default to undefined]
**output** | **{ [key: string]: any; }** |  | [optional] [default to undefined]
**error** | **string** |  | [optional] [default to undefined]
**execution_time** | **number** |  | [optional] [default to undefined]
**logs** | **Array&lt;string&gt;** |  | [optional] [default to undefined]
**node_id** | **string** |  | [optional] [default to undefined]
**agent_id** | **string** |  | [optional] [default to undefined]

## Example

```typescript
import { NodeExecutionResult } from '@mel-agent/api-client';

const instance: NodeExecutionResult = {
    success,
    output,
    error,
    execution_time,
    logs,
    node_id,
    agent_id,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
