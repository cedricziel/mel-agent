package api

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// ListCredentials lists credentials for selection in nodes
func (h *OpenAPIHandlers) ListCredentials(ctx context.Context, request ListCredentialsRequestObject) (ListCredentialsResponseObject, error) {
	query := "SELECT id, name, type, created_at FROM connections ORDER BY name"
	args := []interface{}{}

	// Add filter if credential_type is provided
	if request.Params.CredentialType != nil {
		query = "SELECT id, name, type, created_at FROM connections WHERE type = $1 ORDER BY name"
		args = append(args, *request.Params.CredentialType)
	}

	rows, err := h.db.QueryContext(ctx, query, args...)
	if err != nil {
		errorMsg := "database error"
		message := err.Error()
		return ListCredentials500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}
	defer rows.Close()

	var credentials []Credential
	for rows.Next() {
		var credential Credential
		var id, name, credType string
		var createdAt time.Time

		err := rows.Scan(&id, &name, &credType, &createdAt)
		if err != nil {
			errorMsg := "scan error"
			message := err.Error()
			return ListCredentials500JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}

		credentialUUID, err := uuid.Parse(id)
		if err != nil {
			errorMsg := "uuid parse error"
			message := err.Error()
			return ListCredentials500JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}

		credential.Id = &credentialUUID
		credential.Name = &name
		credential.Type = &credType
		credential.CreatedAt = &createdAt
		credential.Status = func() *CredentialStatus { s := CredentialStatusValid; return &s }() // Default status

		credentials = append(credentials, credential)
	}

	return ListCredentials200JSONResponse(credentials), nil
}
