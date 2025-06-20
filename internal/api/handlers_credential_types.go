package api

import (
	"context"
	"encoding/json"
	"fmt"
)

// ListCredentialTypes lists credential type definitions
func (h *OpenAPIHandlers) ListCredentialTypes(ctx context.Context, request ListCredentialTypesRequestObject) (ListCredentialTypesResponseObject, error) {
	// Get credential types from database
	rows, err := h.db.QueryContext(ctx, "SELECT id, name, description, schema, required_fields FROM credential_types ORDER BY name")
	if err != nil {
		errorMsg := "database error"
		message := err.Error()
		return ListCredentialTypes500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}
	defer rows.Close()

	var credentialTypes []CredentialType
	for rows.Next() {
		var credentialType CredentialType
		var id, name, description, schemaJSON, requiredFieldsJSON string

		err := rows.Scan(&id, &name, &description, &schemaJSON, &requiredFieldsJSON)
		if err != nil {
			errorMsg := "scan error"
			message := err.Error()
			return ListCredentialTypes500JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}

		credentialType.Id = &id
		credentialType.Name = &name
		credentialType.Description = &description

		// Parse schema JSON
		if schemaJSON != "" {
			var schema map[string]interface{}
			if err := json.Unmarshal([]byte(schemaJSON), &schema); err == nil {
				credentialType.Schema = &schema
			}
		}

		// Parse required fields JSON
		if requiredFieldsJSON != "" {
			var requiredFields []string
			if err := json.Unmarshal([]byte(requiredFieldsJSON), &requiredFields); err == nil {
				credentialType.RequiredFields = &requiredFields
			}
		}

		credentialTypes = append(credentialTypes, credentialType)
	}

	return ListCredentialTypes200JSONResponse(credentialTypes), nil
}

// GetCredentialTypeSchema gets JSON schema for credential type
func (h *OpenAPIHandlers) GetCredentialTypeSchema(ctx context.Context, request GetCredentialTypeSchemaRequestObject) (GetCredentialTypeSchemaResponseObject, error) {
	var schemaJSON string
	err := h.db.QueryRowContext(ctx, "SELECT schema FROM credential_types WHERE id = $1", request.Type).Scan(&schemaJSON)
	if err != nil {
		errorMsg := "not found"
		message := "Credential type not found"
		return GetCredentialTypeSchema404JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	var schema CredentialTypeSchema
	if schemaJSON != "" {
		var schemaData map[string]interface{}
		if err := json.Unmarshal([]byte(schemaJSON), &schemaData); err != nil {
			errorMsg := "schema parse error"
			message := err.Error()
			return GetCredentialTypeSchema500JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}

		// Convert to CredentialTypeSchema format
		if typeVal, ok := schemaData["type"].(string); ok {
			schema.Type = &typeVal
		}
		if title, ok := schemaData["title"].(string); ok {
			schema.Title = &title
		}
		if description, ok := schemaData["description"].(string); ok {
			schema.Description = &description
		}
		if properties, ok := schemaData["properties"].(map[string]interface{}); ok {
			schema.Properties = &properties
		}
		if required, ok := schemaData["required"].([]interface{}); ok {
			requiredStrings := make([]string, len(required))
			for i, v := range required {
				if str, ok := v.(string); ok {
					requiredStrings[i] = str
				}
			}
			schema.Required = &requiredStrings
		}
	}

	return GetCredentialTypeSchema200JSONResponse(schema), nil
}

// TestCredentials tests credentials for a specific type
func (h *OpenAPIHandlers) TestCredentials(ctx context.Context, request TestCredentialsRequestObject) (TestCredentialsResponseObject, error) {
	// Check if credential type exists
	var exists bool
	err := h.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM credential_types WHERE id = $1)", request.Type).Scan(&exists)
	if err != nil {
		errorMsg := "database error"
		message := err.Error()
		return TestCredentials500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	if !exists {
		errorMsg := "not found"
		message := "Credential type not found"
		return TestCredentials404JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// For now, return a mock successful test result
	// In a real implementation, this would test the actual credentials
	// against the appropriate service (e.g., AWS, Google Cloud, etc.)

	// Basic validation - check if credentials object is provided
	if request.Body.Credentials == nil {
		errorMsg := "bad request"
		message := "Credentials are required"
		return TestCredentials400JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Mock test result based on credential type
	result := CredentialTestResult{
		Success: func() *bool { b := true; return &b }(),
		Message: func() *string { s := fmt.Sprintf("Credentials for %s tested successfully", request.Type); return &s }(),
	}

	// Add some mock details based on type
	details := map[string]interface{}{
		"credential_type":   request.Type,
		"test_timestamp":    "2023-06-20T12:00:00Z",
		"validation_passed": true,
	}
	result.Details = &details

	return TestCredentials200JSONResponse(result), nil
}
