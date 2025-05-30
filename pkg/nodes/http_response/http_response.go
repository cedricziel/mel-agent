package http_response

import api "github.com/cedricziel/mel-agent/pkg/api"

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

// Execute returns the input unchanged (response handled elsewhere).
func (httpResponseDefinition) Execute(ctx api.ExecutionContext, node api.Node, input interface{}) (interface{}, error) {
	return input, nil
}

func (httpResponseDefinition) Initialize(mel api.Mel) error {
	return nil
}

func init() {
	api.RegisterNodeDefinition(httpResponseDefinition{})
}

// assert that httpResponseDefinition implements the NodeDefinition interface
var _ api.NodeDefinition = (*httpResponseDefinition)(nil)
