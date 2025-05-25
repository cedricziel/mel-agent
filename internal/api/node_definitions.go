package api

import (
	"net/http"

	"github.com/cedricziel/mel-agent/pkg/api"
	_ "github.com/cedricziel/mel-agent/pkg/nodes"
)

// listNodeTypes returns all registered node type metadata.
func listNodeTypes(w http.ResponseWriter, r *http.Request) {
	metas := api.ListNodeTypes()
	writeJSON(w, http.StatusOK, metas)
}
