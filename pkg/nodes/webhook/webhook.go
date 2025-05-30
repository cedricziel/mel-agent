package webhook

import "github.com/cedricziel/mel-agent/pkg/api"

type webhookDefinition struct{}

func (webhookDefinition) Meta() api.NodeType {
	return api.NodeType{
		Type:       "webhook",
		Label:      "Webhook",
		Icon:       "🔌",
		Category:   "Triggers",
		EntryPoint: true,
		Parameters: []api.ParameterDefinition{
			// Allowed HTTP method for this webhook (ANY to accept all)
			api.NewEnumParameter("method", "HTTP Method", []string{"ANY", "GET", "POST", "PUT", "PATCH", "DELETE"}, true).WithDefault("POST").WithGroup("HTTP").WithDescription("Allowed HTTP method for this webhook trigger"),
			api.NewStringParameter("secret", "Secret", false).WithDefault("").WithGroup("Security").WithDescription("HMAC or token to validate requests"),
			api.NewEnumParameter("mode", "Mode", []string{"async", "sync"}, true).WithDefault("async").WithGroup("Execution").WithDescription("Async (enqueue run) or Sync (inline) execution"),
			api.NewNumberParameter("statusCode", "Response Status", false).WithDefault(202).WithGroup("Response").WithVisibilityCondition("mode=='sync'").WithDescription("HTTP status code returned by trigger"),
			api.NewStringParameter("responseBody", "Response Body", false).WithDefault("").WithGroup("Response").WithVisibilityCondition("mode=='sync'").WithDescription("HTTP body returned by trigger"),
		},
	}
}
func (webhookDefinition) Execute(ctx api.ExecutionContext, node api.Node, input interface{}) (interface{}, error) {
	// Execution for webhook is handled via HTTP endpoint, default no-op
	return input, nil
}

func (webhookDefinition) Initialize(mel api.Mel) error {
	return nil
}

func init() {
	api.RegisterNodeDefinition(webhookDefinition{})
}
