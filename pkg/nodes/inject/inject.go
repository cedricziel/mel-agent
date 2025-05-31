package inject

import (
	"github.com/cedricziel/mel-agent/pkg/api"
)

type injectDefinition struct{}

func (injectDefinition) Meta() api.NodeType {
	return api.NodeType{
		Type:       "inject",
		Label:      "Inject",
		Icon:       "▶️",
		Category:   "Debug",
		EntryPoint: true,
		Parameters: []api.ParameterDefinition{
			api.NewObjectParameter("payload", "Payload", false).WithDefault("{}").WithGroup("Inject").WithDescription("Data to inject"),
		},
	}
}

// ExecuteEnvelope returns configured payload as the output using envelopes.
func (d injectDefinition) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[interface{}]) (*api.Envelope[interface{}], error) {
	result := envelope.Clone()
	result.Trace = envelope.Trace.Next(node.ID)

	// Return configured payload as the output
	if p, ok := node.Data["payload"]; ok {
		result.Data = p
	} else {
		result.Data = nil
	}

	return result, nil
}

func (injectDefinition) Initialize(mel api.Mel) error {
	return nil
}

func init() {
	api.RegisterNodeDefinition(injectDefinition{})
}

// assert that injectDefinition implements the interface
var _ api.NodeDefinition = (*injectDefinition)(nil)
