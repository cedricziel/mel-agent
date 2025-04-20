package plugin

import (
   "sync"
)

// Registry holds registered plugins and supports lookup by category or ID.
type Registry struct {
   mu         sync.RWMutex
   plugins    map[string]Plugin                 // id -> Plugin
   byCategory map[string]map[string]Plugin      // category -> (id -> Plugin)
}

// defaultRegistry is the global plugin registry.
var defaultRegistry = NewRegistry()

// NewRegistry creates a new empty Registry.
func NewRegistry() *Registry {
   return &Registry{
       plugins:    make(map[string]Plugin),
       byCategory: make(map[string]map[string]Plugin),
   }
}

// Register adds a plugin to the global registry.
func Register(p Plugin) {
   defaultRegistry.Register(p)
}

// Register adds a plugin to this registry instance.
func (r *Registry) Register(p Plugin) {
   meta := p.Meta()
   id := meta.ID
   r.mu.Lock()
   defer r.mu.Unlock()
   r.plugins[id] = p
   for _, cat := range meta.Categories {
       if _, ok := r.byCategory[cat]; !ok {
           r.byCategory[cat] = make(map[string]Plugin)
       }
       r.byCategory[cat][id] = p
   }
}

// GetPlugins returns metadata for all plugins in the given category.
func GetPlugins(category string) []PluginMeta {
   return defaultRegistry.GetPlugins(category)
}

// GetPlugins returns metadata for all plugins in this registry for a category.
func (r *Registry) GetPlugins(category string) []PluginMeta {
   r.mu.RLock()
   defer r.mu.RUnlock()
   var metas []PluginMeta
   if m, ok := r.byCategory[category]; ok {
       for _, p := range m {
           metas = append(metas, p.Meta())
       }
   }
   return metas
}

// GetAllPlugins returns metadata for all registered plugins.
func GetAllPlugins() []PluginMeta {
   return defaultRegistry.GetAllPlugins()
}

// GetAllPlugins returns metadata for all plugins in this registry.
func (r *Registry) GetAllPlugins() []PluginMeta {
   r.mu.RLock()
   defer r.mu.RUnlock()
   var metas []PluginMeta
   for _, p := range r.plugins {
       metas = append(metas, p.Meta())
   }
   return metas
}

// GetNodePlugin retrieves a NodePlugin implementation by plugin ID.
func GetNodePlugin(id string) (NodePlugin, bool) {
   return defaultRegistry.GetNodePlugin(id)
}

// GetNodePlugin retrieves a NodePlugin by ID from this registry.
func (r *Registry) GetNodePlugin(id string) (NodePlugin, bool) {
   r.mu.RLock()
   defer r.mu.RUnlock()
   p, ok := r.plugins[id]
   if !ok {
       return nil, false
   }
   np, ok := p.(NodePlugin)
   return np, ok
}

// Similar helper methods (GetTriggerPlugin, GetConnectionPlugin, etc.) can be added as needed.