package slack

import "github.com/cedricziel/mel-agent/pkg/api"

type slackDefinition struct{}

func (slackDefinition) Meta() api.NodeType {
	return api.NodeType{
		Type:       "slack",
		Label:      "Slack Slash Command",
		Icon:       "💬",
		Category:   "Triggers",
		EntryPoint: true,
		Parameters: []api.ParameterDefinition{
			{Name: "command", Label: "Command", Type: "string", Required: true, Default: "", Group: "Trigger", Description: "Slash command to respond to"},
			{Name: "mode", Label: "Mode", Type: "enum", Required: true, Default: "async", Options: []string{"async", "sync"}, Group: "Execution"},
			{Name: "statusCode", Label: "Response Status", Type: "number", Required: false, Default: 200, Group: "Response"},
			{Name: "responseBody", Label: "Response Body", Type: "string", Required: false, Default: "", Group: "Response"},
		},
	}
}
func (slackDefinition) Execute(ctx api.ExecutionContext, node api.Node, input interface{}) (interface{}, error) {
	return input, nil
}

func (slackDefinition) Initialize(mel api.Mel) error {
	return nil
}

func init() {
	api.RegisterNodeDefinition(slackDefinition{})
}
