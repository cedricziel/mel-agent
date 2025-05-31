package noop

import api "github.com/cedricziel/mel-agent/pkg/api"

// noopDefinition provides the built-in "No-Op" node.
type noopDefinition struct{}

// Meta returns metadata for the No-Op node.
func (noopDefinition) Meta() api.NodeType {
	return api.NodeType{Type: "noop", Label: "No-Op", Category: "Control"}
}

// ExecuteEnvelope returns the input envelope unchanged.
func (d noopDefinition) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[interface{}]) (*api.Envelope[interface{}], error) {
	result := envelope.Clone()
	result.Trace = envelope.Trace.Next(node.ID)
	return result, nil
}

func (noopDefinition) Initialize(mel api.Mel) error {
	return nil
}

func init() {
	api.RegisterNodeDefinition(noopDefinition{})
}

// assert that noopDefinition implements the NodeDefinition interface
var _ api.NodeDefinition = (*noopDefinition)(nil)
