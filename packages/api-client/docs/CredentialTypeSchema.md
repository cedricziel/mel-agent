# CredentialTypeSchema


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**type** | **string** |  | [optional] [default to undefined]
**properties** | **{ [key: string]: any; }** |  | [optional] [default to undefined]
**required** | **Array&lt;string&gt;** |  | [optional] [default to undefined]
**title** | **string** |  | [optional] [default to undefined]
**description** | **string** |  | [optional] [default to undefined]

## Example

```typescript
import { CredentialTypeSchema } from '@mel-agent/api-client';

const instance: CredentialTypeSchema = {
    type,
    properties,
    required,
    title,
    description,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
