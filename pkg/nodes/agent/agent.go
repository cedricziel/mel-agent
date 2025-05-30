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
func (agentDefinition) Execute(ctx api.ExecutionContext, node api.Node, input interface{}) (interface{}, error) {
	return input, nil
}

func (agentDefinition) Initialize(mel api.Mel) error {
	return nil
}

func init() {
	api.RegisterNodeDefinition(agentDefinition{})
}

// assert that agentDefinition implements the NodeDefinition interface
var _ api.NodeDefinition = (*agentDefinition)(nil)
