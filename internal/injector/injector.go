package injector

import (
	// import all builder node definitions in one sweep
	_ "github.com/cedricziel/mel-agent/pkg/nodes"
	"github.com/cedricziel/mel-agent/pkg/plugin"
	adapters "github.com/cedricziel/mel-agent/pkg/plugin/adapters"
)

// InitializeNodePlugins returns all NodePlugin implementations
// by combining core and builder definitions.
func InitializeNodePlugins() []plugin.NodePlugin {
	// adapters.ProvideNodePlugins wraps core and builder definitions into NodePlugins
	return adapters.ProvideNodePlugins()
}
