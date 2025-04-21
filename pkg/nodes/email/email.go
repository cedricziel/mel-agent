package email

import api "github.com/cedricziel/mel-agent/pkg/api"

// emailDefinition provides the built-in "Email" node.
type emailDefinition struct{}

// Meta returns metadata for the Email node.
func (emailDefinition) Meta() api.NodeType {
	return api.NodeType{
		Type:     "email",
		Label:    "Email",
		Category: "Integration",
		Parameters: []api.ParameterDefinition{
			{Name: "to", Label: "To", Type: "string", Required: true, Group: "Settings", Description: "Recipient address(es)"},
			{Name: "subject", Label: "Subject", Type: "string", Required: true, Group: "Settings"},
			{Name: "body", Label: "Body", Type: "string", Required: true, Group: "Settings"},
		},
	}
}

// Execute returns the input unchanged (sending handled elsewhere).
func (emailDefinition) Execute(agentID string, node api.Node, input interface{}) (interface{}, error) {
	return input, nil
}

func init() {
	api.RegisterNodeDefinition(emailDefinition{})
}
