package api

import (
  "net/http"
)

// NodeType defines metadata for a builder node.
type NodeType struct {
  Type       string            `json:"type"`
  Label      string            `json:"label"`
  Category   string            `json:"category"`
  Defaults   map[string]string `json:"defaults,omitempty"`
  EntryPoint bool              `json:"entry_point,omitempty"`
  Branching  bool              `json:"branching,omitempty"`
}

// NodeDefinition contains metadata and execution logic for a node type.
type NodeDefinition interface {
  Meta() NodeType
  Execute(agentID string, node Node, input interface{}) (interface{}, error)
}

var definitions []NodeDefinition

// RegisterNodeDefinition registers a node type definition.
func RegisterNodeDefinition(def NodeDefinition) {
  definitions = append(definitions, def)
}

// listNodeTypes returns all registered node type metadata.
func listNodeTypes(w http.ResponseWriter, r *http.Request) {
  metas := []NodeType{}
  for _, def := range definitions {
    metas = append(metas, def.Meta())
  }
  writeJSON(w, http.StatusOK, metas)
}

// FindDefinition retrieves the NodeDefinition for a given type.
func FindDefinition(typ string) NodeDefinition {
  for _, def := range definitions {
    if def.Meta().Type == typ {
      return def
    }
  }
  return nil
}

// Register built-in node definitions.
func init() {
  RegisterNodeDefinition(timerDefinition{})
  RegisterNodeDefinition(scheduleDefinition{})
  RegisterNodeDefinition(webhookDefinition{})
  RegisterNodeDefinition(slackDefinition{})
  RegisterNodeDefinition(httpRequestDefinition{})
  RegisterNodeDefinition(ifDefinition{})
  RegisterNodeDefinition(switchDefinition{})
  RegisterNodeDefinition(agentDefinition{})
}

// --- Timer Node Definition ---
type timerDefinition struct{}
func (timerDefinition) Meta() NodeType {
  return NodeType{Type: "timer", Label: "Timer", Category: "Triggers", Defaults: map[string]string{"mode": "async", "statusCode": "202", "responseBody": ""}, EntryPoint: true}
}
func (timerDefinition) Execute(agentID string, node Node, input interface{}) (interface{}, error) {
  // Delegate to existing executor
  return DefaultExecutor{}.Execute(agentID, node, input)
}

// --- Schedule Node Definition ---
type scheduleDefinition struct{}
func (scheduleDefinition) Meta() NodeType {
  return NodeType{Type: "schedule", Label: "Schedule", Category: "Triggers", Defaults: map[string]string{"cron": "", "mode": "async", "statusCode": "202", "responseBody": ""}, EntryPoint: true}
}
func (scheduleDefinition) Execute(agentID string, node Node, input interface{}) (interface{}, error) {
  return ScheduleExecutor{}.Execute(agentID, node, input)
}

// --- Webhook Node Definition ---
type webhookDefinition struct{}
func (webhookDefinition) Meta() NodeType {
  return NodeType{Type: "webhook", Label: "Webhook", Category: "Triggers", Defaults: map[string]string{"secret": "", "mode": "async", "statusCode": "202", "responseBody": ""}, EntryPoint: true}
}
func (webhookDefinition) Execute(agentID string, node Node, input interface{}) (interface{}, error) {
  // Execution for webhook is handled via HTTP endpoint, default no-op
  return DefaultExecutor{}.Execute(agentID, node, input)
}

// --- Slack Node Definition ---
type slackDefinition struct{}
func (slackDefinition) Meta() NodeType {
  return NodeType{Type: "slack", Label: "Slack Slash Command", Category: "Triggers", Defaults: map[string]string{"command": "", "mode": "async", "statusCode": "202", "responseBody": ""}, EntryPoint: true}
}
func (slackDefinition) Execute(agentID string, node Node, input interface{}) (interface{}, error) {
  return DefaultExecutor{}.Execute(agentID, node, input)
}

// --- HTTP Request Node Definition ---
type httpRequestDefinition struct{}
func (httpRequestDefinition) Meta() NodeType {
  return NodeType{Type: "http_request", Label: "HTTP Request", Category: "Integration", Defaults: map[string]string{"url": "", "method": "GET", "headers": "", "body": ""}}
}
func (httpRequestDefinition) Execute(agentID string, node Node, input interface{}) (interface{}, error) {
  return HTTPRequestExecutor{}.Execute(agentID, node, input)
}

// --- If Node Definition ---
type ifDefinition struct{}
func (ifDefinition) Meta() NodeType {
  return NodeType{Type: "if", Label: "If", Category: "Basic", Defaults: map[string]string{"condition": ""}, Branching: true}
}
func (ifDefinition) Execute(agentID string, node Node, input interface{}) (interface{}, error) {
  return IfExecutor{}.Execute(agentID, node, input)
}

// --- Switch Node Definition ---
type switchDefinition struct{}
func (switchDefinition) Meta() NodeType {
  return NodeType{Type: "switch", Label: "Switch", Category: "Basic"}
}
func (switchDefinition) Execute(agentID string, node Node, input interface{}) (interface{}, error) {
  return DefaultExecutor{}.Execute(agentID, node, input)
}

// --- Agent Node Definition ---
type agentDefinition struct{}
func (agentDefinition) Meta() NodeType {
  return NodeType{Type: "agent", Label: "Agent", Category: "LLM"}
}
func (agentDefinition) Execute(agentID string, node Node, input interface{}) (interface{}, error) {
  return DefaultExecutor{}.Execute(agentID, node, input)
}