# Payload Envelope Design Document

## Overview

This document proposes a new typed, versioned payload system for the MEL Agent platform to replace the current `interface{}` approach with a strongly-typed envelope pattern that maintains flexibility while providing better guarantees, traceability, and evolution capabilities.

## Current State

The MEL Agent currently uses `interface{}` for data flowing between nodes:

```go
func (d MyNodeDefinition) Execute(ctx api.ExecutionContext, node api.Node, input interface{}) (interface{}, error)
```

While this provides maximum flexibility, it lacks:
- Type safety and compile-time checks
- Schema versioning and evolution
- Built-in traceability and error tracking
- Standardized metadata handling
- Binary data support

## Design Goals

| Goal | Why It Matters | Implementation Impact |
|------|----------------|----------------------|
| **Traceability** | Distributed engines need to correlate retries, timers and child workflows | Add trace IDs and step tracking to envelope |
| **Schema Evolution** | Workflows live for months; must accept old versions while emitting new | Version fields and backward compatibility |
| **Split/Aggregate** | Fan-out lists and aggregate results like Iterator/Aggregator in Make.com | Support for batch processing patterns |
| **Binary Friendliness** | Avoid base64-inflated JSON but still bind files to logical items | Separate binary map like n8n |
| **Language Ergonomics** | Developers get autocompletion and compile-time checks | Generic envelope with typed Data field |

## Core Envelope Design

### 1. Data Flow Units

```
Envelope ⟶ Item ⟶ Batch
```

- **Envelope**: One JSON object with metadata + one logical piece of business data
- **Item**: Synonym for envelope (echoing n8n terminology)  
- **Batch**: Array of envelopes for fan-out operations

### 2. Go Implementation

```go
// pkg/api/envelope.go (public interface)
package api

import "time"

// Trace holds IDs for logging, metrics, and correlation
type Trace struct {
    AgentID    string `json:"agentId"`           // Workflow/Agent identifier
    RunID      string `json:"runId"`             // Execution run identifier
    ParentID   string `json:"parentId,omitempty"` // Parent workflow for sub-flows
    NodeID     string `json:"nodeId"`            // Current processing node
    Step       string `json:"step"`              // Step name/identifier
    Attempt    int    `json:"attempt"`           // Retry attempt number
}

// ExecutionError captures step-level failures for DLQ & observability
type ExecutionError struct {
    Time     time.Time `json:"time"`
    NodeID   string    `json:"nodeId"`
    Message  string    `json:"message"`
    Stack    string    `json:"stack,omitempty"`
    Code     string    `json:"code,omitempty"`
}

// Envelope is the generic unit that flows through the MEL engine
type Envelope[T any] struct {
    ID        string                 `json:"id"`              // UUID v4
    IssuedAt  time.Time             `json:"issuedAt"`
    Version   int                   `json:"version"`         // Schema version
    DataType  string                `json:"dataType"`        // e.g. "HttpRequest", "UserCreated"
    Data      T                     `json:"data"`            // Strongly-typed business payload
    Binary    map[string][]byte     `json:"binary,omitempty"` // Optional binary attachments
    Meta      map[string]string     `json:"meta,omitempty"`   // Arbitrary metadata
    Variables map[string]interface{} `json:"variables,omitempty"` // Context variables
    Trace     Trace                 `json:"trace"`
    Errors    []ExecutionError      `json:"errors,omitempty"`
}
```

### 3. Common Data Types

```go
// pkg/core/payload_types.go
package core

// HTTPPayload for HTTP request/response data
type HTTPPayload struct {
    Method   string            `json:"method"`
    URL      string            `json:"url"`
    Headers  map[string]string `json:"headers"`
    Body     interface{}       `json:"body"`
    Status   int              `json:"status,omitempty"`
}

// DatabasePayload for database operations
type DatabasePayload struct {
    Query     string                 `json:"query"`
    Params    []interface{}         `json:"params,omitempty"`
    Results   []map[string]interface{} `json:"results,omitempty"`
    RowCount  int                   `json:"rowCount,omitempty"`
}

// FilePayload for file operations
type FilePayload struct {
    Path     string `json:"path"`
    Name     string `json:"name"`
    MimeType string `json:"mimeType"`
    Size     int64  `json:"size"`
    BinaryKey string `json:"binaryKey,omitempty"` // Reference to Binary map
}

// GenericPayload for flexible data
type GenericPayload map[string]interface{}
```

## Node Interface Evolution

### Current Interface
```go
type NodeDefinition interface {
    Execute(ctx ExecutionContext, node Node, input interface{}) (interface{}, error)
}
```

