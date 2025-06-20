package api

import (
	"context"

	"github.com/cedricziel/mel-agent/internal/plugin"
)

// ListExtensions implements the listExtensions operation.
func (h *OpenAPIHandlers) ListExtensions(ctx context.Context, request ListExtensionsRequestObject) (ListExtensionsResponseObject, error) {
	// Get all plugin metadata from the registry
	metas := plugin.GetAllPlugins()

	// Convert PluginMeta to Extension schema
	extensions := make([]Extension, len(metas))
	for i, meta := range metas {
		// Convert ParamSpec to schema format
		var params *[]ParamSpec
		if len(meta.Params) > 0 {
			paramSpecs := make([]ParamSpec, len(meta.Params))
			for j, param := range meta.Params {
				// Convert ValidatorSpec to schema format
				var validators *[]ValidatorSpec
				if len(param.Validators) > 0 {
					validatorSpecs := make([]ValidatorSpec, len(param.Validators))
					for k, validator := range param.Validators {
						validatorSpecs[k] = ValidatorSpec{
							Type:   validator.Type,
							Params: &validator.Params,
						}
					}
					validators = &validatorSpecs
				}

				// Convert options to pointer type
				var options *[]string
				if len(param.Options) > 0 {
					options = &param.Options
				}

				paramSpecs[j] = ParamSpec{
					Name:                param.Name,
					Label:               &param.Label,
					Type:                param.Type,
					Required:            &param.Required,
					Default:             &param.Default,
					Group:               &param.Group,
					VisibilityCondition: &param.VisibilityCondition,
					Options:             options,
					Validators:          validators,
					Description:         &param.Description,
				}
			}
			params = &paramSpecs
		}

		// Convert UI component to pointer type
		var uiComponent *string
		if meta.UIComponent != "" {
			uiComponent = &meta.UIComponent
		}

		extensions[i] = Extension{
			Id:          meta.ID,
			Version:     meta.Version,
			Categories:  meta.Categories,
			Params:      params,
			UiComponent: uiComponent,
		}
	}

	return ListExtensions200JSONResponse(extensions), nil
}
