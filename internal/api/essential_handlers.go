package api

import (
	"encoding/json"
	"net/http"

	"github.com/cedricziel/mel-agent/internal/plugin"
)

// writeJSON is a helper to write JSON responses.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// listExtensionsHandler returns the catalog of registered plugins (GoPlugins and MCP servers).
func listExtensionsHandler(w http.ResponseWriter, r *http.Request) {
	metas := plugin.GetAllPlugins()
	writeJSON(w, http.StatusOK, metas)
}
