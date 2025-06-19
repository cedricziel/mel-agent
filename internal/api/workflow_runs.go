package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/cedricziel/mel-agent/pkg/execution"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// WorkflowRunsHandler handles workflow execution API endpoints
type WorkflowRunsHandler struct {
	db     *sql.DB
	engine execution.ExecutionEngine
}

// NewWorkflowRunsHandler creates a new workflow runs handler
func NewWorkflowRunsHandler(db *sql.DB, engine execution.ExecutionEngine) *WorkflowRunsHandler {
	return &WorkflowRunsHandler{
		db:     db,
		engine: engine,
	}
}

// listWorkflowRuns returns a paginated list of workflow runs for UI
func (h *WorkflowRunsHandler) listWorkflowRuns(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	agentID := r.URL.Query().Get("agent_id")
	status := r.URL.Query().Get("status")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50 // Default limit
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Build query with filters
	baseQuery := `
		SELECT id, agent_id, status, created_at, started_at, completed_at,
		       total_steps, completed_steps, failed_steps,
		       CASE 
		           WHEN completed_at IS NOT NULL AND started_at IS NOT NULL THEN
		               EXTRACT(EPOCH FROM (completed_at - started_at))
		           WHEN started_at IS NOT NULL THEN
		               EXTRACT(EPOCH FROM (NOW() - started_at))
		           ELSE NULL
		       END as duration_seconds,
		       CASE 
		           WHEN total_steps > 0 THEN 
		               ROUND((completed_steps::decimal / total_steps::decimal) * 100, 1)
		           ELSE 0 
		       END as progress_percentage
		FROM workflow_runs`

	var conditions []string
	var args []interface{}
	argIndex := 1

	if agentID != "" {
		if agentUUID, err := uuid.Parse(agentID); err == nil {
			conditions = append(conditions, fmt.Sprintf("agent_id = $%d", argIndex))
			args = append(args, agentUUID)
			argIndex++
		}
	}

	if status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, status)
		argIndex++
	}

	query := baseQuery
	if len(conditions) > 0 {
		query += " WHERE " + joinStrings(conditions, " AND ")
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	// Execute query
	rows, err := h.db.QueryContext(r.Context(), query, args...)
	if err != nil {
		http.Error(w, fmt.Sprintf("Database query failed: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var runs []WorkflowRunSummary
	for rows.Next() {
		var run WorkflowRunSummary
		var durationSeconds *float64
		var progressPercentage *float64

		err := rows.Scan(
			&run.ID, &run.AgentID, &run.Status, &run.CreatedAt,
			&run.StartedAt, &run.CompletedAt, &run.TotalSteps,
			&run.CompletedSteps, &run.FailedSteps,
			&durationSeconds, &progressPercentage,
		)
		if err != nil {
			continue
		}

		if durationSeconds != nil {
			run.DurationSeconds = *durationSeconds
		}
		if progressPercentage != nil {
			run.ProgressPercentage = *progressPercentage
		}

		runs = append(runs, run)
	}

	// Get total count for pagination
	countQuery := "SELECT COUNT(*) FROM workflow_runs"
	if len(conditions) > 0 {
		countQuery += " WHERE " + joinStrings(conditions, " AND ")
	}

	var totalCount int
	if err := h.db.QueryRowContext(r.Context(), countQuery, args[:len(args)-2]...).Scan(&totalCount); err != nil {
		totalCount = 0
	}

	response := PaginatedResponse{
		Data: runs,
		Pagination: PaginationInfo{
			Limit:   limit,
			Offset:  offset,
			Total:   totalCount,
			HasMore: offset+len(runs) < totalCount,
		},
	}

	writeJSON(w, http.StatusOK, response)
}

// getWorkflowRun returns detailed information about a specific workflow run
func (h *WorkflowRunsHandler) getWorkflowRun(w http.ResponseWriter, r *http.Request) {
	runIDStr := chi.URLParam(r, "runID")
	runID, err := uuid.Parse(runIDStr)
	if err != nil {
		http.Error(w, "Invalid run ID", http.StatusBadRequest)
		return
	}

	// Use the database function for detailed run information
	query := `SELECT * FROM get_workflow_run_details($1)`
	row := h.db.QueryRowContext(r.Context(), query, runID)

	var result WorkflowRunDetails
	var durationSeconds *float64
	var progressPercentage *float64
	var stepsJSON, checkpointsJSON []byte

	err = row.Scan(
		&result.ID, &result.AgentID, &result.AgentName, &result.Status,
		&result.CreatedAt, &result.StartedAt, &result.CompletedAt,
		&durationSeconds, &progressPercentage,
		&result.TotalSteps, &result.CompletedSteps, &result.FailedSteps,
		&result.InputData, &result.OutputData, &result.ErrorData,
		&stepsJSON, &checkpointsJSON,
	)

	if err == sql.ErrNoRows {
		http.Error(w, "Workflow run not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		return
	}

	if durationSeconds != nil {
		result.DurationSeconds = *durationSeconds
	}
	if progressPercentage != nil {
		result.ProgressPercentage = *progressPercentage
	}

	// Parse steps and checkpoints JSON
	if stepsJSON != nil {
		json.Unmarshal(stepsJSON, &result.Steps)
	}
	if checkpointsJSON != nil {
		json.Unmarshal(checkpointsJSON, &result.Checkpoints)
	}

	writeJSON(w, http.StatusOK, result)
}

// createWorkflowRun starts a new workflow execution
func (h *WorkflowRunsHandler) createWorkflowRun(w http.ResponseWriter, r *http.Request) {
	var req CreateWorkflowRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.AgentID == uuid.Nil {
		http.Error(w, "agent_id is required", http.StatusBadRequest)
		return
	}

	// Create workflow run
	run := &execution.WorkflowRun{
		ID:             uuid.New(),
		AgentID:        req.AgentID,
		VersionID:      req.VersionID,
		TriggerID:      req.TriggerID,
		Status:         execution.RunStatusPending,
		InputData:      req.InputData,
		Variables:      req.Variables,
		TimeoutSeconds: req.TimeoutSeconds,
		RetryPolicy:    req.RetryPolicy,
	}

	// Set defaults
	if run.TimeoutSeconds == 0 {
		run.TimeoutSeconds = 3600 // 1 hour default
	}
	if run.Variables == nil {
		run.Variables = make(map[string]any)
	}

	// Start the workflow execution
	if err := h.engine.StartRun(r.Context(), run); err != nil {
		http.Error(w, fmt.Sprintf("Failed to start workflow: %v", err), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, run)
}

// controlWorkflowRun handles pause/resume/cancel operations
func (h *WorkflowRunsHandler) controlWorkflowRun(w http.ResponseWriter, r *http.Request) {
	runIDStr := chi.URLParam(r, "runID")
	runID, err := uuid.Parse(runIDStr)
	if err != nil {
		http.Error(w, "Invalid run ID", http.StatusBadRequest)
		return
	}

	var req ControlWorkflowRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	switch req.Action {
	case "pause":
		err = h.engine.PauseRun(r.Context(), runID)
	case "resume":
		err = h.engine.ResumeRun(r.Context(), runID)
	case "cancel":
		err = h.engine.CancelRun(r.Context(), runID)
	default:
		http.Error(w, "Invalid action. Must be 'pause', 'resume', or 'cancel'", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to %s workflow: %v", req.Action, err), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "success", "action": req.Action})
}

// retryWorkflowStep retries a failed step
func (h *WorkflowRunsHandler) retryWorkflowStep(w http.ResponseWriter, r *http.Request) {
	stepIDStr := chi.URLParam(r, "stepID")
	stepID, err := uuid.Parse(stepIDStr)
	if err != nil {
		http.Error(w, "Invalid step ID", http.StatusBadRequest)
		return
	}

	if err := h.engine.RetryStep(r.Context(), stepID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to retry step: %v", err), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "success", "action": "retry"})
}

// Data structures for API responses

type WorkflowRunSummary struct {
	ID                 uuid.UUID  `json:"id"`
	AgentID            uuid.UUID  `json:"agent_id"`
	Status             string     `json:"status"`
	CreatedAt          time.Time  `json:"created_at"`
	StartedAt          *time.Time `json:"started_at,omitempty"`
	CompletedAt        *time.Time `json:"completed_at,omitempty"`
	TotalSteps         int        `json:"total_steps"`
	CompletedSteps     int        `json:"completed_steps"`
	FailedSteps        int        `json:"failed_steps"`
	DurationSeconds    float64    `json:"duration_seconds"`
	ProgressPercentage float64    `json:"progress_percentage"`
}

type WorkflowRunDetails struct {
	WorkflowRunSummary
	AgentName   *string               `json:"agent_name,omitempty"`
	InputData   map[string]any        `json:"input_data,omitempty"`
	OutputData  map[string]any        `json:"output_data,omitempty"`
	ErrorData   map[string]any        `json:"error_data,omitempty"`
	Steps       []WorkflowStepDetails `json:"steps,omitempty"`
	Checkpoints []WorkflowCheckpoint  `json:"checkpoints,omitempty"`
}

type WorkflowStepDetails struct {
	ID           uuid.UUID      `json:"id"`
	NodeID       string         `json:"node_id"`
	NodeType     string         `json:"node_type"`
	Status       string         `json:"status"`
	StepNumber   int            `json:"step_number"`
	AttemptCount int            `json:"attempt_count"`
	MaxAttempts  int            `json:"max_attempts"`
	CreatedAt    time.Time      `json:"created_at"`
	StartedAt    *time.Time     `json:"started_at,omitempty"`
	CompletedAt  *time.Time     `json:"completed_at,omitempty"`
	NextRetryAt  *time.Time     `json:"next_retry_at,omitempty"`
	NodeConfig   map[string]any `json:"node_config,omitempty"`
	ErrorDetails map[string]any `json:"error_details,omitempty"`
	DurationMs   *float64       `json:"duration_ms,omitempty"`
}

type WorkflowCheckpoint struct {
	ID               uuid.UUID      `json:"id"`
	StepID           uuid.UUID      `json:"step_id"`
	CheckpointType   string         `json:"checkpoint_type"`
	CreatedAt        time.Time      `json:"created_at"`
	ExecutionContext map[string]any `json:"execution_context,omitempty"`
}

type CreateWorkflowRunRequest struct {
	AgentID        uuid.UUID             `json:"agent_id"`
	VersionID      uuid.UUID             `json:"version_id,omitempty"`
	TriggerID      *uuid.UUID            `json:"trigger_id,omitempty"`
	InputData      map[string]any        `json:"input_data,omitempty"`
	Variables      map[string]any        `json:"variables,omitempty"`
	TimeoutSeconds int                   `json:"timeout_seconds,omitempty"`
	RetryPolicy    execution.RetryPolicy `json:"retry_policy,omitempty"`
}

type ControlWorkflowRunRequest struct {
	Action string `json:"action"` // "pause", "resume", "cancel"
}

type PaginatedResponse struct {
	Data       any            `json:"data"`
	Pagination PaginationInfo `json:"pagination"`
}

type PaginationInfo struct {
	Limit   int  `json:"limit"`
	Offset  int  `json:"offset"`
	Total   int  `json:"total"`
	HasMore bool `json:"has_more"`
}

// Helper function to join strings
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}

	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
