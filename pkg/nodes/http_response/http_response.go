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
			{Name: "statusCode", Label: "Status Code", Type: "number", Required: true, Default: 200, Group: "Settings"},
			{Name: "body", Label: "Body", Type: "string", Required: false, Default: "", Group: "Settings"},
		},
	}
}

// Execute returns the input unchanged (response handled elsewhere).
func (httpResponseDefinition) Execute(agentID string, node api.Node, input interface{}) (interface{}, error) {
	return input, nil
}

func init() {
	api.RegisterNodeDefinition(httpResponseDefinition{})
}
