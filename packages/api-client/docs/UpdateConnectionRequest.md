# UpdateConnectionRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**name** | **string** |  | [optional] [default to undefined]
**secret** | **{ [key: string]: any; }** | Connection secret configuration containing authentication credentials | [optional] [default to undefined]
**config** | **{ [key: string]: any; }** | Connection configuration containing non-sensitive connection parameters | [optional] [default to undefined]
**usage_limit_month** | **number** |  | [optional] [default to undefined]
**is_default** | **boolean** |  | [optional] [default to undefined]
**status** | [**ConnectionStatus**](ConnectionStatus.md) |  | [optional] [default to undefined]

## Example

```typescript
import { UpdateConnectionRequest } from '@mel-agent/api-client';

const instance: UpdateConnectionRequest = {
    name,
    secret,
    config,
    usage_limit_month,
    is_default,
    status,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
