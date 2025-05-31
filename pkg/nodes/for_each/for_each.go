package for_each

import (
	api "github.com/cedricziel/mel-agent/pkg/api"
)

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

// ExecuteEnvelope iterates over input arrays. Currently passthrough.
func (d forEachDefinition) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[interface{}]) (*api.Envelope[interface{}], error) {
	result := envelope.Clone()
	result.Trace = envelope.Trace.Next(node.ID)
	return result, nil
}

func (forEachDefinition) Initialize(mel api.Mel) error {
	return nil
}

func init() {
	api.RegisterNodeDefinition(forEachDefinition{})
}

// assert that forEachDefinition implements the interface
var _ api.NodeDefinition = (*forEachDefinition)(nil)
