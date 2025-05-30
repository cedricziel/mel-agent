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
			api.NewStringParameter("url", "URL", true).
				WithGroup("Request").
				WithDescription("Endpoint to call").
				WithValidators(api.ValidatorSpec{Type: "notEmpty"}, api.ValidatorSpec{Type: "url"}),
			api.NewEnumParameter("method", "Method", []string{"GET", "POST", "PUT", "DELETE"}, true).
				WithDefault("GET").
				WithGroup("Request"),
			api.NewObjectParameter("headers", "Headers", false).
				WithDefault("{}").
				WithGroup("Request").
				WithValidators(api.ValidatorSpec{Type: "json"}),
			api.NewStringParameter("body", "Body", false).
				WithGroup("Request").
				WithDescription("Request body").
				WithVisibilityCondition("method!='GET'"),
			api.NewNumberParameter("timeout", "Timeout", false).
				WithDefault(30).
				WithGroup("Advanced").
				WithDescription("Timeout in seconds"),
		},
	}
}

func (httpRequestDefinition) Execute(ctx api.ExecutionContext, node api.Node, input interface{}) (interface{}, error) {
	return input, nil
}

func (httpRequestDefinition) Initialize(mel api.Mel) error {
	return nil
}

func init() {
	api.RegisterNodeDefinition(httpRequestDefinition{})
}

// assert that httpRequestDefinition implements the NodeDefinition interface
var _ api.NodeDefinition = (*httpRequestDefinition)(nil)
