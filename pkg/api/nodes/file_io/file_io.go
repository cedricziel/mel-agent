package file_io

import (
   internalapi "github.com/cedricziel/mel-agent/internal/api"
)

// fileIODefinition provides the built-in "File I/O" node.
type fileIODefinition struct{}

// Meta returns metadata for the File I/O node.
func (fileIODefinition) Meta() internalapi.NodeType {
   return internalapi.NodeType{
       Type:     "file_io",
       Label:    "File I/O",
       Category: "Integration",
       Parameters: []internalapi.ParameterDefinition{
           {Name: "operation", Label: "Operation", Type: "enum", Required: true, Default: "read", Options: []string{"read", "write"}, Group: "Settings"},
           {Name: "path", Label: "Path", Type: "string", Required: true, Group: "Settings"},
       },
   }
}

// Execute performs the file I/O operation (currently no-op).
func (fileIODefinition) Execute(agentID string, node internalapi.Node, input interface{}) (interface{}, error) {
   return input, nil
}

func init() {
   internalapi.RegisterNodeDefinition(fileIODefinition{})
}