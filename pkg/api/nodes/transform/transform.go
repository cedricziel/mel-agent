package transform

import (
   internalapi "github.com/cedricziel/mel-agent/internal/api"
)

// transformDefinition provides the built-in "Transform" node.
type transformDefinition struct{}

// Meta returns metadata for the Transform node.
func (transformDefinition) Meta() internalapi.NodeType {
   return internalapi.NodeType{
       Type:     "transform",
       Label:    "Transform",
       Category: "Utility",
       Parameters: []internalapi.ParameterDefinition{
           {Name: "expression", Label: "Expression", Type: "string", Required: true, Group: "Settings", Description: "Transform input via expression"},
       },
   }
}

// Execute applies the expression to the input (currently passthrough).
func (transformDefinition) Execute(agentID string, node internalapi.Node, input interface{}) (interface{}, error) {
   return input, nil
}

func init() {
   internalapi.RegisterNodeDefinition(transformDefinition{})
}