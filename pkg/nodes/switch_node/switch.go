package switch_node

import api "github.com/cedricziel/mel-agent/pkg/api"

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

// Execute returns the input unchanged (branching logic handled elsewhere).
func (switchDefinition) Execute(ctx api.ExecutionContext, node api.Node, input interface{}) (interface{}, error) {
	return input, nil
}

func (switchDefinition) Initialize(mel api.Mel) error {
	return nil
}

func init() {
	api.RegisterNodeDefinition(switchDefinition{})
}
