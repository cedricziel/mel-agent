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

	// Check for type filter query parameter
	typeFilter := r.URL.Query().Get("type")
	if typeFilter != "" {
		// Split by comma to support multiple types
		requestedTypes := strings.Split(typeFilter, ",")
		requestedTypeMap := make(map[string]bool)
		for _, t := range requestedTypes {
			requestedTypeMap[strings.TrimSpace(t)] = true
		}

		// Filter metas based on category mapping
		var filteredMetas []api.NodeType
		for _, meta := range metas {
			nodeTypeCategory := getNodeTypeCategory(meta)
			if requestedTypeMap[nodeTypeCategory] {
				filteredMetas = append(filteredMetas, meta)
			}
		}
		metas = filteredMetas
	}

	writeJSON(w, http.StatusOK, metas)
}

// getNodeTypeCategory maps node categories to API filter types
func getNodeTypeCategory(nodeType api.NodeType) string {
	switch nodeType.Category {
	case "Configuration":
		// Determine subcategory based on node type
		switch {
		case strings.Contains(nodeType.Type, "model"):
			return "model"
		case strings.Contains(nodeType.Type, "memory"):
			return "memory"
		case strings.Contains(nodeType.Type, "tools"):
			return "tools"
		default:
			return "model" // Default config nodes to model category
		}
	case "Actions":
		return "action"
	case "Triggers":
		return "trigger"
	case "Logic":
		return "action" // Logic nodes are typically actions
	case "Data":
		return "action" // Data nodes are typically actions
	case "Integration":
		return "action" // Integration nodes are typically actions
	default:
		return "action" // Default to action
	}
}
