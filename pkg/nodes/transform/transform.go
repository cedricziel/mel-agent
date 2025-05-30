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

// Execute applies the expression to the input (currently passthrough).
func (transformDefinition) Execute(ctx api.ExecutionContext, node api.Node, input interface{}) (interface{}, error) {
	return input, nil
}

func (transformDefinition) Initialize(mel api.Mel) error {
	return nil
}

func init() {
	api.RegisterNodeDefinition(transformDefinition{})
}
