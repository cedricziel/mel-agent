# AgentVersion


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**id** | **string** |  | [optional] [default to undefined]
**agent_id** | **string** |  | [optional] [default to undefined]
**version_number** | **number** |  | [optional] [default to undefined]
**name** | **string** |  | [optional] [default to undefined]
**description** | **string** |  | [optional] [default to undefined]
**definition** | [**WorkflowDefinition**](WorkflowDefinition.md) |  | [optional] [default to undefined]
**created_at** | **string** |  | [optional] [default to undefined]
**is_current** | **boolean** |  | [optional] [default to undefined]

## Example

```typescript
import { AgentVersion } from '@mel-agent/api-client';

const instance: AgentVersion = {
    id,
    agent_id,
    version_number,
    name,
    description,
    definition,
    created_at,
    is_current,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
