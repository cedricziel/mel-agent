package injector

import (
   // blank-import legacy builder nodes so they register with internal/api
   _ "github.com/cedricziel/mel-agent/internal/api/nodes"
   // import migrated Delay node definition
   _ "github.com/cedricziel/mel-agent/pkg/api/nodes/delay"
   // import migrated Random node definition
   _ "github.com/cedricziel/mel-agent/pkg/api/nodes/random"
   // import migrated File I/O node definition
   _ "github.com/cedricziel/mel-agent/pkg/api/nodes/file_io"
   // import migrated For Each node definition
   _ "github.com/cedricziel/mel-agent/pkg/api/nodes/for_each"
   // import migrated Merge node definition
   _ "github.com/cedricziel/mel-agent/pkg/api/nodes/merge"
   // import migrated DB Query node definition
   _ "github.com/cedricziel/mel-agent/pkg/api/nodes/db_query"
   // import migrated Email node definition
   _ "github.com/cedricziel/mel-agent/pkg/api/nodes/email"
   // import migrated Log node definition
   _ "github.com/cedricziel/mel-agent/pkg/api/nodes/log"
   // import migrated No-Op node definition
   _ "github.com/cedricziel/mel-agent/pkg/api/nodes/noop"
   // import migrated Set Variable node
   _ "github.com/cedricziel/mel-agent/pkg/api/nodes/set_variable"
   // import migrated HTTP Response node definition
   _ "github.com/cedricziel/mel-agent/pkg/api/nodes/http_response"
   // import migrated transform node to register via internal/API
   _ "github.com/cedricziel/mel-agent/pkg/api/nodes/transform"
   // import migrated Script node definition
   _ "github.com/cedricziel/mel-agent/pkg/api/nodes/script"
   // import migrated Switch node definition
   _ "github.com/cedricziel/mel-agent/pkg/api/nodes/switch_node"
   adapters "github.com/cedricziel/mel-agent/pkg/plugin/adapters"
   "github.com/cedricziel/mel-agent/pkg/plugin"
)

// InitializeNodePlugins returns all NodePlugin implementations
// by combining core and builder definitions.
func InitializeNodePlugins() []plugin.NodePlugin {
   // adapters.ProvideNodePlugins wraps core and builder definitions into NodePlugins
   return adapters.ProvideNodePlugins()
}