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

// ListTriggers retrieves all triggers
func (h *OpenAPIHandlers) ListTriggers(ctx context.Context, request ListTriggersRequestObject) (ListTriggersResponseObject, error) {
	rows, err := h.db.QueryContext(ctx,
		"SELECT id, name, type, workflow_id, config, enabled, created_at, updated_at FROM triggers ORDER BY created_at DESC")
	if err != nil {
		errorMsg := "database error"
		message := err.Error()
		return ListTriggers500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}
	defer rows.Close()

	var triggers []Trigger
	for rows.Next() {
		var trigger Trigger
		var id, name, triggerType, workflowID string
		var configJson []byte
		var enabled bool
		var createdAt, updatedAt time.Time

		err := rows.Scan(&id, &name, &triggerType, &workflowID, &configJson, &enabled, &createdAt, &updatedAt)
		if err != nil {
			errorMsg := "scan error"
			message := err.Error()
			return ListTriggers500JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}

		triggerUUID, err := uuid.Parse(id)
		if err != nil {
			errorMsg := "uuid parse error"
			message := err.Error()
			return ListTriggers500JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}

		workflowUUID, err := uuid.Parse(workflowID)
		if err != nil {
			errorMsg := "workflow uuid parse error"
			message := err.Error()
			return ListTriggers500JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}

		// Parse config JSON
		var config map[string]interface{}
		if len(configJson) > 0 {
			err = json.Unmarshal(configJson, &config)
			if err != nil {
				errorMsg := "config parse error"
				message := err.Error()
				return ListTriggers500JSONResponse{
					Error:   &errorMsg,
					Message: &message,
				}, nil
			}
		}

		trigger.Id = &triggerUUID
		trigger.Name = &name
		trigger.Type = func() *TriggerType {
			if triggerType == "schedule" {
				t := TriggerTypeSchedule
				return &t
			} else if triggerType == "webhook" {
				t := TriggerTypeWebhook
				return &t
			}
			return nil
		}()
		trigger.WorkflowId = &workflowUUID
		trigger.Config = &config
		trigger.Enabled = &enabled
		trigger.CreatedAt = &createdAt
		trigger.UpdatedAt = &updatedAt

		triggers = append(triggers, trigger)
	}

	return ListTriggers200JSONResponse(triggers), nil
}

