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
		"SELECT id, user_id, integration_id, name, secret, config, usage_limit_month, is_default, created_at, last_validated, status FROM connections ORDER BY created_at DESC")
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
		var id, userID, integrationID, name, status string
		var secretJson, configJson []byte
		var usageLimitMonth sql.NullInt32
		var isDefault bool
		var createdAt time.Time
		var lastValidated sql.NullTime

		err := rows.Scan(&id, &userID, &integrationID, &name, &secretJson, &configJson, &usageLimitMonth, &isDefault, &createdAt, &lastValidated, &status)
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

		userUUID, err := uuid.Parse(userID)
		if err != nil {
			errorMsg := "user uuid parse error"
			message := err.Error()
			return ListConnections500JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}

		integrationUUID, err := uuid.Parse(integrationID)
		if err != nil {
			errorMsg := "integration uuid parse error"
			message := err.Error()
			return ListConnections500JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}

		// Parse secret JSON
		var secret map[string]interface{}
		if len(secretJson) > 0 {
			err = json.Unmarshal(secretJson, &secret)
			if err != nil {
				errorMsg := "secret parse error"
				message := err.Error()
				return ListConnections500JSONResponse{
					Error:   &errorMsg,
					Message: &message,
				}, nil
			}
		}

		// Parse config JSON
		var config map[string]interface{}
		if len(configJson) > 0 {
			err = json.Unmarshal(configJson, &config)
			if err != nil {
				errorMsg := "config parse error"
				message := err.Error()
				return ListConnections500JSONResponse{
					Error:   &errorMsg,
					Message: &message,
				}, nil
			}
		}

		connection.Id = &connectionUUID
		connection.UserId = &userUUID
		connection.IntegrationId = &integrationUUID
		connection.Name = &name
		connection.Secret = &secret
		connection.Config = &config
		if usageLimitMonth.Valid {
			usageLimit := int(usageLimitMonth.Int32)
			connection.UsageLimitMonth = &usageLimit
		}
		connection.IsDefault = &isDefault
		connection.CreatedAt = &createdAt
		if lastValidated.Valid {
			connection.LastValidated = &lastValidated.Time
		}
		connectionStatus := ConnectionStatus(status)
		connection.Status = &connectionStatus

		connections = append(connections, connection)
	}

	return ListConnections200JSONResponse(connections), nil
}

