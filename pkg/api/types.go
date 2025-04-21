package api

import (
   internal "github.com/cedricziel/mel-agent/internal/api"
   "net/http"
)

// Handler returns the HTTP handler with all API routes mounted.
func Handler() http.Handler {
   return internal.Handler()
}
// WebhookHandler handles external webhook events.
func WebhookHandler(w http.ResponseWriter, r *http.Request) {
   internal.WebhookHandler(w, r)
}

// NodeType describes metadata for a builder node.
type NodeType = internal.NodeType

// ParameterDefinition defines a single configuration parameter for a node.
type ParameterDefinition = internal.ParameterDefinition

// Trigger represents a user-configured event trigger.
type Trigger = internal.Trigger

// Connection represents a stored integration connection.
type Connection = internal.Connection

// Agent describes an agent record.
type Agent = internal.Agent
// NodeDefinition is the interface for node definitions.
type NodeDefinition = internal.NodeDefinition

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