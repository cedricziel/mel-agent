# Node Kinds and Interface System

MEL Agent uses a kind-based architecture that allows nodes to implement multiple capabilities through different interfaces. This provides flexibility and type safety while enabling nodes to serve multiple purposes within the platform.

## Overview

### Terminology
- **Type**: Specific node implementation (e.g., `openai_model`, `http_request`, `local_memory`)
- **Kind**: Functional capability category (e.g., `action`, `model`, `memory`, `tool`, `trigger`)

A single node type can implement multiple kinds by conforming to their respective interfaces.

## Available Node Kinds

### 1. Action (`action`)
**Default capability for all nodes** - Can execute as workflow steps.

```go
type ActionNode interface {
    NodeDefinition
    // ExecuteEnvelope is inherited from NodeDefinition
}
```

All nodes are action nodes by default through the `NodeDefinition` interface.

### 2. Model (`model`)  
**AI Model Interaction** - Provides language model capabilities.

```go
type ModelNode interface {
    NodeDefinition
    InteractWith(ctx ExecutionContext, node Node, input string, options map[string]any) (string, error)
}
```

**Use cases:**
- Chat completion
- Text generation
- Model configuration for agents
- Direct model invocation in workflows

### 3. Memory (`memory`)
**Memory Storage and Retrieval** - Manages persistent data and semantic search.

```go
type MemoryNode interface {
    NodeDefinition
    Store(ctx ExecutionContext, node Node, key string, data any) error
    Retrieve(ctx ExecutionContext, node Node, key string) (any, error)
    Search(ctx ExecutionContext, node Node, query string, limit int) ([]MemoryResult, error)
}
```

**Use cases:**
- Context preservation across workflow runs
- Semantic search for relevant information
- User preference storage
- Knowledge base integration

### 4. Tool (`tool`)
**Tool Execution** - Provides external tool and function calling capabilities.

```go
type ToolNode interface {
    NodeDefinition
    CallTool(ctx ExecutionContext, node Node, toolName string, parameters map[string]any) (any, error)
    ListTools(ctx ExecutionContext, node Node) ([]ToolDefinition, error)
}
```

**Use cases:**
- Function calling for AI agents
- External API integration
- Code execution environments
- System tool access

### 5. Trigger (`trigger`)
**Workflow Initiation** - Can start workflow execution based on events.

```go
type TriggerNode interface {
    NodeDefinition
    StartListening(ctx ExecutionContext, node Node) error
    StopListening(ctx ExecutionContext, node Node) error
}
```

**Use cases:**
- Webhook endpoints
- Scheduled execution
- File system monitoring
- External event subscriptions

## Multi-Kind Implementation

Nodes can implement multiple interfaces to provide versatile functionality:

### Example: OpenAI Model Node

```go
type OpenAIModelNode struct{}

// Interface compliance checks
var _ api.ActionNode = (*OpenAIModelNode)(nil)
var _ api.ModelNode = (*OpenAIModelNode)(nil)

func (n *OpenAIModelNode) Meta() api.NodeType {
    return api.NodeType{
        Type:     "openai_model",
        Label:    "OpenAI Model",
        Category: "Configuration",
        // Kinds are automatically determined by interface implementation
        Parameters: []api.ParameterDefinition{
            api.NewEnumParameter("model", "Model", []string{"gpt-4", "gpt-3.5-turbo"}, true),
            api.NewNumberParameter("temperature", "Temperature", false).WithDefault(0.7),
            api.NewCredentialParameter("credential", "API Key", "openai_api_key", true),
        },
    }
}

// ActionNode capability - can be used in workflows
func (n *OpenAIModelNode) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[any]) (*api.Envelope[any], error) {
    // Workflow execution logic
    return envelope, nil
}

// ModelNode capability - can be used for AI interaction
func (n *OpenAIModelNode) InteractWith(ctx api.ExecutionContext, node api.Node, input string, options map[string]any) (string, error) {
    // OpenAI API interaction logic
    return "AI response to: " + input, nil
}
```

