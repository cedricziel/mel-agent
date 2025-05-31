package merge

import (
	api "github.com/cedricziel/mel-agent/pkg/api"
)

// mergeDefinition provides the built-in "Merge" node.
type mergeDefinition struct{}

// Meta returns metadata for the Merge node.
func (mergeDefinition) Meta() api.NodeType {
	return api.NodeType{
		Type:     "merge",
		Label:    "Merge",
		Category: "Control",
		Parameters: []api.ParameterDefinition{
			api.NewEnumParameter("strategy", "Strategy", []string{"concat", "union"}, true).WithDefault("concat").WithGroup("Settings").WithDescription("Merge strategy"),
		},
	}
}

// ExecuteEnvelope merges data based on strategy. Currently passthrough.
func (d mergeDefinition) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[interface{}]) (*api.Envelope[interface{}], error) {
	// TODO: Implement actual merge logic based on strategy
	result := envelope.Clone()
	result.Trace = envelope.Trace.Next(node.ID)
	return result, nil
}

func (mergeDefinition) Initialize(mel api.Mel) error {
	return nil
}

func init() {
	api.RegisterNodeDefinition(mergeDefinition{})
}

// assert that mergeDefinition implements the interface
var _ api.NodeDefinition = (*mergeDefinition)(nil)
