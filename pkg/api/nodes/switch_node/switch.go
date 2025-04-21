package switch_node

import internalapi "github.com/cedricziel/mel-agent/internal/api"

// switchDefinition provides the built-in "Switch" node.
type switchDefinition struct{}

// Meta returns metadata for the Switch node.
func (switchDefinition) Meta() internalapi.NodeType {
   return internalapi.NodeType{
       Type:      "switch",
       Label:     "Switch",
       Category:  "Control",
       Branching: true,
       Parameters: []internalapi.ParameterDefinition{
           {Name: "expression", Label: "Expression", Type: "string", Required: true, Group: "Settings"},
       },
   }
}

// Execute returns the input unchanged (branching logic handled elsewhere).
func (switchDefinition) Execute(agentID string, node internalapi.Node, input interface{}) (interface{}, error) {
   return input, nil
}

func init() {
   internalapi.RegisterNodeDefinition(switchDefinition{})
}