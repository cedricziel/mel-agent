package random

import (
   "fmt"
   "time"
   "github.com/google/uuid"
   internalapi "github.com/cedricziel/mel-agent/internal/api"
)

// randomDefinition provides the built-in "Random" node.
type randomDefinition struct{}

// Meta returns metadata for the Random node.
func (randomDefinition) Meta() internalapi.NodeType {
   return internalapi.NodeType{
       Type:     "random",
       Label:    "Random",
       Category: "Utility",
       Parameters: []internalapi.ParameterDefinition{
           {Name: "type", Label: "Type", Type: "enum", Required: true, Default: "uuid", Options: []string{"uuid", "number"}, Group: "Settings"},
       },
   }
}

// Execute generates a random value based on the configured type.
func (randomDefinition) Execute(agentID string, node internalapi.Node, input interface{}) (interface{}, error) {
   typ, _ := node.Data["type"].(string)
   switch typ {
   case "uuid":
       return fmt.Sprintf("%s", uuid.New()), nil
   case "number":
       // Simple random number: use nanosecond timestamp
       return time.Now().UnixNano(), nil
   default:
       return input, nil
   }
}

func init() {
   internalapi.RegisterNodeDefinition(randomDefinition{})
}