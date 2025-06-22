# NodeTestResult


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**success** | **boolean** |  | [optional] [default to undefined]
**output** | **{ [key: string]: any; }** |  | [optional] [default to undefined]
**error** | **string** |  | [optional] [default to undefined]
**execution_time** | **number** |  | [optional] [default to undefined]
**logs** | **Array&lt;string&gt;** |  | [optional] [default to undefined]

## Example

```typescript
import { NodeTestResult } from '@mel-agent/api-client';

const instance: NodeTestResult = {
    success,
    output,
    error,
    execution_time,
    logs,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
