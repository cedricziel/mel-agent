package http_response

import internalapi "github.com/cedricziel/mel-agent/internal/api"

// httpResponseDefinition provides the built-in "HTTP Response" node.
type httpResponseDefinition struct{}

// Meta returns metadata for the HTTP Response node.
func (httpResponseDefinition) Meta() internalapi.NodeType {
   return internalapi.NodeType{
       Type:     "http_response",
       Label:    "HTTP Response",
       Category: "Integration",
       Parameters: []internalapi.ParameterDefinition{
           {Name: "statusCode", Label: "Status Code", Type: "number", Required: true, Default: 200, Group: "Settings"},
           {Name: "body", Label: "Body", Type: "string", Required: false, Default: "", Group: "Settings"},
       },
   }
}

// Execute returns the input unchanged (response handled elsewhere).
func (httpResponseDefinition) Execute(agentID string, node internalapi.Node, input interface{}) (interface{}, error) {
   return input, nil
}

func init() {
   internalapi.RegisterNodeDefinition(httpResponseDefinition{})
}