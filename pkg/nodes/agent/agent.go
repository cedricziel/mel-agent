package agent

import (
	"github.com/cedricziel/mel-agent/pkg/api"
)

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

// ExecuteEnvelope returns the input unchanged (agent logic handled elsewhere).
func (d agentDefinition) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[interface{}]) (*api.Envelope[interface{}], error) {
	result := envelope.Clone()
	result.Trace = envelope.Trace.Next(node.ID)
	return result, nil
}

func (agentDefinition) Initialize(mel api.Mel) error {
	return nil
}

func init() {
	api.RegisterNodeDefinition(agentDefinition{})
}

// assert that agentDefinition implements both interfaces
var _ api.NodeDefinition = (*agentDefinition)(nil)
