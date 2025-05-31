package variable_get

import (
	"github.com/cedricziel/mel-agent/pkg/api"
)

// getVariableDefinition provides the built-in "Get Variable" node.
type getVariableDefinition struct{}

// Meta returns metadata for the Get Variable node.
func (getVariableDefinition) Meta() api.NodeType {
	return api.NodeType{
		Type:     "variable_get",
		Label:    "Get Variable",
		Icon:     "📋",
		Category: "Variables",
		Parameters: []api.ParameterDefinition{
			api.NewStringParameter("key", "Variable Name", true).
				WithGroup("Settings").
				WithDescription("Name of the variable to retrieve"),
			api.NewEnumParameter("scope", "Scope", []string{"run", "workflow", "global"}, true).
				WithDefault("run").
				WithGroup("Settings").
				WithDescription("Variable scope: run (current execution), workflow (all runs), global (all workflows)"),
			api.NewStringParameter("defaultValue", "Default Value", false).
				WithGroup("Settings").
				WithDescription("Value to return if variable doesn't exist"),
			api.NewBooleanParameter("failIfMissing", "Fail If Missing", false).
				WithDefault(false).
				WithGroup("Settings").
				WithDescription("Throw error if variable doesn't exist"),
			api.NewEnumParameter("outputMode", "Output Mode", []string{"value_only", "with_metadata", "merge_input"}, false).
				WithDefault("merge_input").
				WithGroup("Settings").
				WithDescription("How to format the output"),
		},
	}
}

// Execute retrieves a variable from the specified scope.
func (getVariableDefinition) Execute(ctx api.ExecutionContext, node api.Node, input interface{}) (interface{}, error) {
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

	// Create context for variable operations
	varCtx := api.CreateVariableContext(ctx.AgentID, ctx.RunID, node.ID)

	// Get the variable
	value, exists, err := api.GetVariable(varCtx, scope, key)
	if err != nil {
		return input, api.NewNodeError(node.ID, node.Type, "failed to get variable: "+err.Error())
	}

	// Handle missing variable
	if !exists {
		failIfMissing, _ := node.Data["failIfMissing"].(bool)
		if failIfMissing {
			return input, api.NewNodeError(node.ID, node.Type, "variable '"+key+"' not found in "+scopeStr+" scope")
		}

		// Use default value if provided
		if defaultValue, hasDefault := node.Data["defaultValue"]; hasDefault {
			value = defaultValue
		} else {
			value = nil
		}
	}

	// Format output based on output mode
	outputMode, _ := node.Data["outputMode"].(string)
	if outputMode == "" {
		outputMode = "merge_input"
	}

	switch outputMode {
	case "value_only":
		return value, nil

	case "with_metadata":
		return map[string]interface{}{
			"value":  value,
			"exists": exists,
			"key":    key,
			"scope":  scopeStr,
		}, nil

	case "merge_input":
		// Merge with input data
		result := make(map[string]interface{})
		
		// Add input data if it's a map
		if inputMap, ok := input.(map[string]interface{}); ok {
			for k, v := range inputMap {
				result[k] = v
			}
		} else if input != nil {
			result["input"] = input
		}

		// Add the variable value
		result[key] = value
		return result, nil

	default:
		return input, api.NewNodeError(node.ID, node.Type, "invalid output mode: "+outputMode)
	}
}

func (getVariableDefinition) Initialize(mel api.Mel) error {
	return nil
}

func init() {
	api.RegisterNodeDefinition(getVariableDefinition{})
}

// assert that getVariableDefinition implements the NodeDefinition interface
var _ api.NodeDefinition = (*getVariableDefinition)(nil)