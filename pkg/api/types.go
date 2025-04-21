package api

import (
	"net/http"

	internal "github.com/cedricziel/mel-agent/internal/api"
)

// Handler returns the HTTP handler with all API routes mounted.
func Handler() http.Handler {
	return internal.Handler()
}

// WebhookHandler handles external webhook events.
func WebhookHandler(w http.ResponseWriter, r *http.Request) {
	internal.WebhookHandler(w, r)
}

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

// NodeDefinition contains metadata and execution logic for a node type.
type NodeDefinition interface {
	Meta() NodeType
	Execute(agentID string, node Node, input interface{}) (interface{}, error)
}

// Trigger represents a user-configured event trigger.
type Trigger = internal.Trigger

// Connection represents a stored integration connection.
type Connection = internal.Connection

// Agent describes an agent record.
type Agent = internal.Agent

// Node is the runtime node struct with ID, Type, Data.
type Node = internal.Node

// RegisterNodeDefinition registers a node type for the builder.
func RegisterNodeDefinition(def NodeDefinition) {
	internal.RegisterNodeDefinition(def)
}

// ListNodeDefinitions returns all registered node definitions.
func ListNodeDefinitions() []NodeDefinition {
	return internal.ListNodeDefinitions()
}

// AllCoreDefinitions returns the built-in core trigger and utility node definitions.
func AllCoreDefinitions() []NodeDefinition {
	return internal.AllCoreDefinitions()
}
