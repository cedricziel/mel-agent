package email

import internalapi "github.com/cedricziel/mel-agent/internal/api"

// emailDefinition provides the built-in "Email" node.
type emailDefinition struct{}

// Meta returns metadata for the Email node.
func (emailDefinition) Meta() internalapi.NodeType {
   return internalapi.NodeType{
       Type:     "email",
       Label:    "Email",
       Category: "Integration",
       Parameters: []internalapi.ParameterDefinition{
           {Name: "to", Label: "To", Type: "string", Required: true, Group: "Settings", Description: "Recipient address(es)"},
           {Name: "subject", Label: "Subject", Type: "string", Required: true, Group: "Settings"},
           {Name: "body", Label: "Body", Type: "string", Required: true, Group: "Settings"},
       },
   }
}

// Execute returns the input unchanged (sending handled elsewhere).
func (emailDefinition) Execute(agentID string, node internalapi.Node, input interface{}) (interface{}, error) {
   return input, nil
}

func init() {
   internalapi.RegisterNodeDefinition(emailDefinition{})
}