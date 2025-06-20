# CreateConnectionRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**name** | **string** |  | [default to undefined]
**integration_id** | **string** |  | [default to undefined]
**secret** | **{ [key: string]: any; }** |  | [optional] [default to undefined]
**config** | **{ [key: string]: any; }** |  | [optional] [default to undefined]
**usage_limit_month** | **number** |  | [optional] [default to undefined]
**is_default** | **boolean** |  | [optional] [default to undefined]

## Example

```typescript
import { CreateConnectionRequest } from '@mel-agent/api-client';

const instance: CreateConnectionRequest = {
    name,
    integration_id,
    secret,
    config,
    usage_limit_month,
    is_default,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
