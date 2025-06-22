# Extension


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**id** | **string** | Unique plugin identifier | [default to undefined]
**version** | **string** | Semver-compliant version string | [default to undefined]
**categories** | **Array&lt;string&gt;** | Extension types provided (e.g. \&quot;node\&quot;, \&quot;trigger\&quot;) | [default to undefined]
**params** | [**Array&lt;ParamSpec&gt;**](ParamSpec.md) | Parameter schema for configuration/UI | [optional] [default to undefined]
**ui_component** | **string** | Optional path or URL to React bundle | [optional] [default to undefined]

## Example

```typescript
import { Extension } from '@mel-agent/api-client';

const instance: Extension = {
    id,
    version,
    categories,
    params,
    ui_component,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
