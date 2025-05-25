package timer

import "github.com/cedricziel/mel-agent/pkg/api"

type timerDefinition struct{}

func (timerDefinition) Meta() api.NodeType {
	return api.NodeType{
		Type:       "timer",
		Label:      "Timer",
		Icon:       "⏰",
		Category:   "Triggers",
		EntryPoint: true,
		Parameters: []api.ParameterDefinition{
			{Name: "mode", Label: "Mode", Type: "enum", Required: true, Default: "async", Options: []string{"async", "sync"}, Group: "Execution", Description: "Async (enqueue run) or Sync (inline) execution"},
			{Name: "statusCode", Label: "Response Status", Type: "number", Required: false, Default: 202, Group: "Response", Description: "HTTP status code returned by trigger"},
			{Name: "responseBody", Label: "Response Body", Type: "string", Required: false, Default: "", Group: "Response", Description: "HTTP body returned by trigger"},
		},
	}
}

func (timerDefinition) Execute(ctx api.ExecutionContext, node api.Node, input interface{}) (interface{}, error) {
	// Delegate to existing executor
	return input, nil
}

func (timerDefinition) Initialize(mel api.Mel) error {
	return nil
}
