package inject

import "github.com/cedricziel/mel-agent/pkg/api"

type injectDefinition struct{}

func (injectDefinition) Meta() api.NodeType {
	return api.NodeType{
		Type:       "inject",
		Label:      "Inject",
		Icon:       "▶️",
		Category:   "Debug",
		EntryPoint: true,
		Parameters: []api.ParameterDefinition{
			{Name: "payload", Label: "Payload", Type: "json", Required: false, Default: "{}", Group: "Inject", Description: "Data to inject"},
		},
	}
}

func (injectDefinition) Execute(agentID string, node api.Node, input interface{}) (interface{}, error) {
	// Return configured payload as the output
	if p, ok := node.Data["payload"]; ok {
		return p, nil
	}
	return nil, nil
}
