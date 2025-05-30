package schedule

import "github.com/cedricziel/mel-agent/pkg/api"

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
func (scheduleDefinition) Execute(ctx api.ExecutionContext, node api.Node, input interface{}) (interface{}, error) {
	return input, nil
}

func (scheduleDefinition) Initialize(mel api.Mel) error {
	return nil
}

func init() {
	api.RegisterNodeDefinition(scheduleDefinition{})
}
