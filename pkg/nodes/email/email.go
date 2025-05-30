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
			api.NewStringParameter("to", "To", true).WithGroup("Settings").WithDescription("Recipient address(es)"),
			api.NewStringParameter("subject", "Subject", true).WithGroup("Settings"),
			api.NewStringParameter("body", "Body", true).WithGroup("Settings"),
		},
	}
}

// Execute returns the input unchanged (sending handled elsewhere).
func (emailDefinition) Execute(ctx api.ExecutionContext, node api.Node, input interface{}) (interface{}, error) {
	return input, nil
}

func (emailDefinition) Initialize(mel api.Mel) error {
	return nil
}

func init() {
	api.RegisterNodeDefinition(emailDefinition{})
}

// assert that emailDefinition implements the NodeDefinition interface
var _ api.NodeDefinition = (*emailDefinition)(nil)
