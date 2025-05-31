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

// ExecuteEnvelope generates a random value based on the configured type using envelopes.
func (d randomDefinition) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[interface{}]) (*api.Envelope[interface{}], error) {
	typ, _ := node.Data["type"].(string)

	var randomValue interface{}
	switch typ {
	case "uuid":
		randomValue = fmt.Sprintf("%s", uuid.New())
	case "number":
		// Simple random number: use nanosecond timestamp
		randomValue = time.Now().UnixNano()
	default:
		randomValue = envelope.Data
	}

	result := envelope.Clone()
	result.Trace = envelope.Trace.Next(node.ID)
	result.Data = randomValue

	return result, nil
}

func (randomDefinition) Initialize(mel api.Mel) error {
	return nil
}

func init() {
	api.RegisterNodeDefinition(randomDefinition{})
}

// assert that randomDefinition implements the interface
var _ api.NodeDefinition = (*randomDefinition)(nil)
