package memory

import (
	"github.com/cedricziel/mel-agent/pkg/api"
)

type memoryDefinition struct{}

func (memoryDefinition) Meta() api.NodeType {
	return api.NodeType{
		Type:     "memory",
		Label:    "Memory Configuration",
		Icon:     "🧠",
		Category: "Configuration",
		Parameters: []api.ParameterDefinition{
			api.NewEnumParameter("memoryType", "Memory Type", []string{"short_term", "long_term", "semantic", "episodic"}, true).
				WithDescription("Type of memory to configure").
				WithDefault("short_term"),
			api.NewIntegerParameter("maxMessages", "Max Messages", false).
				WithDescription("Maximum number of messages to retain in memory").
				WithDefault(100),
			api.NewIntegerParameter("maxTokens", "Max Tokens", false).
				WithDescription("Maximum number of tokens to retain in memory").
				WithDefault(8000),
			api.NewBooleanParameter("enableSummarization", "Enable Summarization", false).
				WithDescription("Automatically summarize old conversations").
				WithDefault(true),
			api.NewIntegerParameter("summarizationThreshold", "Summarization Threshold", false).
				WithDescription("Number of messages after which to trigger summarization").
				WithDefault(50),
			api.NewEnumParameter("retentionPolicy", "Retention Policy", []string{"fifo", "importance_based", "time_based", "manual"}, false).
				WithDescription("Policy for removing old memories").
				WithDefault("fifo"),
			api.NewObjectParameter("vectorStore", "Vector Store Configuration", false).
				WithDescription("Configuration for semantic memory storage"),
			api.NewBooleanParameter("enablePersonalization", "Enable Personalization", false).
				WithDescription("Learn and adapt to user preferences").
				WithDefault(false),
		},
	}
}

// ExecuteEnvelope returns memory configuration data for use by agent nodes
func (d memoryDefinition) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[interface{}]) (*api.Envelope[interface{}], error) {
	result := envelope.Clone()
	result.Trace = envelope.Trace.Next(node.ID)

	// Extract memory configuration from node data
	memoryConfig := map[string]interface{}{
		"memoryType":             getStringValue(node.Data, "memoryType", "short_term"),
		"maxMessages":            getIntValue(node.Data, "maxMessages", 100),
		"maxTokens":              getIntValue(node.Data, "maxTokens", 8000),
		"enableSummarization":    getBoolValue(node.Data, "enableSummarization", true),
		"summarizationThreshold": getIntValue(node.Data, "summarizationThreshold", 50),
		"retentionPolicy":        getStringValue(node.Data, "retentionPolicy", "fifo"),
		"vectorStore":            getObjectValue(node.Data, "vectorStore", map[string]interface{}{}),
		"enablePersonalization":  getBoolValue(node.Data, "enablePersonalization", false),
	}

	result.Data = memoryConfig
	return result, nil
}

func (memoryDefinition) Initialize(mel api.Mel) error {
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

func getBoolValue(data map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := data[key]; ok {
		if b, ok := val.(bool); ok {
			return b
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

func getObjectValue(data map[string]interface{}, key string, defaultValue map[string]interface{}) map[string]interface{} {
	if val, ok := data[key]; ok {
		if obj, ok := val.(map[string]interface{}); ok {
			return obj
		}
	}
	return defaultValue
}

func init() {
	api.RegisterNodeDefinition(memoryDefinition{})
}

// assert that memoryDefinition implements the interface
var _ api.NodeDefinition = (*memoryDefinition)(nil)
