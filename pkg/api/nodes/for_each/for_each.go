package for_each

import internalapi "github.com/cedricziel/mel-agent/internal/api"

// forEachDefinition provides the built-in "For Each" node.
type forEachDefinition struct{}

// Meta returns metadata for the For Each node.
func (forEachDefinition) Meta() internalapi.NodeType {
   return internalapi.NodeType{
       Type:     "for_each",
       Label:    "For Each",
       Category: "Control",
       Parameters: []internalapi.ParameterDefinition{
           {Name: "path", Label: "Array Path", Type: "string", Required: true, Group: "Settings", Description: "JSONPath to array"},
       },
   }
}

// Execute iterates over input arrays. Currently passthrough.
func (forEachDefinition) Execute(agentID string, node internalapi.Node, input interface{}) (interface{}, error) {
   return input, nil
}

func init() {
   internalapi.RegisterNodeDefinition(forEachDefinition{})
}