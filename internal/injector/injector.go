package injector

import (
   // blank-import legacy builder nodes so they register with internal/api
   _ "github.com/cedricziel/mel-agent/internal/api/nodes"
   adapters "github.com/cedricziel/mel-agent/pkg/plugin/adapters"
   "github.com/cedricziel/mel-agent/pkg/plugin"
)

// InitializeNodePlugins returns all NodePlugin implementations
// by combining core and builder definitions.
func InitializeNodePlugins() []plugin.NodePlugin {
   // adapters.ProvideNodePlugins wraps core and builder definitions into NodePlugins
   return adapters.ProvideNodePlugins()
}