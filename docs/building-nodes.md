# Building Custom Nodes

This guide explains how to create custom node types for the MEL Agent platform. Nodes are the building blocks of workflows, handling data processing, external integrations, and control flow.

## Table of Contents

1. [Quick Start](#quick-start)
2. [Node Architecture](#node-architecture)
3. [Creating a Node](#creating-a-node)
4. [Parameter Definitions](#parameter-definitions)
5. [Envelope-Based Execution](#envelope-based-execution)
6. [Platform Utilities (Mel Interface)](#platform-utilities-mel-interface)
7. [Advanced Features](#advanced-features)
8. [Testing Your Node](#testing-your-node)
9. [Best Practices](#best-practices)

## Quick Start

Here's a minimal example of a custom node:

```go
package hello

import (
    "github.com/cedricziel/mel-agent/pkg/api"
)

type helloDefinition struct{}

func (helloDefinition) Meta() api.NodeType {
    return api.NodeType{
        Type:     "hello",
        Label:    "Hello World",
        Category: "Custom",
        Parameters: []api.ParameterDefinition{
            api.NewStringParameter("name", "Name", true).WithDefault("World"),
        },
    }
}

func (d helloDefinition) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[interface{}]) (*api.Envelope[interface{}], error) {
    name := "World"
    if n, ok := node.Data["name"].(string); ok && n != "" {
        name = n
    }

    result := envelope.Clone()
    result.Trace = envelope.Trace.Next(node.ID)
    result.Data = map[string]interface{}{
        "message": "Hello, " + name + "!",
    }
    
    return result, nil
}

func (helloDefinition) Initialize(mel api.Mel) error {
    return nil
}

func init() {
    api.RegisterNodeDefinition(helloDefinition{})
}

// Compile-time interface check
var _ api.NodeDefinition = (*helloDefinition)(nil)
```

## Node Architecture

Every node in MEL Agent consists of three main components:

### 1. Node Definition Structure
A struct implementing the `api.NodeDefinition` interface with three required methods:
- `Meta()` - Returns node metadata and parameters
- `ExecuteEnvelope()` - Handles node execution with envelope-based data flow
- `Initialize()` - Sets up node dependencies and resources

### 2. Registration
Nodes are registered globally using `api.RegisterNodeDefinition()` in the `init()` function.

### 3. Package Organization
Each node type should live in its own package under `pkg/nodes/[type]/`.

## Creating a Node

### Step 1: Set up the package structure

```
pkg/nodes/your_node/
├── your_node.go
└── your_node_test.go
```

### Step 2: Implement the NodeDefinition interface

```go
package your_node

import "github.com/cedricziel/mel-agent/pkg/api"

type yourNodeDefinition struct{}

func (yourNodeDefinition) Meta() api.NodeType {
    return api.NodeType{
        Type:     "your_node",
        Label:    "Your Node",
        Category: "Custom",
        Parameters: []api.ParameterDefinition{
            // Define your parameters here
        },
    }
}

func (d yourNodeDefinition) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[interface{}]) (*api.Envelope[interface{}], error) {
    // Your execution logic here
    result := envelope.Clone()
    result.Trace = envelope.Trace.Next(node.ID)
    
    // Process data and set result
    result.Data = map[string]interface{}{
        "processed": true,
    }
    
    return result, nil
}

func (yourNodeDefinition) Initialize(mel api.Mel) error {
    // Initialize resources, validate dependencies, etc.
    return nil
}

func init() {
    api.RegisterNodeDefinition(yourNodeDefinition{})
}

var _ api.NodeDefinition = (*yourNodeDefinition)(nil)
```

### Step 3: Register with the build system

Add your node to the build by importing it in a file that's already imported by the main application:

```go
import _ "github.com/cedricziel/mel-agent/pkg/nodes/your_node"
```

## Parameter Definitions

Parameters define the configuration options for your node. MEL Agent provides helper functions for common parameter types:

### Basic Parameter Types

```go
// String parameter
api.NewStringParameter("message", "Message", true).WithDefault("Hello")

// Number parameter
api.NewNumberParameter("timeout", "Timeout (seconds)", true).WithDefault(30)

// Integer parameter
api.NewIntegerParameter("retries", "Max Retries", false).WithDefault(3)

// Boolean parameter
api.NewBooleanParameter("enabled", "Enable Feature", false).WithDefault(true)

// Enum parameter
api.NewEnumParameter("level", "Log Level", []string{"debug", "info", "warn", "error"}, true)

// Object parameter (JSON)
api.NewObjectParameter("config", "Configuration", false)

// Array parameter
api.NewArrayParameter("items", "Items", false)

// Credential parameter
api.NewCredentialParameter("auth", "Authentication", "api_key", true)
```

### Parameter Modifiers

Chain modifiers to customize parameter behavior:

```go
api.NewStringParameter("url", "API URL", true).
    WithDescription("The endpoint URL for the API").
    WithGroup("Connection").
    WithValidators(
        api.ValidatorSpec{Type: "url"},
        api.ValidatorSpec{Type: "notEmpty"},
    ).
    WithVisibilityCondition("enabled == true")
```

### Parameter Groups

Group related parameters together in the UI:

```go
Parameters: []api.ParameterDefinition{
    api.NewStringParameter("host", "Host", true).WithGroup("Connection"),
    api.NewNumberParameter("port", "Port", true).WithGroup("Connection").WithDefault(80),
    api.NewBooleanParameter("ssl", "Use SSL", false).WithGroup("Connection"),
    
    api.NewStringParameter("username", "Username", false).WithGroup("Authentication"),
    api.NewCredentialParameter("password", "Password", "password", false).WithGroup("Authentication"),
}
```

### Dynamic Options

For parameters that need dynamic option loading:

```go
api.NewEnumParameter("database", "Database", []string{}, true).
    WithDynamicOptions().
    WithDescription("Select a database from your connection")
```

Then implement the `DynamicOptionsProvider` interface:

```go
func (d yourNodeDefinition) GetDynamicOptions(ctx api.ExecutionContext, parameterName string, dependencies map[string]interface{}) ([]api.OptionChoice, error) {
    if parameterName == "database" {
        // Fetch databases from your connection
        return []api.OptionChoice{
            {Value: "db1", Label: "Database 1", Description: "Primary database"},
            {Value: "db2", Label: "Database 2", Description: "Secondary database"},
        }, nil
    }
    return nil, nil
}
```

## Envelope-Based Execution

MEL Agent uses an envelope-based architecture for data flow. Envelopes carry data, tracing information, and variables between nodes.

### Envelope Structure

```go
type Envelope[T any] struct {
    ID        string                 `json:"id"`
    IssuedAt  time.Time             `json:"issuedAt"`
    Version   int                   `json:"version"`
    DataType  string                `json:"dataType"`
    Data      T                     `json:"data"`
    Trace     Trace                 `json:"trace"`
    Variables map[string]interface{} `json:"variables,omitempty"`
}
```

### Key Principles

1. **Immutability**: Always clone the input envelope before modifying
2. **Tracing**: Update the trace to record node execution
3. **Type Safety**: Use generic envelopes for compile-time type checking when possible

### Basic Envelope Processing

```go
func (d yourNodeDefinition) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[interface{}]) (*api.Envelope[interface{}], error) {
    // Always clone the input envelope
    result := envelope.Clone()
    
    // Update the trace
    result.Trace = envelope.Trace.Next(node.ID)
    
    // Process the data
    inputData := envelope.Data
    outputData := processData(inputData)
    
    // Set the result data
    result.Data = outputData
    
    return result, nil
}
```

### Working with Variables

Envelopes can carry variables that persist across the workflow:

```go
func (d yourNodeDefinition) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[interface{}]) (*api.Envelope[interface{}], error) {
    result := envelope.Clone()
    result.Trace = envelope.Trace.Next(node.ID)
    
    // Read variables
    if userID, exists := envelope.Variables["user_id"]; exists {
        // Use the user ID
    }
    
    // Set variables for downstream nodes
    if result.Variables == nil {
        result.Variables = make(map[string]interface{})
    }
    result.Variables["processed_at"] = time.Now()
    
    return result, nil
}
```

### Error Handling

Return descriptive errors for debugging:

```go
func (d yourNodeDefinition) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[interface{}]) (*api.Envelope[interface{}], error) {
    config, ok := node.Data["config"].(string)
    if !ok || config == "" {
        return nil, api.NewNodeError(node.ID, "your_node", "config parameter is required")
    }
    
    // Process with error handling
    if err := processConfig(config); err != nil {
        return nil, api.NewNodeErrorWithCode(node.ID, "your_node", "failed to process config: "+err.Error(), "CONFIG_ERROR")
    }
    
    result := envelope.Clone()
    result.Trace = envelope.Trace.Next(node.ID)
    return result, nil
}
```

## Platform Utilities (Mel Interface)

The `ctx.Mel` interface provides access to platform utilities:

### HTTP Requests

```go
func (d yourNodeDefinition) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[interface{}]) (*api.Envelope[interface{}], error) {
    url := node.Data["url"].(string)
    
    httpReq := api.HTTPRequest{
        Method:  "GET",
        URL:     url,
        Headers: map[string]string{
            "Authorization": "Bearer " + token,
            "Content-Type":  "application/json",
        },
        Timeout: 30 * time.Second,
    }
    
    response, err := ctx.Mel.HTTPRequest(context.Background(), httpReq)
    if err != nil {
        return nil, api.NewNodeError(node.ID, "your_node", "HTTP request failed: "+err.Error())
    }
    
    result := envelope.Clone()
    result.Trace = envelope.Trace.Next(node.ID)
    result.Data = map[string]interface{}{
        "status": response.StatusCode,
        "body":   string(response.Body),
    }
    
    return result, nil
}
```

### Workflow Communication

```go
// Call another workflow
req := api.WorkflowCallRequest{
    TargetWorkflowID: targetID,
    CallData:         map[string]interface{}{"input": "data"},
    CallMode:         "sync",
    TimeoutSeconds:   30,
    SourceContext:    ctx,
}

response, err := ctx.Mel.CallWorkflow(context.Background(), req)

// Return data to calling workflow (from workflow_return node)
err := ctx.Mel.ReturnToWorkflow(context.Background(), callID, returnData, "success")
```

### Data Storage

```go
// Store data with TTL
err := ctx.Mel.StoreData(context.Background(), "cache-key", data, 1*time.Hour)

// Retrieve data
data, err := ctx.Mel.RetrieveData(context.Background(), "cache-key")

// Delete data
err := ctx.Mel.DeleteData(context.Background(), "cache-key")
```

## Advanced Features

### Typed Node Definitions

For compile-time type safety, use `TypedNodeDefinition`:

```go
type typedNodeDefinition struct{}

func (d typedNodeDefinition) ExecuteTyped(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[InputType]) (*api.Envelope[OutputType], error) {
    // Strongly typed input and output
    inputData := envelope.Data
    outputData := processTypedData(inputData)
    
    result := &api.Envelope[OutputType]{
        ID:        envelope.ID,
        IssuedAt:  envelope.IssuedAt,
        Version:   envelope.Version + 1,
        DataType:  "OutputType",
        Data:      outputData,
        Trace:     envelope.Trace.Next(node.ID),
        Variables: envelope.Variables,
    }
    
    return result, nil
}
```

### Conditional Parameter Visibility

Use CEL expressions to show/hide parameters based on other values:

```go
api.NewStringParameter("api_key", "API Key", true).
    WithVisibilityCondition("auth_type == 'api_key'"),
api.NewCredentialParameter("oauth_token", "OAuth Token", "oauth", true).
    WithVisibilityCondition("auth_type == 'oauth'")
```

### Complex Validation

Define custom validators:

```go
api.NewStringParameter("email", "Email Address", true).
    WithValidators(
        api.ValidatorSpec{
            Type: "regex",
            Params: map[string]interface{}{
                "pattern": `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
            },
        },
    )
```

## Testing Your Node

Create comprehensive tests for your node:

```go
package your_node

import (
    "testing"
    "time"
    
    "github.com/cedricziel/mel-agent/pkg/api"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestYourNodeDefinition_ExecuteEnvelope(t *testing.T) {
    def := yourNodeDefinition{}
    
    ctx := api.ExecutionContext{
        AgentID: "test-agent",
        RunID:   "test-run",
    }
    
    node := api.Node{
        ID:   "test-node",
        Type: "your_node",
        Data: map[string]interface{}{
            "param1": "value1",
        },
    }
    
    envelope := &api.Envelope[interface{}]{
        ID:       "test-envelope",
        IssuedAt: time.Now(),
        Version:  1,
        Data:     map[string]interface{}{"input": "test"},
        Trace:    api.NewTrace(),
    }
    
    result, err := def.ExecuteEnvelope(ctx, node, envelope)
    
    require.NoError(t, err)
    assert.NotNil(t, result)
    assert.Equal(t, envelope.ID, result.ID)
    assert.Equal(t, envelope.Version+1, result.Version)
    
    // Test your specific output
    output, ok := result.Data.(map[string]interface{})
    require.True(t, ok)
    assert.Equal(t, "expected_value", output["key"])
}

func TestYourNodeDefinition_Meta(t *testing.T) {
    def := yourNodeDefinition{}
    meta := def.Meta()
    
    assert.Equal(t, "your_node", meta.Type)
    assert.Equal(t, "Your Node", meta.Label)
    assert.Equal(t, "Custom", meta.Category)
    assert.Len(t, meta.Parameters, 1)
}
```

## Best Practices

### 1. Error Handling
- Use `api.NewNodeError()` for user-facing errors
- Include the node ID and type in errors
- Provide helpful error messages
- Use error codes for programmatic handling

### 2. Parameter Design
- Use descriptive labels and help text
- Group related parameters logically
- Provide sensible defaults
- Validate input parameters thoroughly

### 3. Data Processing
- Always clone input envelopes
- Update the trace correctly
- Handle nil and missing data gracefully
- Preserve variable context when appropriate

### 4. Performance
- Avoid blocking operations in `ExecuteEnvelope()`
- Use context cancellation for long-running operations
- Implement proper timeouts for external calls
- Cache expensive computations when possible

### 5. Security
- Validate all input parameters
- Sanitize user-provided data
- Use credentials securely
- Never log sensitive information

### 6. Documentation
- Document parameter behavior
- Provide usage examples
- Explain side effects and dependencies
- Include performance characteristics

### 7. Backward Compatibility
- Version your node types when making breaking changes
- Provide migration paths for existing workflows
- Use parameter defaults to maintain compatibility
- Test with existing workflow data

### 8. Testing
- Test with various input types
- Test error conditions
- Test parameter validation
- Test envelope transformations
- Use table-driven tests for multiple scenarios

This documentation should provide a comprehensive guide for building custom nodes in the MEL Agent platform. Each node you create extends the platform's capabilities and can be combined with others to create powerful workflows.