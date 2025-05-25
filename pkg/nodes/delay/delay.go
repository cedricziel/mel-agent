package delay

import (
	"time"

	"github.com/cedricziel/mel-agent/pkg/api"
)

// delayDefinition provides the built-in "Delay" node.
type delayDefinition struct{}

// Meta returns metadata for the Delay node.
func (delayDefinition) Meta() api.NodeType {
	return api.NodeType{
		Type:     "delay",
		Label:    "Delay",
		Category: "Control",
		Parameters: []api.ParameterDefinition{
			{Name: "duration", Label: "Duration (ms)", Type: "number", Required: true, Default: 1000, Group: "Settings"},
		},
	}
}

// Execute pauses execution for the specified duration.
func (delayDefinition) Execute(ctx api.ExecutionContext, node api.Node, input interface{}) (interface{}, error) {
	dur, ok := node.Data["duration"].(float64)
	if !ok {
		return input, nil
	}
	time.Sleep(time.Duration(dur) * time.Millisecond)
	return input, nil
}

func (delayDefinition) Initialize(mel api.Mel) error {
	// No initialization needed for this node.
	return nil
}

func init() {
	api.RegisterNodeDefinition(delayDefinition{})
}

// assert that delayDefinition implements the NodeExecutor interface
var _ api.NodeDefinition = (*delayDefinition)(nil)
