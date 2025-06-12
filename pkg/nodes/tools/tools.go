package tools

import (
	"github.com/cedricziel/mel-agent/pkg/api"
)

type toolsDefinition struct{}

func (toolsDefinition) Meta() api.NodeType {
	return api.NodeType{
		Type:     "tools",
		Label:    "Tools Configuration",
		Icon:     "🔧",
		Category: "Configuration",
		Parameters: []api.ParameterDefinition{
			api.NewArrayParameter("toolList", "Available Tools", true).
				WithDescription("List of tools available to the agent"),
			api.NewBooleanParameter("allowCodeExecution", "Allow Code Execution", false).
				WithDescription("Enable code execution capabilities").
				WithDefault(false),
			api.NewBooleanParameter("allowWebSearch", "Allow Web Search", false).
				WithDescription("Enable web search capabilities").
				WithDefault(false),
			api.NewBooleanParameter("allowFileAccess", "Allow File Access", false).
				WithDescription("Enable file system access").
				WithDefault(false),
			api.NewObjectParameter("customTools", "Custom Tools", false).
				WithDescription("Custom tool definitions and configurations"),
			api.NewStringParameter("toolRestrictions", "Tool Restrictions", false).
				WithDescription("Restrictions or limitations on tool usage"),
		},
	}
}

// ExecuteEnvelope returns tools configuration data for use by agent nodes
func (d toolsDefinition) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[interface{}]) (*api.Envelope[interface{}], error) {
	result := envelope.Clone()
	result.Trace = envelope.Trace.Next(node.ID)

	// Extract tools configuration from node data
	toolsConfig := map[string]interface{}{
		"toolList":           getArrayValue(node.Data, "toolList", []interface{}{}),
		"allowCodeExecution": getBoolValue(node.Data, "allowCodeExecution", false),
		"allowWebSearch":     getBoolValue(node.Data, "allowWebSearch", false),
		"allowFileAccess":    getBoolValue(node.Data, "allowFileAccess", false),
		"customTools":        getObjectValue(node.Data, "customTools", map[string]interface{}{}),
		"toolRestrictions":   getStringValue(node.Data, "toolRestrictions", ""),
	}

	result.Data = toolsConfig
	return result, nil
}

func (toolsDefinition) Initialize(mel api.Mel) error {
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

func getArrayValue(data map[string]interface{}, key string, defaultValue []interface{}) []interface{} {
	if val, ok := data[key]; ok {
		if arr, ok := val.([]interface{}); ok {
			return arr
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
	api.RegisterNodeDefinition(toolsDefinition{})
}

// assert that toolsDefinition implements the interface
var _ api.NodeDefinition = (*toolsDefinition)(nil)
