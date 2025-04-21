package script

import internalapi "github.com/cedricziel/mel-agent/internal/api"

// scriptDefinition provides the built-in "Script" node.
type scriptDefinition struct{}

// Meta returns metadata for the Script node.
func (scriptDefinition) Meta() internalapi.NodeType {
   return internalapi.NodeType{
       Type:     "script",
       Label:    "Script",
       Category: "Utility",
       Parameters: []internalapi.ParameterDefinition{
           {Name: "language", Label: "Language", Type: "enum", Required: true, Default: "javascript", Options: []string{"javascript", "python"}, Group: "Settings"},
           {Name: "code", Label: "Code", Type: "string", Required: true, Group: "Settings", Description: "Your script code"},
       },
   }
}

// Execute runs the user-provided script. Currently passthrough.
func (scriptDefinition) Execute(agentID string, node internalapi.Node, input interface{}) (interface{}, error) {
   return input, nil
}

func init() {
   internalapi.RegisterNodeDefinition(scriptDefinition{})
}