// CreateConnection creates a new connection
func (h *OpenAPIHandlers) CreateConnection(ctx context.Context, request CreateConnectionRequestObject) (CreateConnectionResponseObject, error) {
	connectionID := uuid.New()
	now := time.Now()

	// For now, use a default user_id (in real implementation, this would come from auth context)
	defaultUserID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	// Marshal secret to JSON (use nil for database if not provided)
	var secretJson interface{}
	var err error
	if request.Body.Secret != nil {
		secretJson, err = json.Marshal(*request.Body.Secret)
		if err != nil {
			errorMsg := "failed to marshal secret"
			message := err.Error()
			return CreateConnection500JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}
	}

	// Marshal config to JSON (use nil for database if not provided)
	var configJson interface{}
	if request.Body.Config != nil {
		configJson, err = json.Marshal(*request.Body.Config)
		if err != nil {
			errorMsg := "failed to marshal config"
			message := err.Error()
			return CreateConnection500JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}
	}

	// Set defaults
	isDefault := false
	if request.Body.IsDefault != nil {
		isDefault = *request.Body.IsDefault
	}

	var usageLimitMonth *int
	if request.Body.UsageLimitMonth != nil {
		usageLimitMonth = request.Body.UsageLimitMonth
	}

	// Insert connection into database
	_, err = h.db.ExecContext(ctx,
		"INSERT INTO connections (id, user_id, integration_id, name, secret, config, usage_limit_month, is_default, created_at, status) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)",
		connectionID, defaultUserID, request.Body.IntegrationId, request.Body.Name, secretJson, configJson, usageLimitMonth, isDefault, now, string(ConnectionStatusValid))
	if err != nil {
		errorMsg := "failed to create connection"
		message := err.Error()
		return CreateConnection500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	status := ConnectionStatusValid
	connection := Connection{
		Id:            &connectionID,
		UserId:        &defaultUserID,
		IntegrationId: &request.Body.IntegrationId,
		Name:          &request.Body.Name,
		Secret:        request.Body.Secret,
		Config:        request.Body.Config,
		IsDefault:     &isDefault,
		CreatedAt:     &now,
		Status:        &status,
	}

	if usageLimitMonth != nil {
		connection.UsageLimitMonth = usageLimitMonth
	}

	return CreateConnection201JSONResponse(connection), nil
}

// GetConnection retrieves a single connection by ID
func (h *OpenAPIHandlers) GetConnection(ctx context.Context, request GetConnectionRequestObject) (GetConnectionResponseObject, error) {
	var connection Connection
	var id, userID, integrationID, name, status string
	var secretJson, configJson []byte
	var usageLimitMonth sql.NullInt32
	var isDefault bool
	var createdAt time.Time
	var lastValidated sql.NullTime

	err := h.db.QueryRowContext(ctx,
		"SELECT id, user_id, integration_id, name, secret, config, usage_limit_month, is_default, created_at, last_validated, status FROM connections WHERE id = $1",
		request.Id.String()).Scan(&id, &userID, &integrationID, &name, &secretJson, &configJson, &usageLimitMonth, &isDefault, &createdAt, &lastValidated, &status)
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

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		errorMsg := "user uuid parse error"
		message := err.Error()
		return GetConnection500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	integrationUUID, err := uuid.Parse(integrationID)
	if err != nil {
		errorMsg := "integration uuid parse error"
		message := err.Error()
		return GetConnection500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Parse secret JSON
	var secret map[string]interface{}
	if len(secretJson) > 0 {
		err = json.Unmarshal(secretJson, &secret)
		if err != nil {
			errorMsg := "secret parse error"
			message := err.Error()
			return GetConnection500JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}
	}

	// Parse config JSON
	var config map[string]interface{}
	if len(configJson) > 0 {
		err = json.Unmarshal(configJson, &config)
		if err != nil {
			errorMsg := "config parse error"
			message := err.Error()
			return GetConnection500JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}
	}

	connection.Id = &connectionUUID
	connection.UserId = &userUUID
	connection.IntegrationId = &integrationUUID
	connection.Name = &name
	connection.Secret = &secret
	connection.Config = &config
	if usageLimitMonth.Valid {
		usageLimit := int(usageLimitMonth.Int32)
		connection.UsageLimitMonth = &usageLimit
	}
	connection.IsDefault = &isDefault
	connection.CreatedAt = &createdAt
	if lastValidated.Valid {
		connection.LastValidated = &lastValidated.Time
	}
	connectionStatus := ConnectionStatus(status)
	connection.Status = &connectionStatus

	return GetConnection200JSONResponse(connection), nil
}

// UpdateConnection updates an existing connection
func (h *OpenAPIHandlers) UpdateConnection(ctx context.Context, request UpdateConnectionRequestObject) (UpdateConnectionResponseObject, error) {
	// Start building the update query
	var setParts []string
	var args []interface{}
	argIndex := 1

	if request.Body.Name != nil {
		setParts = append(setParts, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, *request.Body.Name)
		argIndex++
	}

	if request.Body.Secret != nil {
		secretJson, err := json.Marshal(*request.Body.Secret)
		if err != nil {
			errorMsg := "failed to marshal secret"
			message := err.Error()
			return UpdateConnection500JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}
		setParts = append(setParts, fmt.Sprintf("secret = $%d", argIndex))
		args = append(args, secretJson)
		argIndex++
	}

	if request.Body.Config != nil {
		configJson, err := json.Marshal(*request.Body.Config)
		if err != nil {
			errorMsg := "failed to marshal config"
			message := err.Error()
			return UpdateConnection500JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}
		setParts = append(setParts, fmt.Sprintf("config = $%d", argIndex))
		args = append(args, configJson)
		argIndex++
	}

	if request.Body.UsageLimitMonth != nil {
		setParts = append(setParts, fmt.Sprintf("usage_limit_month = $%d", argIndex))
		args = append(args, *request.Body.UsageLimitMonth)
		argIndex++
	}

	if request.Body.IsDefault != nil {
		setParts = append(setParts, fmt.Sprintf("is_default = $%d", argIndex))
		args = append(args, *request.Body.IsDefault)
		argIndex++
	}

	if request.Body.Status != nil {
		setParts = append(setParts, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, string(*request.Body.Status))
		argIndex++
	}

	// Check if we have anything to update
	if len(setParts) == 0 {
		// No fields to update, just return the existing connection
		getRequest := GetConnectionRequestObject{Id: request.Id}
		getResponse, err := h.GetConnection(ctx, getRequest)
		if err != nil {
			errorMsg := "failed to fetch connection"
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

		errorMsg := "unexpected response type"
		message := "Failed to convert response"
		return UpdateConnection500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
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
