package nodes

import (
   "fmt"
   "time"

   "github.com/cedricziel/mel-agent/pkg/api"
   "github.com/google/uuid"
)

// AllNodeDefinitions returns all built-in builder node definitions.
func AllNodeDefinitions() []api.NodeDefinition {
   return []api.NodeDefinition{
       setVariableDefinition{},
       transformDefinition{},
       scriptDefinition{},
       switchDefinition{},
       forEachDefinition{},
       mergeDefinition{},
       delayDefinition{},
       httpResponseDefinition{},
       dbQueryDefinition{},
       emailDefinition{},
       fileIODefinition{},
       randomDefinition{},
       logDefinition{},
       noopDefinition{},
   }
}

// --- Set Variable Node ---
// Note: package aliases api types from pkg/api
// Each definition implements api.NodeDefinition.
// See pkg/api for NodeDefinition interface.
// Below are the individual node definitions.

// ... [omitting full definitions for brevity] ...