### New Envelope Interface
```go
// pkg/api/types.go (updated)
type NodeDefinition interface {
    Initialize(mel Mel) error
    Meta() NodeType
    Execute(ctx ExecutionContext, node Node, envelope *Envelope[interface{}]) (*Envelope[interface{}], error)
}

// Strongly-typed nodes can use specific types  
type TypedNodeDefinition[TIn, TOut any] interface {
    Initialize(mel Mel) error
    Meta() NodeType
    Execute(ctx ExecutionContext, node Node, envelope *Envelope[TIn]) (*Envelope[TOut], error)
}
```

### Backward Compatibility Layer
```go
// pkg/api/adapters.go
package api

// LegacyNodeAdapter wraps old interface{} nodes  
type LegacyNodeAdapter struct {
    legacy NodeDefinition // old interface{} based definition
}

func (a *LegacyNodeAdapter) Execute(ctx ExecutionContext, node Node, envelope *Envelope[interface{}]) (*Envelope[interface{}], error) {
    // Extract data for legacy node
    output, err := a.legacy.Execute(ctx, node, envelope.Data)
    if err != nil {
        return nil, err
    }
    
    // Wrap output in new envelope
    return &Envelope[interface{}]{
        ID:       generateID(),
        IssuedAt: time.Now(),
        Version:  envelope.Version,
        DataType: inferDataType(output),
        Data:     output,
        Trace:    envelope.Trace.Next(node.ID),
    }, nil
}
```

## Serialization Strategy

### PayloadConverter Interface
```go
// pkg/core/converter.go
package core

type PayloadConverter interface {
    Marshal(envelope *Envelope[interface{}]) ([]byte, error)
    Unmarshal(data []byte) (*Envelope[interface{}], error)
    ContentType() string
}

type JSONConverter struct{}
type ProtobufConverter struct{}

// Factory function
func NewConverter(format string) PayloadConverter {
    switch format {
    case "json":
        return &JSONConverter{}
    case "protobuf":
        return &ProtobufConverter{}
    default:
        return &JSONConverter{} // Default to JSON
    }
}
```

## Split/Aggregate Pattern

### Splitter Node
```go
// Splits an envelope containing an array into multiple envelopes
type SplitterNode struct{}

func (s *SplitterNode) Execute(ctx ExecutionContext, node Node, envelope *Envelope[[]interface{}]) ([]*Envelope[interface{}], error) {
    var results []*Envelope[interface{}]
    
    for i, item := range envelope.Data {
        childEnvelope := &Envelope[interface{}]{
            ID:       generateID(),
            IssuedAt: time.Now(),
            Version:  envelope.Version,
            DataType: envelope.DataType + "Item",
            Data:     item,
            Trace:    envelope.Trace.Child(node.ID, i),
            Meta:     copyMap(envelope.Meta),
        }
        results = append(results, childEnvelope)
    }
    
    return results, nil
}
```

### Aggregator Node
```go
// Collects envelopes and merges them when quorum is met
type AggregatorNode struct {
    storage map[string][]*Envelope[interface{}] // Keyed by workflow+step
}

func (a *AggregatorNode) Execute(ctx ExecutionContext, node Node, envelope *Envelope[interface{}]) (*Envelope[[]interface{}], error) {
    key := fmt.Sprintf("%s:%s", envelope.Trace.RunID, envelope.Trace.Step)
    
    // Add to collection
    a.storage[key] = append(a.storage[key], envelope)
    
    // Check if quorum met (implementation specific)
    if a.isQuorumMet(key) {
        items := a.storage[key]
        var data []interface{}
        for _, item := range items {
            data = append(data, item.Data)
        }
        
        result := &Envelope[[]interface{}]{
            ID:       generateID(),
            IssuedAt: time.Now(),
            Version:  envelope.Version,
            DataType: envelope.DataType + "Collection",
            Data:     data,
            Trace:    envelope.Trace.Next(node.ID),
        }
        
        delete(a.storage, key) // Cleanup
        return result, nil
    }
    
    return nil, nil // Not ready yet
}
```

## Error Handling and Retry

### Error Accumulation
```go
func (e *Envelope[T]) AddError(nodeID, message string, err error) {
    execError := ExecutionError{
        Time:    time.Now(),
        NodeID:  nodeID,
        Message: message,
        Stack:   getStackTrace(err),
    }
    
    if apiErr, ok := err.(*NodeError); ok {
        execError.Code = apiErr.Code
    }
    
    e.Errors = append(e.Errors, execError)
}

func (e *Envelope[T]) HasErrors() bool {
    return len(e.Errors) > 0
}

func (e *Envelope[T]) LastError() *ExecutionError {
    if len(e.Errors) == 0 {
        return nil
    }
    return &e.Errors[len(e.Errors)-1]
}
```

