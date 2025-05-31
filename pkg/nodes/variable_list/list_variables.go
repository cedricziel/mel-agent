package variable_list

import (
	"github.com/cedricziel/mel-agent/pkg/api"
)

// listVariablesDefinition provides the built-in "List Variables" node.
type listVariablesDefinition struct{}

// Meta returns metadata for the List Variables node.
func (listVariablesDefinition) Meta() api.NodeType {
	return api.NodeType{
		Type:     "variable_list",
		Label:    "List Variables",
		Icon:     "📑",
		Category: "Variables",
		Parameters: []api.ParameterDefinition{
			api.NewEnumParameter("scope", "Scope", []string{"run", "workflow", "global", "all"}, false).
				WithDefault("all").
				WithGroup("Settings").
				WithDescription("Variable scope to list: run, workflow, global, or all"),
			api.NewBooleanParameter("includeEmpty", "Include Empty Scopes", false).
				WithDefault(false).
				WithGroup("Settings").
				WithDescription("Include scopes that have no variables"),
		},
	}
}

// Execute lists variables from the specified scope(s).
func (listVariablesDefinition) Execute(ctx api.ExecutionContext, node api.Node, input interface{}) (interface{}, error) {
	scopeStr, _ := node.Data["scope"].(string)
	if scopeStr == "" {
		scopeStr = "all"
	}

	includeEmpty, _ := node.Data["includeEmpty"].(bool)

	// Create context for variable operations
	varCtx := api.CreateVariableContext(ctx.AgentID, ctx.RunID, node.ID)

	result := map[string]interface{}{
		"input": input,
	}

	if scopeStr == "all" {
		// List all scopes
		scopes := []api.VariableScope{api.RunScope, api.WorkflowScope, api.GlobalScope}
		scopeNames := []string{"run", "workflow", "global"}

		for i, scope := range scopes {
			scopeName := scopeNames[i]
			vars, err := api.ListVariables(varCtx, scope)
			if err != nil {
				return input, api.NewNodeError(node.ID, node.Type, "failed to list "+scopeName+" variables: "+err.Error())
			}

			if len(vars) > 0 || includeEmpty {
				result[scopeName+"_variables"] = vars
			}
		}
	} else {
		// List specific scope
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

		vars, err := api.ListVariables(varCtx, scope)
		if err != nil {
			return input, api.NewNodeError(node.ID, node.Type, "failed to list variables: "+err.Error())
		}

		result["variables"] = vars
		result["scope"] = scopeStr
	}

	return result, nil
}

func (listVariablesDefinition) Initialize(mel api.Mel) error {
	return nil
}

func init() {
	api.RegisterNodeDefinition(listVariablesDefinition{})
}

// assert that listVariablesDefinition implements the NodeDefinition interface
var _ api.NodeDefinition = (*listVariablesDefinition)(nil)