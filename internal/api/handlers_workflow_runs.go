package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ListWorkflowRuns retrieves workflow runs with optional filtering
func (h *OpenAPIHandlers) ListWorkflowRuns(ctx context.Context, request ListWorkflowRunsRequestObject) (ListWorkflowRunsResponseObject, error) {
	page := 1
	limit := 20

	if request.Params.Page != nil {
		page = *request.Params.Page
	}
	if request.Params.Limit != nil {
		limit = *request.Params.Limit
	}

	offset := (page - 1) * limit

	// Build query with optional filters
	whereClause := ""
	args := []interface{}{}
	argIndex := 1

	if request.Params.WorkflowId != nil {
		whereClause = "WHERE workflow_id = $" + fmt.Sprintf("%d", argIndex)
		args = append(args, request.Params.WorkflowId.String())
		argIndex++
	}

	if request.Params.Status != nil {
		if whereClause != "" {
			whereClause += " AND "
		} else {
			whereClause = "WHERE "
		}
		whereClause += "status = $" + fmt.Sprintf("%d", argIndex)
		args = append(args, string(*request.Params.Status))
		argIndex++
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM workflow_runs " + whereClause
	var total int
	err := h.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		errorMsg := "database error"
		message := err.Error()
		return ListWorkflowRuns500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	// Get workflow runs with pagination
	// Use COALESCE to handle both legacy (agent_id) and new (workflow_id) schemas
	query := "SELECT id, COALESCE(workflow_id, agent_id) as workflow_id, status, COALESCE(started_at, created_at) as started_at, completed_at, COALESCE(context, variables) as context, COALESCE(error, error_data::text) as error FROM workflow_runs " +
		whereClause + " ORDER BY COALESCE(started_at, created_at) DESC LIMIT $" + fmt.Sprintf("%d", argIndex) + " OFFSET $" + fmt.Sprintf("%d", argIndex+1)
	args = append(args, limit, offset)

	rows, err := h.db.QueryContext(ctx, query, args...)
	if err != nil {
		errorMsg := "database error"
		message := err.Error()
		return ListWorkflowRuns500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}
	defer rows.Close()

	runs := make([]WorkflowRun, 0)
	for rows.Next() {
		var run WorkflowRun
		var id, workflowID, status string
		var startedAt time.Time
		var completedAt sql.NullTime
		var contextJson sql.NullString
		var errorMsg sql.NullString

		err := rows.Scan(&id, &workflowID, &status, &startedAt, &completedAt, &contextJson, &errorMsg)
		if err != nil {
			errorMsg := "scan error"
			message := err.Error()
			return ListWorkflowRuns500JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}

		runUUID, err := uuid.Parse(id)
		if err != nil {
			errorMsg := "uuid parse error"
			message := err.Error()
			return ListWorkflowRuns500JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}

		workflowUUID, err := uuid.Parse(workflowID)
		if err != nil {
			errorMsg := "workflow uuid parse error"
			message := err.Error()
			return ListWorkflowRuns500JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}

		run.Id = &runUUID
		run.WorkflowId = &workflowUUID
		run.Status = func() *WorkflowRunStatus {
			switch status {
			case "pending":
				s := WorkflowRunStatusPending
				return &s
			case "running":
				s := WorkflowRunStatusRunning
				return &s
			case "completed":
				s := WorkflowRunStatusCompleted
				return &s
			case "failed":
				s := WorkflowRunStatusFailed
				return &s
			}
			return nil
		}()
		run.StartedAt = &startedAt

		if completedAt.Valid {
			run.CompletedAt = &completedAt.Time
		}

		// Parse context JSON if present
		if contextJson.Valid && contextJson.String != "" {
			var context map[string]interface{}
			err = json.Unmarshal([]byte(contextJson.String), &context)
			if err != nil {
				errorMsg := "context parse error"
				message := err.Error()
				return ListWorkflowRuns500JSONResponse{
					Error:   &errorMsg,
					Message: &message,
				}, nil
			}
			run.Context = &context
		}

		if errorMsg.Valid {
			run.Error = &errorMsg.String
		}

		runs = append(runs, run)
	}

	return ListWorkflowRuns200JSONResponse{
		Runs:  runs,
		Total: total,
		Page:  page,
		Limit: limit,
	}, nil
}

// GetWorkflowRun retrieves a single workflow run by ID
func (h *OpenAPIHandlers) GetWorkflowRun(ctx context.Context, request GetWorkflowRunRequestObject) (GetWorkflowRunResponseObject, error) {
	var run WorkflowRun
	var id, workflowID, status string
	var startedAt time.Time
	var completedAt sql.NullTime
	var contextJson sql.NullString
	var errorMsg sql.NullString

	err := h.db.QueryRowContext(ctx,
		"SELECT id, COALESCE(workflow_id, agent_id) as workflow_id, status, COALESCE(started_at, created_at) as started_at, completed_at, COALESCE(context, variables) as context, COALESCE(error, error_data::text) as error FROM workflow_runs WHERE id = $1",
		request.Id.String()).Scan(&id, &workflowID, &status, &startedAt, &completedAt, &contextJson, &errorMsg)
	if err != nil {
		if err == sql.ErrNoRows {
			errorMsg := "not found"
			message := "Workflow run not found"
			return GetWorkflowRun404JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}
		errorMsg := "database error"
		message := err.Error()
		return GetWorkflowRun500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	runUUID, err := uuid.Parse(id)
	if err != nil {
		errorMsg := "uuid parse error"
		message := err.Error()
		return GetWorkflowRun500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	workflowUUID, err := uuid.Parse(workflowID)
	if err != nil {
		errorMsg := "workflow uuid parse error"
		message := err.Error()
		return GetWorkflowRun500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}

	run.Id = &runUUID
	run.WorkflowId = &workflowUUID
	run.Status = func() *WorkflowRunStatus {
		switch status {
		case "pending":
			s := WorkflowRunStatusPending
			return &s
		case "running":
			s := WorkflowRunStatusRunning
			return &s
		case "completed":
			s := WorkflowRunStatusCompleted
			return &s
		case "failed":
			s := WorkflowRunStatusFailed
			return &s
		}
		return nil
	}()
	run.StartedAt = &startedAt

	if completedAt.Valid {
		run.CompletedAt = &completedAt.Time
	}

	// Parse context JSON if present
	if contextJson.Valid && contextJson.String != "" {
		var context map[string]interface{}
		err = json.Unmarshal([]byte(contextJson.String), &context)
		if err != nil {
			errorMsg := "context parse error"
			message := err.Error()
			return GetWorkflowRun500JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}
		run.Context = &context
	}

	if errorMsg.Valid {
		run.Error = &errorMsg.String
	}

	return GetWorkflowRun200JSONResponse(run), nil
}

// GetWorkflowRunSteps retrieves steps for a workflow run
func (h *OpenAPIHandlers) GetWorkflowRunSteps(ctx context.Context, request GetWorkflowRunStepsRequestObject) (GetWorkflowRunStepsResponseObject, error) {
	rows, err := h.db.QueryContext(ctx,
		"SELECT id, run_id, node_id, status, COALESCE(started_at, created_at) as started_at, completed_at, COALESCE(input_envelope, '{}') as input, COALESCE(output_envelope, '{}') as output, COALESCE(error_details::text, '') as error FROM workflow_steps WHERE run_id = $1 ORDER BY COALESCE(started_at, created_at)",
		request.Id.String())
	if err != nil {
		errorMsg := "database error"
		message := err.Error()
		return GetWorkflowRunSteps500JSONResponse{
			Error:   &errorMsg,
			Message: &message,
		}, nil
	}
	defer rows.Close()

	var steps []WorkflowStep
	for rows.Next() {
		var step WorkflowStep
		var id, runID, nodeID, status string
		var startedAt time.Time
		var completedAt sql.NullTime
		var inputJson, outputJson sql.NullString
		var errorMsg sql.NullString

		err := rows.Scan(&id, &runID, &nodeID, &status, &startedAt, &completedAt, &inputJson, &outputJson, &errorMsg)
		if err != nil {
			errorMsg := "scan error"
			message := err.Error()
			return GetWorkflowRunSteps500JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}

		stepUUID, err := uuid.Parse(id)
		if err != nil {
			errorMsg := "uuid parse error"
			message := err.Error()
			return GetWorkflowRunSteps500JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}

		runUUID, err := uuid.Parse(runID)
		if err != nil {
			errorMsg := "run uuid parse error"
			message := err.Error()
			return GetWorkflowRunSteps500JSONResponse{
				Error:   &errorMsg,
				Message: &message,
			}, nil
		}

		step.Id = &stepUUID
		step.RunId = &runUUID
		step.NodeId = &nodeID
		step.Status = func() *WorkflowStepStatus {
			switch status {
			case "pending":
				s := WorkflowStepStatusPending
				return &s
			case "running":
				s := WorkflowStepStatusRunning
				return &s
			case "completed":
				s := WorkflowStepStatusCompleted
				return &s
			case "failed":
				s := WorkflowStepStatusFailed
				return &s
			case "skipped":
				s := WorkflowStepStatusSkipped
				return &s
			}
			return nil
		}()
		step.StartedAt = &startedAt

		if completedAt.Valid {
			step.CompletedAt = &completedAt.Time
		}

		// Parse input JSON if present
		if inputJson.Valid && inputJson.String != "" {
			var input map[string]interface{}
			err = json.Unmarshal([]byte(inputJson.String), &input)
			if err != nil {
				errorMsg := "input parse error"
				message := err.Error()
				return GetWorkflowRunSteps500JSONResponse{
					Error:   &errorMsg,
					Message: &message,
				}, nil
			}
			genericInput := GenericInput(input)
			step.Input = &genericInput
		}

		// Parse output JSON if present
		if outputJson.Valid && outputJson.String != "" {
			var output map[string]interface{}
			err = json.Unmarshal([]byte(outputJson.String), &output)
			if err != nil {
				errorMsg := "output parse error"
				message := err.Error()
				return GetWorkflowRunSteps500JSONResponse{
					Error:   &errorMsg,
					Message: &message,
				}, nil
			}
			genericOutput := GenericOutput(output)
			step.Output = &genericOutput
		}

		if errorMsg.Valid {
			step.Error = &errorMsg.String
		}

		steps = append(steps, step)
	}

	return GetWorkflowRunSteps200JSONResponse(steps), nil
}
