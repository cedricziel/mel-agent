package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ListConnections retrieves all connections
func (h *OpenAPIHandlers) ListConnections(ctx context.Context, request ListConnectionsRequestObject) (ListConnectionsResponseObject, error) {
	rows, err := h.db.QueryContext(ctx,
		"SELECT id, name, type, credentials, created_at, updated_at FROM connections ORDER BY created_at DESC")
	if err != nil {
		errorMsg := "database error"
		message := err.Error()
		return ListConnections500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}
	defer rows.Close()

	var connections []Connection
	for rows.Next() {
		var connection Connection
		var id, name, connType string
		var credentialsJson []byte
		var createdAt, updatedAt time.Time

		err := rows.Scan(&id, &name, &connType, &credentialsJson, &createdAt, &updatedAt)
		if err != nil {
			errorMsg := "scan error"
			message := err.Error()
			return ListConnections500JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}

		connectionUUID, err := uuid.Parse(id)
		if err != nil {
			errorMsg := "uuid parse error"
			message := err.Error()
			return ListConnections500JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}

		// Parse credentials JSON
		var credentials map[string]interface{}
		if len(credentialsJson) > 0 {
			err = json.Unmarshal(credentialsJson, &credentials)
			if err != nil {
				errorMsg := "credentials parse error"
				message := err.Error()
				return ListConnections500JSONResponse{
					Error:   &errorMsg,
					Message: &message,
				}, nil
			}
		}

		connection.Id = &connectionUUID
		connection.Name = &name
		connection.Type = &connType
		connection.Credentials = &credentials
		connection.CreatedAt = &createdAt
		connection.UpdatedAt = &updatedAt

		connections = append(connections, connection)
	}

	return ListConnections200JSONResponse(connections), nil
}

// CreateConnection creates a new connection
func (h *OpenAPIHandlers) CreateConnection(ctx context.Context, request CreateConnectionRequestObject) (CreateConnectionResponseObject, error) {
	connectionID := uuid.New()
	now := time.Now()

	// Marshal credentials to JSON
	credentialsJson, err := json.Marshal(request.Body.Credentials)
	if err != nil {
		errorMsg := "failed to marshal credentials"
		message := err.Error()
		return CreateConnection500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Insert connection into database
	_, err = h.db.ExecContext(ctx,
		"INSERT INTO connections (id, name, type, credentials, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)",
		connectionID, request.Body.Name, request.Body.Type, credentialsJson, now, now)
	if err != nil {
		errorMsg := "failed to create connection"
		message := err.Error()
		return CreateConnection500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	connection := Connection{
		Id:          &connectionID,
		Name:        &request.Body.Name,
		Type:        &request.Body.Type,
		Credentials: &request.Body.Credentials,
		CreatedAt:   &now,
		UpdatedAt:   &now,
	}

	return CreateConnection201JSONResponse(connection), nil
}

// GetConnection retrieves a single connection by ID
func (h *OpenAPIHandlers) GetConnection(ctx context.Context, request GetConnectionRequestObject) (GetConnectionResponseObject, error) {
	var connection Connection
	var id, name, connType string
	var credentialsJson []byte
	var createdAt, updatedAt time.Time

	err := h.db.QueryRowContext(ctx,
		"SELECT id, name, type, credentials, created_at, updated_at FROM connections WHERE id = $1",
		request.Id.String()).Scan(&id, &name, &connType, &credentialsJson, &createdAt, &updatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			errorMsg := "not found"
			message := "Connection not found"
			return GetConnection404JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}
		errorMsg := "database error"
		message := err.Error()
		return GetConnection500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	connectionUUID, err := uuid.Parse(id)
	if err != nil {
		errorMsg := "uuid parse error"
		message := err.Error()
		return GetConnection500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Parse credentials JSON
	var credentials map[string]interface{}
	if len(credentialsJson) > 0 {
		err = json.Unmarshal(credentialsJson, &credentials)
		if err != nil {
			errorMsg := "credentials parse error"
			message := err.Error()
			return GetConnection500JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}
	}

	connection.Id = &connectionUUID
	connection.Name = &name
	connection.Type = &connType
	connection.Credentials = &credentials
	connection.CreatedAt = &createdAt
	connection.UpdatedAt = &updatedAt

	return GetConnection200JSONResponse(connection), nil
}

// UpdateConnection updates an existing connection
func (h *OpenAPIHandlers) UpdateConnection(ctx context.Context, request UpdateConnectionRequestObject) (UpdateConnectionResponseObject, error) {
	now := time.Now()

	// Start building the update query
	setParts := []string{"updated_at = $1"}
	args := []interface{}{now}
	argIndex := 2

	if request.Body.Name != nil {
		setParts = append(setParts, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, *request.Body.Name)
		argIndex++
	}

	if request.Body.Credentials != nil {
		credentialsJson, err := json.Marshal(*request.Body.Credentials)
		if err != nil {
			errorMsg := "failed to marshal credentials"
			message := err.Error()
			return UpdateConnection500JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}
		setParts = append(setParts, fmt.Sprintf("credentials = $%d", argIndex))
		args = append(args, credentialsJson)
		argIndex++
	}

	// Add the ID as the last parameter
	args = append(args, request.Id.String())

	query := fmt.Sprintf("UPDATE connections SET %s WHERE id = $%d",
		strings.Join(setParts, ", "), argIndex)

	result, err := h.db.ExecContext(ctx, query, args...)
	if err != nil {
		errorMsg := "failed to update connection"
		message := err.Error()
		return UpdateConnection500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		errorMsg := "failed to check update result"
		message := err.Error()
		return UpdateConnection500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	if rowsAffected == 0 {
		errorMsg := "not found"
		message := "Connection not found"
		return UpdateConnection404JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Fetch the updated connection
	getRequest := GetConnectionRequestObject{Id: request.Id}
	getResponse, err := h.GetConnection(ctx, getRequest)
	if err != nil {
		errorMsg := "failed to fetch updated connection"
		message := err.Error()
		return UpdateConnection500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Convert GetConnectionResponseObject to UpdateConnectionResponseObject
	if connectionResponse, ok := getResponse.(GetConnection200JSONResponse); ok {
		return UpdateConnection200JSONResponse(connectionResponse), nil
	}

	// If we get here, something went wrong
	errorMsg := "unexpected response type"
	message := "Failed to convert response"
	return UpdateConnection500JSONResponse{
		Error:   &errorMsg,
		Message: &message,
	}, nil
}

// DeleteConnection removes a connection
func (h *OpenAPIHandlers) DeleteConnection(ctx context.Context, request DeleteConnectionRequestObject) (DeleteConnectionResponseObject, error) {
	result, err := h.db.ExecContext(ctx, "DELETE FROM connections WHERE id = $1", request.Id.String())
	if err != nil {
		errorMsg := "failed to delete connection"
		message := err.Error()
		return DeleteConnection500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		errorMsg := "failed to check delete result"
		message := err.Error()
		return DeleteConnection500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	if rowsAffected == 0 {
		errorMsg := "not found"
		message := "Connection not found"
		return DeleteConnection404JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	return DeleteConnection204Response{}, nil
}