// CreateTrigger creates a new trigger
func (h *OpenAPIHandlers) CreateTrigger(ctx context.Context, request CreateTriggerRequestObject) (CreateTriggerResponseObject, error) {
	triggerID := uuid.New()
	now := time.Now()

	enabled := true
	if request.Body.Enabled != nil {
		enabled = *request.Body.Enabled
	}

	// Marshal config to JSON
	var configJson []byte
	var err error
	if request.Body.Config != nil {
		configJson, err = json.Marshal(*request.Body.Config)
		if err != nil {
			errorMsg := "failed to marshal config"
			message := err.Error()
			return CreateTrigger500JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}
	}

	// Insert trigger into database
	_, err = h.db.ExecContext(ctx,
		"INSERT INTO triggers (id, name, type, workflow_id, config, enabled, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		triggerID, request.Body.Name, string(request.Body.Type), request.Body.WorkflowId.String(), configJson, enabled, now, now)
	if err != nil {
		errorMsg := "failed to create trigger"
		message := err.Error()
		return CreateTrigger500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	triggerType := TriggerType(request.Body.Type)
	trigger := Trigger{
		Id:         &triggerID,
		Name:       &request.Body.Name,
		Type:       &triggerType,
		WorkflowId: &request.Body.WorkflowId,
		Config:     request.Body.Config,
		Enabled:    &enabled,
		CreatedAt:  &now,
		UpdatedAt:  &now,
	}

	return CreateTrigger201JSONResponse(trigger), nil
}

// GetTrigger retrieves a single trigger by ID
func (h *OpenAPIHandlers) GetTrigger(ctx context.Context, request GetTriggerRequestObject) (GetTriggerResponseObject, error) {
	var trigger Trigger
	var id, name, triggerType, workflowID string
	var configJson []byte
	var enabled bool
	var createdAt, updatedAt time.Time

	err := h.db.QueryRowContext(ctx,
		"SELECT id, name, type, workflow_id, config, enabled, created_at, updated_at FROM triggers WHERE id = $1",
		request.Id.String()).Scan(&id, &name, &triggerType, &workflowID, &configJson, &enabled, &createdAt, &updatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			errorMsg := "not found"
			message := "Trigger not found"
			return GetTrigger404JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}
		errorMsg := "database error"
		message := err.Error()
		return GetTrigger500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	triggerUUID, err := uuid.Parse(id)
	if err != nil {
		errorMsg := "uuid parse error"
		message := err.Error()
		return GetTrigger500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	workflowUUID, err := uuid.Parse(workflowID)
	if err != nil {
		errorMsg := "workflow uuid parse error"
		message := err.Error()
		return GetTrigger500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Parse config JSON
	var config map[string]interface{}
	if len(configJson) > 0 {
		err = json.Unmarshal(configJson, &config)
		if err != nil {
			errorMsg := "config parse error"
			message := err.Error()
			return GetTrigger500JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}
	}

	trigger.Id = &triggerUUID
	trigger.Name = &name
	trigger.Type = func() *TriggerType {
		if triggerType == "schedule" {
			t := TriggerTypeSchedule
			return &t
		} else if triggerType == "webhook" {
			t := TriggerTypeWebhook
			return &t
		}
		return nil
	}()
	trigger.WorkflowId = &workflowUUID
	trigger.Config = &config
	trigger.Enabled = &enabled
	trigger.CreatedAt = &createdAt
	trigger.UpdatedAt = &updatedAt

	return GetTrigger200JSONResponse(trigger), nil
}

// UpdateTrigger updates an existing trigger
func (h *OpenAPIHandlers) UpdateTrigger(ctx context.Context, request UpdateTriggerRequestObject) (UpdateTriggerResponseObject, error) {
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

	if request.Body.Config != nil {
		configJson, err := json.Marshal(*request.Body.Config)
		if err != nil {
			errorMsg := "failed to marshal config"
			message := err.Error()
			return UpdateTrigger500JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}
		setParts = append(setParts, fmt.Sprintf("config = $%d", argIndex))
		args = append(args, configJson)
		argIndex++
	}

	if request.Body.Enabled != nil {
		setParts = append(setParts, fmt.Sprintf("enabled = $%d", argIndex))
		args = append(args, *request.Body.Enabled)
		argIndex++
	}

	// Add the ID as the last parameter
	args = append(args, request.Id.String())

	query := fmt.Sprintf("UPDATE triggers SET %s WHERE id = $%d",
		strings.Join(setParts, ", "), argIndex)

	result, err := h.db.ExecContext(ctx, query, args...)
	if err != nil {
		errorMsg := "failed to update trigger"
		message := err.Error()
		return UpdateTrigger500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		errorMsg := "failed to check update result"
		message := err.Error()
		return UpdateTrigger500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	if rowsAffected == 0 {
		errorMsg := "not found"
		message := "Trigger not found"
		return UpdateTrigger404JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Fetch the updated trigger
	getRequest := GetTriggerRequestObject{Id: request.Id}
	getResponse, err := h.GetTrigger(ctx, getRequest)
	if err != nil {
		errorMsg := "failed to fetch updated trigger"
		message := err.Error()
		return UpdateTrigger500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Convert GetTriggerResponseObject to UpdateTriggerResponseObject
	if triggerResponse, ok := getResponse.(GetTrigger200JSONResponse); ok {
		return UpdateTrigger200JSONResponse(triggerResponse), nil
	}

	// If we get here, something went wrong
	errorMsg := "unexpected response type"
	message := "Failed to convert response"
	return UpdateTrigger500JSONResponse{
		Error:   &errorMsg,
		Message: &message,
	}, nil
}

// DeleteTrigger removes a trigger
func (h *OpenAPIHandlers) DeleteTrigger(ctx context.Context, request DeleteTriggerRequestObject) (DeleteTriggerResponseObject, error) {
	result, err := h.db.ExecContext(ctx, "DELETE FROM triggers WHERE id = $1", request.Id.String())
	if err != nil {
		errorMsg := "failed to delete trigger"
		message := err.Error()
		return DeleteTrigger500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		errorMsg := "failed to check delete result"
		message := err.Error()
		return DeleteTrigger500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	if rowsAffected == 0 {
		errorMsg := "not found"
		message := "Trigger not found"
		return DeleteTrigger404JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	return DeleteTrigger204Response{}, nil
}
