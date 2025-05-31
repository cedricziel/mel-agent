package timer

import (
	"github.com/cedricziel/mel-agent/pkg/api"
)

type timerDefinition struct{}

func (timerDefinition) Meta() api.NodeType {
	return api.NodeType{
		Type:       "timer",
		Label:      "Timer",
		Icon:       "⏰",
		Category:   "Triggers",
		EntryPoint: true,
		Parameters: []api.ParameterDefinition{
			api.NewEnumParameter("mode", "Mode", []string{"async", "sync"}, true).WithDefault("async").WithGroup("Execution").WithDescription("Async (enqueue run) or Sync (inline) execution"),
			api.NewNumberParameter("statusCode", "Response Status", false).WithDefault(202).WithGroup("Response").WithDescription("HTTP status code returned by trigger"),
			api.NewStringParameter("responseBody", "Response Body", false).WithDefault("").WithGroup("Response").WithDescription("HTTP body returned by trigger"),
		},
	}
}

// ExecuteEnvelope delegates to existing executor.
func (d timerDefinition) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[interface{}]) (*api.Envelope[interface{}], error) {
	result := envelope.Clone()
	result.Trace = envelope.Trace.Next(node.ID)
	return result, nil
}

func (timerDefinition) Initialize(mel api.Mel) error {
	return nil
}

func init() {
	api.RegisterNodeDefinition(timerDefinition{})
}

// assert that timerDefinition implements both interfaces
var _ api.NodeDefinition = (*timerDefinition)(nil)
