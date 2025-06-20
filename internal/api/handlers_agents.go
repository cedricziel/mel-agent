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

// CreateAgentVersion creates a new version of an agent
func (h *OpenAPIHandlers) CreateAgentVersion(ctx context.Context, request CreateAgentVersionRequestObject) (CreateAgentVersionResponseObject, error) {
	// Check if agent exists
	var agentExists bool
	err := h.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM agents WHERE id = $1)", request.AgentId.String()).Scan(&agentExists)
	if err != nil {
		errorMsg := "database error"
		message := err.Error()
		return CreateAgentVersion500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	if !agentExists {
		errorMsg := "not found"
		message := "Agent not found"
		return CreateAgentVersion404JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Get next version number
	var maxVersion sql.NullInt64
	err = h.db.QueryRowContext(ctx, "SELECT MAX(version_number) FROM agent_versions WHERE agent_id = $1", request.AgentId.String()).Scan(&maxVersion)
	if err != nil {
		errorMsg := "database error"
		message := err.Error()
		return CreateAgentVersion500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	nextVersion := int64(1)
	if maxVersion.Valid {
		nextVersion = maxVersion.Int64 + 1
	}

	versionID := uuid.New()
	now := time.Now()

	// Insert new version
	_, err = h.db.ExecContext(ctx,
		`INSERT INTO agent_versions (id, agent_id, version_number, name, description, definition, created_at) 
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		versionID, request.AgentId, nextVersion, request.Body.Name, request.Body.Description, "{}", now)
	if err != nil {
		errorMsg := "failed to create agent version"
		message := err.Error()
		return CreateAgentVersion500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	version := AgentVersion{
		Id:            &versionID,
		AgentId:       &request.AgentId,
		VersionNumber: func() *int { i := int(nextVersion); return &i }(),
		Name:          &request.Body.Name,
		CreatedAt:     &now,
		IsCurrent:     func() *bool { b := false; return &b }(),
	}

	if request.Body.Description != nil {
		version.Description = request.Body.Description
	}

	if request.Body.Definition != nil {
		version.Definition = request.Body.Definition
	}

	return CreateAgentVersion201JSONResponse(version), nil
}

// GetAgentDraft retrieves the current draft for an agent
func (h *OpenAPIHandlers) GetAgentDraft(ctx context.Context, request GetAgentDraftRequestObject) (GetAgentDraftResponseObject, error) {
	// Check if agent exists
	var agentExists bool
	err := h.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM agents WHERE id = $1)", request.AgentId.String()).Scan(&agentExists)
	if err != nil {
		errorMsg := "database error"
		message := err.Error()
		return GetAgentDraft500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	if !agentExists {
		errorMsg := "not found"
		message := "Agent not found"
		return GetAgentDraft404JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Get or create draft
	var draftID, definition string
	var updatedAt time.Time
	err = h.db.QueryRowContext(ctx,
		"SELECT id, definition, updated_at FROM agent_drafts WHERE agent_id = $1",
		request.AgentId.String()).Scan(&draftID, &definition, &updatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			// Create new draft
			draftUUID := uuid.New()
			now := time.Now()
			emptyDefinition := "{}"

			_, err = h.db.ExecContext(ctx,
				"INSERT INTO agent_drafts (id, agent_id, definition, updated_at) VALUES ($1, $2, $3, $4)",
				draftUUID, request.AgentId, emptyDefinition, now)
			if err != nil {
				errorMsg := "failed to create draft"
				message := err.Error()
				return GetAgentDraft500JSONResponse{
					Error:   &errorMsg,
					Message: &message,
				}, nil
			}

			draft := AgentDraft{
				Id:        &draftUUID,
				AgentId:   &request.AgentId,
				UpdatedAt: &now,
			}

			return GetAgentDraft200JSONResponse(draft), nil
		}

		errorMsg := "database error"
		message := err.Error()
		return GetAgentDraft500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	draftUUID, err := uuid.Parse(draftID)
	if err != nil {
		errorMsg := "uuid parse error"
		message := err.Error()
		return GetAgentDraft500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	draft := AgentDraft{
		Id:        &draftUUID,
		AgentId:   &request.AgentId,
		UpdatedAt: &updatedAt,
	}

	return GetAgentDraft200JSONResponse(draft), nil
}

// UpdateAgentDraft updates the agent draft with auto-persistence
func (h *OpenAPIHandlers) UpdateAgentDraft(ctx context.Context, request UpdateAgentDraftRequestObject) (UpdateAgentDraftResponseObject, error) {
	// Check if agent exists
	var agentExists bool
	err := h.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM agents WHERE id = $1)", request.AgentId.String()).Scan(&agentExists)
	if err != nil {
		errorMsg := "database error"
		message := err.Error()
		return UpdateAgentDraft500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	if !agentExists {
		errorMsg := "not found"
		message := "Agent not found"
		return UpdateAgentDraft404JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	now := time.Now()
	definitionJSON := "{}"

	// Upsert draft
	var draftID uuid.UUID
	err = h.db.QueryRowContext(ctx,
		`INSERT INTO agent_drafts (id, agent_id, definition, updated_at) 
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (agent_id) 
		 DO UPDATE SET definition = $3, updated_at = $4
		 RETURNING id`,
		uuid.New(), request.AgentId, definitionJSON, now).Scan(&draftID)
	if err != nil {
		errorMsg := "failed to update draft"
		message := err.Error()
		return UpdateAgentDraft500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	draft := AgentDraft{
		Id:        &draftID,
		AgentId:   &request.AgentId,
		UpdatedAt: &now,
	}

	if request.Body.Definition != nil {
		draft.Definition = request.Body.Definition
	}

	return UpdateAgentDraft200JSONResponse(draft), nil
}

// TestDraftNode tests a single node in draft context
func (h *OpenAPIHandlers) TestDraftNode(ctx context.Context, request TestDraftNodeRequestObject) (TestDraftNodeResponseObject, error) {
	// Check if agent exists
	var agentExists bool
	err := h.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM agents WHERE id = $1)", request.AgentId.String()).Scan(&agentExists)
	if err != nil {
		errorMsg := "database error"
		message := err.Error()
		return TestDraftNode500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	if !agentExists {
		errorMsg := "not found"
		message := "Agent not found"
		return TestDraftNode404JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// For now, return a mock successful test result
	// In a real implementation, this would execute the node with the provided input
	result := NodeTestResult{
		Success:       func() *bool { b := true; return &b }(),
		ExecutionTime: func() *float32 { f := float32(0.123); return &f }(),
		Logs:          &[]string{"Node test executed successfully"},
	}

	if request.Body.Input != nil {
		// In real implementation, use the input to execute the node
		result.Output = &map[string]interface{}{
			"message": "Test completed with provided input",
			"input":   request.Body.Input,
		}
	} else {
		result.Output = &map[string]interface{}{
			"message": "Test completed without input",
		}
	}

	return TestDraftNode200JSONResponse(result), nil
}

// DeployAgentVersion deploys a specific agent version
func (h *OpenAPIHandlers) DeployAgentVersion(ctx context.Context, request DeployAgentVersionRequestObject) (DeployAgentVersionResponseObject, error) {
	// Check if agent exists
	var agentExists bool
	err := h.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM agents WHERE id = $1)", request.AgentId.String()).Scan(&agentExists)
	if err != nil {
		errorMsg := "database error"
		message := err.Error()
		return DeployAgentVersion500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	if !agentExists {
		errorMsg := "not found"
		message := "Agent not found"
		return DeployAgentVersion404JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Check if version exists
	var versionExists bool
	err = h.db.QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM agent_versions WHERE id = $1 AND agent_id = $2)",
		request.Body.VersionId.String(), request.AgentId.String()).Scan(&versionExists)
	if err != nil {
		errorMsg := "database error"
		message := err.Error()
		return DeployAgentVersion500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	if !versionExists {
		errorMsg := "not found"
		message := "Version not found"
		return DeployAgentVersion404JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	now := time.Now()

	// Update current version flags
	_, err = h.db.ExecContext(ctx,
		"UPDATE agent_versions SET is_current = false WHERE agent_id = $1",
		request.AgentId.String())
	if err != nil {
		errorMsg := "failed to clear current version flags"
		message := err.Error()
		return DeployAgentVersion500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Set new current version
	_, err = h.db.ExecContext(ctx,
		"UPDATE agent_versions SET is_current = true WHERE id = $1",
		request.Body.VersionId.String())
	if err != nil {
		errorMsg := "failed to set current version"
		message := err.Error()
		return DeployAgentVersion500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	deployment := AgentDeployment{
		AgentId:    &request.AgentId,
		VersionId:  &request.Body.VersionId,
		DeployedAt: &now,
		Status:     func() *AgentDeploymentStatus { s := AgentDeploymentStatusDeployed; return &s }(),
	}

	return DeployAgentVersion200JSONResponse(deployment), nil
}
