package api

import (
	"net/http"

	"github.com/cedricziel/mel-agent/pkg/api"
	_ "github.com/cedricziel/mel-agent/pkg/api/nodes"
)

var definitions []api.NodeDefinition

// RegisterNodeDefinition registers a node type definition.
func RegisterNodeDefinition(def api.NodeDefinition) {
	definitions = append(definitions, def)
}

// ListNodeDefinitions returns all registered NodeDefinition implementations.
func ListNodeDefinitions() []api.NodeDefinition {
	return definitions
}

// listNodeTypes returns all registered node type metadata.
func listNodeTypes(w http.ResponseWriter, r *http.Request) {
	metas := []api.NodeType{}
	for _, def := range definitions {
		metas = append(metas, def.Meta())
	}
	writeJSON(w, http.StatusOK, metas)
}

// FindDefinition retrieves the NodeDefinition for a given type.
func FindDefinition(typ string) api.NodeDefinition {
	for _, def := range definitions {
		if def.Meta().Type == typ {
			return def
		}
	}
	return nil
}

// AllCoreDefinitions returns the built-in core trigger and utility node definitions.
func AllCoreDefinitions() []api.NodeDefinition {
	return []NodeDefinition{
		slackDefinition{},
		ifDefinition{},
		switchDefinition{},
		agentDefinition{},
		llmDefinition{},
		injectDefinition{},
	}
}

// init registers all core NodeDefinitions for the /node-types endpoint.
// Builder definitions are registered by blank-importing pkg/api/nodes.
func init() {
	for _, def := range AllCoreDefinitions() {
		RegisterNodeDefinition(def)
	}
}
