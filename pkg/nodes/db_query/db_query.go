package db_query

import internalapi "github.com/cedricziel/mel-agent/pkg/api"

// dbQueryDefinition provides the built-in "DB Query" node.
type dbQueryDefinition struct{}

// Meta returns metadata for the DB Query node.
func (dbQueryDefinition) Meta() internalapi.NodeType {
	return internalapi.NodeType{
		Type:     "db_query",
		Label:    "DB Query",
		Category: "Integration",
		Parameters: []internalapi.ParameterDefinition{
			internalapi.NewStringParameter("connectionId", "Connection ID", true).WithGroup("Settings").WithDescription("Select a connection"),
			internalapi.NewStringParameter("query", "SQL Query", true).WithGroup("Settings").WithDescription("Your SQL query"),
		},
	}
}

// Execute returns the input unchanged (actual I/O happens at runtime).
func (dbQueryDefinition) Execute(ctx internalapi.ExecutionContext, node internalapi.Node, input interface{}) (interface{}, error) {
	return input, nil
}

func (dbQueryDefinition) Initialize(mel internalapi.Mel) error {
	return nil
}

func init() {
	internalapi.RegisterNodeDefinition(dbQueryDefinition{})
}

// assert that dbQueryDefinition implements the NodeDefinition interface
var _ internalapi.NodeDefinition = (*dbQueryDefinition)(nil)
