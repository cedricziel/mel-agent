package delay

import (
   internalapi "github.com/cedricziel/mel-agent/internal/api"
   "time"
)

// delayDefinition provides the built-in "Delay" node.
type delayDefinition struct{}

// Meta returns metadata for the Delay node.
func (delayDefinition) Meta() internalapi.NodeType {
   return internalapi.NodeType{
       Type:     "delay",
       Label:    "Delay",
       Category: "Control",
       Parameters: []internalapi.ParameterDefinition{
           {Name: "duration", Label: "Duration (ms)", Type: "number", Required: true, Default: 1000, Group: "Settings"},
       },
   }
}

// Execute pauses execution for the specified duration.
func (delayDefinition) Execute(agentID string, node internalapi.Node, input interface{}) (interface{}, error) {
   dur, ok := node.Data["duration"].(float64)
   if !ok {
       return input, nil
   }
   time.Sleep(time.Duration(dur) * time.Millisecond)
   return input, nil
}

func init() {
   internalapi.RegisterNodeDefinition(delayDefinition{})
}