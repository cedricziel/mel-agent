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
			{Name: "operation", Label: "Operation", Type: "enum", Required: true, Default: "read", Options: []string{"read", "write"}, Group: "Settings"},
			{Name: "path", Label: "Path", Type: "string", Required: true, Group: "Settings"},
		},
	}
}

// Execute performs the file I/O operation (currently no-op).
func (fileIODefinition) Execute(ctx api.ExecutionContext, node api.Node, input interface{}) (interface{}, error) {
	return input, nil
}

func (fileIODefinition) Initialize(mel api.Mel) error {
	return nil
}

func init() {
	api.RegisterNodeDefinition(fileIODefinition{})
}
