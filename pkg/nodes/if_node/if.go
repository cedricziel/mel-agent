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
			{Name: "condition", Label: "Condition", Type: "string", Required: true, Default: "", Group: "Expression", Description: "Boolean CEL expression to evaluate"},
		},
	}
}
func (ifDefinition) Execute(agentID string, node api.Node, input interface{}) (interface{}, error) {
	return input, nil
}
