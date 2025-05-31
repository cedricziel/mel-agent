package schedule

import (
	"github.com/cedricziel/mel-agent/pkg/api"
)

type scheduleDefinition struct{}

func (scheduleDefinition) Meta() api.NodeType {
	return api.NodeType{
		Type:       "schedule",
		Label:      "Schedule",
		Icon:       "🗓️",
		Category:   "Triggers",
		EntryPoint: true,
		Parameters: []api.ParameterDefinition{
			api.NewStringParameter("cron", "Cron Expression", true).WithDefault("").WithGroup("Schedule").WithDescription("Cron schedule to run"),
			api.NewEnumParameter("mode", "Mode", []string{"async", "sync"}, true).WithDefault("async").WithGroup("Execution"),
			api.NewNumberParameter("statusCode", "Response Status", false).WithDefault(202).WithGroup("Response"),
			api.NewStringParameter("responseBody", "Response Body", false).WithDefault("").WithGroup("Response"),
		},
	}
}

// ExecuteEnvelope returns the input unchanged (schedule logic handled elsewhere).
func (d scheduleDefinition) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[interface{}]) (*api.Envelope[interface{}], error) {
	result := envelope.Clone()
	result.Trace = envelope.Trace.Next(node.ID)
	return result, nil
}

func (scheduleDefinition) Initialize(mel api.Mel) error {
	return nil
}

func init() {
	api.RegisterNodeDefinition(scheduleDefinition{})
}

// assert that scheduleDefinition implements both interfaces
var _ api.NodeDefinition = (*scheduleDefinition)(nil)
