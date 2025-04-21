package httprequest

import "github.com/cedricziel/mel-agent/pkg/api"

type httpRequestDefinition struct{}

func (httpRequestDefinition) Meta() api.NodeType {
	return api.NodeType{
		Type:     "http_request",
		Label:    "HTTP Request",
		Icon:     "🌐",
		Category: "Integration",
		Parameters: []api.ParameterDefinition{
			{Name: "url", Label: "URL", Type: "string", Required: true, Default: "", Group: "Request", Description: "Endpoint to call", Validators: []api.ValidatorSpec{{Type: "notEmpty"}, {Type: "url"}}},
			{Name: "method", Label: "Method", Type: "enum", Required: true, Default: "GET", Options: []string{"GET", "POST", "PUT", "DELETE"}, Group: "Request"},
			{Name: "headers", Label: "Headers", Type: "json", Required: false, Default: "{}", Group: "Request", Validators: []api.ValidatorSpec{{Type: "json"}}},
			{Name: "body", Label: "Body", Type: "string", Required: false, Default: "", Group: "Request", VisibilityCondition: "method!='GET'"},
			{Name: "timeout", Label: "Timeout", Type: "number", Required: false, Default: 30, Group: "Advanced", Description: "Timeout in seconds"},
		},
	}
}

func (httpRequestDefinition) Execute(agentID string, node api.Node, input interface{}) (interface{}, error) {
	return input, nil
}
