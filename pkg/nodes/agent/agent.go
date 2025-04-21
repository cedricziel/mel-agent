package agent

import "github.com/cedricziel/mel-agent/pkg/api"

type agentDefinition struct{}

func (agentDefinition) Meta() api.NodeType {
	return api.NodeType{
		Type:       "agent",
		Label:      "Agent",
		Icon:       "🤖",
		Category:   "LLM",
		Parameters: []api.ParameterDefinition{},
	}
}
func (agentDefinition) Execute(agentID string, node api.Node, input interface{}) (interface{}, error) {
	return input, nil
}
