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

// GetNodeParameterOptions gets dynamic options for node parameters
func (h *OpenAPIHandlers) GetNodeParameterOptions(ctx context.Context, request GetNodeParameterOptionsRequestObject) (GetNodeParameterOptionsResponseObject, error) {
	// For now, return mock options based on the node type and parameter
	// In a real implementation, this would:
	// 1. Look up the node type definition
	// 2. Check if the parameter supports dynamic options
	// 3. Generate options based on context (e.g., available connections, models, etc.)

	var options []struct {
		Description *string `json:"description,omitempty"`
		Label       *string `json:"label,omitempty"`
		Value       *string `json:"value,omitempty"`
	}

	// Mock different options based on parameter type
	switch request.Parameter {
	case "model":
		options = []struct {
			Description *string `json:"description,omitempty"`
			Label       *string `json:"label,omitempty"`
			Value       *string `json:"value,omitempty"`
		}{
			{
				Description: func() *string { s := "Most capable model"; return &s }(),
				Label:       func() *string { s := "GPT-4"; return &s }(),
				Value:       func() *string { s := "gpt-4"; return &s }(),
			},
			{
				Description: func() *string { s := "Fast and efficient"; return &s }(),
				Label:       func() *string { s := "GPT-3.5 Turbo"; return &s }(),
				Value:       func() *string { s := "gpt-3.5-turbo"; return &s }(),
			},
			{
				Description: func() *string { s := "Anthropic's latest model"; return &s }(),
				Label:       func() *string { s := "Claude 3"; return &s }(),
				Value:       func() *string { s := "claude-3"; return &s }(),
			},
		}
	case "connection":
		options = []struct {
			Description *string `json:"description,omitempty"`
			Label       *string `json:"label,omitempty"`
			Value       *string `json:"value,omitempty"`
		}{
			{
				Description: func() *string { s := "Production OpenAI API"; return &s }(),
				Label:       func() *string { s := "OpenAI Connection"; return &s }(),
				Value:       func() *string { s := "conn-1"; return &s }(),
			},
			{
				Description: func() *string { s := "Claude API connection"; return &s }(),
				Label:       func() *string { s := "Anthropic Connection"; return &s }(),
				Value:       func() *string { s := "conn-2"; return &s }(),
			},
		}
	case "temperature":
		options = []struct {
			Description *string `json:"description,omitempty"`
			Label       *string `json:"label,omitempty"`
			Value       *string `json:"value,omitempty"`
		}{
			{
				Description: func() *string { s := "Very deterministic"; return &s }(),
				Label:       func() *string { s := "Low (0.1)"; return &s }(),
				Value:       func() *string { s := "0.1"; return &s }(),
			},
			{
				Description: func() *string { s := "Balanced creativity"; return &s }(),
				Label:       func() *string { s := "Medium (0.7)"; return &s }(),
				Value:       func() *string { s := "0.7"; return &s }(),
			},
			{
				Description: func() *string { s := "Very creative"; return &s }(),
				Label:       func() *string { s := "High (1.0)"; return &s }(),
				Value:       func() *string { s := "1.0"; return &s }(),
			},
		}
	default:
		// Generic options for unknown parameters
		options = []struct {
			Description *string `json:"description,omitempty"`
			Label       *string `json:"label,omitempty"`
			Value       *string `json:"value,omitempty"`
		}{
			{
				Description: func() *string { s := "First option"; return &s }(),
				Label:       func() *string { s := "Option 1"; return &s }(),
				Value:       func() *string { s := "option1"; return &s }(),
			},
			{
				Description: func() *string { s := "Second option"; return &s }(),
				Label:       func() *string { s := "Option 2"; return &s }(),
				Value:       func() *string { s := "option2"; return &s }(),
			},
		}
	}

	result := NodeParameterOptions{
		Options: &options,
		Dynamic: func() *bool { b := true; return &b }(),
		ContextDependent: func() *bool {
			// Parameters like "connection" are context-dependent
			b := request.Parameter == "connection" || request.Parameter == "model"
			return &b
		}(),
	}

	return GetNodeParameterOptions200JSONResponse(result), nil
}
