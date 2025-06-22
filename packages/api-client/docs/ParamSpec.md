# ParamSpec


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**name** | **string** | Key name | [default to undefined]
**label** | **string** | User-facing label | [optional] [default to undefined]
**type** | **string** | Data type (e.g. \&quot;string\&quot;, \&quot;number\&quot;, \&quot;enum\&quot;) | [default to undefined]
**required** | **boolean** | Must be provided | [optional] [default to undefined]
**_default** | **any** | Default value | [optional] [default to undefined]
**group** | **string** | UI grouping | [optional] [default to undefined]
**visibility_condition** | **string** | Conditional display expression | [optional] [default to undefined]
**_options** | **Array&lt;string&gt;** | Enum options | [optional] [default to undefined]
**validators** | [**Array&lt;ValidatorSpec&gt;**](ValidatorSpec.md) | Validation rules | [optional] [default to undefined]
**description** | **string** | Help text | [optional] [default to undefined]

## Example

```typescript
import { ParamSpec } from '@mel-agent/api-client';

const instance: ParamSpec = {
    name,
    label,
    type,
    required,
    _default,
    group,
    visibility_condition,
    _options,
    validators,
    description,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
