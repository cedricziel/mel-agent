package plugin

import (
   "context"

   "github.com/cedricziel/mel-agent/internal/api"
)

// PluginMeta defines the metadata for a plugin, including its ID, version,
// supported categories, parameter schema, and optional UI component path.
type PluginMeta struct {
   ID          string                   `json:"id"`                              // unique plugin identifier
   Version     string                   `json:"version"`                         // semver-compliant version string
   Categories  []string                 `json:"categories"`                      // extension types provided (e.g. "node", "trigger")
   Params      []api.ParameterDefinition `json:"params,omitempty"`               // parameter schema for configuration/UI
   UIComponent string                   `json:"ui_component,omitempty"`          // optional path or URL to React bundle
}

// Plugin is the base interface that all plugins must implement.
type Plugin interface {
   Meta() PluginMeta
}

// NodePlugin represents an executable unit of work (LLM call, HTTP, transform, script).
type NodePlugin interface {
   Plugin
   Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error)
}

// TriggerPlugin defines inbound events, polling or webhook scheduling.
type TriggerPlugin interface {
   Plugin
   // OnTrigger is invoked when the trigger fires, with optional payload.
   OnTrigger(ctx context.Context, payload interface{}) (interface{}, error)
}

// ConnectionPlugin manages external integrations, credential flows, capability discovery.
type ConnectionPlugin interface {
   Plugin
   // Connect establishes a connection/resource given configuration data.
   Connect(ctx context.Context, config map[string]interface{}) (interface{}, error)
}

// AgentProtocolPlugin overrides or extends the agent control loop (memory, tool selection).
type AgentProtocolPlugin interface {
   Plugin
   // TODO: define agent control methods (Initialize, Plan, Execute, etc.)
}

// ModelSpec describes a model or tool exposed by an external MCP server.
type ModelSpec struct {
   ID          string                   `json:"id"`          // unique model or tool identifier
   Description string                   `json:"description"` // human-readable summary
   Params      []api.ParameterDefinition `json:"params,omitempty"` // invocation parameters (e.g. temperature)
}

// ModelServerPlugin integrates external Model Context Protocol (MCP) servers,
// registering available models, tools, and resources as dynamic NodePlugins.
type ModelServerPlugin interface {
   Plugin
   // Models returns metadata for each model or tool supported by the server.
   Models() []ModelSpec
   // InvokeModel sends input to the specified model or tool and returns the response.
   InvokeModel(ctx context.Context, modelID string, input interface{}) (interface{}, error)
}