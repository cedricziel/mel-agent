package openai_model

import (
	"github.com/cedricziel/mel-agent/pkg/api"
)

// OpenAIModelNode represents an OpenAI model configuration node
type OpenAIModelNode struct{}

// Meta returns the node type metadata
func (n *OpenAIModelNode) Meta() api.NodeType {
	return api.NodeType{
		Type:     "openai_model",
		Label:    "OpenAI Model",
		Icon:     "🤖",
		Category: "Configuration",
		Parameters: []api.ParameterDefinition{
			api.NewEnumParameter("model", "Model", []string{"gpt-4", "gpt-3.5-turbo", "gpt-4-turbo", "gpt-4o", "gpt-4o-mini"}, true).
				WithDefault("gpt-4").
				WithDescription("OpenAI model to use"),
			api.NewNumberParameter("temperature", "Temperature", false).
				WithDefault(0.7).
				WithDescription("Controls randomness in output (0.0 to 2.0)"),
			api.NewIntegerParameter("maxTokens", "Max Tokens", false).
				WithDefault(1000).
				WithDescription("Maximum number of tokens to generate"),
			api.NewCredentialParameter("credential", "API Credential", "openai_api_key", true).
				WithDescription("OpenAI API key credential"),
		},
	}
}

// Initialize sets up the node
func (n *OpenAIModelNode) Initialize(mel api.Mel) error {
	return nil
}

// ExecuteEnvelope executes the node (config nodes typically don't execute)
func (n *OpenAIModelNode) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[any]) (*api.Envelope[any], error) {
	// Config nodes typically don't execute, they provide configuration
	return envelope, nil
}

func init() {
	api.RegisterNodeDefinition(&OpenAIModelNode{})
}
