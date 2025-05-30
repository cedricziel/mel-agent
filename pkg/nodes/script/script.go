package script

import api "github.com/cedricziel/mel-agent/pkg/api"

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

// Execute runs the user-provided script. Currently passthrough.
func (scriptDefinition) Execute(ctx api.ExecutionContext, node api.Node, input interface{}) (interface{}, error) {
	return input, nil
}

func (scriptDefinition) Initialize(mel api.Mel) error {
	return nil
}

func init() {
	api.RegisterNodeDefinition(scriptDefinition{})
}

// assert that scriptDefinition implements the NodeDefinition interface
var _ api.NodeDefinition = (*scriptDefinition)(nil)
