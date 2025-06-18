package api

// NodeKind represents the functional category of a node
type NodeKind string

const (
	NodeKindAction  NodeKind = "action"
	NodeKindModel   NodeKind = "model"
	NodeKindMemory  NodeKind = "memory"
	NodeKindTool    NodeKind = "tool"
	NodeKindTrigger NodeKind = "trigger"
)

// ActionNode represents nodes that can execute workflow actions
type ActionNode interface {
	NodeDefinition
	// ExecuteEnvelope is inherited from NodeDefinition
	// This interface serves as a marker for action capabilities
}

// ModelNode represents nodes that provide AI model interaction capabilities
type ModelNode interface {
	NodeDefinition
	// InteractWith handles model interaction with given input
	InteractWith(ctx ExecutionContext, node Node, input string, options map[string]any) (string, error)
}

// MemoryNode represents nodes that provide memory storage and retrieval capabilities
type MemoryNode interface {
	NodeDefinition
	// Store saves data to memory
	Store(ctx ExecutionContext, node Node, key string, data any) error
	// Retrieve gets data from memory
	Retrieve(ctx ExecutionContext, node Node, key string) (any, error)
	// Search performs semantic search in memory
	Search(ctx ExecutionContext, node Node, query string, limit int) ([]MemoryResult, error)
}

// ToolNode represents nodes that provide tool capabilities
type ToolNode interface {
	NodeDefinition
	// CallTool executes a tool with given parameters
	CallTool(ctx ExecutionContext, node Node, toolName string, parameters map[string]any) (any, error)
	// ListTools returns available tools
	ListTools(ctx ExecutionContext, node Node) ([]ToolDefinition, error)
}

// TriggerNode represents nodes that can initiate workflow execution
type TriggerNode interface {
	NodeDefinition
	// StartListening begins listening for trigger events
	StartListening(ctx ExecutionContext, node Node) error
	// StopListening stops listening for trigger events
	StopListening(ctx ExecutionContext, node Node) error
}

// MemoryResult represents a single memory search result
type MemoryResult struct {
	Key       string  `json:"key"`
	Data      any     `json:"data"`
	Score     float64 `json:"score"`
	Timestamp string  `json:"timestamp"`
}

// ToolDefinition represents a tool that can be called
type ToolDefinition struct {
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Parameters  []ParameterDefinition `json:"parameters"`
}

// GetNodeKinds determines what kinds a node implements based on its interfaces
func GetNodeKinds(def NodeDefinition) []NodeKind {
	var kinds []NodeKind

	// Check for ActionNode (all nodes are actions by default via ExecuteEnvelope)
	if _, ok := def.(ActionNode); ok || def != nil {
		kinds = append(kinds, NodeKindAction)
	}

	// Check for ModelNode
	if _, ok := def.(ModelNode); ok {
		kinds = append(kinds, NodeKindModel)
	}

	// Check for MemoryNode
	if _, ok := def.(MemoryNode); ok {
		kinds = append(kinds, NodeKindMemory)
	}

	// Check for ToolNode
	if _, ok := def.(ToolNode); ok {
		kinds = append(kinds, NodeKindTool)
	}

	// Check for TriggerNode
	if _, ok := def.(TriggerNode); ok {
		kinds = append(kinds, NodeKindTrigger)
	}

	return kinds
}
