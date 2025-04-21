package nodes

import (
   "github.com/cedricziel/mel-agent/internal/api"
)

// init registers all built-in builder node definitions.
func init() {
   // setVariableDefinition migrated out to pkg/api/nodes/set_variable
   // mergeDefinition migrated to pkg/api/nodes/merge
   // dbQueryDefinition migrated to pkg/api/nodes/db_query
   // emailDefinition migrated to pkg/api/nodes/email
   // logDefinition migrated to pkg/api/nodes/log
   // noopDefinition migrated to pkg/api/nodes/noop
}

// --- Set Variable Node ---
type setVariableDefinition struct{}
func (setVariableDefinition) Meta() api.NodeType {
   return api.NodeType{
       Type:     "set_variable",
       Label:    "Set Variable",
       Category: "Utility",
       Parameters: []api.ParameterDefinition{
           {Name: "key", Label: "Key", Type: "string", Required: true, Group: "Settings"},
           {Name: "value", Label: "Value", Type: "string", Required: true, Group: "Settings"},
       },
   }
}
func (setVariableDefinition) Execute(agentID string, node api.Node, input interface{}) (interface{}, error) {
   data, ok := input.(map[string]interface{})
   if !ok {
       data = map[string]interface{}{"input": input}
   }
   key, _ := node.Data["key"].(string)
   val := node.Data["value"]
   if key != "" {
       data[key] = val
   }
   return data, nil
}



// --- Script Node ---
// Migrated to pkg/api/nodes/script

// --- Switch Node ---
// Migrated to pkg/api/nodes/switch_node



  // --- Delay Node ---
  // Migrated to pkg/api/nodes/delay

// --- HTTP Response Node ---
// Migrated to pkg/api/nodes/http_response

// --- DB Query Node ---
// Migrated to pkg/api/nodes/db_query

// --- Email Notification Node ---
// Migrated to pkg/api/nodes/email

   // --- File I/O Node ---
   // Migrated to pkg/api/nodes/file_io

// --- Random Node ---
// Migrated to pkg/api/nodes/random

// --- Log Node ---
// Migrated to pkg/api/nodes/log

// --- No-Op Node ---
// Migrated to pkg/api/nodes/noop