package switch_node

import (
	api "github.com/cedricziel/mel-agent/pkg/api"
)

// switchDefinition provides the built-in "Switch" node.
type switchDefinition struct{}

// Meta returns metadata for the Switch node.
func (switchDefinition) Meta() api.NodeType {
	return api.NodeType{
		Type:      "switch",
		Label:     "Switch",
		Category:  "Control",
		Branching: true,
		Parameters: []api.ParameterDefinition{
			api.NewStringParameter("expression", "Expression", true).WithGroup("Settings"),
		},
	}
}

// ExecuteEnvelope returns the input unchanged (branching logic handled elsewhere).
func (d switchDefinition) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[interface{}]) (*api.Envelope[interface{}], error) {
	result := envelope.Clone()
	result.Trace = envelope.Trace.Next(node.ID)
	return result, nil
}

func (switchDefinition) Initialize(mel api.Mel) error {
	return nil
}

func init() {
	api.RegisterNodeDefinition(switchDefinition{})
}

// assert that switchDefinition implements the interface
var _ api.NodeDefinition = (*switchDefinition)(nil)