This node has `kinds: ["action", "model"]` and can be used:
1. As a workflow step (action capability)
2. As an AI model provider for agent configurations (model capability)

## Kind Detection

Kinds are automatically determined at runtime using Go interface assertions:

```go
func GetNodeKinds(def NodeDefinition) []NodeKind {
    var kinds []NodeKind

    // All nodes are actions by default
    kinds = append(kinds, NodeKindAction)

    // Check for additional capabilities
    if _, ok := def.(ModelNode); ok {
        kinds = append(kinds, NodeKindModel)
    }
    if _, ok := def.(MemoryNode); ok {
        kinds = append(kinds, NodeKindMemory)
    }
    if _, ok := def.(ToolNode); ok {
        kinds = append(kinds, NodeKindTool)
    }
    if _, ok := def.(TriggerNode); ok {
        kinds = append(kinds, NodeKindTrigger)
    }

    return kinds
}
```

## API Usage

### Frontend Filtering

The frontend can filter nodes by kind:

```javascript
import { nodeTypesApi } from './api/nodeTypesApi';

// Get all model nodes
const modelNodes = await nodeTypesApi.getNodeTypes('model');

// Get multiple kinds
const configNodes = await nodeTypesApi.getNodeTypes(['model', 'memory', 'tool']);

// Get all nodes
const allNodes = await nodeTypesApi.getAllNodeTypes();
```

### Backend Endpoints

```bash
# Get all node types
GET /api/node-types

# Filter by single kind
GET /api/node-types?kind=model

# Filter by multiple kinds
GET /api/node-types?kind=model,memory,tool

# Response includes kinds array
{
  "type": "openai_model",
  "label": "OpenAI Model", 
  "kinds": ["action", "model"],
  "parameters": [...]
}
```

## Best Practices

### 1. Interface Compliance
Always use compile-time checks to ensure interface compliance:

```go
var _ api.ActionNode = (*YourNode)(nil)
var _ api.ModelNode = (*YourNode)(nil)
```

### 2. Kind-Specific Logic
Separate concerns between different kind implementations:

```go
// Action-specific logic in ExecuteEnvelope
func (n *YourNode) ExecuteEnvelope(...) {
    // Workflow step logic
}

// Model-specific logic in InteractWith  
func (n *YourNode) InteractWith(...) {
    // AI interaction logic
}
```

### 3. Parameter Design
Design parameters that make sense for all implemented kinds:

```go
// Good: Parameters useful for both action and model capabilities
api.NewCredentialParameter("api_key", "API Key", "openai", true),
api.NewStringParameter("model", "Model Name", true),

// Avoid: Kind-specific parameters that don't apply to all capabilities
```

### 4. Error Handling
Provide clear errors for unsupported operations:

```go
func (n *YourNode) InteractWith(...) (string, error) {
    if !n.supportsModelInteraction() {
        return "", fmt.Errorf("node %s not configured for model interaction", node.ID)
    }
    // Implementation...
}
```

### 5. Documentation
Document which kinds your node implements and their use cases:

```go
// YourNode implements both ActionNode and ModelNode interfaces.
// As an ActionNode, it can process data in workflows.
// As a ModelNode, it provides AI text generation capabilities.
type YourNode struct{}
```

## Migration Guide

### From Category-Based to Kind-Based

**Old approach (deprecated):**
```javascript
// Frontend filtering by category inference
const modelNodes = await nodeTypesApi.getNodeTypes('type=model'); // ❌
```

**New approach:**
```javascript
// Frontend filtering by explicit kinds
const modelNodes = await nodeTypesApi.getNodeTypes('model'); // ✅
```

**Backend changes:**
- URL parameter changed from `?type=` to `?kind=`
- Response includes `kinds` array instead of category inference
- Nodes automatically expose their capabilities through interface detection

This kind-based system provides better type safety, clearer capabilities, and more flexible node designs while maintaining backward compatibility through the action kind default.