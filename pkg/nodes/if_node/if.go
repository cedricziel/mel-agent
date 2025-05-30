package if_node

import "github.com/cedricziel/mel-agent/pkg/api"

type ifDefinition struct{}

func (ifDefinition) Meta() api.NodeType {
	return api.NodeType{
		Type:      "if",
		Label:     "If",
		Icon:      "❓",
		Category:  "Basic",
		Branching: true,
		Parameters: []api.ParameterDefinition{
			api.NewStringParameter("condition", "Condition", true).
				WithGroup("Expression").
				WithDescription("Boolean CEL expression to evaluate"),
		},
	}
}
func (ifDefinition) Execute(ctx api.ExecutionContext, node api.Node, input interface{}) (interface{}, error) {
	return input, nil
}

func (ifDefinition) Initialize(mel api.Mel) error {
	return nil
}

func init() {
	api.RegisterNodeDefinition(ifDefinition{})
}
