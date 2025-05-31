package log

import (
	api "github.com/cedricziel/mel-agent/pkg/api"
)

// logDefinition provides the built-in "Log" node.
type logDefinition struct{}

// Meta returns metadata for the Log node.
func (logDefinition) Meta() api.NodeType {
	return api.NodeType{
		Type:     "log",
		Label:    "Log",
		Category: "Utility",
		Parameters: []api.ParameterDefinition{
			api.NewEnumParameter("level", "Level", []string{"debug", "info", "warn", "error"}, true).WithDefault("info").WithGroup("Settings"),
			api.NewStringParameter("message", "Message", true).WithGroup("Settings"),
		},
	}
}

// ExecuteEnvelope returns the input envelope unchanged (logging handled separately).
func (d logDefinition) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[interface{}]) (*api.Envelope[interface{}], error) {
	result := envelope.Clone()
	result.Trace = envelope.Trace.Next(node.ID)
	return result, nil
}

func (logDefinition) Initialize(mel api.Mel) error {
	return nil
}

func init() {
	api.RegisterNodeDefinition(logDefinition{})
}

// assert that logDefinition implements the interface
var _ api.NodeDefinition = (*logDefinition)(nil)
