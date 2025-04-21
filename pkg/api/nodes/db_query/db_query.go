package db_query

import internalapi "github.com/cedricziel/mel-agent/internal/api"

// dbQueryDefinition provides the built-in "DB Query" node.
type dbQueryDefinition struct{}

// Meta returns metadata for the DB Query node.
func (dbQueryDefinition) Meta() internalapi.NodeType {
   return internalapi.NodeType{
       Type:     "db_query",
       Label:    "DB Query",
       Category: "Integration",
       Parameters: []internalapi.ParameterDefinition{
           {Name: "connectionId", Label: "Connection ID", Type: "string", Required: true, Group: "Settings", Description: "Select a connection"},
           {Name: "query", Label: "SQL Query", Type: "string", Required: true, Group: "Settings", Description: "Your SQL query"},
       },
   }
}

// Execute returns the input unchanged (actual I/O happens at runtime).
func (dbQueryDefinition) Execute(agentID string, node internalapi.Node, input interface{}) (interface{}, error) {
   return input, nil
}

func init() {
   internalapi.RegisterNodeDefinition(dbQueryDefinition{})
}