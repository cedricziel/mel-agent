package webhook

import (
	"github.com/cedricziel/mel-agent/pkg/api"
)

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

// ExecuteEnvelope for webhook is handled via HTTP endpoint, default no-op
func (d webhookDefinition) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[interface{}]) (*api.Envelope[interface{}], error) {
	result := envelope.Clone()
	result.Trace = envelope.Trace.Next(node.ID)
	return result, nil
}

func (webhookDefinition) Initialize(mel api.Mel) error {
	return nil
}

func init() {
	api.RegisterNodeDefinition(webhookDefinition{})
}

// assert that webhookDefinition implements both interfaces
var _ api.NodeDefinition = (*webhookDefinition)(nil)
