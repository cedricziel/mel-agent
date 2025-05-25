package nodes

// Register all builder node definitions by blank-importing their packages.
import (
	_ "github.com/cedricziel/mel-agent/pkg/nodes/db_query"
	_ "github.com/cedricziel/mel-agent/pkg/nodes/delay"
	_ "github.com/cedricziel/mel-agent/pkg/nodes/email"
	_ "github.com/cedricziel/mel-agent/pkg/nodes/file_io"
	_ "github.com/cedricziel/mel-agent/pkg/nodes/for_each"
	_ "github.com/cedricziel/mel-agent/pkg/nodes/http_response"
	_ "github.com/cedricziel/mel-agent/pkg/nodes/log"
	_ "github.com/cedricziel/mel-agent/pkg/nodes/merge"
	_ "github.com/cedricziel/mel-agent/pkg/nodes/noop"
	_ "github.com/cedricziel/mel-agent/pkg/nodes/random"
	_ "github.com/cedricziel/mel-agent/pkg/nodes/script"
	_ "github.com/cedricziel/mel-agent/pkg/nodes/set_variable"
	_ "github.com/cedricziel/mel-agent/pkg/nodes/switch_node"
	_ "github.com/cedricziel/mel-agent/pkg/nodes/transform"
)
