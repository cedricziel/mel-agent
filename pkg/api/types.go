package api


// ParameterType represents the type of a parameter with JSON schema support.
type ParameterType string

const (
	TypeString   ParameterType = "string"
	TypeNumber   ParameterType = "number"
	TypeInteger  ParameterType = "integer"
	TypeBoolean  ParameterType = "boolean"
	TypeEnum     ParameterType = "enum"
	TypeObject   ParameterType = "object"
	TypeArray    ParameterType = "array"
	TypeJSON     ParameterType = "json" // backward compatibility alias for object
)

// JSONSchema represents a JSON schema definition.
type JSONSchema struct {
	Type        string                 `json:"type,omitempty"`
	Format      string                 `json:"format,omitempty"`
	Enum        []interface{}          `json:"enum,omitempty"`
	Properties  map[string]*JSONSchema `json:"properties,omitempty"`
	Items       *JSONSchema            `json:"items,omitempty"`
	Required    []string               `json:"required,omitempty"`
	Minimum     *float64               `json:"minimum,omitempty"`
	Maximum     *float64               `json:"maximum,omitempty"`
	MinLength   *int                   `json:"minLength,omitempty"`
	MaxLength   *int                   `json:"maxLength,omitempty"`
	Pattern     string                 `json:"pattern,omitempty"`
	Description string                 `json:"description,omitempty"`
	Default     interface{}            `json:"default,omitempty"`
}

// ToJSONSchema converts a ParameterType to its JSON schema representation.
func (pt ParameterType) ToJSONSchema() *JSONSchema {
	switch pt {
	case TypeString:
		return &JSONSchema{Type: "string"}
	case TypeNumber:
		return &JSONSchema{Type: "number"}
	case TypeInteger:
		return &JSONSchema{Type: "integer"}
	case TypeBoolean:
		return &JSONSchema{Type: "boolean"}
	case TypeEnum:
		return &JSONSchema{Type: "string"} // enum constraint added separately
	case TypeObject, TypeJSON:
		return &JSONSchema{Type: "object"}
	case TypeArray:
		return &JSONSchema{Type: "array"}
	default:
		return &JSONSchema{Type: "string"} // fallback
	}
}

// String returns the string representation of the parameter type.
func (pt ParameterType) String() string {
	return string(pt)
}

// IsValid checks if the parameter type is valid.
func (pt ParameterType) IsValid() bool {
	switch pt {
	case TypeString, TypeNumber, TypeInteger, TypeBoolean, TypeEnum, TypeObject, TypeArray, TypeJSON:
		return true
	default:
		return false
	}
}

