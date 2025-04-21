package nodes

// Register all builder node definitions by blank-importing their packages.
import (
	_ "github.com/cedricziel/mel-agent/pkg/api/nodes/db_query"
	_ "github.com/cedricziel/mel-agent/pkg/api/nodes/delay"
	_ "github.com/cedricziel/mel-agent/pkg/api/nodes/email"
	_ "github.com/cedricziel/mel-agent/pkg/api/nodes/file_io"
	_ "github.com/cedricziel/mel-agent/pkg/api/nodes/for_each"
	_ "github.com/cedricziel/mel-agent/pkg/api/nodes/http_response"
	_ "github.com/cedricziel/mel-agent/pkg/api/nodes/log"
	_ "github.com/cedricziel/mel-agent/pkg/api/nodes/merge"
	_ "github.com/cedricziel/mel-agent/pkg/api/nodes/noop"
	_ "github.com/cedricziel/mel-agent/pkg/api/nodes/random"
	_ "github.com/cedricziel/mel-agent/pkg/api/nodes/script"
	_ "github.com/cedricziel/mel-agent/pkg/api/nodes/set_variable"
	_ "github.com/cedricziel/mel-agent/pkg/api/nodes/switch_node"
	_ "github.com/cedricziel/mel-agent/pkg/api/nodes/transform"
)
