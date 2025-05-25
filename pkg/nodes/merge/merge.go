package merge

import api "github.com/cedricziel/mel-agent/pkg/api"

// mergeDefinition provides the built-in "Merge" node.
type mergeDefinition struct{}

// Meta returns metadata for the Merge node.
func (mergeDefinition) Meta() api.NodeType {
	return api.NodeType{
		Type:     "merge",
		Label:    "Merge",
		Category: "Control",
		Parameters: []api.ParameterDefinition{
			{Name: "strategy", Label: "Strategy", Type: "enum", Required: true, Default: "concat", Options: []string{"concat", "union"}, Group: "Settings", Description: "Merge strategy"},
		},
	}
}

// Execute merges data based on strategy. Currently passthrough.
func (mergeDefinition) Execute(ctx api.ExecutionContext, node api.Node, input interface{}) (interface{}, error) {
	return input, nil
}

func (mergeDefinition) Initialize(mel api.Mel) error {
	return nil
}

func init() {
	api.RegisterNodeDefinition(mergeDefinition{})
}
