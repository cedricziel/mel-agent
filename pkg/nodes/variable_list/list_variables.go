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

// ExecuteEnvelope lists variables from the specified scope(s) using envelopes.
func (d listVariablesDefinition) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[interface{}]) (*api.Envelope[interface{}], error) {
	input := envelope.Data
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
				err := api.NewNodeError(node.ID, node.Type, "failed to list "+scopeName+" variables: "+err.Error())
				envelope.AddError(node.ID, "failed to list "+scopeName+" variables: "+err.Error(), err)
				return envelope, err
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
			err := api.NewNodeError(node.ID, node.Type, "invalid scope: "+scopeStr)
			envelope.AddError(node.ID, "invalid scope: "+scopeStr, err)
			return envelope, err
		}

		vars, err := api.ListVariables(varCtx, scope)
		if err != nil {
			err := api.NewNodeError(node.ID, node.Type, "failed to list variables: "+err.Error())
			envelope.AddError(node.ID, "failed to list variables: "+err.Error(), err)
			return envelope, err
		}

		result["variables"] = vars
		result["scope"] = scopeStr
	}

	// Create result envelope
	resultEnvelope := envelope.Clone()
	resultEnvelope.Trace = envelope.Trace.Next(node.ID)
	resultEnvelope.Data = result
	resultEnvelope.DataType = "object"

	return resultEnvelope, nil
}

func (listVariablesDefinition) Initialize(mel api.Mel) error {
	return nil
}

func init() {
	api.RegisterNodeDefinition(listVariablesDefinition{})
}

// assert that listVariablesDefinition implements the interface
var _ api.NodeDefinition = (*listVariablesDefinition)(nil)

