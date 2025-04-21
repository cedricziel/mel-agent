package set_variable

import (
   internalapi "github.com/cedricziel/mel-agent/internal/api"
)

// setVariableDefinition provides the built-in "Set Variable" node.
type setVariableDefinition struct{}

// Meta returns metadata for the Set Variable node.
func (setVariableDefinition) Meta() internalapi.NodeType {
   return internalapi.NodeType{
       Type:     "set_variable",
       Label:    "Set Variable",
       Category: "Utility",
       Parameters: []internalapi.ParameterDefinition{
           {Name: "key", Label: "Key", Type: "string", Required: true, Group: "Settings"},
           {Name: "value", Label: "Value", Type: "string", Required: true, Group: "Settings"},
       },
   }
}

// Execute sets a variable in the data map under the given key.
func (setVariableDefinition) Execute(agentID string, node internalapi.Node, input interface{}) (interface{}, error) {
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

func init() {
   internalapi.RegisterNodeDefinition(setVariableDefinition{})
}