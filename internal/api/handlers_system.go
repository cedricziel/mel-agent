package api

import (
	"context"
	"strings"

	"github.com/cedricziel/mel-agent/pkg/api"
	_ "github.com/cedricziel/mel-agent/pkg/nodes"
)

// GetHealth implements the health check endpoint
func (h *OpenAPIHandlers) GetHealth(ctx context.Context, request GetHealthRequestObject) (GetHealthResponseObject, error) {
	status := "ok"
	return GetHealth200JSONResponse{
		Status: &status,
	}, nil
}

// ListNodeTypes returns all registered node type metadata
func (h *OpenAPIHandlers) ListNodeTypes(ctx context.Context, request ListNodeTypesRequestObject) (ListNodeTypesResponseObject, error) {
	metas := api.ListNodeTypes()

	// Check for kind filter query parameter
	if request.Params.Kind != nil {
		kindFilter := *request.Params.Kind
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

	// Convert to OpenAPI types
	var nodeTypes []NodeType
	for _, meta := range metas {
		nodeType := NodeType{
			Id:          &meta.Type,
			Name:        &meta.Label,
			Description: &meta.Category,
		}

		// Convert kinds
		if meta.Kinds != nil {
			kinds := make([]NodeTypeKinds, len(meta.Kinds))
			for i, kind := range meta.Kinds {
				kinds[i] = NodeTypeKinds(kind)
			}
			nodeType.Kinds = &kinds
		}

		// Convert parameters to inputs
		if meta.Parameters != nil {
			inputs := make([]NodeInput, len(meta.Parameters))
			for i, param := range meta.Parameters {
				inputs[i] = NodeInput{
					Name:        &param.Name,
					Type:        &param.Type,
					Required:    &param.Required,
					Description: &param.Description,
				}
			}
			nodeType.Inputs = &inputs
		}

		// For now, we don't have outputs defined in the meta, so leave empty

		nodeTypes = append(nodeTypes, nodeType)
	}

	return ListNodeTypes200JSONResponse(nodeTypes), nil
}
