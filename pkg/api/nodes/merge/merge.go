package merge

import internalapi "github.com/cedricziel/mel-agent/internal/api"

// mergeDefinition provides the built-in "Merge" node.
type mergeDefinition struct{}

// Meta returns metadata for the Merge node.
func (mergeDefinition) Meta() internalapi.NodeType {
   return internalapi.NodeType{
       Type:     "merge",
       Label:    "Merge",
       Category: "Control",
       Parameters: []internalapi.ParameterDefinition{
           {Name: "strategy", Label: "Strategy", Type: "enum", Required: true, Default: "concat", Options: []string{"concat", "union"}, Group: "Settings", Description: "Merge strategy"},
       },
   }
}

// Execute merges data based on strategy. Currently passthrough.
func (mergeDefinition) Execute(agentID string, node internalapi.Node, input interface{}) (interface{}, error) {
   return input, nil
}

func init() {
   internalapi.RegisterNodeDefinition(mergeDefinition{})
}