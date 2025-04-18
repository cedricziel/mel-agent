package api

import (
  "net/http"
)

// ParameterDefinition defines a single configuration parameter for a node.
type ParameterDefinition struct {
  Name                string                 `json:"name"`                           // key in node.Data
  Label               string                 `json:"label"`                          // user-facing label
  Type                string                 `json:"type"`                           // "string", "number", "boolean", "enum", "json"
  Required            bool                   `json:"required"`                       // must be provided (non-empty)
  Default             interface{}            `json:"default,omitempty"`             // default value
  Group               string                 `json:"group,omitempty"`               // logical grouping in UI
  VisibilityCondition string                 `json:"visibilityCondition,omitempty"` // CEL expression for conditional display
  Options             []string               `json:"options,omitempty"`             // for enum types
  Validators          []ValidatorSpec        `json:"validators,omitempty"`          // validation rules to apply
  Description         string                 `json:"description,omitempty"`          // help text or tooltip
}

// ValidatorSpec defines a validation rule and its parameters.
type ValidatorSpec struct {
  Type   string                 `json:"type"`   // e.g. "notEmpty", "url", "regex"
  Params map[string]interface{} `json:"params,omitempty"`
}

// NodeType defines metadata for a builder node, including parameters.
type NodeType struct {
  Type        string                 `json:"type"`
  Label       string                 `json:"label"`
  Category    string                 `json:"category"`
  EntryPoint  bool                   `json:"entry_point,omitempty"`
  Branching   bool                   `json:"branching,omitempty"`
  Parameters  []ParameterDefinition  `json:"parameters,omitempty"`
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
  return NodeType{
    Type:       "timer",
    Label:      "Timer",
    Category:   "Triggers",
    EntryPoint: true,
    Parameters: []ParameterDefinition{
      {Name: "mode", Label: "Mode", Type: "enum", Required: true, Default: "async", Options: []string{"async", "sync"}, Group: "Execution", Description: "Async (enqueue run) or Sync (inline) execution"},
      {Name: "statusCode", Label: "Response Status", Type: "number", Required: false, Default: 202, Group: "Response", Description: "HTTP status code returned by trigger"},
      {Name: "responseBody", Label: "Response Body", Type: "string", Required: false, Default: "", Group: "Response", Description: "HTTP body returned by trigger"},
    },
  }
}
func (timerDefinition) Execute(agentID string, node Node, input interface{}) (interface{}, error) {
  // Delegate to existing executor
  return DefaultExecutor{}.Execute(agentID, node, input)
}

// --- Schedule Node Definition ---
type scheduleDefinition struct{}
func (scheduleDefinition) Meta() NodeType {
  return NodeType{
    Type:       "schedule",
    Label:      "Schedule",
    Category:   "Triggers",
    EntryPoint: true,
    Parameters: []ParameterDefinition{
      {Name: "cron", Label: "Cron Expression", Type: "string", Required: true, Default: "", Group: "Schedule", Description: "Cron schedule to run"},
      {Name: "mode", Label: "Mode", Type: "enum", Required: true, Default: "async", Options: []string{"async", "sync"}, Group: "Execution"},
      {Name: "statusCode", Label: "Response Status", Type: "number", Required: false, Default: 202, Group: "Response"},
      {Name: "responseBody", Label: "Response Body", Type: "string", Required: false, Default: "", Group: "Response"},
    },
  }
}
func (scheduleDefinition) Execute(agentID string, node Node, input interface{}) (interface{}, error) {
  return ScheduleExecutor{}.Execute(agentID, node, input)
}

// --- Webhook Node Definition ---
type webhookDefinition struct{}
func (webhookDefinition) Meta() NodeType {
  return NodeType{
    Type:       "webhook",
    Label:      "Webhook",
    Category:   "Triggers",
    EntryPoint: true,
    Parameters: []ParameterDefinition{
      {Name: "secret", Label: "Secret", Type: "string", Required: false, Default: "", Group: "Security", Description: "HMAC or token to validate requests"},
      {Name: "mode", Label: "Mode", Type: "enum", Required: true, Default: "async", Options: []string{"async", "sync"}, Group: "Execution"},
      {Name: "statusCode", Label: "Response Status", Type: "number", Required: false, Default: 202, Group: "Response"},
      {Name: "responseBody", Label: "Response Body", Type: "string", Required: false, Default: "", Group: "Response"},
    },
  }
}
func (webhookDefinition) Execute(agentID string, node Node, input interface{}) (interface{}, error) {
  // Execution for webhook is handled via HTTP endpoint, default no-op
  return DefaultExecutor{}.Execute(agentID, node, input)
}

