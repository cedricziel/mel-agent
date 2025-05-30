package set_variable

import (
	api "github.com/cedricziel/mel-agent/pkg/api"
)

// setVariableDefinition provides the built-in "Set Variable" node.
type setVariableDefinition struct{}

// Meta returns metadata for the Set Variable node.
func (setVariableDefinition) Meta() api.NodeType {
	return api.NodeType{
		Type:     "set_variable",
		Label:    "Set Variable",
		Category: "Utility",
		Parameters: []api.ParameterDefinition{
			api.NewStringParameter("key", "Key", true).
				WithGroup("Settings").
				WithDescription("Variable name to set"),
			api.NewStringParameter("value", "Value", true).
				WithGroup("Settings").
				WithDescription("Value to assign to the variable"),
		},
	}
}

// Execute sets a variable in the data map under the given key.
func (setVariableDefinition) Execute(ctx api.ExecutionContext, node api.Node, input interface{}) (interface{}, error) {
	data, ok := input.(map[string]interface{})
	if !ok {
		data = map[string]interface{}{"input": input}
	}
	key, _ := node.Data["key"].(string)
	val := node.Data["value"]
	if key != "" {
		data[key] = val
	}
	return data, nil
}

func (setVariableDefinition) Initialize(mel api.Mel) error {
	return nil
}

func init() {
	api.RegisterNodeDefinition(setVariableDefinition{})
}

// assert that setVariableDefinition implements the NodeDefinition interface
var _ api.NodeDefinition = (*setVariableDefinition)(nil)