## Variable Integration

### Enhanced Variable Context
```go
// pkg/api/envelope.go (methods on Envelope)
func (e *Envelope[T]) GetVariable(scope VariableScope, key string) (interface{}, bool) {
    // Try envelope-local variables first
    if scope == RunScope {
        if val, exists := e.Variables[key]; exists {
            return val, true
        }
    }
    
    // Fall back to global variable store
    ctx := CreateVariableContext(e.Trace.AgentID, e.Trace.RunID, e.Trace.NodeID)
    val, exists, _ := GetVariable(ctx, scope, key)
    return val, exists
}

func (e *Envelope[T]) SetVariable(scope VariableScope, key string, value interface{}) error {
    if scope == RunScope {
        if e.Variables == nil {
            e.Variables = make(map[string]interface{})
        }
        e.Variables[key] = value
        return nil
    }
    
    // Use global variable store for other scopes
    ctx := CreateVariableContext(e.Trace.AgentID, e.Trace.RunID, e.Trace.NodeID)
    return SetVariable(ctx, scope, key, value)
}
```

## Migration Strategy

### Phase 1: Parallel Implementation
1. Add envelope types alongside existing interface{} system
2. Create adapter layer for backward compatibility
3. Implement core envelope utilities (ID generation, tracing, etc.)

### Phase 2: Node-by-Node Migration
1. Start with new nodes using typed envelopes
2. Gradually migrate existing nodes using adapters
3. Update variable nodes to work with envelope variables

### Phase 3: Full Transition
1. Remove interface{} from public APIs
2. Optimize serialization performance
3. Add schema validation and evolution tools

## Benefits

1. **Type Safety**: Compile-time checks instead of runtime type assertions
2. **Traceability**: Built-in correlation IDs and execution tracking
3. **Schema Evolution**: Versioned payloads support backward compatibility
4. **Binary Support**: Efficient handling of files and large data
5. **Error Context**: Rich error information travels with data
6. **Variable Integration**: Seamless variable access within envelopes
7. **Performance**: Option to switch to binary formats for production

## Implementation Files

### Public API (pkg/)
```
pkg/api/
├── envelope.go         # Core envelope and trace types (public interface)
├── types.go           # Updated to include envelope types
├── node.go            # Updated node interfaces with envelope support
└── adapters.go        # Backward compatibility layer

pkg/core/
├── envelope.go         # Core envelope implementation
├── payload_types.go    # Common payload type definitions
├── converter.go        # Serialization interfaces and implementations
├── utils.go           # ID generation, tracing utilities
└── patterns.go        # Split/aggregate node implementations
```

### Internal Implementation (internal/)
```
internal/api/
├── envelope_handlers.go # HTTP handlers for envelope-based APIs
├── router.go           # Updated to support envelope endpoints
└── schema.go          # Envelope schema validation

internal/runs/
├── runner.go          # Updated to process envelopes
└── envelope_processor.go # Envelope-specific execution logic

internal/models/
└── envelope_models.go  # Database models for envelope storage

migrations/
└── envelope_migration.go # Database schema for envelope storage
```

## API Package Separation

The envelope design respects MEL Agent's clean separation between public and internal APIs:

### Public API (`pkg/api/`)
- **Envelope types and interfaces**: Core `Envelope[T]`, `Trace`, `ExecutionError` types
- **Node interfaces**: Updated `NodeDefinition` with envelope support  
- **Backward compatibility**: Adapters for gradual migration
- **Type definitions**: Common payload types for node developers
- **Stability promise**: These interfaces maintain backward compatibility

### Internal Implementation (`internal/`)
- **HTTP handlers**: REST endpoints for envelope-based workflows
- **Execution engine**: Runner logic for processing envelopes
- **Database models**: Persistence layer for envelope storage
- **Business logic**: Workflow orchestration, retry logic, etc.
- **Implementation details**: Can change without affecting public API

### Core Utilities (`pkg/core/`)
- **Envelope implementations**: Concrete types implementing public interfaces
- **Serialization**: JSON/Protobuf converters for different environments
- **Common patterns**: Split/aggregate node implementations
- **Utilities**: ID generation, tracing helpers

This separation ensures that:
1. **Node developers** only depend on stable public interfaces
2. **Internal changes** don't break existing workflows
3. **API evolution** can happen without breaking compatibility
4. **Testing** can mock public interfaces easily

This design provides a clear evolution path from the current `interface{}` system to a robust, typed envelope pattern that supports the long-term needs of the MEL Agent platform.