package log

import internalapi "github.com/cedricziel/mel-agent/internal/api"

// logDefinition provides the built-in "Log" node.
type logDefinition struct{}

// Meta returns metadata for the Log node.
func (logDefinition) Meta() internalapi.NodeType {
   return internalapi.NodeType{
       Type:     "log",
       Label:    "Log",
       Category: "Utility",
       Parameters: []internalapi.ParameterDefinition{
           {Name: "level", Label: "Level", Type: "enum", Required: true, Default: "info", Options: []string{"debug", "info", "warn", "error"}, Group: "Settings"},
           {Name: "message", Label: "Message", Type: "string", Required: true, Group: "Settings"},
       },
   }
}

// Execute returns the input unchanged (logging handled separately).
func (logDefinition) Execute(agentID string, node internalapi.Node, input interface{}) (interface{}, error) {
   return input, nil
}

func init() {
   internalapi.RegisterNodeDefinition(logDefinition{})
}