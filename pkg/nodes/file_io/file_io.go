package file_io

import (
	api "github.com/cedricziel/mel-agent/pkg/api"
)

// fileIODefinition provides the built-in "File I/O" node.
type fileIODefinition struct{}

// Meta returns metadata for the File I/O node.
func (fileIODefinition) Meta() api.NodeType {
	return api.NodeType{
		Type:     "file_io",
		Label:    "File I/O",
		Category: "Integration",
		Parameters: []api.ParameterDefinition{
			api.NewEnumParameter("operation", "Operation", []string{"read", "write"}, true).WithDefault("read").WithGroup("Settings"),
			api.NewStringParameter("path", "Path", true).WithGroup("Settings"),
		},
	}
}

// ExecuteEnvelope performs the file I/O operation (currently no-op).
func (d fileIODefinition) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[interface{}]) (*api.Envelope[interface{}], error) {
	result := envelope.Clone()
	result.Trace = envelope.Trace.Next(node.ID)
	return result, nil
}

func (fileIODefinition) Initialize(mel api.Mel) error {
	return nil
}

func init() {
	api.RegisterNodeDefinition(fileIODefinition{})
}

// assert that fileIODefinition implements both interfaces
var _ api.NodeDefinition = (*fileIODefinition)(nil)
