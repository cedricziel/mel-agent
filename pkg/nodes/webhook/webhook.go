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
			{Name: "method", Label: "HTTP Method", Type: "enum", Required: true, Default: "POST", Options: []string{"ANY", "GET", "POST", "PUT", "PATCH", "DELETE"}, Group: "HTTP", Description: "Allowed HTTP method for this webhook trigger"},
			{Name: "secret", Label: "Secret", Type: "string", Required: false, Default: "", Group: "Security", Description: "HMAC or token to validate requests"},
			{Name: "mode", Label: "Mode", Type: "enum", Required: true, Default: "async", Options: []string{"async", "sync"}, Group: "Execution", Description: "Async (enqueue run) or Sync (inline) execution"},
			{Name: "statusCode", Label: "Response Status", Type: "number", Required: false, Default: 202, Group: "Response", VisibilityCondition: "mode=='sync'", Description: "HTTP status code returned by trigger"},
			{Name: "responseBody", Label: "Response Body", Type: "string", Required: false, Default: "", Group: "Response", VisibilityCondition: "mode=='sync'", Description: "HTTP body returned by trigger"},
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
