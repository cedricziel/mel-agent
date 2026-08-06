# CredentialType


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**id** | **string** |  | [optional] [default to undefined]
**name** | **string** |  | [optional] [default to undefined]
**description** | **string** |  | [optional] [default to undefined]
**schema** | **{ [key: string]: any; }** |  | [optional] [default to undefined]
**required_fields** | **Array&lt;string&gt;** |  | [optional] [default to undefined]
**test_endpoint** | **string** |  | [optional] [default to undefined]

## Example

```typescript
import { CredentialType } from '@mel-agent/api-client';

const instance: CredentialType = {
    id,
    name,
    description,
    schema,
    required_fields,
    test_endpoint,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
