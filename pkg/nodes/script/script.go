package script

import (
	api "github.com/cedricziel/mel-agent/pkg/api"
)

// scriptDefinition provides the built-in "Script" node.
type scriptDefinition struct{}

// Meta returns metadata for the Script node.
func (scriptDefinition) Meta() api.NodeType {
	return api.NodeType{
		Type:     "script",
		Label:    "Script",
		Category: "Utility",
		Parameters: []api.ParameterDefinition{
			api.NewEnumParameter("language", "Language", []string{"javascript", "python"}, true).WithDefault("javascript").WithGroup("Settings"),
			api.NewStringParameter("code", "Code", true).WithGroup("Settings").WithDescription("Your script code"),
		},
	}
}

// ExecuteEnvelope runs the user-provided script. Currently passthrough.
func (d scriptDefinition) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[interface{}]) (*api.Envelope[interface{}], error) {
	result := envelope.Clone()
	result.Trace = envelope.Trace.Next(node.ID)
	return result, nil
}

func (scriptDefinition) Initialize(mel api.Mel) error {
	return nil
}

func init() {
	api.RegisterNodeDefinition(scriptDefinition{})
}

// assert that scriptDefinition implements both interfaces
var _ api.NodeDefinition = (*scriptDefinition)(nil)
