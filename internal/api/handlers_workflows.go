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

// ListWorkflows retrieves all workflows with pagination
func (h *OpenAPIHandlers) ListWorkflows(ctx context.Context, request ListWorkflowsRequestObject) (ListWorkflowsResponseObject, error) {
	page := 1
	limit := 20

	if request.Params.Page != nil {
		page = *request.Params.Page
	}
	if request.Params.Limit != nil {
		limit = *request.Params.Limit
	}

	offset := (page - 1) * limit

	// Get total count
	var total int
	err := h.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM workflows").Scan(&total)
	if err != nil {
		errorMsg := "database error"
		message := err.Error()
		return ListWorkflows500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Get workflows with pagination
	rows, err := h.db.QueryContext(ctx,
		"SELECT id, name, description, definition, created_at, updated_at FROM workflows ORDER BY created_at DESC LIMIT $1 OFFSET $2",
		limit, offset)
	if err != nil {
		errorMsg := "database error"
		message := err.Error()
		return ListWorkflows500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}
	defer rows.Close()

	var workflows []Workflow
	for rows.Next() {
		var workflow Workflow
		var description sql.NullString
		var definitionJson sql.NullString
		var id, name string
		var createdAt, updatedAt time.Time

		err := rows.Scan(&id, &name, &description, &definitionJson, &createdAt, &updatedAt)
		if err != nil {
			errorMsg := "scan error"
			message := err.Error()
			return ListWorkflows500JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}

		workflowUUID, err := uuid.Parse(id)
		if err != nil {
			errorMsg := "uuid parse error"
			message := err.Error()
			return ListWorkflows500JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}

		workflow.Id = workflowUUID
		workflow.Name = name
		if description.Valid {
			workflow.Description = &description.String
		}
		workflow.CreatedAt = createdAt
		workflow.UpdatedAt = updatedAt

		// Parse definition JSON if present
		if definitionJson.Valid && definitionJson.String != "" {
			var definition WorkflowDefinition
			err = json.Unmarshal([]byte(definitionJson.String), &definition)
			if err != nil {
				errorMsg := "definition parse error"
				message := err.Error()
				return ListWorkflows500JSONResponse{
					Error:   &errorMsg,
					Message: &message,
				}, nil
			}
			workflow.Definition = &definition
		}

		workflows = append(workflows, workflow)
	}

	return ListWorkflows200JSONResponse{
		Workflows: &workflows,
		Total:     &total,
		Page:      &page,
		Limit:     &limit,
	}, nil
}

// CreateWorkflow creates a new workflow
func (h *OpenAPIHandlers) CreateWorkflow(ctx context.Context, request CreateWorkflowRequestObject) (CreateWorkflowResponseObject, error) {
	workflowID := uuid.New()
	now := time.Now()

	var description *string
	if request.Body.Description != nil {
		description = request.Body.Description
	}

	var definitionJson interface{}
	var err error
	if request.Body.Definition != nil {
		definitionJson, err = json.Marshal(request.Body.Definition)
		if err != nil {
			errorMsg := "failed to marshal definition"
			message := err.Error()
			return CreateWorkflow500JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}
	}

	// For now, use a default user_id (in real implementation, this would come from auth context)
	defaultUserID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	// Insert workflow into database
	_, err = h.db.ExecContext(ctx,
		"INSERT INTO workflows (id, user_id, name, description, definition, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		workflowID, defaultUserID, request.Body.Name, description, definitionJson, now, now)
	if err != nil {
		errorMsg := "failed to create workflow"
		message := err.Error()
		return CreateWorkflow500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	workflow := Workflow{
		Id:          workflowID,
		Name:        request.Body.Name,
		Description: description,
		Definition:  request.Body.Definition,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	return CreateWorkflow201JSONResponse(workflow), nil
}

// GetWorkflow retrieves a single workflow by ID
func (h *OpenAPIHandlers) GetWorkflow(ctx context.Context, request GetWorkflowRequestObject) (GetWorkflowResponseObject, error) {
	var workflow Workflow
	var description sql.NullString
	var definitionJson sql.NullString
	var id, name string
	var createdAt, updatedAt time.Time

	err := h.db.QueryRowContext(ctx,
		"SELECT id, name, description, definition, created_at, updated_at FROM workflows WHERE id = $1",
		request.Id.String()).Scan(&id, &name, &description, &definitionJson, &createdAt, &updatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			errorMsg := "not found"
			message := "Workflow not found"
			return GetWorkflow404JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}
		errorMsg := "database error"
		message := err.Error()
		return GetWorkflow500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	workflowUUID, err := uuid.Parse(id)
	if err != nil {
		errorMsg := "uuid parse error"
		message := err.Error()
		return GetWorkflow500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	workflow.Id = workflowUUID
	workflow.Name = name
	if description.Valid {
		workflow.Description = &description.String
	}
	workflow.CreatedAt = createdAt
	workflow.UpdatedAt = updatedAt

	// Parse definition JSON if present
	if definitionJson.Valid && definitionJson.String != "" {
		var definition WorkflowDefinition
		err = json.Unmarshal([]byte(definitionJson.String), &definition)
		if err != nil {
			errorMsg := "definition parse error"
			message := err.Error()
			return GetWorkflow500JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}
		workflow.Definition = &definition
	}

	return GetWorkflow200JSONResponse(workflow), nil
}

// UpdateWorkflow updates an existing workflow
func (h *OpenAPIHandlers) UpdateWorkflow(ctx context.Context, request UpdateWorkflowRequestObject) (UpdateWorkflowResponseObject, error) {
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

	if request.Body.Description != nil {
		setParts = append(setParts, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, *request.Body.Description)
		argIndex++
	}

	if request.Body.Definition != nil {
		definitionJson, err := json.Marshal(*request.Body.Definition)
		if err != nil {
			errorMsg := "failed to marshal definition"
			message := err.Error()
			return UpdateWorkflow500JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}
		setParts = append(setParts, fmt.Sprintf("definition = $%d", argIndex))
		args = append(args, definitionJson)
		argIndex++
	}

	// Add the ID as the last parameter
	args = append(args, request.Id.String())

	query := fmt.Sprintf("UPDATE workflows SET %s WHERE id = $%d",
		strings.Join(setParts, ", "), argIndex)

	result, err := h.db.ExecContext(ctx, query, args...)
	if err != nil {
		errorMsg := "failed to update workflow"
		message := err.Error()
		return UpdateWorkflow500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		errorMsg := "failed to check update result"
		message := err.Error()
		return UpdateWorkflow500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	if rowsAffected == 0 {
		errorMsg := "not found"
		message := "Workflow not found"
		return UpdateWorkflow404JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Fetch the updated workflow
	getRequest := GetWorkflowRequestObject{Id: request.Id}
	getResponse, err := h.GetWorkflow(ctx, getRequest)
	if err != nil {
		errorMsg := "failed to fetch updated workflow"
		message := err.Error()
		return UpdateWorkflow500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Convert GetWorkflowResponseObject to UpdateWorkflowResponseObject
	if workflowResponse, ok := getResponse.(GetWorkflow200JSONResponse); ok {
		return UpdateWorkflow200JSONResponse(workflowResponse), nil
	}

	// If we get here, something went wrong
	errorMsg := "unexpected response type"
	message := "Failed to convert response"
	return UpdateWorkflow500JSONResponse{
		Error:   &errorMsg,
		Message: &message,
	}, nil
}

// DeleteWorkflow removes a workflow
func (h *OpenAPIHandlers) DeleteWorkflow(ctx context.Context, request DeleteWorkflowRequestObject) (DeleteWorkflowResponseObject, error) {
	result, err := h.db.ExecContext(ctx, "DELETE FROM workflows WHERE id = $1", request.Id.String())
	if err != nil {
		errorMsg := "failed to delete workflow"
		message := err.Error()
		return DeleteWorkflow500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		errorMsg := "failed to check delete result"
		message := err.Error()
		return DeleteWorkflow500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	if rowsAffected == 0 {
		errorMsg := "not found"
		message := "Workflow not found"
		return DeleteWorkflow404JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	return DeleteWorkflow204Response{}, nil
}

// ExecuteWorkflow executes a workflow
func (h *OpenAPIHandlers) ExecuteWorkflow(ctx context.Context, request ExecuteWorkflowRequestObject) (ExecuteWorkflowResponseObject, error) {
	// First, verify the workflow exists
	var workflowName string
	err := h.db.QueryRowContext(ctx,
		"SELECT name FROM workflows WHERE id = $1",
		request.Id.String()).Scan(&workflowName)
	if err != nil {
		if err == sql.ErrNoRows {
			errorMsg := "not found"
			message := "Workflow not found"
			return ExecuteWorkflow404JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}
		errorMsg := "database error"
		message := err.Error()
		return ExecuteWorkflow500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Create a workflow execution record
	executionID := uuid.New()
	now := time.Now()

	var inputJson interface{}
	if request.Body != nil && request.Body.Input != nil {
		inputJson, err = json.Marshal(*request.Body.Input)
		if err != nil {
			errorMsg := "failed to marshal input"
			message := err.Error()
			return ExecuteWorkflow500JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}
	}

	// For now, we'll create a simple execution record
	// In a real implementation, this would trigger the workflow engine
	_, err = h.db.ExecContext(ctx,
		"INSERT INTO workflow_executions (id, workflow_id, status, started_at, input) VALUES ($1, $2, $3, $4, $5)",
		executionID, request.Id.String(), "pending", now, inputJson)
	if err != nil {
		errorMsg := "failed to create execution"
		message := err.Error()
		return ExecuteWorkflow500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// TODO: Actually trigger the workflow execution via the engine
	// This would involve calling h.engine.Execute() or similar

	execution := WorkflowExecution{
		Id:         &executionID,
		WorkflowId: &request.Id,
		Status:     func() *WorkflowExecutionStatus { s := WorkflowExecutionStatusPending; return &s }(),
		StartedAt:  &now,
	}

	if request.Body != nil && request.Body.Input != nil {
		execution.Result = request.Body.Input
	}

	return ExecuteWorkflow200JSONResponse(execution), nil
}
