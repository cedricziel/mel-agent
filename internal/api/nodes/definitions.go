package nodes

import (
   "github.com/cedricziel/mel-agent/internal/api"
)

// init registers all built-in builder node definitions.
func init() {
   // setVariableDefinition migrated out to pkg/api/nodes/set_variable
   api.RegisterNodeDefinition(scriptDefinition{})
   api.RegisterNodeDefinition(switchDefinition{})
   api.RegisterNodeDefinition(forEachDefinition{})
   api.RegisterNodeDefinition(mergeDefinition{})
   api.RegisterNodeDefinition(dbQueryDefinition{})
   api.RegisterNodeDefinition(emailDefinition{})
   api.RegisterNodeDefinition(logDefinition{})
   api.RegisterNodeDefinition(noopDefinition{})
}

// --- Set Variable Node ---
type setVariableDefinition struct{}
func (setVariableDefinition) Meta() api.NodeType {
   return api.NodeType{
       Type:     "set_variable",
       Label:    "Set Variable",
       Category: "Utility",
       Parameters: []api.ParameterDefinition{
           {Name: "key", Label: "Key", Type: "string", Required: true, Group: "Settings"},
           {Name: "value", Label: "Value", Type: "string", Required: true, Group: "Settings"},
       },
   }
}
func (setVariableDefinition) Execute(agentID string, node api.Node, input interface{}) (interface{}, error) {
   data, ok := input.(map[string]interface{})
   if !ok {
       data = map[string]interface{}{"input": input}
   }
   key, _ := node.Data["key"].(string)
   val := node.Data["value"]
   if key != "" {
       data[key] = val
   }
   return data, nil
}



// --- Script Node ---
type scriptDefinition struct{}
func (scriptDefinition) Meta() api.NodeType {
   return api.NodeType{
       Type:     "script",
       Label:    "Script",
       Category: "Utility",
       Parameters: []api.ParameterDefinition{
           {Name: "language", Label: "Language", Type: "enum", Required: true, Default: "javascript", Options: []string{"javascript", "python"}, Group: "Settings"},
           {Name: "code", Label: "Code", Type: "string", Required: true, Group: "Settings", Description: "Your script code"},
       },
   }
}
func (scriptDefinition) Execute(agentID string, node api.Node, input interface{}) (interface{}, error) {
   return input, nil
}

// --- Switch Node ---
type switchDefinition struct{}
func (switchDefinition) Meta() api.NodeType {
   return api.NodeType{
       Type:      "switch",
       Label:     "Switch",
       Category:  "Control",
       Branching: true,
       Parameters: []api.ParameterDefinition{
           {Name: "expression", Label: "Expression", Type: "string", Required: true, Group: "Settings"},
       },
   }
}
func (switchDefinition) Execute(agentID string, node api.Node, input interface{}) (interface{}, error) {
   return input, nil
}

// --- For Each Node ---
type forEachDefinition struct{}
func (forEachDefinition) Meta() api.NodeType {
   return api.NodeType{
       Type:     "for_each",
       Label:    "For Each",
       Category: "Control",
       Parameters: []api.ParameterDefinition{
           {Name: "path", Label: "Array Path", Type: "string", Required: true, Group: "Settings"},
       },
   }
}
func (forEachDefinition) Execute(agentID string, node api.Node, input interface{}) (interface{}, error) {
   return input, nil
}

// --- Merge Node ---
type mergeDefinition struct{}
func (mergeDefinition) Meta() api.NodeType {
   return api.NodeType{
       Type:     "merge",
       Label:    "Merge",
       Category: "Control",
       Parameters: []api.ParameterDefinition{
           {Name: "strategy", Label: "Strategy", Type: "enum", Required: true, Default: "concat", Options: []string{"concat", "union"}, Group: "Settings"},
       },
   }
}
func (mergeDefinition) Execute(agentID string, node api.Node, input interface{}) (interface{}, error) {
   return input, nil
}

  // --- Delay Node ---
  // Migrated to pkg/api/nodes/delay

// --- HTTP Response Node ---
// Migrated to pkg/api/nodes/http_response

// --- DB Query Node ---
type dbQueryDefinition struct{}
func (dbQueryDefinition) Meta() api.NodeType {
   return api.NodeType{
       Type:     "db_query",
       Label:    "DB Query",
       Category: "Integration",
       Parameters: []api.ParameterDefinition{
           {Name: "connectionId", Label: "Connection ID", Type: "string", Required: true, Group: "Settings"},
           {Name: "query", Label: "SQL Query", Type: "string", Required: true, Group: "Settings"},
       },
   }
}
func (dbQueryDefinition) Execute(agentID string, node api.Node, input interface{}) (interface{}, error) {
   return input, nil
}

// --- Email Notification Node ---
type emailDefinition struct{}
func (emailDefinition) Meta() api.NodeType {
   return api.NodeType{
       Type:     "email",
       Label:    "Email",
       Category: "Integration",
       Parameters: []api.ParameterDefinition{
           {Name: "to", Label: "To", Type: "string", Required: true, Group: "Settings"},
           {Name: "subject", Label: "Subject", Type: "string", Required: true, Group: "Settings"},
           {Name: "body", Label: "Body", Type: "string", Required: true, Group: "Settings"},
       },
   }
}
func (emailDefinition) Execute(agentID string, node api.Node, input interface{}) (interface{}, error) {
   return input, nil
}

   // --- File I/O Node ---
   // Migrated to pkg/api/nodes/file_io

// --- Random Node ---
// Migrated to pkg/api/nodes/random

// --- Log Node ---
type logDefinition struct{}
func (logDefinition) Meta() api.NodeType {
   return api.NodeType{Type: "log", Label: "Log", Category: "Utility", Parameters: []api.ParameterDefinition{
           {Name: "level", Label: "Level", Type: "enum", Required: true, Default: "info", Options: []string{"debug", "info", "warn", "error"}, Group: "Settings"},
           {Name: "message", Label: "Message", Type: "string", Required: true, Group: "Settings"},
       }}
}
func (logDefinition) Execute(agentID string, node api.Node, input interface{}) (interface{}, error) {
   return input, nil
}

// --- No-Op Node ---
type noopDefinition struct{}
func (noopDefinition) Meta() api.NodeType { return api.NodeType{Type: "noop", Label: "No-Op", Category: "Control"} }
func (noopDefinition) Execute(agentID string, node api.Node, input interface{}) (interface{}, error) { return input, nil }