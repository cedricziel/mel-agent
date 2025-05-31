package transform

import (
	api "github.com/cedricziel/mel-agent/pkg/api"
)

// transformDefinition provides the built-in "Transform" node.
type transformDefinition struct{}

// Meta returns metadata for the Transform node.
func (transformDefinition) Meta() api.NodeType {
	return api.NodeType{
		Type:     "transform",
		Label:    "Transform",
		Category: "Utility",
		Parameters: []api.ParameterDefinition{
			api.NewStringParameter("expression", "Expression", true).WithGroup("Settings").WithDescription("Transform input via expression"),
		},
	}
}

// ExecuteEnvelope applies the expression to the input envelope (currently passthrough).
func (d transformDefinition) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[interface{}]) (*api.Envelope[interface{}], error) {
	// TODO: Implement actual transformation logic based on expression
	result := envelope.Clone()
	result.Trace = envelope.Trace.Next(node.ID)
	return result, nil
}

func (transformDefinition) Initialize(mel api.Mel) error {
	return nil
}

func init() {
	api.RegisterNodeDefinition(transformDefinition{})
}

// assert that transformDefinition implements the interface
var _ api.NodeDefinition = (*transformDefinition)(nil)
