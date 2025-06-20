package api

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ListAgents retrieves all agents with pagination
func (h *OpenAPIHandlers) ListAgents(ctx context.Context, request ListAgentsRequestObject) (ListAgentsResponseObject, error) {
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
	err := h.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM agents").Scan(&total)
	if err != nil {
		errorMsg := "database error"
		message := err.Error()
		return ListAgents500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Get agents with pagination
	rows, err := h.db.QueryContext(ctx,
		"SELECT id, name, description, created_at FROM agents ORDER BY created_at DESC LIMIT $1 OFFSET $2",
		limit, offset)
	if err != nil {
		errorMsg := "database error"
		message := err.Error()
		return ListAgents500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}
	defer rows.Close()

	var agents []Agent
	for rows.Next() {
		var agent Agent
		var description sql.NullString
		var id, name string
		var createdAt time.Time

		err := rows.Scan(&id, &name, &description, &createdAt)
		if err != nil {
			errorMsg := "scan error"
			message := err.Error()
			return ListAgents500JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}

		agentUUID, err := uuid.Parse(id)
		if err != nil {
			errorMsg := "uuid parse error"
			message := err.Error()
			return ListAgents500JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}

		agent.Id = agentUUID
		agent.Name = name
		if description.Valid {
			agent.Description = &description.String
		}
		agent.CreatedAt = createdAt

		agents = append(agents, agent)
	}

	return ListAgents200JSONResponse{
		Agents: &agents,
		Total:  &total,
		Page:   &page,
		Limit:  &limit,
	}, nil
}

// CreateAgent creates a new agent
func (h *OpenAPIHandlers) CreateAgent(ctx context.Context, request CreateAgentRequestObject) (CreateAgentResponseObject, error) {
	agentID := uuid.New()
	now := time.Now()

	var description *string
	if request.Body.Description != nil {
		description = request.Body.Description
	}

	// For now, use a default user_id (in real implementation, this would come from auth context)
	defaultUserID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	// Insert agent into database
	_, err := h.db.ExecContext(ctx,
		"INSERT INTO agents (id, user_id, name, description, created_at) VALUES ($1, $2, $3, $4, $5)",
		agentID, defaultUserID, request.Body.Name, description, now)
	if err != nil {
		errorMsg := "failed to create agent"
		message := err.Error()
		return CreateAgent500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	agent := Agent{
		Id:          agentID,
		Name:        request.Body.Name,
		Description: description,
		CreatedAt:   now,
	}

	// Handle workflow definition if provided
	if request.Body.Definition != nil {
		agent.Definition = request.Body.Definition
	}

	return CreateAgent201JSONResponse(agent), nil
}

// GetAgent retrieves a single agent by ID
func (h *OpenAPIHandlers) GetAgent(ctx context.Context, request GetAgentRequestObject) (GetAgentResponseObject, error) {
	var agent Agent
	var description sql.NullString
	var id, name string
	var createdAt time.Time

	err := h.db.QueryRowContext(ctx,
		"SELECT id, name, description, created_at FROM agents WHERE id = $1",
		request.Id.String()).Scan(&id, &name, &description, &createdAt)
	if err != nil {
		if err == sql.ErrNoRows {
			errorMsg := "not found"
			message := "Agent not found"
			return GetAgent404JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}
		errorMsg := "database error"
		message := err.Error()
		return GetAgent500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	agentUUID, err := uuid.Parse(id)
	if err != nil {
		errorMsg := "uuid parse error"
		message := err.Error()
		return GetAgent500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	agent.Id = agentUUID
	agent.Name = name
	if description.Valid {
		agent.Description = &description.String
	}
	agent.CreatedAt = createdAt

	return GetAgent200JSONResponse(agent), nil
}

// UpdateAgent updates an existing agent
func (h *OpenAPIHandlers) UpdateAgent(ctx context.Context, request UpdateAgentRequestObject) (UpdateAgentResponseObject, error) {
	// Start building the update query
	var setParts []string
	var args []interface{}
	argIndex := 1

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

	// Check if we have anything to update
	if len(setParts) == 0 {
		// No fields to update, just return the existing agent
		getRequest := GetAgentRequestObject{Id: request.Id}
		getResponse, err := h.GetAgent(ctx, getRequest)
		if err != nil {
			errorMsg := "failed to fetch agent"
			message := err.Error()
			return UpdateAgent500JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}

		// Convert GetAgentResponseObject to UpdateAgentResponseObject
		if agentResponse, ok := getResponse.(GetAgent200JSONResponse); ok {
			return UpdateAgent200JSONResponse(agentResponse), nil
		}

		errorMsg := "unexpected response type"
		message := "Failed to convert response"
		return UpdateAgent500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Add the ID as the last parameter
	args = append(args, request.Id.String())

	query := fmt.Sprintf("UPDATE agents SET %s WHERE id = $%d",
		strings.Join(setParts, ", "), argIndex)

	result, err := h.db.ExecContext(ctx, query, args...)
	if err != nil {
		errorMsg := "failed to update agent"
		message := err.Error()
		return UpdateAgent500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		errorMsg := "failed to check update result"
		message := err.Error()
		return UpdateAgent500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	if rowsAffected == 0 {
		errorMsg := "not found"
		message := "Agent not found"
		return UpdateAgent404JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Fetch the updated agent
	getRequest := GetAgentRequestObject{Id: request.Id}
	getResponse, err := h.GetAgent(ctx, getRequest)
	if err != nil {
		errorMsg := "failed to fetch updated agent"
		message := err.Error()
		return UpdateAgent500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Convert GetAgentResponseObject to UpdateAgentResponseObject
	if agentResponse, ok := getResponse.(GetAgent200JSONResponse); ok {
		return UpdateAgent200JSONResponse(agentResponse), nil
	}

	// If we get here, something went wrong
	errorMsg := "unexpected response type"
	message := "Failed to convert response"
	return UpdateAgent500JSONResponse{
		Error:   &errorMsg,
		Message: &message,
	}, nil
}

// DeleteAgent removes an agent
func (h *OpenAPIHandlers) DeleteAgent(ctx context.Context, request DeleteAgentRequestObject) (DeleteAgentResponseObject, error) {
	result, err := h.db.ExecContext(ctx, "DELETE FROM agents WHERE id = $1", request.Id.String())
	if err != nil {
		errorMsg := "failed to delete agent"
		message := err.Error()
		return DeleteAgent500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		errorMsg := "failed to check delete result"
		message := err.Error()
		return DeleteAgent500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	if rowsAffected == 0 {
		errorMsg := "not found"
		message := "Agent not found"
		return DeleteAgent404JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	return DeleteAgent204Response{}, nil
}
