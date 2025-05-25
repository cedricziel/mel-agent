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
			{Name: "cron", Label: "Cron Expression", Type: "string", Required: true, Default: "", Group: "Schedule", Description: "Cron schedule to run"},
			{Name: "mode", Label: "Mode", Type: "enum", Required: true, Default: "async", Options: []string{"async", "sync"}, Group: "Execution"},
			{Name: "statusCode", Label: "Response Status", Type: "number", Required: false, Default: 202, Group: "Response"},
			{Name: "responseBody", Label: "Response Body", Type: "string", Required: false, Default: "", Group: "Response"},
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
