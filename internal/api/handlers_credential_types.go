package api

import (
	"context"
	"fmt"

	apiPkg "github.com/cedricziel/mel-agent/pkg/api"
)

// ListCredentialTypes lists credential type definitions
func (h *OpenAPIHandlers) ListCredentialTypes(ctx context.Context, request ListCredentialTypesRequestObject) (ListCredentialTypesResponseObject, error) {
	// Get credential types from code registry instead of database
	credentialTypeDefs := apiPkg.ListCredentialDefinitions()

	var credentialTypes []CredentialType
	for _, def := range credentialTypeDefs {
		credentialType := CredentialType{
			Id:          &def.Type,
			Name:        &def.Name,
			Description: &def.Description,
		}

		// Convert parameters to schema format if needed
		if len(def.Parameters) > 0 {
			schema := map[string]interface{}{
				"type":       "object",
				"properties": make(map[string]interface{}),
				"required":   []string{},
			}

			properties := make(map[string]interface{})
			var required []string

			for _, param := range def.Parameters {
				paramSchema := map[string]interface{}{
					"type":        param.Type,
					"description": param.Description,
				}

				if param.Default != nil {
					paramSchema["default"] = param.Default
				}

				properties[param.Name] = paramSchema

				if param.Required {
					required = append(required, param.Name)
				}
			}

			schema["properties"] = properties
			if len(required) > 0 {
				schema["required"] = required
				credentialType.RequiredFields = &required
			}

			credentialType.Schema = &schema
		}

		credentialTypes = append(credentialTypes, credentialType)
	}

	return ListCredentialTypes200JSONResponse(credentialTypes), nil
}

// GetCredentialTypeSchema gets JSON schema for credential type
func (h *OpenAPIHandlers) GetCredentialTypeSchema(ctx context.Context, request GetCredentialTypeSchemaRequestObject) (GetCredentialTypeSchemaResponseObject, error) {
	// Find credential definition in code registry
	def := apiPkg.FindCredentialDefinition(request.Type)
	if def == nil {
		errorMsg := "not found"
		message := "Credential type not found"
		return GetCredentialTypeSchema404JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Build schema from credential definition
	schema := CredentialTypeSchema{}

	schemaType := "object"
	schema.Type = &schemaType

	title := def.Name()
	schema.Title = &title

	description := def.Description()
	schema.Description = &description

	// Build properties from parameters
	if params := def.Parameters(); len(params) > 0 {
		properties := make(map[string]interface{})
		var required []string

		for _, param := range params {
			paramSchema := map[string]interface{}{
				"type":        param.Type,
				"description": param.Description,
			}

			if param.Default != nil {
				paramSchema["default"] = param.Default
			}

			properties[param.Name] = paramSchema

			if param.Required {
				required = append(required, param.Name)
			}
		}

		schema.Properties = &properties
		if len(required) > 0 {
			schema.Required = &required
		}
	}

	return GetCredentialTypeSchema200JSONResponse(schema), nil
}

// TestCredentials tests credentials for a specific type
func (h *OpenAPIHandlers) TestCredentials(ctx context.Context, request TestCredentialsRequestObject) (TestCredentialsResponseObject, error) {
	// Find credential definition in code registry
	def := apiPkg.FindCredentialDefinition(request.Type)
	if def == nil {
		errorMsg := "not found"
		message := "Credential type not found"
		return TestCredentials404JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Basic validation - check if credentials object is provided
	if request.Body.Credentials == nil {
		errorMsg := "bad request"
		message := "Credentials are required"
		return TestCredentials400JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Test credentials using the credential definition
	credentials := *request.Body.Credentials
	err := def.Test(credentials)

	result := CredentialTestResult{}

	if err != nil {
		// Test failed
		success := false
		result.Success = &success

		message := fmt.Sprintf("Credential test failed: %s", err.Error())
		result.Message = &message

		details := map[string]interface{}{
			"credential_type": request.Type,
			"error":           err.Error(),
		}
		result.Details = &details
	} else {
		// Test succeeded
		success := true
		result.Success = &success

		message := fmt.Sprintf("Credentials for %s tested successfully", request.Type)
		result.Message = &message

		details := map[string]interface{}{
			"credential_type": request.Type,
		}
		result.Details = &details
	}

	return TestCredentials200JSONResponse(result), nil
}
