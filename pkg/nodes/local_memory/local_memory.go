package local_memory

import (
	"github.com/cedricziel/mel-agent/pkg/api"
)

// LocalMemoryNode represents a local memory configuration node
type LocalMemoryNode struct{}

// Meta returns the node type metadata
func (n *LocalMemoryNode) Meta() api.NodeType {
	return api.NodeType{
		Type:     "local_memory",
		Label:    "Local Memory",
		Icon:     "🧠",
		Category: "Configuration",
		Parameters: []api.ParameterDefinition{
			api.NewStringParameter("storageType", "Storage Type", true).
				WithDefault("local").
				WithDescription("Type of storage to use for memory"),
			api.NewStringParameter("namespace", "Namespace", false).
				WithDefault("default").
				WithDescription("Namespace for memory isolation"),
			api.NewIntegerParameter("maxEntries", "Max Entries", false).
				WithDefault(1000).
				WithDescription("Maximum number of memory entries to store"),
			api.NewBooleanParameter("persistent", "Persistent", false).
				WithDefault(true).
				WithDescription("Whether memory should persist across sessions"),
		},
	}
}

// Initialize sets up the node
func (n *LocalMemoryNode) Initialize(mel api.Mel) error {
	return nil
}

// ExecuteEnvelope executes the node (config nodes typically don't execute)
func (n *LocalMemoryNode) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[any]) (*api.Envelope[any], error) {
	// Config nodes typically don't execute, they provide configuration
	return envelope, nil
}

func init() {
	api.RegisterNodeDefinition(&LocalMemoryNode{})
}