// ParameterDefinition defines a single configuration parameter for a node.
type ParameterDefinition struct {
	Name                string          `json:"name"`                          // key in node.Data
	Label               string          `json:"label"`                         // user-facing label
	Type                string          `json:"type"`                          // "string", "number", "boolean", "enum", "json" (backward compatibility)
	ParameterType       ParameterType   `json:"parameterType,omitempty"`       // typed version of Type field
	Required            bool            `json:"required"`                      // must be provided (non-empty)
	Default             interface{}     `json:"default,omitempty"`             // default value
	Group               string          `json:"group,omitempty"`               // logical grouping in UI
	VisibilityCondition string          `json:"visibilityCondition,omitempty"` // CEL expression for conditional display
	Options             []string        `json:"options,omitempty"`             // for enum types
	Validators          []ValidatorSpec `json:"validators,omitempty"`          // validation rules to apply
	Description         string          `json:"description,omitempty"`         // help text or tooltip
	
	// JSON Schema specific fields
	JSONSchema          *JSONSchema     `json:"jsonSchema,omitempty"`          // explicit JSON schema override
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

// GetEffectiveType returns the parameter type, preferring ParameterType over Type for backward compatibility.
func (pd ParameterDefinition) GetEffectiveType() ParameterType {
	if pd.ParameterType != "" && pd.ParameterType.IsValid() {
		return pd.ParameterType
	}
	return ParameterType(pd.Type) // fallback to string Type field
}

// ToJSONSchema generates a JSON schema for this parameter definition.
func (pd ParameterDefinition) ToJSONSchema() *JSONSchema {
	// If explicit JSON schema is provided, use it
	if pd.JSONSchema != nil {
		schema := *pd.JSONSchema // copy
		// Apply parameter-level overrides
		if schema.Description == "" && pd.Description != "" {
			schema.Description = pd.Description
		}
		if schema.Default == nil && pd.Default != nil {
			schema.Default = pd.Default
		}
		return &schema
	}

	// Generate schema from parameter type
	effectiveType := pd.GetEffectiveType()
	schema := effectiveType.ToJSONSchema()

	// Apply parameter definition properties
	if pd.Description != "" {
		schema.Description = pd.Description
	}
	if pd.Default != nil {
		schema.Default = pd.Default
	}

	// Handle enum options
	if effectiveType == TypeEnum && len(pd.Options) > 0 {
		schema.Enum = make([]interface{}, len(pd.Options))
		for i, opt := range pd.Options {
			schema.Enum[i] = opt
		}
	}

	// Apply validators
	for _, validator := range pd.Validators {
		switch validator.Type {
		case "notEmpty":
			if schema.Type == "string" {
				minLen := 1
				schema.MinLength = &minLen
			}
		case "url":
			if schema.Type == "string" {
				schema.Format = "uri"
			}
		case "regex":
			if pattern, ok := validator.Params["pattern"].(string); ok && schema.Type == "string" {
				schema.Pattern = pattern
			}
		case "min":
			if min, ok := validator.Params["value"].(float64); ok && (schema.Type == "number" || schema.Type == "integer") {
				schema.Minimum = &min
			}
		case "max":
			if max, ok := validator.Params["value"].(float64); ok && (schema.Type == "number" || schema.Type == "integer") {
				schema.Maximum = &max
			}
		}
	}

	return schema
}

// NewStringParameter creates a parameter definition for a string type.
func NewStringParameter(name, label string, required bool) ParameterDefinition {
	return ParameterDefinition{
		Name:          name,
		Label:         label,
		Type:          string(TypeString),
		ParameterType: TypeString,
		Required:      required,
	}
}

// NewNumberParameter creates a parameter definition for a number type.
func NewNumberParameter(name, label string, required bool) ParameterDefinition {
	return ParameterDefinition{
		Name:          name,
		Label:         label,
		Type:          string(TypeNumber),
		ParameterType: TypeNumber,
		Required:      required,
	}
}

// NewIntegerParameter creates a parameter definition for an integer type.
func NewIntegerParameter(name, label string, required bool) ParameterDefinition {
	return ParameterDefinition{
		Name:          name,
		Label:         label,
		Type:          string(TypeInteger),
		ParameterType: TypeInteger,
		Required:      required,
	}
}

// NewBooleanParameter creates a parameter definition for a boolean type.
func NewBooleanParameter(name, label string, required bool) ParameterDefinition {
	return ParameterDefinition{
		Name:          name,
		Label:         label,
		Type:          string(TypeBoolean),
		ParameterType: TypeBoolean,
		Required:      required,
	}
}

// NewEnumParameter creates a parameter definition for an enum type.
func NewEnumParameter(name, label string, options []string, required bool) ParameterDefinition {
	return ParameterDefinition{
		Name:          name,
		Label:         label,
		Type:          string(TypeEnum),
		ParameterType: TypeEnum,
		Options:       options,
		Required:      required,
	}
}

// NewObjectParameter creates a parameter definition for an object type.
func NewObjectParameter(name, label string, required bool) ParameterDefinition {
	return ParameterDefinition{
		Name:          name,
		Label:         label,
		Type:          string(TypeObject),
		ParameterType: TypeObject,
		Required:      required,
	}
}

// WithDefault sets the default value for a parameter definition.
func (pd ParameterDefinition) WithDefault(value interface{}) ParameterDefinition {
	pd.Default = value
	return pd
}

// WithDescription sets the description for a parameter definition.
func (pd ParameterDefinition) WithDescription(desc string) ParameterDefinition {
	pd.Description = desc
	return pd
}

// WithGroup sets the group for a parameter definition.
func (pd ParameterDefinition) WithGroup(group string) ParameterDefinition {
	pd.Group = group
	return pd
}

// WithValidators sets the validators for a parameter definition.
func (pd ParameterDefinition) WithValidators(validators ...ValidatorSpec) ParameterDefinition {
	pd.Validators = validators
	return pd
}

// WithVisibilityCondition sets the visibility condition for a parameter definition.
func (pd ParameterDefinition) WithVisibilityCondition(condition string) ParameterDefinition {
	pd.VisibilityCondition = condition
	return pd
}
