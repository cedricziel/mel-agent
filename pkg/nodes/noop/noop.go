package noop

import api "github.com/cedricziel/mel-agent/pkg/api"

// noopDefinition provides the built-in "No-Op" node.
type noopDefinition struct{}

// Meta returns metadata for the No-Op node.
func (noopDefinition) Meta() api.NodeType {
	return api.NodeType{Type: "noop", Label: "No-Op", Category: "Control"}
}

// Execute returns the input unchanged.
func (noopDefinition) Execute(ctx api.ExecutionContext, node api.Node, input interface{}) (interface{}, error) {
	return input, nil
}

func (noopDefinition) Initialize(mel api.Mel) error {
	return nil
}

func init() {
	api.RegisterNodeDefinition(noopDefinition{})
}
