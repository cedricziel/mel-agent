package api

import (
  "net/http"
)

// NodeType defines metadata for a builder node.
// NodeType defines metadata for a builder node.
type NodeType struct {
  // Unique identifier for the node type
  Type       string            `json:"type"`
  // Display label in the builder palette
  Label      string            `json:"label"`
  // Category groups node types in the palette
  Category   string            `json:"category"`
  // Default param values for new nodes
  Defaults   map[string]string `json:"defaults,omitempty"`
  // EntryPoint marks nodes that have no inputs (trigger nodes)
  EntryPoint bool              `json:"entry_point,omitempty"`
  // Branching marks nodes that have multiple outputs (e.g. If node)
  Branching bool              `json:"branching,omitempty"`
}

// all node types; extendable via plugins
var nodeTypes = []NodeType{
  // Trigger nodes
  // Trigger nodes (entry points)
  {Type: "timer", Label: "Timer", Category: "Triggers", Defaults: map[string]string{"mode": "async", "statusCode": "202", "responseBody": ""}, EntryPoint: true},
  {Type: "schedule", Label: "Schedule", Category: "Triggers", Defaults: map[string]string{"cron": "", "mode": "async", "statusCode": "202", "responseBody": ""}, EntryPoint: true},
  {Type: "webhook", Label: "Webhook", Category: "Triggers", Defaults: map[string]string{"secret": "", "mode": "async", "statusCode": "202", "responseBody": ""}, EntryPoint: true},
  {Type: "slack", Label: "Slack Slash Command", Category: "Triggers", Defaults: map[string]string{"command": "", "mode": "async", "statusCode": "202", "responseBody": ""}, EntryPoint: true},
  {Type: "http", Label: "HTTP Request", Category: "Triggers", Defaults: map[string]string{"mode": "async", "statusCode": "202", "responseBody": ""}, EntryPoint: true},
  {Type: "if", Label: "If", Category: "Basic", Defaults: map[string]string{"condition": ""}, Branching: true},
  {Type: "switch", Label: "Switch", Category: "Basic"},
  {Type: "agent", Label: "Agent", Category: "LLM"},
}

// RegisterNodeType adds a new node definition, allowing plugins to register types.
func RegisterNodeType(def NodeType) {
  nodeTypes = append(nodeTypes, def)
}

// listNodeTypes returns all registered node type definitions.
func listNodeTypes(w http.ResponseWriter, r *http.Request) {
  writeJSON(w, http.StatusOK, nodeTypes)
}