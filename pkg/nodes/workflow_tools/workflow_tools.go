package workflow_tools

import (
	"github.com/cedricziel/mel-agent/pkg/api"
)

// WorkflowToolsNode represents a workflow tools configuration node
type WorkflowToolsNode struct{}

// Meta returns the node type metadata
func (n *WorkflowToolsNode) Meta() api.NodeType {
	return api.NodeType{
		Type:     "workflow_tools",
		Label:    "Workflow Tools",
		Icon:     "🛠️",
		Category: "Configuration",
		Parameters: []api.ParameterDefinition{
			api.NewArrayParameter("enabledTools", "Enabled Tools", true).
				WithDescription("List of tools available to the workflow"),
			api.NewBooleanParameter("allowHttpRequests", "Allow HTTP Requests", false).
				WithDefault(true).
				WithDescription("Allow workflow to make HTTP requests"),
			api.NewBooleanParameter("allowFileSystem", "Allow File System", false).
				WithDefault(false).
				WithDescription("Allow workflow to access file system"),
			api.NewBooleanParameter("allowShellCommands", "Allow Shell Commands", false).
				WithDefault(false).
				WithDescription("Allow workflow to execute shell commands"),
			api.NewStringParameter("sandboxMode", "Sandbox Mode", false).
				WithDefault("restricted").
				WithDescription("Security sandbox level: none, restricted, strict"),
		},
	}
}

// Initialize sets up the node
func (n *WorkflowToolsNode) Initialize(mel api.Mel) error {
	return nil
}

// ExecuteEnvelope executes the node (config nodes typically don't execute)
func (n *WorkflowToolsNode) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[any]) (*api.Envelope[any], error) {
	// Config nodes typically don't execute, they provide configuration
	return envelope, nil
}

func init() {
	api.RegisterNodeDefinition(&WorkflowToolsNode{})
}
