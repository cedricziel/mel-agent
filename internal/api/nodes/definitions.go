package nodes

import (
   "context"
   "fmt"
   "time"

   "github.com/cedricziel/mel-agent/internal/api"
   "github.com/cedricziel/mel-agent/internal/plugin"
   "github.com/google/uuid"
)

// This file registers a set of basic node types (Utility, Transform, I/O, etc.)
// that users can drag into their builder flows.
// AllNodeDefinitions returns all built-in builder node definitions.
func AllNodeDefinitions() []api.NodeDefinition {
   return []api.NodeDefinition{
       setVariableDefinition{},
       transformDefinition{},
       scriptDefinition{},
       switchDefinition{},
       forEachDefinition{},
       mergeDefinition{},
       delayDefinition{},
       httpResponseDefinition{},
       dbQueryDefinition{},
       emailDefinition{},
       fileIODefinition{},
       randomDefinition{},
       logDefinition{},
       noopDefinition{},
   }
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

// --- Transform Node (Map) ---
type transformDefinition struct{}
func (transformDefinition) Meta() api.NodeType {
   return api.NodeType{
       Type:     "transform",
       Label:    "Transform",
       Category: "Utility",
       Parameters: []api.ParameterDefinition{
           {Name: "expression", Label: "Expression", Type: "string", Required: true, Group: "Settings", Description: "CEL or JS expression to map input"},
       },
   }
}
func (transformDefinition) Execute(agentID string, node api.Node, input interface{}) (interface{}, error) {
   // Placeholder: currently passthrough
   return input, nil
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
   // Placeholder: passthrough
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
   // Placeholder: passthrough
   return input, nil
}

// --- For-Each Node ---
type forEachDefinition struct{}
func (forEachDefinition) Meta() api.NodeType {
   return api.NodeType{
       Type:     "for_each",
       Label:    "For Each",
       Category: "Control",
       Parameters: []api.ParameterDefinition{
           {Name: "path", Label: "Array Path", Type: "string", Required: true, Group: "Settings", Description: "JSONPath to array"},
       },
   }
}
func (forEachDefinition) Execute(agentID string, node api.Node, input interface{}) (interface{}, error) {
   // Placeholder: passthrough
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
   // Placeholder: passthrough
   return input, nil
}

// --- Delay Node ---
type delayDefinition struct{}
func (delayDefinition) Meta() api.NodeType {
   return api.NodeType{
       Type:     "delay",
       Label:    "Delay",
       Category: "Control",
       Parameters: []api.ParameterDefinition{
           {Name: "duration", Label: "Duration (ms)", Type: "number", Required: true, Default: 1000, Group: "Settings"},
       },
   }
}
func (delayDefinition) Execute(agentID string, node api.Node, input interface{}) (interface{}, error) {
   dms, _ := node.Data["duration"].(float64)
   time.Sleep(time.Duration(dms) * time.Millisecond)
   return input, nil
}

// --- HTTP Response Node ---
type httpResponseDefinition struct{}
func (httpResponseDefinition) Meta() api.NodeType {
   return api.NodeType{
       Type:     "http_response",
       Label:    "HTTP Response",
       Category: "Integration",
       Parameters: []api.ParameterDefinition{
           {Name: "statusCode", Label: "Status Code", Type: "number", Required: true, Default: 200, Group: "Settings"},
           {Name: "body", Label: "Body", Type: "string", Required: false, Default: "", Group: "Settings"},
       },
   }
}
func (httpResponseDefinition) Execute(agentID string, node api.Node, input interface{}) (interface{}, error) {
   return input, nil
}

// --- Database Query Node ---
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
   // Resolve the connection plugin by ID
   connID, _ := node.Data["connectionId"].(string)
   cp, ok := plugin.GetConnectionPlugin(connID)
   if !ok {
       return nil, fmt.Errorf("db_query: connection plugin %s not found", connID)
   }
   // Establish resource (e.g., DB client)
   resource, err := cp.Connect(context.Background(), node.Data)
   if err != nil {
       return nil, err
   }
   // Placeholder: return the resource handle and query
   return map[string]interface{}{"connection": resource, "query": node.Data["query"]}, nil
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
type fileIODefinition struct{}
func (fileIODefinition) Meta() api.NodeType {
   return api.NodeType{Type: "file_io", Label: "File I/O", Category: "Integration", Parameters: []api.ParameterDefinition{
           {Name: "operation", Label: "Operation", Type: "enum", Required: true, Default: "read", Options: []string{"read", "write"}, Group: "Settings"},
           {Name: "path", Label: "Path", Type: "string", Required: true, Group: "Settings"},
       }}
}
func (fileIODefinition) Execute(agentID string, node api.Node, input interface{}) (interface{}, error) {
   return input, nil
}

// --- Random / Utility Node ---
type randomDefinition struct{}
func (randomDefinition) Meta() api.NodeType {
   return api.NodeType{Type: "random", Label: "Random", Category: "Utility", Parameters: []api.ParameterDefinition{
           {Name: "type", Label: "Type", Type: "enum", Required: true, Default: "uuid", Options: []string{"uuid", "number"}, Group: "Settings"},
       }}
}
func (randomDefinition) Execute(agentID string, node api.Node, input interface{}) (interface{}, error) {
   typ, _ := node.Data["type"].(string)
   switch typ {
   case "uuid":
       return fmt.Sprintf("%s", uuid.New()), nil
   default:
       return input, nil
   }
}

// --- Log / Audit Node ---
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