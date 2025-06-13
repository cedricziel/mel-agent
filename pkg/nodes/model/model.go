package model

import (
	"github.com/cedricziel/mel-agent/pkg/api"
)

type modelDefinition struct{}

func (modelDefinition) Meta() api.NodeType {
	return api.NodeType{
		Type:     "model",
		Label:    "Model Configuration",
		Icon:     "🧠",
		Category: "Configuration",
		Parameters: []api.ParameterDefinition{
			api.NewStringParameter("provider", "Provider", true).
				WithDescription("AI model provider (e.g., openai, anthropic, cohere)").
				WithDefault("openai"),
			api.NewStringParameter("model", "Model", true).
				WithDescription("Specific model name (e.g., gpt-4, claude-3-sonnet)").
				WithDefault("gpt-4"),
			api.NewCredentialParameter("connectionId", "Connection", "openai", false).
				WithDescription("API credentials for the model provider"),
			api.NewNumberParameter("temperature", "Temperature", false).
				WithDescription("Controls randomness in responses (0.0-2.0)").
				WithDefault(0.7),
			api.NewIntegerParameter("maxTokens", "Max Tokens", false).
				WithDescription("Maximum number of tokens to generate").
				WithDefault(1000),
			api.NewNumberParameter("topP", "Top P", false).
				WithDescription("Nucleus sampling parameter (0.0-1.0)").
				WithDefault(0.9),
			api.NewIntegerParameter("seed", "Seed", false).
				WithDescription("Random seed for reproducible outputs"),
		},
	}
}

// ExecuteEnvelope returns model configuration data for use by agent nodes
func (d modelDefinition) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[interface{}]) (*api.Envelope[interface{}], error) {
	result := envelope.Clone()
	result.Trace = envelope.Trace.Next(node.ID)

	// Extract model configuration from node data
	modelConfig := map[string]interface{}{
		"provider":     getStringValue(node.Data, "provider", "openai"),
		"model":        getStringValue(node.Data, "model", "gpt-4"),
		"connectionId": getStringValue(node.Data, "connectionId", ""),
		"temperature":  getFloatValue(node.Data, "temperature", 0.7),
		"maxTokens":    getIntValue(node.Data, "maxTokens", 1000),
		"topP":         getFloatValue(node.Data, "topP", 0.9),
		"seed":         getIntValue(node.Data, "seed", 0),
	}

	result.Data = modelConfig
	return result, nil
}

func (modelDefinition) Initialize(mel api.Mel) error {
	return nil
}

// Helper functions to safely extract values from node data
func getStringValue(data map[string]interface{}, key string, defaultValue string) string {
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

func getFloatValue(data map[string]interface{}, key string, defaultValue float64) float64 {
	if val, ok := data[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case float32:
			return float64(v)
		case int:
			return float64(v)
		case int64:
			return float64(v)
		}
	}
	return defaultValue
}

func getIntValue(data map[string]interface{}, key string, defaultValue int) int {
	if val, ok := data[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case int64:
			return int(v)
		case float64:
			return int(v)
		case float32:
			return int(v)
		}
	}
	return defaultValue
}

func init() {
	api.RegisterNodeDefinition(modelDefinition{})
}

// assert that modelDefinition implements the interface
var _ api.NodeDefinition = (*modelDefinition)(nil)
