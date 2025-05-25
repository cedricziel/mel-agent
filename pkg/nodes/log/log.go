package log

import api "github.com/cedricziel/mel-agent/pkg/api"

// logDefinition provides the built-in "Log" node.
type logDefinition struct{}

// Meta returns metadata for the Log node.
func (logDefinition) Meta() api.NodeType {
	return api.NodeType{
		Type:     "log",
		Label:    "Log",
		Category: "Utility",
		Parameters: []api.ParameterDefinition{
			{Name: "level", Label: "Level", Type: "enum", Required: true, Default: "info", Options: []string{"debug", "info", "warn", "error"}, Group: "Settings"},
			{Name: "message", Label: "Message", Type: "string", Required: true, Group: "Settings"},
		},
	}
}

// Execute returns the input unchanged (logging handled separately).
func (logDefinition) Execute(ctx api.ExecutionContext, node api.Node, input interface{}) (interface{}, error) {
	return input, nil
}

func (logDefinition) Initialize(mel api.Mel) error {
	return nil
}

func init() {
	api.RegisterNodeDefinition(logDefinition{})
}
