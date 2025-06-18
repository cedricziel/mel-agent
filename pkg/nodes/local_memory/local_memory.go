package local_memory

import (
	"github.com/cedricziel/mel-agent/pkg/api"
)

// LocalMemoryNode represents a local memory configuration node
// It implements both ActionNode and MemoryNode interfaces
type LocalMemoryNode struct{}

// Ensure LocalMemoryNode implements both ActionNode and MemoryNode
var _ api.ActionNode = (*LocalMemoryNode)(nil)
var _ api.MemoryNode = (*LocalMemoryNode)(nil)

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

// Store saves data to memory (MemoryNode interface)
func (n *LocalMemoryNode) Store(ctx api.ExecutionContext, node api.Node, key string, data any) error {
	// TODO: Implement actual memory storage
	return nil
}

// Retrieve gets data from memory (MemoryNode interface)
func (n *LocalMemoryNode) Retrieve(ctx api.ExecutionContext, node api.Node, key string) (any, error) {
	// TODO: Implement actual memory retrieval
	return nil, nil
}

// Search performs semantic search in memory (MemoryNode interface)
func (n *LocalMemoryNode) Search(ctx api.ExecutionContext, node api.Node, query string, limit int) ([]api.MemoryResult, error) {
	// TODO: Implement actual memory search
	return []api.MemoryResult{}, nil
}

func init() {
	api.RegisterNodeDefinition(&LocalMemoryNode{})
}
