package api

import (
	"net/http"
	"strings"

	"github.com/cedricziel/mel-agent/pkg/api"
	_ "github.com/cedricziel/mel-agent/pkg/nodes"
)

// listNodeTypes returns all registered node type metadata.
func listNodeTypes(w http.ResponseWriter, r *http.Request) {
	metas := api.ListNodeTypes()

	// Check for kind filter query parameter
	kindFilter := r.URL.Query().Get("kind")
	if kindFilter != "" {
		// Split by comma to support multiple kinds
		requestedKinds := strings.Split(kindFilter, ",")
		requestedKindMap := make(map[string]bool)
		for _, k := range requestedKinds {
			requestedKindMap[strings.TrimSpace(k)] = true
		}

		// Filter metas based on kinds
		var filteredMetas []api.NodeType
		for _, meta := range metas {
			// Check if any of the node's kinds match the requested kinds
			for _, kind := range meta.Kinds {
				if requestedKindMap[string(kind)] {
					filteredMetas = append(filteredMetas, meta)
					break // Only add once even if multiple kinds match
				}
			}
		}
		metas = filteredMetas
	}

	writeJSON(w, http.StatusOK, metas)
}
