package plugin

import internal "github.com/cedricziel/mel-agent/internal/plugin"

// Plugin is the base interface for all plugins.
type Plugin = internal.Plugin

// NodePlugin represents an executable workflow node.
type NodePlugin = internal.NodePlugin

// TriggerPlugin defines inbound event triggers.
type TriggerPlugin = internal.TriggerPlugin

// ConnectionPlugin manages external integrations.
type ConnectionPlugin = internal.ConnectionPlugin

// PluginMeta defines plugin metadata.
type PluginMeta = internal.PluginMeta

// ParamSpec defines a parameter spec in plugin metadata.
type ParamSpec = internal.ParamSpec

// ValidatorSpec defines a validation rule spec.
type ValidatorSpec = internal.ValidatorSpec

// Register adds a plugin to the global registry.
func Register(p Plugin) {
   internal.Register(p)
}

// GetNodePlugin retrieves a NodePlugin by ID.
func GetNodePlugin(id string) (NodePlugin, bool) {
   return internal.GetNodePlugin(id)
}

// GetTriggerPlugin retrieves a TriggerPlugin by ID.
func GetTriggerPlugin(id string) (TriggerPlugin, bool) {
   return internal.GetTriggerPlugin(id)
}

// GetConnectionPlugin retrieves a ConnectionPlugin by ID.
func GetConnectionPlugin(id string) (ConnectionPlugin, bool) {
   return internal.GetConnectionPlugin(id)
}

// GetAllPlugins returns metadata for all registered plugins.
func GetAllPlugins() []PluginMeta {
   return internal.GetAllPlugins()
}

// RegisterConnectionPlugins loads and registers all connections from the DB.
func RegisterConnectionPlugins() {
   internal.RegisterConnectionPlugins()
}