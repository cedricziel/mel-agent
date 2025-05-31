package email

import (
	api "github.com/cedricziel/mel-agent/pkg/api"
)

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

// ExecuteEnvelope returns the input unchanged (sending handled elsewhere).
func (d emailDefinition) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[interface{}]) (*api.Envelope[interface{}], error) {
	result := envelope.Clone()
	result.Trace = envelope.Trace.Next(node.ID)
	return result, nil
}

func (emailDefinition) Initialize(mel api.Mel) error {
	return nil
}

func init() {
	api.RegisterNodeDefinition(emailDefinition{})
}

// assert that emailDefinition implements both interfaces
var _ api.NodeDefinition = (*emailDefinition)(nil)
