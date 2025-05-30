package for_each

import api "github.com/cedricziel/mel-agent/pkg/api"

// forEachDefinition provides the built-in "For Each" node.
type forEachDefinition struct{}

// Meta returns metadata for the For Each node.
func (forEachDefinition) Meta() api.NodeType {
	return api.NodeType{
		Type:     "for_each",
		Label:    "For Each",
		Category: "Control",
		Parameters: []api.ParameterDefinition{
			api.NewStringParameter("path", "Array Path", true).WithGroup("Settings").WithDescription("JSONPath to array"),
		},
	}
}

// Execute iterates over input arrays. Currently passthrough.
func (forEachDefinition) Execute(ctx api.ExecutionContext, node api.Node, input interface{}) (interface{}, error) {
	return input, nil
}

func (forEachDefinition) Initialize(mel api.Mel) error {
	return nil
}

func init() {
	api.RegisterNodeDefinition(forEachDefinition{})
}
