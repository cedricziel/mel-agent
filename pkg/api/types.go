package api

// ParameterDefinition defines a single configuration parameter for a node.
type ParameterDefinition struct {
	Name                string          `json:"name"`                          // key in node.Data
	Label               string          `json:"label"`                         // user-facing label
	Type                string          `json:"type"`                          // "string", "number", "boolean", "enum", "json"
	Required            bool            `json:"required"`                      // must be provided (non-empty)
	Default             interface{}     `json:"default,omitempty"`             // default value
	Group               string          `json:"group,omitempty"`               // logical grouping in UI
	VisibilityCondition string          `json:"visibilityCondition,omitempty"` // CEL expression for conditional display
	Options             []string        `json:"options,omitempty"`             // for enum types
	Validators          []ValidatorSpec `json:"validators,omitempty"`          // validation rules to apply
	Description         string          `json:"description,omitempty"`         // help text or tooltip
}

// ValidatorSpec defines a validation rule and its parameters.
type ValidatorSpec struct {
	Type   string                 `json:"type"` // e.g. "notEmpty", "url", "regex"
	Params map[string]interface{} `json:"params,omitempty"`
}

// NodeType defines metadata for a builder node, including parameters.
type NodeType struct {
	Type       string                `json:"type"`
	Label      string                `json:"label"`
	Icon       string                `json:"icon,omitempty"`
	Category   string                `json:"category"`
	EntryPoint bool                  `json:"entry_point,omitempty"`
	Branching  bool                  `json:"branching,omitempty"`
	Parameters []ParameterDefinition `json:"parameters,omitempty"`
}

// ExecutionContext provides context for node execution.
type ExecutionContext struct {
	AgentID   string                 `json:"agent_id"`
	RunID     string                 `json:"run_id,omitempty"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// ExecutionResult represents the result of node execution.
type ExecutionResult struct {
	Output interface{} `json:"output"`
	Error  error       `json:"error,omitempty"`
}

// NodeError represents node execution errors with context.
type NodeError struct {
	NodeID  string `json:"node_id"`
	Type    string `json:"type"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

func (e NodeError) Error() string {
	return e.Message
}

// NewNodeError creates a new NodeError.
func NewNodeError(nodeID, nodeType, message string) *NodeError {
	return &NodeError{
		NodeID:  nodeID,
		Type:    nodeType,
		Message: message,
	}
}

// NewNodeErrorWithCode creates a new NodeError with an error code.
func NewNodeErrorWithCode(nodeID, nodeType, message, code string) *NodeError {
	return &NodeError{
		NodeID:  nodeID,
		Type:    nodeType,
		Message: message,
		Code:    code,
	}
}

// NodeDefinition contains metadata and execution logic for a node type.
type NodeDefinition interface {
	Initialize(mel Mel) error
	Meta() NodeType
	Execute(ctx ExecutionContext, node Node, input interface{}) (interface{}, error)
}
