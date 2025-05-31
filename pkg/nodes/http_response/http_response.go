package http_response

import (
	api "github.com/cedricziel/mel-agent/pkg/api"
)

// httpResponseDefinition provides the built-in "HTTP Response" node.
type httpResponseDefinition struct{}

// Meta returns metadata for the HTTP Response node.
func (httpResponseDefinition) Meta() api.NodeType {
	return api.NodeType{
		Type:     "http_response",
		Label:    "HTTP Response",
		Category: "Integration",
		Parameters: []api.ParameterDefinition{
			api.NewNumberParameter("statusCode", "Status Code", true).WithDefault(200).WithGroup("Settings"),
			api.NewStringParameter("body", "Body", false).WithDefault("").WithGroup("Settings"),
		},
	}
}

// ExecuteEnvelope returns the input unchanged (response handled elsewhere).
func (d httpResponseDefinition) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[interface{}]) (*api.Envelope[interface{}], error) {
	result := envelope.Clone()
	result.Trace = envelope.Trace.Next(node.ID)
	return result, nil
}

func (httpResponseDefinition) Initialize(mel api.Mel) error {
	return nil
}

func init() {
	api.RegisterNodeDefinition(httpResponseDefinition{})
}

// assert that httpResponseDefinition implements the interface
var _ api.NodeDefinition = (*httpResponseDefinition)(nil)
