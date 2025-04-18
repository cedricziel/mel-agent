package api

import (
  "net/http"
)

// NodeType defines metadata for a builder node.
type NodeType struct {
  Type     string            `json:"type"`
  Label    string            `json:"label"`
  Category string            `json:"category"`
  Defaults map[string]string `json:"defaults,omitempty"`
}

// all node types; extendable via plugins
var nodeTypes = []NodeType{
  {Type: "timer", Label: "Timer", Category: "Triggers"},
  {Type: "http", Label: "HTTP Request", Category: "Triggers"},
  {Type: "if", Label: "If", Category: "Basic", Defaults: map[string]string{"condition": ""}},
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