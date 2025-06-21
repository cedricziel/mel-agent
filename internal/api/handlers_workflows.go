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

	workflows := make([]Workflow, 0)
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
		Workflows: workflows,
		Total:     total,
		Page:      page,
		Limit:     limit,
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

// ListWorkflowNodes lists all nodes in a workflow
func (h *OpenAPIHandlers) ListWorkflowNodes(ctx context.Context, request ListWorkflowNodesRequestObject) (ListWorkflowNodesResponseObject, error) {
	// Check if workflow exists
	var workflowExists bool
	err := h.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM workflows WHERE id = $1)", request.WorkflowId.String()).Scan(&workflowExists)
	if err != nil {
		errorMsg := "database error"
		message := err.Error()
		return ListWorkflowNodes500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	if !workflowExists {
		errorMsg := "not found"
		message := "Workflow not found"
		return ListWorkflowNodes404JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Get workflow definition and extract nodes
	var definitionJson sql.NullString
	err = h.db.QueryRowContext(ctx, "SELECT definition FROM workflows WHERE id = $1", request.WorkflowId.String()).Scan(&definitionJson)
	if err != nil {
		errorMsg := "database error"
		message := err.Error()
		return ListWorkflowNodes500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	var nodes []WorkflowNode
	if definitionJson.Valid && definitionJson.String != "" {
		var definition WorkflowDefinition
		err = json.Unmarshal([]byte(definitionJson.String), &definition)
		if err != nil {
			errorMsg := "definition parse error"
			message := err.Error()
			return ListWorkflowNodes500JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}

		if definition.Nodes != nil {
			nodes = *definition.Nodes
		}
	}

	return ListWorkflowNodes200JSONResponse(nodes), nil
}

// CreateWorkflowNode creates a new node in workflow
func (h *OpenAPIHandlers) CreateWorkflowNode(ctx context.Context, request CreateWorkflowNodeRequestObject) (CreateWorkflowNodeResponseObject, error) {
	// Check if workflow exists
	var workflowExists bool
	err := h.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM workflows WHERE id = $1)", request.WorkflowId.String()).Scan(&workflowExists)
	if err != nil {
		errorMsg := "database error"
		message := err.Error()
		return CreateWorkflowNode500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	if !workflowExists {
		errorMsg := "not found"
		message := "Workflow not found"
		return CreateWorkflowNode404JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Get current workflow definition
	var definitionJson sql.NullString
	err = h.db.QueryRowContext(ctx, "SELECT definition FROM workflows WHERE id = $1", request.WorkflowId.String()).Scan(&definitionJson)
	if err != nil {
		errorMsg := "database error"
		message := err.Error()
		return CreateWorkflowNode500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	var definition WorkflowDefinition
	if definitionJson.Valid && definitionJson.String != "" {
		err = json.Unmarshal([]byte(definitionJson.String), &definition)
		if err != nil {
			errorMsg := "definition parse error"
			message := err.Error()
			return CreateWorkflowNode500JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}
	}

	// Create new node
	newNode := WorkflowNode{
		Id:     request.Body.Id,
		Name:   request.Body.Name,
		Type:   request.Body.Type,
		Config: request.Body.Config,
	}

	// Add to nodes array
	if definition.Nodes == nil {
		definition.Nodes = &[]WorkflowNode{}
	}
	*definition.Nodes = append(*definition.Nodes, newNode)

	// Update workflow definition in database
	updatedDefinitionJson, err := json.Marshal(definition)
	if err != nil {
		errorMsg := "definition marshal error"
		message := err.Error()
		return CreateWorkflowNode500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	_, err = h.db.ExecContext(ctx,
		"UPDATE workflows SET definition = $1, updated_at = $2 WHERE id = $3",
		string(updatedDefinitionJson), time.Now(), request.WorkflowId.String())
	if err != nil {
		errorMsg := "failed to update workflow"
		message := err.Error()
		return CreateWorkflowNode500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	return CreateWorkflowNode201JSONResponse(newNode), nil
}

// GetWorkflowNode retrieves a specific workflow node
func (h *OpenAPIHandlers) GetWorkflowNode(ctx context.Context, request GetWorkflowNodeRequestObject) (GetWorkflowNodeResponseObject, error) {
	// Check if workflow exists
	var workflowExists bool
	err := h.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM workflows WHERE id = $1)", request.WorkflowId.String()).Scan(&workflowExists)
	if err != nil {
		errorMsg := "database error"
		message := err.Error()
		return GetWorkflowNode500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	if !workflowExists {
		errorMsg := "not found"
		message := "Workflow not found"
		return GetWorkflowNode404JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Get workflow definition
	var definitionJson sql.NullString
	err = h.db.QueryRowContext(ctx, "SELECT definition FROM workflows WHERE id = $1", request.WorkflowId.String()).Scan(&definitionJson)
	if err != nil {
		errorMsg := "database error"
		message := err.Error()
		return GetWorkflowNode500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	if !definitionJson.Valid || definitionJson.String == "" {
		errorMsg := "not found"
		message := "Node not found"
		return GetWorkflowNode404JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	var definition WorkflowDefinition
	err = json.Unmarshal([]byte(definitionJson.String), &definition)
	if err != nil {
		errorMsg := "definition parse error"
		message := err.Error()
		return GetWorkflowNode500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Find the node
	if definition.Nodes != nil {
		for _, node := range *definition.Nodes {
			if node.Id == request.NodeId {
				return GetWorkflowNode200JSONResponse(node), nil
			}
		}
	}

	errorMsg := "not found"
	message := "Node not found"
	return GetWorkflowNode404JSONResponse{
		Error:   &errorMsg,
		Message: &message,
	}, nil
}

// UpdateWorkflowNode updates a workflow node
func (h *OpenAPIHandlers) UpdateWorkflowNode(ctx context.Context, request UpdateWorkflowNodeRequestObject) (UpdateWorkflowNodeResponseObject, error) {
	// Check if workflow exists
	var workflowExists bool
	err := h.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM workflows WHERE id = $1)", request.WorkflowId.String()).Scan(&workflowExists)
	if err != nil {
		errorMsg := "database error"
		message := err.Error()
		return UpdateWorkflowNode500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	if !workflowExists {
		errorMsg := "not found"
		message := "Workflow not found"
		return UpdateWorkflowNode404JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Get current workflow definition
	var definitionJson sql.NullString
	err = h.db.QueryRowContext(ctx, "SELECT definition FROM workflows WHERE id = $1", request.WorkflowId.String()).Scan(&definitionJson)
	if err != nil {
		errorMsg := "database error"
		message := err.Error()
		return UpdateWorkflowNode500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	if !definitionJson.Valid || definitionJson.String == "" {
		errorMsg := "not found"
		message := "Node not found"
		return UpdateWorkflowNode404JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	var definition WorkflowDefinition
	err = json.Unmarshal([]byte(definitionJson.String), &definition)
	if err != nil {
		errorMsg := "definition parse error"
		message := err.Error()
		return UpdateWorkflowNode500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Find and update the node
	var updatedNode *WorkflowNode
	if definition.Nodes != nil {
		for i, node := range *definition.Nodes {
			if node.Id == request.NodeId {
				// Update fields if provided
				if request.Body.Name != nil {
					(*definition.Nodes)[i].Name = *request.Body.Name
				}
				if request.Body.Config != nil {
					(*definition.Nodes)[i].Config = *request.Body.Config
				}
				updatedNode = &(*definition.Nodes)[i]
				break
			}
		}
	}

	if updatedNode == nil {
		errorMsg := "not found"
		message := "Node not found"
		return UpdateWorkflowNode404JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Update workflow definition in database
	updatedDefinitionJson, err := json.Marshal(definition)
	if err != nil {
		errorMsg := "definition marshal error"
		message := err.Error()
		return UpdateWorkflowNode500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	_, err = h.db.ExecContext(ctx,
		"UPDATE workflows SET definition = $1, updated_at = $2 WHERE id = $3",
		string(updatedDefinitionJson), time.Now(), request.WorkflowId.String())
	if err != nil {
		errorMsg := "failed to update workflow"
		message := err.Error()
		return UpdateWorkflowNode500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	return UpdateWorkflowNode200JSONResponse(*updatedNode), nil
}

// DeleteWorkflowNode deletes a workflow node
func (h *OpenAPIHandlers) DeleteWorkflowNode(ctx context.Context, request DeleteWorkflowNodeRequestObject) (DeleteWorkflowNodeResponseObject, error) {
	// Check if workflow exists
	var workflowExists bool
	err := h.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM workflows WHERE id = $1)", request.WorkflowId.String()).Scan(&workflowExists)
	if err != nil {
		errorMsg := "database error"
		message := err.Error()
		return DeleteWorkflowNode500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	if !workflowExists {
		errorMsg := "not found"
		message := "Workflow not found"
		return DeleteWorkflowNode404JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Get current workflow definition
	var definitionJson sql.NullString
	err = h.db.QueryRowContext(ctx, "SELECT definition FROM workflows WHERE id = $1", request.WorkflowId.String()).Scan(&definitionJson)
	if err != nil {
		errorMsg := "database error"
		message := err.Error()
		return DeleteWorkflowNode500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	if !definitionJson.Valid || definitionJson.String == "" {
		errorMsg := "not found"
		message := "Node not found"
		return DeleteWorkflowNode404JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	var definition WorkflowDefinition
	err = json.Unmarshal([]byte(definitionJson.String), &definition)
	if err != nil {
		errorMsg := "definition parse error"
		message := err.Error()
		return DeleteWorkflowNode500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Find and remove the node
	found := false
	if definition.Nodes != nil {
		for i, node := range *definition.Nodes {
			if node.Id == request.NodeId {
				// Remove node from slice
				*definition.Nodes = append((*definition.Nodes)[:i], (*definition.Nodes)[i+1:]...)
				found = true
				break
			}
		}
	}

	if !found {
		errorMsg := "not found"
		message := "Node not found"
		return DeleteWorkflowNode404JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Update workflow definition in database
	updatedDefinitionJson, err := json.Marshal(definition)
	if err != nil {
		errorMsg := "definition marshal error"
		message := err.Error()
		return DeleteWorkflowNode500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	_, err = h.db.ExecContext(ctx,
		"UPDATE workflows SET definition = $1, updated_at = $2 WHERE id = $3",
		string(updatedDefinitionJson), time.Now(), request.WorkflowId.String())
	if err != nil {
		errorMsg := "failed to update workflow"
		message := err.Error()
		return DeleteWorkflowNode500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	return DeleteWorkflowNode204Response{}, nil
}

// ListWorkflowEdges lists all edges in a workflow
func (h *OpenAPIHandlers) ListWorkflowEdges(ctx context.Context, request ListWorkflowEdgesRequestObject) (ListWorkflowEdgesResponseObject, error) {
	// Check if workflow exists
	var workflowExists bool
	err := h.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM workflows WHERE id = $1)", request.WorkflowId.String()).Scan(&workflowExists)
	if err != nil {
		errorMsg := "database error"
		message := err.Error()
		return ListWorkflowEdges500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	if !workflowExists {
		errorMsg := "not found"
		message := "Workflow not found"
		return ListWorkflowEdges404JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Get workflow definition and extract edges
	var definitionJson sql.NullString
	err = h.db.QueryRowContext(ctx, "SELECT definition FROM workflows WHERE id = $1", request.WorkflowId.String()).Scan(&definitionJson)
	if err != nil {
		errorMsg := "database error"
		message := err.Error()
		return ListWorkflowEdges500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	var edges []WorkflowEdge
	if definitionJson.Valid && definitionJson.String != "" {
		var definition WorkflowDefinition
		err = json.Unmarshal([]byte(definitionJson.String), &definition)
		if err != nil {
			errorMsg := "definition parse error"
			message := err.Error()
			return ListWorkflowEdges500JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}

		if definition.Edges != nil {
			edges = *definition.Edges
		}
	}

	return ListWorkflowEdges200JSONResponse(edges), nil
}

// CreateWorkflowEdge creates a new edge in workflow
func (h *OpenAPIHandlers) CreateWorkflowEdge(ctx context.Context, request CreateWorkflowEdgeRequestObject) (CreateWorkflowEdgeResponseObject, error) {
	// Check if workflow exists
	var workflowExists bool
	err := h.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM workflows WHERE id = $1)", request.WorkflowId.String()).Scan(&workflowExists)
	if err != nil {
		errorMsg := "database error"
		message := err.Error()
		return CreateWorkflowEdge500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	if !workflowExists {
		errorMsg := "not found"
		message := "Workflow not found"
		return CreateWorkflowEdge404JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Get current workflow definition
	var definitionJson sql.NullString
	err = h.db.QueryRowContext(ctx, "SELECT definition FROM workflows WHERE id = $1", request.WorkflowId.String()).Scan(&definitionJson)
	if err != nil {
		errorMsg := "database error"
		message := err.Error()
		return CreateWorkflowEdge500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	var definition WorkflowDefinition
	if definitionJson.Valid && definitionJson.String != "" {
		err = json.Unmarshal([]byte(definitionJson.String), &definition)
		if err != nil {
			errorMsg := "definition parse error"
			message := err.Error()
			return CreateWorkflowEdge500JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}
	}

	// Create new edge
	newEdge := WorkflowEdge{
		Id:     request.Body.Id,
		Source: request.Body.Source,
		Target: request.Body.Target,
	}

	if request.Body.SourceOutput != nil {
		newEdge.SourceOutput = request.Body.SourceOutput
	}

	if request.Body.TargetInput != nil {
		newEdge.TargetInput = request.Body.TargetInput
	}

	// Add to edges array
	if definition.Edges == nil {
		definition.Edges = &[]WorkflowEdge{}
	}
	*definition.Edges = append(*definition.Edges, newEdge)

	// Update workflow definition in database
	updatedDefinitionJson, err := json.Marshal(definition)
	if err != nil {
		errorMsg := "definition marshal error"
		message := err.Error()
		return CreateWorkflowEdge500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	_, err = h.db.ExecContext(ctx,
		"UPDATE workflows SET definition = $1, updated_at = $2 WHERE id = $3",
		string(updatedDefinitionJson), time.Now(), request.WorkflowId.String())
	if err != nil {
		errorMsg := "failed to update workflow"
		message := err.Error()
		return CreateWorkflowEdge500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	return CreateWorkflowEdge201JSONResponse(newEdge), nil
}

// DeleteWorkflowEdge deletes a workflow edge
func (h *OpenAPIHandlers) DeleteWorkflowEdge(ctx context.Context, request DeleteWorkflowEdgeRequestObject) (DeleteWorkflowEdgeResponseObject, error) {
	// Check if workflow exists
	var workflowExists bool
	err := h.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM workflows WHERE id = $1)", request.WorkflowId.String()).Scan(&workflowExists)
	if err != nil {
		errorMsg := "database error"
		message := err.Error()
		return DeleteWorkflowEdge500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	if !workflowExists {
		errorMsg := "not found"
		message := "Workflow not found"
		return DeleteWorkflowEdge404JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Get current workflow definition
	var definitionJson sql.NullString
	err = h.db.QueryRowContext(ctx, "SELECT definition FROM workflows WHERE id = $1", request.WorkflowId.String()).Scan(&definitionJson)
	if err != nil {
		errorMsg := "database error"
		message := err.Error()
		return DeleteWorkflowEdge500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	if !definitionJson.Valid || definitionJson.String == "" {
		errorMsg := "not found"
		message := "Edge not found"
		return DeleteWorkflowEdge404JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	var definition WorkflowDefinition
	err = json.Unmarshal([]byte(definitionJson.String), &definition)
	if err != nil {
		errorMsg := "definition parse error"
		message := err.Error()
		return DeleteWorkflowEdge500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Find and remove the edge
	found := false
	if definition.Edges != nil {
		for i, edge := range *definition.Edges {
			if edge.Id == request.EdgeId {
				// Remove edge from slice
				*definition.Edges = append((*definition.Edges)[:i], (*definition.Edges)[i+1:]...)
				found = true
				break
			}
		}
	}

	if !found {
		errorMsg := "not found"
		message := "Edge not found"
		return DeleteWorkflowEdge404JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Update workflow definition in database
	updatedDefinitionJson, err := json.Marshal(definition)
	if err != nil {
		errorMsg := "definition marshal error"
		message := err.Error()
		return DeleteWorkflowEdge500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	_, err = h.db.ExecContext(ctx,
		"UPDATE workflows SET definition = $1, updated_at = $2 WHERE id = $3",
		string(updatedDefinitionJson), time.Now(), request.WorkflowId.String())
	if err != nil {
		errorMsg := "failed to update workflow"
		message := err.Error()
		return DeleteWorkflowEdge500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	return DeleteWorkflowEdge204Response{}, nil
}

// AutoLayoutWorkflow auto-layouts workflow nodes
func (h *OpenAPIHandlers) AutoLayoutWorkflow(ctx context.Context, request AutoLayoutWorkflowRequestObject) (AutoLayoutWorkflowResponseObject, error) {
	// Check if workflow exists
	var workflowExists bool
	err := h.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM workflows WHERE id = $1)", request.WorkflowId.String()).Scan(&workflowExists)
	if err != nil {
		errorMsg := "database error"
		message := err.Error()
		return AutoLayoutWorkflow500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	if !workflowExists {
		errorMsg := "not found"
		message := "Workflow not found"
		return AutoLayoutWorkflow404JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// For now, return a simple mock layout
	// In a real implementation, this would apply graph layout algorithms
	nodes := []struct {
		Id       *string `json:"id,omitempty"`
		Position *struct {
			X *float32 `json:"x,omitempty"`
			Y *float32 `json:"y,omitempty"`
		} `json:"position,omitempty"`
	}{
		{
			Id: func() *string { s := "start"; return &s }(),
			Position: &struct {
				X *float32 `json:"x,omitempty"`
				Y *float32 `json:"y,omitempty"`
			}{
				X: func() *float32 { f := float32(100); return &f }(),
				Y: func() *float32 { f := float32(100); return &f }(),
			},
		},
		{
			Id: func() *string { s := "process"; return &s }(),
			Position: &struct {
				X *float32 `json:"x,omitempty"`
				Y *float32 `json:"y,omitempty"`
			}{
				X: func() *float32 { f := float32(300); return &f }(),
				Y: func() *float32 { f := float32(100); return &f }(),
			},
		},
		{
			Id: func() *string { s := "end"; return &s }(),
			Position: &struct {
				X *float32 `json:"x,omitempty"`
				Y *float32 `json:"y,omitempty"`
			}{
				X: func() *float32 { f := float32(500); return &f }(),
				Y: func() *float32 { f := float32(100); return &f }(),
			},
		},
	}

	result := WorkflowLayoutResult{
		Nodes: &nodes,
	}

	return AutoLayoutWorkflow200JSONResponse(result), nil
}

// CreateWorkflowVersion creates a new version of a workflow
func (h *OpenAPIHandlers) CreateWorkflowVersion(ctx context.Context, request CreateWorkflowVersionRequestObject) (CreateWorkflowVersionResponseObject, error) {
	// Check if workflow exists
	var workflowExists bool
	err := h.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM workflows WHERE id = $1)", request.WorkflowId.String()).Scan(&workflowExists)
	if err != nil {
		errorMsg := "database error"
		message := err.Error()
		return CreateWorkflowVersion500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	if !workflowExists {
		errorMsg := "not found"
		message := "Workflow not found"
		return CreateWorkflowVersion404JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Validate request body
	if request.Body == nil {
		errorMsg := "bad request"
		message := "Request body is required"
		return CreateWorkflowVersion400JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Generate new version number
	var nextVersion int
	err = h.db.QueryRowContext(ctx, 
		"SELECT COALESCE(MAX(version_number), 0) + 1 FROM workflow_versions WHERE workflow_id = $1", 
		request.WorkflowId.String()).Scan(&nextVersion)
	if err != nil {
		errorMsg := "database error"
		message := err.Error()
		return CreateWorkflowVersion500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Get current workflow definition to save as version
	var currentDefinition []byte
	err = h.db.QueryRowContext(ctx, "SELECT definition FROM workflows WHERE id = $1", request.WorkflowId.String()).Scan(&currentDefinition)
	if err != nil {
		errorMsg := "database error"
		message := err.Error()
		return CreateWorkflowVersion500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Create new version
	versionID := uuid.New()
	name := request.Body.Name
	description := ""
	if request.Body.Description != nil {
		description = *request.Body.Description
	}
	isCurrent := false // New versions are not current by default

	createdAt := time.Now()
	_, err = h.db.ExecContext(ctx, `
		INSERT INTO workflow_versions 
		(id, workflow_id, version_number, name, description, definition, is_current, created_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		versionID, request.WorkflowId.String(), nextVersion, name, description, currentDefinition, isCurrent, createdAt)

	if err != nil {
		errorMsg := "failed to create version"
		message := err.Error()
		return CreateWorkflowVersion500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Return the created version
	version := WorkflowVersion{
		Id:            &versionID,
		WorkflowId:    &request.WorkflowId,
		VersionNumber: &nextVersion,
		Name:          &name,
		Description:   &description,
		IsCurrent:     &isCurrent,
		CreatedAt:     &createdAt,
	}

	return CreateWorkflowVersion201JSONResponse(version), nil
}

// DeployWorkflowVersion deploys a specific version of a workflow (makes it current)
func (h *OpenAPIHandlers) DeployWorkflowVersion(ctx context.Context, request DeployWorkflowVersionRequestObject) (DeployWorkflowVersionResponseObject, error) {
	// Check if workflow exists
	var workflowExists bool
	err := h.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM workflows WHERE id = $1)", request.WorkflowId.String()).Scan(&workflowExists)
	if err != nil {
		errorMsg := "database error"
		message := err.Error()
		return DeployWorkflowVersion500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	if !workflowExists {
		errorMsg := "not found"
		message := "Workflow not found"
		return DeployWorkflowVersion404JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Check if version exists
	var versionExists bool
	err = h.db.QueryRowContext(ctx, 
		"SELECT EXISTS(SELECT 1 FROM workflow_versions WHERE workflow_id = $1 AND version_number = $2)", 
		request.WorkflowId.String(), request.VersionNumber).Scan(&versionExists)
	if err != nil {
		errorMsg := "database error"
		message := err.Error()
		return DeployWorkflowVersion500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	if !versionExists {
		errorMsg := "not found"
		message := "Version not found"
		return DeployWorkflowVersion404JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Start transaction to ensure atomicity
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		errorMsg := "database error"
		message := err.Error()
		return DeployWorkflowVersion500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}
	defer tx.Rollback()

	// Unset current flag from all versions
	_, err = tx.ExecContext(ctx, 
		"UPDATE workflow_versions SET is_current = false WHERE workflow_id = $1", 
		request.WorkflowId.String())
	if err != nil {
		errorMsg := "database error"
		message := err.Error()
		return DeployWorkflowVersion500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Set the specified version as current
	_, err = tx.ExecContext(ctx, 
		"UPDATE workflow_versions SET is_current = true WHERE workflow_id = $1 AND version_number = $2", 
		request.WorkflowId.String(), request.VersionNumber)
	if err != nil {
		errorMsg := "database error"
		message := err.Error()
		return DeployWorkflowVersion500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		errorMsg := "database error"
		message := err.Error()
		return DeployWorkflowVersion500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Fetch and return the deployed version
	var version WorkflowVersion
	var versionID uuid.UUID
	var name, description string
	var isCurrent bool
	var createdAt time.Time

	err = h.db.QueryRowContext(ctx, `
		SELECT id, workflow_id, version_number, name, description, is_current, created_at 
		FROM workflow_versions 
		WHERE workflow_id = $1 AND version_number = $2`,
		request.WorkflowId.String(), request.VersionNumber).Scan(
		&versionID, &version.WorkflowId, &version.VersionNumber, 
		&name, &description, &isCurrent, &createdAt)

	if err != nil {
		errorMsg := "database error"
		message := err.Error()
		return DeployWorkflowVersion500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	version.Id = &versionID
	version.WorkflowId = &request.WorkflowId
	version.Name = &name
	version.Description = &description
	version.IsCurrent = &isCurrent
	version.CreatedAt = &createdAt

	return DeployWorkflowVersion200JSONResponse(version), nil
}

// GetLatestWorkflowVersion gets the latest version of a workflow
func (h *OpenAPIHandlers) GetLatestWorkflowVersion(ctx context.Context, request GetLatestWorkflowVersionRequestObject) (GetLatestWorkflowVersionResponseObject, error) {
	// Check if workflow exists
	var workflowExists bool
	err := h.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM workflows WHERE id = $1)", request.WorkflowId.String()).Scan(&workflowExists)
	if err != nil {
		errorMsg := "database error"
		message := err.Error()
		return GetLatestWorkflowVersion500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	if !workflowExists {
		errorMsg := "not found"
		message := "Workflow not found"
		return GetLatestWorkflowVersion404JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Get the latest version (highest version_number)
	var version WorkflowVersion
	var versionID uuid.UUID
	var name, description string
	var isCurrent bool
	var createdAt time.Time

	err = h.db.QueryRowContext(ctx, `
		SELECT id, workflow_id, version_number, name, description, is_current, created_at 
		FROM workflow_versions 
		WHERE workflow_id = $1 
		ORDER BY version_number DESC 
		LIMIT 1`,
		request.WorkflowId.String()).Scan(
		&versionID, &version.WorkflowId, &version.VersionNumber, 
		&name, &description, &isCurrent, &createdAt)

	if err != nil {
		if err == sql.ErrNoRows {
			errorMsg := "not found"
			message := "No versions found for this workflow"
			return GetLatestWorkflowVersion404JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}
		errorMsg := "database error"
		message := err.Error()
		return GetLatestWorkflowVersion500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	version.Id = &versionID
	version.WorkflowId = &request.WorkflowId
	version.Name = &name
	version.Description = &description
	version.IsCurrent = &isCurrent
	version.CreatedAt = &createdAt

	return GetLatestWorkflowVersion200JSONResponse(version), nil
}

// GetWorkflowDraft gets the current draft of a workflow
func (h *OpenAPIHandlers) GetWorkflowDraft(ctx context.Context, request GetWorkflowDraftRequestObject) (GetWorkflowDraftResponseObject, error) {
	// Check if workflow exists
	var workflowExists bool
	err := h.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM workflows WHERE id = $1)", request.WorkflowId.String()).Scan(&workflowExists)
	if err != nil {
		errorMsg := "database error"
		message := err.Error()
		return GetWorkflowDraft500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	if !workflowExists {
		errorMsg := "not found"
		message := "Workflow not found"
		return GetWorkflowDraft404JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Try to get existing draft
	var draft WorkflowDraft
	var workflowID uuid.UUID
	var definitionJSON []byte
	var createdAt, updatedAt time.Time

	err = h.db.QueryRowContext(ctx, `
		SELECT workflow_id, definition, created_at, updated_at 
		FROM workflow_drafts 
		WHERE workflow_id = $1`,
		request.WorkflowId.String()).Scan(&workflowID, &definitionJSON, &createdAt, &updatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			// No draft exists, create empty one based on current workflow definition
			var currentDefinition []byte
			err = h.db.QueryRowContext(ctx, "SELECT definition FROM workflows WHERE id = $1", request.WorkflowId.String()).Scan(&currentDefinition)
			if err != nil {
				errorMsg := "database error"
				message := err.Error()
				return GetWorkflowDraft500JSONResponse{
					Error:   &errorMsg,
					Message: &message,
				}, nil
			}

			// Parse current definition
			var workflowDef WorkflowDefinition
			if err := json.Unmarshal(currentDefinition, &workflowDef); err != nil {
				errorMsg := "definition parse error"
				message := err.Error()
				return GetWorkflowDraft500JSONResponse{
					Error:   &errorMsg,
					Message: &message,
				}, nil
			}

			// Return draft based on current workflow
			now := time.Now()
			draft = WorkflowDraft{
				WorkflowId: &request.WorkflowId,
				Definition: &workflowDef,
				UpdatedAt:  &now,
			}
		} else {
			errorMsg := "database error"
			message := err.Error()
			return GetWorkflowDraft500JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}
	} else {
		// Parse existing draft definition
		var workflowDef WorkflowDefinition
		if err := json.Unmarshal(definitionJSON, &workflowDef); err != nil {
			errorMsg := "definition parse error"
			message := err.Error()
			return GetWorkflowDraft500JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}

		draft = WorkflowDraft{
			WorkflowId: &workflowID,
			Definition: &workflowDef,
			UpdatedAt:  &updatedAt,
		}
	}

	return GetWorkflowDraft200JSONResponse(draft), nil
}

// UpdateWorkflowDraft updates the draft of a workflow
func (h *OpenAPIHandlers) UpdateWorkflowDraft(ctx context.Context, request UpdateWorkflowDraftRequestObject) (UpdateWorkflowDraftResponseObject, error) {
	// Check if workflow exists
	var workflowExists bool
	err := h.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM workflows WHERE id = $1)", request.WorkflowId.String()).Scan(&workflowExists)
	if err != nil {
		errorMsg := "database error"
		message := err.Error()
		return UpdateWorkflowDraft500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	if !workflowExists {
		errorMsg := "not found"
		message := "Workflow not found"
		return UpdateWorkflowDraft404JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Validate request body
	if request.Body == nil || request.Body.Definition == nil {
		errorMsg := "bad request"
		message := "Definition is required"
		return UpdateWorkflowDraft400JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Serialize definition
	definitionJSON, err := json.Marshal(request.Body.Definition)
	if err != nil {
		errorMsg := "definition marshal error"
		message := err.Error()
		return UpdateWorkflowDraft500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	now := time.Now()

	// Upsert draft
	_, err = h.db.ExecContext(ctx, `
		INSERT INTO workflow_drafts (workflow_id, definition, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (workflow_id) 
		DO UPDATE SET 
			definition = EXCLUDED.definition,
			updated_at = EXCLUDED.updated_at`,
		request.WorkflowId.String(), definitionJSON, now, now)

	if err != nil {
		errorMsg := "failed to update draft"
		message := err.Error()
		return UpdateWorkflowDraft500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Return updated draft
	draft := WorkflowDraft{
		WorkflowId: &request.WorkflowId,
		Definition: request.Body.Definition,
		UpdatedAt:  &now,
	}

	return UpdateWorkflowDraft200JSONResponse(draft), nil
}

// ListWorkflowVersions lists all versions of a workflow
func (h *OpenAPIHandlers) ListWorkflowVersions(ctx context.Context, request ListWorkflowVersionsRequestObject) (ListWorkflowVersionsResponseObject, error) {
	// Check if workflow exists
	var workflowExists bool
	err := h.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM workflows WHERE id = $1)", request.WorkflowId.String()).Scan(&workflowExists)
	if err != nil {
		errorMsg := "database error"
		message := err.Error()
		return ListWorkflowVersions500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	if !workflowExists {
		errorMsg := "not found"
		message := "Workflow not found"
		return ListWorkflowVersions404JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Get all versions for the workflow
	rows, err := h.db.QueryContext(ctx, `
		SELECT id, workflow_id, version_number, name, description, is_current, created_at 
		FROM workflow_versions 
		WHERE workflow_id = $1 
		ORDER BY version_number DESC`,
		request.WorkflowId.String())
	if err != nil {
		errorMsg := "database error"
		message := err.Error()
		return ListWorkflowVersions500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}
	defer rows.Close()

	versions := make([]WorkflowVersion, 0)
	for rows.Next() {
		var version WorkflowVersion
		var versionID uuid.UUID
		var name, description string
		var isCurrent bool
		var createdAt time.Time

		err := rows.Scan(&versionID, &version.WorkflowId, &version.VersionNumber, 
			&name, &description, &isCurrent, &createdAt)
		if err != nil {
			errorMsg := "scan error"
			message := err.Error()
			return ListWorkflowVersions500JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}

		version.Id = &versionID
		version.WorkflowId = &request.WorkflowId
		version.Name = &name
		version.Description = &description
		version.IsCurrent = &isCurrent
		version.CreatedAt = &createdAt

		versions = append(versions, version)
	}

	total := len(versions)
	versionList := WorkflowVersionList{
		Total:    total,
		Versions: versions,
	}
	return ListWorkflowVersions200JSONResponse(versionList), nil
}

// TestWorkflowDraftNode tests a single node in the workflow draft
func (h *OpenAPIHandlers) TestWorkflowDraftNode(ctx context.Context, request TestWorkflowDraftNodeRequestObject) (TestWorkflowDraftNodeResponseObject, error) {
	// Check if workflow exists
	var workflowExists bool
	err := h.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM workflows WHERE id = $1)", request.WorkflowId.String()).Scan(&workflowExists)
	if err != nil {
		errorMsg := "database error"
		message := err.Error()
		return TestWorkflowDraftNode500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	if !workflowExists {
		errorMsg := "not found"
		message := "Workflow not found"
		return TestWorkflowDraftNode404JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Get draft definition
	var definitionJSON []byte
	err = h.db.QueryRowContext(ctx, `
		SELECT definition 
		FROM workflow_drafts 
		WHERE workflow_id = $1`,
		request.WorkflowId.String()).Scan(&definitionJSON)

	if err != nil {
		if err == sql.ErrNoRows {
			errorMsg := "not found"
			message := "No draft found for this workflow"
			return TestWorkflowDraftNode404JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}
		errorMsg := "database error"
		message := err.Error()
		return TestWorkflowDraftNode500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Parse draft definition
	var workflowDef WorkflowDefinition
	if err := json.Unmarshal(definitionJSON, &workflowDef); err != nil {
		errorMsg := "definition parse error"
		message := err.Error()
		return TestWorkflowDraftNode500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Find the node in the draft
	var targetNode *WorkflowNode
	if workflowDef.Nodes != nil {
		for i := range *workflowDef.Nodes {
			if (*workflowDef.Nodes)[i].Id == request.NodeId {
				targetNode = &(*workflowDef.Nodes)[i]
				break
			}
		}
	}

	if targetNode == nil {
		errorMsg := "not found"
		message := "Node not found in draft"
		return TestWorkflowDraftNode404JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// For now, return a simple test result
	// In a real implementation, this would execute the node with test data
	success := true
	executionTime := float32(0.1) // Mock execution time
	output := map[string]interface{}{
		"status": "test_passed",
		"nodeId": request.NodeId,
	}
	result := NodeTestResult{
		Success:       &success,
		ExecutionTime: &executionTime,
		Output:        &output,
	}

	return TestWorkflowDraftNode200JSONResponse(result), nil
}