// --- Slack Node Definition ---
type slackDefinition struct{}
func (slackDefinition) Meta() NodeType {
  return NodeType{
    Type:       "slack",
    Label:      "Slack Slash Command",
    Category:   "Triggers",
    EntryPoint: true,
    Parameters: []ParameterDefinition{
      {Name: "command", Label: "Command", Type: "string", Required: true, Default: "", Group: "Trigger", Description: "Slash command to respond to"},
      {Name: "mode", Label: "Mode", Type: "enum", Required: true, Default: "async", Options: []string{"async", "sync"}, Group: "Execution"},
      {Name: "statusCode", Label: "Response Status", Type: "number", Required: false, Default: 200, Group: "Response"},
      {Name: "responseBody", Label: "Response Body", Type: "string", Required: false, Default: "", Group: "Response"},
    },
  }
}
func (slackDefinition) Execute(agentID string, node Node, input interface{}) (interface{}, error) {
  return DefaultExecutor{}.Execute(agentID, node, input)
}

// --- HTTP Request Node Definition ---
type httpRequestDefinition struct{}
func (httpRequestDefinition) Meta() NodeType {
  return NodeType{
    Type:      "http_request",
    Label:     "HTTP Request",
    Category:  "Integration",
    Parameters: []ParameterDefinition{
      {Name: "url", Label: "URL", Type: "string", Required: true, Default: "", Group: "Request", Description: "Endpoint to call", Validators: []ValidatorSpec{{Type: "notEmpty"}, {Type: "url"}}},
      {Name: "method", Label: "Method", Type: "enum", Required: true, Default: "GET", Options: []string{"GET","POST","PUT","DELETE"}, Group: "Request"},
      {Name: "headers", Label: "Headers", Type: "json", Required: false, Default: "{}", Group: "Request", Validators: []ValidatorSpec{{Type: "json"}}},
      {Name: "body", Label: "Body", Type: "string", Required: false, Default: "", Group: "Request", VisibilityCondition: "method!='GET'"},
      {Name: "timeout", Label: "Timeout", Type: "number", Required: false, Default: 30, Group: "Advanced", Description: "Timeout in seconds"},
    },
  }
}
func (httpRequestDefinition) Execute(agentID string, node Node, input interface{}) (interface{}, error) {
  return HTTPRequestExecutor{}.Execute(agentID, node, input)
}

// --- If Node Definition ---
type ifDefinition struct{}
func (ifDefinition) Meta() NodeType {
  return NodeType{
    Type:      "if",
    Label:     "If",
    Category:  "Basic",
    Branching: true,
    Parameters: []ParameterDefinition{
      {Name: "condition", Label: "Condition", Type: "string", Required: true, Default: "", Group: "Expression", Description: "Boolean CEL expression to evaluate"},
    },
  }
}
func (ifDefinition) Execute(agentID string, node Node, input interface{}) (interface{}, error) {
  return IfExecutor{}.Execute(agentID, node, input)
}

// --- Switch Node Definition ---
type switchDefinition struct{}
func (switchDefinition) Meta() NodeType {
  return NodeType{
    Type:      "switch",
    Label:     "Switch",
    Category:  "Basic",
    Parameters: []ParameterDefinition{},
  }
}
func (switchDefinition) Execute(agentID string, node Node, input interface{}) (interface{}, error) {
  return DefaultExecutor{}.Execute(agentID, node, input)
}

// --- Agent Node Definition ---
type agentDefinition struct{}
func (agentDefinition) Meta() NodeType {
  return NodeType{
    Type:      "agent",
    Label:     "Agent",
    Category:  "LLM",
    Parameters: []ParameterDefinition{},
  }
}
func (agentDefinition) Execute(agentID string, node Node, input interface{}) (interface{}, error) {
  return DefaultExecutor{}.Execute(agentID, node, input)
}