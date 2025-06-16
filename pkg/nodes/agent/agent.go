package agent

import (
	"fmt"

	"github.com/cedricziel/mel-agent/pkg/api"
)

type agentDefinition struct{}

func (agentDefinition) Meta() api.NodeType {
	return api.NodeType{
		Type:     "agent",
		Label:    "Agent",
		Icon:     "🤖",
		Category: "LLM",
		Parameters: []api.ParameterDefinition{
			api.NewStringParameter("systemPrompt", "System Prompt", false).
				WithDescription("Instructions that define the agent's behavior and role").
				WithDefault("You are a helpful AI assistant."),
			api.NewNodeReferenceParameter("modelConfig", "Model Configuration", false).
				WithDescription("Reference to a model configuration node"),
			api.NewNodeReferenceParameter("toolsConfig", "Tools Configuration", false).
				WithDescription("Reference to a tools configuration node"),
			api.NewNodeReferenceParameter("memoryConfig", "Memory Configuration", false).
				WithDescription("Reference to a memory configuration node"),
			api.NewObjectParameter("additionalSettings", "Additional Settings", false).
				WithDescription("Any additional agent-specific settings"),
		},
	}
}

// ExecuteEnvelope processes the agent with its configuration nodes.
func (d agentDefinition) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[interface{}]) (*api.Envelope[interface{}], error) {
	result := envelope.Clone()
	result.Trace = envelope.Trace.Next(node.ID)

	// Extract system prompt
	systemPrompt := getStringValue(node.Data, "systemPrompt", "You are a helpful AI assistant.")

	// Build agent configuration
	agentConfig := map[string]interface{}{
		"systemPrompt": systemPrompt,
		"nodeId":       node.ID,
	}

	// Resolve model configuration if provided
	if modelNodeId := getStringValue(node.Data, "modelConfig", ""); modelNodeId != "" {
		modelConfig, err := d.resolveNodeReference(ctx, modelNodeId, "model")
		if err != nil {
			return nil, fmt.Errorf("failed to resolve model configuration: %w", err)
		}
		agentConfig["model"] = modelConfig
	}

	// Resolve tools configuration if provided
	if toolsNodeId := getStringValue(node.Data, "toolsConfig", ""); toolsNodeId != "" {
		toolsConfig, err := d.resolveNodeReference(ctx, toolsNodeId, "tools")
		if err != nil {
			return nil, fmt.Errorf("failed to resolve tools configuration: %w", err)
		}
		agentConfig["tools"] = toolsConfig
	}

	// Resolve memory configuration if provided
	if memoryNodeId := getStringValue(node.Data, "memoryConfig", ""); memoryNodeId != "" {
		memoryConfig, err := d.resolveNodeReference(ctx, memoryNodeId, "memory")
		if err != nil {
			return nil, fmt.Errorf("failed to resolve memory configuration: %w", err)
		}
		agentConfig["memory"] = memoryConfig
	}

	// Include additional settings
	if additionalSettings := getObjectValue(node.Data, "additionalSettings", nil); additionalSettings != nil {
		agentConfig["additionalSettings"] = additionalSettings
	}

	// Add input data to the agent configuration
	agentConfig["input"] = envelope.Data

	result.Data = agentConfig
	return result, nil
}

// resolveNodeReference resolves a node reference by calling the platform
// Note: This is a placeholder - actual implementation would need access to workflow context
func (d agentDefinition) resolveNodeReference(ctx api.ExecutionContext, nodeId string, expectedType string) (interface{}, error) {
	// TODO: Implement actual node resolution logic
	// This would typically involve:
	// 1. Looking up the node in the current workflow
	// 2. Executing the referenced node to get its configuration
	// 3. Validating that the node is of the expected type
	// 4. Returning the configuration data

	// For now, return a placeholder
	return map[string]interface{}{
		"nodeId":   nodeId,
		"nodeType": expectedType,
		"resolved": false,
		"message":  "Node resolution not yet implemented",
	}, nil
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

func getObjectValue(data map[string]interface{}, key string, defaultValue map[string]interface{}) map[string]interface{} {
	if val, ok := data[key]; ok {
		if obj, ok := val.(map[string]interface{}); ok {
			return obj
		}
	}
	return defaultValue
}

func (agentDefinition) Initialize(mel api.Mel) error {
	return nil
}

func init() {
	api.RegisterNodeDefinition(agentDefinition{})
}

// assert that agentDefinition implements both interfaces
var _ api.NodeDefinition = (*agentDefinition)(nil)
