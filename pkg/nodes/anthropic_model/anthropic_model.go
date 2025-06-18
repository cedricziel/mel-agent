package anthropic_model

import (
	"github.com/cedricziel/mel-agent/pkg/api"
)

// AnthropicModelNode represents an Anthropic model configuration node
// It implements both ActionNode and ModelNode interfaces
type AnthropicModelNode struct{}

// Ensure AnthropicModelNode implements both ActionNode and ModelNode
var _ api.ActionNode = (*AnthropicModelNode)(nil)
var _ api.ModelNode = (*AnthropicModelNode)(nil)

// Meta returns the node type metadata
func (n *AnthropicModelNode) Meta() api.NodeType {
	return api.NodeType{
		Type:     "anthropic_model",
		Label:    "Anthropic Model",
		Icon:     "🧠",
		Category: "Configuration",
		Parameters: []api.ParameterDefinition{
			api.NewEnumParameter("model", "Model", []string{"claude-3-5-sonnet-20241022", "claude-3-5-haiku-20241022", "claude-3-opus-20240229", "claude-3-sonnet-20240229", "claude-3-haiku-20240307"}, true).
				WithDefault("claude-3-5-sonnet-20241022").
				WithDescription("Anthropic model to use"),
			api.NewNumberParameter("temperature", "Temperature", false).
				WithDefault(0.7).
				WithDescription("Controls randomness in output (0.0 to 1.0)"),
			api.NewIntegerParameter("maxTokens", "Max Tokens", false).
				WithDefault(1000).
				WithDescription("Maximum number of tokens to generate"),
			api.NewCredentialParameter("credential", "API Credential", "anthropic_api_key", true).
				WithDescription("Anthropic API key credential"),
		},
	}
}

// Initialize sets up the node
func (n *AnthropicModelNode) Initialize(mel api.Mel) error {
	return nil
}

// ExecuteEnvelope executes the node (config nodes typically don't execute)
func (n *AnthropicModelNode) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[any]) (*api.Envelope[any], error) {
	// Config nodes typically don't execute, they provide configuration
	return envelope, nil
}

// InteractWith handles model interaction with given input (ModelNode interface)
func (n *AnthropicModelNode) InteractWith(ctx api.ExecutionContext, node api.Node, input string, options map[string]any) (string, error) {
	// TODO: Implement actual Anthropic API interaction
	// For now, return a placeholder response
	return "Anthropic model response: " + input, nil
}

func init() {
	api.RegisterNodeDefinition(&AnthropicModelNode{})
}
