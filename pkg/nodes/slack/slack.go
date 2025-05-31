package slack

import (
	"github.com/cedricziel/mel-agent/pkg/api"
)

type slackDefinition struct{}

func (slackDefinition) Meta() api.NodeType {
	return api.NodeType{
		Type:       "slack",
		Label:      "Slack Slash Command",
		Icon:       "💬",
		Category:   "Triggers",
		EntryPoint: true,
		Parameters: []api.ParameterDefinition{
			api.NewStringParameter("command", "Command", true).WithDefault("").WithGroup("Trigger").WithDescription("Slash command to respond to"),
			api.NewEnumParameter("mode", "Mode", []string{"async", "sync"}, true).WithDefault("async").WithGroup("Execution"),
			api.NewNumberParameter("statusCode", "Response Status", false).WithDefault(200).WithGroup("Response"),
			api.NewStringParameter("responseBody", "Response Body", false).WithDefault("").WithGroup("Response"),
		},
	}
}

// ExecuteEnvelope returns the input unchanged (Slack handling elsewhere).
func (d slackDefinition) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[interface{}]) (*api.Envelope[interface{}], error) {
	result := envelope.Clone()
	result.Trace = envelope.Trace.Next(node.ID)
	return result, nil
}

func (slackDefinition) Initialize(mel api.Mel) error {
	return nil
}

func init() {
	api.RegisterNodeDefinition(slackDefinition{})
}

// assert that slackDefinition implements both interfaces
var _ api.NodeDefinition = (*slackDefinition)(nil)
