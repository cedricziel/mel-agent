package random

import (
	"fmt"
	"time"

	api "github.com/cedricziel/mel-agent/pkg/api"
	"github.com/google/uuid"
)

// randomDefinition provides the built-in "Random" node.
type randomDefinition struct{}

// Meta returns metadata for the Random node.
func (randomDefinition) Meta() api.NodeType {
	return api.NodeType{
		Type:     "random",
		Label:    "Random",
		Category: "Utility",
		Parameters: []api.ParameterDefinition{
			api.NewEnumParameter("type", "Type", []string{"uuid", "number"}, true).WithDefault("uuid").WithGroup("Settings"),
		},
	}
}

// Execute generates a random value based on the configured type.
func (randomDefinition) Execute(ctx api.ExecutionContext, node api.Node, input interface{}) (interface{}, error) {
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

func (randomDefinition) Initialize(mel api.Mel) error {
	return nil
}

func init() {
	api.RegisterNodeDefinition(randomDefinition{})
}
