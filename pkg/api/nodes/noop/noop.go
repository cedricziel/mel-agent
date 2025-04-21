package noop

import internalapi "github.com/cedricziel/mel-agent/internal/api"

// noopDefinition provides the built-in "No-Op" node.
type noopDefinition struct{}

// Meta returns metadata for the No-Op node.
func (noopDefinition) Meta() internalapi.NodeType {
   return internalapi.NodeType{Type: "noop", Label: "No-Op", Category: "Control"}
}

// Execute returns the input unchanged.
func (noopDefinition) Execute(agentID string, node internalapi.Node, input interface{}) (interface{}, error) {
   return input, nil
}

func init() {
   internalapi.RegisterNodeDefinition(noopDefinition{})
}