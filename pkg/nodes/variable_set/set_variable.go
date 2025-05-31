package variable_set

import (
	"github.com/cedricziel/mel-agent/pkg/api"
)

// setVariableDefinition provides the built-in "Set Variable" node.
type setVariableDefinition struct{}

// Meta returns metadata for the Set Variable node.
func (setVariableDefinition) Meta() api.NodeType {
	return api.NodeType{
		Type:     "variable_set",
		Label:    "Set Variable",
		Icon:     "📝",
		Category: "Variables",
		Parameters: []api.ParameterDefinition{
			api.NewStringParameter("key", "Variable Name", true).
				WithGroup("Settings").
				WithDescription("Name of the variable to set"),
			api.NewEnumParameter("scope", "Scope", []string{"run", "workflow", "global"}, true).
				WithDefault("run").
				WithGroup("Settings").
				WithDescription("Variable scope: run (current execution), workflow (all runs), global (all workflows)"),
			api.NewStringParameter("value", "Value", false).
				WithGroup("Settings").
				WithDescription("Static value to assign (leave empty to use input data)"),
			api.NewStringParameter("valueFrom", "Value From Input", false).
				WithGroup("Settings").
				WithDescription("JSONPath to extract value from input (e.g., 'data.result')").
				WithVisibilityCondition("value==''"),
			api.NewBooleanParameter("passthrough", "Pass Through Input", false).
				WithDefault(true).
				WithGroup("Settings").
				WithDescription("Pass the input data through as output"),
		},
	}
}

// Execute sets a variable in the specified scope.
func (setVariableDefinition) Execute(ctx api.ExecutionContext, node api.Node, input interface{}) (interface{}, error) {
	// Get configuration
	key, _ := node.Data["key"].(string)
	if key == "" {
		return input, api.NewNodeError(node.ID, node.Type, "variable name is required")
	}

	scopeStr, _ := node.Data["scope"].(string)
	if scopeStr == "" {
		scopeStr = "run" // default scope
	}

	var scope api.VariableScope
	switch scopeStr {
	case "run":
		scope = api.RunScope
	case "workflow":
		scope = api.WorkflowScope
	case "global":
		scope = api.GlobalScope
	default:
		return input, api.NewNodeError(node.ID, node.Type, "invalid scope: "+scopeStr)
	}

	// Determine the value to set
	var value interface{}
	if staticValue, exists := node.Data["value"]; exists && staticValue != "" {
		// Use static value
		value = staticValue
	} else if valueFrom, exists := node.Data["valueFrom"].(string); exists && valueFrom != "" {
		// Extract value from input using JSONPath-like syntax
		value = extractValueFromInput(input, valueFrom)
	} else {
		// Use entire input as value
		value = input
	}

	// Create context for variable operations
	varCtx := api.CreateVariableContext(ctx.AgentID, ctx.RunID, node.ID)

	// Set the variable
	if err := api.SetVariable(varCtx, scope, key, value); err != nil {
		return input, api.NewNodeError(node.ID, node.Type, "failed to set variable: "+err.Error())
	}

	// Determine output
	passthrough, _ := node.Data["passthrough"].(bool)
	if passthrough {
		return input, nil
	}

	// Return information about the variable that was set
	return map[string]interface{}{
		"variable_set": true,
		"key":          key,
		"scope":        scopeStr,
		"value":        value,
	}, nil
}

// extractValueFromInput extracts a value from input using a simple JSONPath-like syntax
func extractValueFromInput(input interface{}, path string) interface{} {
	if path == "" {
		return input
	}

	// Simple implementation - just support dot notation for now
	// e.g., "data.result" or "response.data.items"
	current := input
	
	// Handle the case where input is not a map
	if path == "." {
		return current
	}

	// Split by dots and traverse
	parts := splitPath(path)
	for _, part := range parts {
		if part == "" {
			continue
		}

		switch v := current.(type) {
		case map[string]interface{}:
			current = v[part]
		case map[interface{}]interface{}:
			current = v[part]
		default:
			// Can't traverse further
			return nil
		}

		if current == nil {
			return nil
		}
	}

	return current
}

// splitPath splits a dot-separated path
func splitPath(path string) []string {
	if path == "" {
		return nil
	}
	
	var parts []string
	var current string
	
	for _, char := range path {
		if char == '.' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	
	if current != "" {
		parts = append(parts, current)
	}
	
	return parts
}

func (setVariableDefinition) Initialize(mel api.Mel) error {
	return nil
}

func init() {
	api.RegisterNodeDefinition(setVariableDefinition{})
}

// assert that setVariableDefinition implements the NodeDefinition interface
var _ api.NodeDefinition = (*setVariableDefinition)(nil)
