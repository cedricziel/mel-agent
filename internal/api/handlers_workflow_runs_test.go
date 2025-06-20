package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cedricziel/mel-agent/internal/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOpenAPIListWorkflowRuns tests listing workflow runs with pagination and filtering
func TestOpenAPIListWorkflowRuns(t *testing.T) {
	db, mockEngine, cleanup := testutil.SetupOpenAPITestDB(t)
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	// Create test workflow first (needed for foreign key)
	createWorkflowReq := CreateWorkflowRequest{
		Name:        "Test Workflow for Runs",
		Description: testutil.StringPtr("A workflow for testing runs"),
	}
	reqBody, _ := json.Marshal(createWorkflowReq)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/workflows", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var createdWorkflow Workflow
	json.NewDecoder(w.Body).Decode(&createdWorkflow)

	// Create test workflow runs with both new schema (workflow_id) and legacy schema (agent_id)
	testRuns := []struct {
		id         uuid.UUID
		workflowID uuid.UUID
		status     string
		context    map[string]interface{}
	}{
		{
			id:         uuid.New(),
			workflowID: createdWorkflow.Id,
			status:     "completed",
			context:    map[string]interface{}{"env": "test", "run": 1},
		},
		{
			id:         uuid.New(),
			workflowID: createdWorkflow.Id,
			status:     "running",
			context:    map[string]interface{}{"env": "test", "run": 2},
		},
		{
			id:         uuid.New(),
			workflowID: createdWorkflow.Id,
			status:     "failed",
			context:    map[string]interface{}{"env": "test", "run": 3},
		},
	}

	// Insert test workflow runs directly into database
	for i, run := range testRuns {
		contextJson, _ := json.Marshal(run.context)

		if i == 0 {
			// Use new schema (workflow_id, context, error)
			_, err := db.Exec(`
				INSERT INTO workflow_runs (id, workflow_id, status, started_at, completed_at, context, error) 
				VALUES ($1, $2, $3, $4, $5, $6, $7)`,
				run.id, run.workflowID, run.status,
				time.Now().Add(-time.Hour),
				func() *time.Time {
					if run.status == "completed" {
						t := time.Now()
						return &t
					}
					return nil
				}(),
				contextJson,
				func() *string {
					if run.status == "failed" {
						s := "Test error"
						return &s
					}
					return nil
				}())
			require.NoError(t, err)
		} else {
			// Use legacy schema (agent_id, variables, error_data) to test compatibility
			// First create a legacy agent for the foreign key constraint
			_, err := db.Exec(`
				INSERT INTO agents (id, name, description) 
				VALUES ($1, $2, $3) ON CONFLICT (id) DO NOTHING`,
				run.workflowID, "Legacy Agent for Test", "Test agent for legacy compatibility")
			require.NoError(t, err)

			errorData := func() interface{} {
				if run.status == "failed" {
					ed, _ := json.Marshal(map[string]interface{}{"message": "Legacy test error"})
					return ed
				}
				return nil
			}()

			_, err = db.Exec(`
				INSERT INTO workflow_runs (id, agent_id, status, created_at, started_at, completed_at, variables, error_data) 
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
				run.id, run.workflowID, run.status,
				time.Now().Add(-time.Hour),
				time.Now().Add(-time.Hour),
				func() *time.Time {
					if run.status == "completed" {
						t := time.Now()
						return &t
					}
					return nil
				}(),
				contextJson, errorData)
			require.NoError(t, err)
		}
	}

	// Test 1: List all workflow runs
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/workflow-runs", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response WorkflowRunList
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.NotNil(t, response.Runs)
	assert.Len(t, *response.Runs, 3)
	assert.Equal(t, 3, *response.Total)
	assert.Equal(t, 1, *response.Page)
	assert.Equal(t, 20, *response.Limit)

	// Verify the runs have the expected data
	runStatuses := make(map[WorkflowRunStatus]bool)
	for _, run := range *response.Runs {
		runStatuses[*run.Status] = true
		assert.NotNil(t, run.Id)
		assert.Equal(t, createdWorkflow.Id, *run.WorkflowId)
		assert.NotNil(t, run.StartedAt)
		assert.NotNil(t, run.Context)
	}
	assert.True(t, runStatuses[WorkflowRunStatusCompleted])
	assert.True(t, runStatuses[WorkflowRunStatusRunning])
	assert.True(t, runStatuses[WorkflowRunStatusFailed])
}

// TestOpenAPIListWorkflowRunsWithFilters tests filtering by workflow_id and status
func TestOpenAPIListWorkflowRunsWithFilters(t *testing.T) {
	db, mockEngine, cleanup := testutil.SetupOpenAPITestDB(t)
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	// Create two test workflows
	var workflow1ID, workflow2ID uuid.UUID

	for i, name := range []string{"Test Workflow 1", "Test Workflow 2"} {
		createWorkflowReq := CreateWorkflowRequest{
			Name: name,
		}
		reqBody, _ := json.Marshal(createWorkflowReq)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/workflows", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)

		var createdWorkflow Workflow
		json.NewDecoder(w.Body).Decode(&createdWorkflow)

		if i == 0 {
			workflow1ID = createdWorkflow.Id
		} else {
			workflow2ID = createdWorkflow.Id
		}
	}

	// Create test runs for both workflows
	testRuns := []struct {
		workflowID uuid.UUID
		status     string
	}{
		{workflow1ID, "completed"},
		{workflow1ID, "running"},
		{workflow2ID, "completed"},
		{workflow2ID, "failed"},
	}

	for _, run := range testRuns {
		_, err := db.Exec(`
			INSERT INTO workflow_runs (id, workflow_id, status, started_at) 
			VALUES ($1, $2, $3, $4)`,
			uuid.New(), run.workflowID, run.status, time.Now())
		require.NoError(t, err)
	}

	// Test filter by workflow_id
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/workflow-runs?workflow_id=%s", workflow1ID), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response WorkflowRunList
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.NotNil(t, response.Runs)
	assert.Len(t, *response.Runs, 2)
	for _, run := range *response.Runs {
		assert.Equal(t, workflow1ID, *run.WorkflowId)
	}

	// Test filter by status
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/workflow-runs?status=completed", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.NotNil(t, response.Runs)
	assert.Len(t, *response.Runs, 2)
	for _, run := range *response.Runs {
		assert.Equal(t, WorkflowRunStatusCompleted, *run.Status)
	}
}

// TestOpenAPIListWorkflowRunsWithPagination tests pagination
func TestOpenAPIListWorkflowRunsWithPagination(t *testing.T) {
	db, mockEngine, cleanup := testutil.SetupOpenAPITestDB(t)
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	// Create test workflow
	createWorkflowReq := CreateWorkflowRequest{
		Name: "Pagination Test Workflow",
	}
	reqBody, _ := json.Marshal(createWorkflowReq)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/workflows", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var createdWorkflow Workflow
	json.NewDecoder(w.Body).Decode(&createdWorkflow)

	// Create 7 test runs
	for i := 0; i < 7; i++ {
		_, err := db.Exec(`
			INSERT INTO workflow_runs (id, workflow_id, status, started_at) 
			VALUES ($1, $2, $3, $4)`,
			uuid.New(), createdWorkflow.Id, "pending", time.Now().Add(-time.Duration(i)*time.Hour))
		require.NoError(t, err)
	}

	// Test page 1 with limit 3
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/workflow-runs?page=1&limit=3", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response WorkflowRunList
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.NotNil(t, response.Runs)
	assert.Len(t, *response.Runs, 3)
	assert.Equal(t, 7, *response.Total)
	assert.Equal(t, 1, *response.Page)
	assert.Equal(t, 3, *response.Limit)

	// Test page 2 with limit 3
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/workflow-runs?page=2&limit=3", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.NotNil(t, response.Runs)
	assert.Len(t, *response.Runs, 3)
	assert.Equal(t, 7, *response.Total)
	assert.Equal(t, 2, *response.Page)
	assert.Equal(t, 3, *response.Limit)
}

// TestOpenAPIGetWorkflowRun tests retrieving a single workflow run
func TestOpenAPIGetWorkflowRun(t *testing.T) {
	db, mockEngine, cleanup := testutil.SetupOpenAPITestDB(t)
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	// Create test workflow
	createWorkflowReq := CreateWorkflowRequest{
		Name:        "Get Test Workflow",
		Description: testutil.StringPtr("A workflow for get testing"),
	}
	reqBody, _ := json.Marshal(createWorkflowReq)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/workflows", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var createdWorkflow Workflow
	json.NewDecoder(w.Body).Decode(&createdWorkflow)

	// Create a test workflow run
	runID := uuid.New()
	startedAt := time.Now().Add(-time.Hour)
	completedAt := time.Now()
	contextData := map[string]interface{}{
		"user_id": "test-user",
		"env":     "testing",
	}
	contextJson, _ := json.Marshal(contextData)

	_, err := db.Exec(`
		INSERT INTO workflow_runs (id, workflow_id, status, started_at, completed_at, context) 
		VALUES ($1, $2, $3, $4, $5, $6)`,
		runID, createdWorkflow.Id, "completed", startedAt, completedAt, contextJson)
	require.NoError(t, err)

	// Get the workflow run
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/workflow-runs/%s", runID), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response WorkflowRun
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, runID, *response.Id)
	assert.Equal(t, createdWorkflow.Id, *response.WorkflowId)
	assert.Equal(t, WorkflowRunStatusCompleted, *response.Status)
	assert.NotNil(t, response.StartedAt)
	assert.NotNil(t, response.CompletedAt)
	assert.NotNil(t, response.Context)
	assert.Equal(t, "test-user", (*response.Context)["user_id"])
	assert.Equal(t, "testing", (*response.Context)["env"])
}

// TestOpenAPIGetWorkflowRunNotFound tests retrieving a non-existent workflow run
func TestOpenAPIGetWorkflowRunNotFound(t *testing.T) {
	db, mockEngine, cleanup := testutil.SetupOpenAPITestDB(t)
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	nonExistentID := uuid.New()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/workflow-runs/%s", nonExistentID), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response Error
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "not found", *response.Error)
}

// TestOpenAPIGetWorkflowRunWithError tests retrieving a workflow run that has an error
func TestOpenAPIGetWorkflowRunWithError(t *testing.T) {
	db, mockEngine, cleanup := testutil.SetupOpenAPITestDB(t)
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	// Create test workflow
	createWorkflowReq := CreateWorkflowRequest{
		Name: "Error Test Workflow",
	}
	reqBody, _ := json.Marshal(createWorkflowReq)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/workflows", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var createdWorkflow Workflow
	json.NewDecoder(w.Body).Decode(&createdWorkflow)

	// Create a failed workflow run with error
	runID := uuid.New()
	errorMessage := "Test execution failed: network timeout"

	_, err := db.Exec(`
		INSERT INTO workflow_runs (id, workflow_id, status, started_at, error) 
		VALUES ($1, $2, $3, $4, $5)`,
		runID, createdWorkflow.Id, "failed", time.Now().Add(-time.Hour), errorMessage)
	require.NoError(t, err)

	// Get the failed workflow run
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/workflow-runs/%s", runID), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response WorkflowRun
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, runID, *response.Id)
	assert.Equal(t, WorkflowRunStatusFailed, *response.Status)
	assert.NotNil(t, response.Error)
	assert.Equal(t, errorMessage, *response.Error)
}

// TestOpenAPIGetWorkflowRunSteps tests retrieving steps for a workflow run
func TestOpenAPIGetWorkflowRunSteps(t *testing.T) {
	db, mockEngine, cleanup := testutil.SetupOpenAPITestDB(t)
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	// Create test workflow
	createWorkflowReq := CreateWorkflowRequest{
		Name: "Steps Test Workflow",
	}
	reqBody, _ := json.Marshal(createWorkflowReq)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/workflows", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var createdWorkflow Workflow
	json.NewDecoder(w.Body).Decode(&createdWorkflow)

	// Create a test workflow run
	runID := uuid.New()
	_, err := db.Exec(`
		INSERT INTO workflow_runs (id, workflow_id, status, started_at) 
		VALUES ($1, $2, $3, $4)`,
		runID, createdWorkflow.Id, "running", time.Now().Add(-time.Hour))
	require.NoError(t, err)

	// Create test workflow steps
	testSteps := []struct {
		nodeID string
		status string
		input  map[string]interface{}
		output map[string]interface{}
	}{
		{
			nodeID: "trigger-node",
			status: "completed",
			input:  map[string]interface{}{"webhook_data": "test"},
			output: map[string]interface{}{"processed": true},
		},
		{
			nodeID: "action-node",
			status: "running",
			input:  map[string]interface{}{"data": "from trigger"},
			output: nil,
		},
		{
			nodeID: "failed-node",
			status: "failed",
			input:  map[string]interface{}{"retry": true},
			output: nil,
		},
	}

	for i, step := range testSteps {
		stepID := uuid.New()
		inputJson, _ := json.Marshal(step.input)

		var outputJson interface{}
		if step.output != nil {
			outputJson, _ = json.Marshal(step.output)
		}

		var errorDetails interface{}
		if step.status == "failed" {
			errorDetails, _ = json.Marshal(map[string]interface{}{
				"message": "Step execution failed",
				"code":    "EXECUTION_ERROR",
			})
		}

		_, err := db.Exec(`
			INSERT INTO workflow_steps (id, run_id, node_id, node_type, step_number, status, created_at, started_at, completed_at, input_envelope, output_envelope, error_details) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
			stepID, runID, step.nodeID, "action", i+1, step.status,
			time.Now().Add(-time.Hour),
			time.Now().Add(-time.Hour),
			func() *time.Time {
				if step.status == "completed" {
					t := time.Now()
					return &t
				}
				return nil
			}(),
			inputJson,
			outputJson,
			errorDetails)
		require.NoError(t, err)
	}

	// Get the workflow run steps
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/workflow-runs/%s/steps", runID), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []WorkflowStep
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Len(t, response, 3)

	// Verify steps are returned in order
	stepsByNodeID := make(map[string]WorkflowStep)
	for _, step := range response {
		stepsByNodeID[*step.NodeId] = step
	}

	// Check trigger step
	triggerStep := stepsByNodeID["trigger-node"]
	assert.Equal(t, runID, *triggerStep.RunId)
	assert.Equal(t, WorkflowStepStatusCompleted, *triggerStep.Status)
	assert.NotNil(t, triggerStep.Input)
	assert.Equal(t, "test", (*triggerStep.Input)["webhook_data"])
	assert.NotNil(t, triggerStep.Output)
	assert.Equal(t, true, (*triggerStep.Output)["processed"])
	assert.NotNil(t, triggerStep.CompletedAt)

	// Check running step
	actionStep := stepsByNodeID["action-node"]
	assert.Equal(t, WorkflowStepStatusRunning, *actionStep.Status)
	assert.NotNil(t, actionStep.Input)
	assert.Equal(t, "from trigger", (*actionStep.Input)["data"])

	// Check failed step
	failedStep := stepsByNodeID["failed-node"]
	assert.Equal(t, WorkflowStepStatusFailed, *failedStep.Status)
	assert.NotNil(t, failedStep.Error)
}

// TestOpenAPIGetWorkflowRunStepsNotFound tests retrieving steps for a non-existent workflow run
func TestOpenAPIGetWorkflowRunStepsNotFound(t *testing.T) {
	db, mockEngine, cleanup := testutil.SetupOpenAPITestDB(t)
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	nonExistentID := uuid.New()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/workflow-runs/%s/steps", nonExistentID), nil)
	router.ServeHTTP(w, req)

	// This should return 200 with empty array, not 404, since steps endpoint doesn't check if run exists
	assert.Equal(t, http.StatusOK, w.Code)

	var response []WorkflowStep
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Len(t, response, 0)
}

// TestOpenAPIWorkflowRunLegacyCompatibility tests that handlers work with legacy schema
func TestOpenAPIWorkflowRunLegacyCompatibility(t *testing.T) {
	db, mockEngine, cleanup := testutil.SetupOpenAPITestDB(t)
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	// Create a workflow run using the legacy schema (agent_id, variables, error_data)
	runID := uuid.New()
	agentID := uuid.New() // This simulates the legacy agent_id

	// Create an agent first for the foreign key constraint
	_, err := db.Exec(`
		INSERT INTO agents (id, name, description) 
		VALUES ($1, $2, $3)`,
		agentID, "Legacy Agent", "Test agent for legacy compatibility")
	require.NoError(t, err)

	variables := map[string]interface{}{
		"legacy_var": "test_value",
		"env":        "legacy",
	}
	variablesJson, _ := json.Marshal(variables)

	errorData := map[string]interface{}{
		"message": "Legacy error format",
		"code":    "LEGACY_ERROR",
	}
	errorDataJson, _ := json.Marshal(errorData)

	_, err = db.Exec(`
		INSERT INTO workflow_runs (id, agent_id, status, created_at, variables, error_data) 
		VALUES ($1, $2, $3, $4, $5, $6)`,
		runID, agentID, "failed", time.Now().Add(-time.Hour), variablesJson, errorDataJson)
	require.NoError(t, err)

	// Test that the OpenAPI handler can read the legacy format
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/workflow-runs/%s", runID), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response WorkflowRun
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, runID, *response.Id)
	assert.Equal(t, agentID, *response.WorkflowId) // agent_id is mapped to workflow_id
	assert.Equal(t, WorkflowRunStatusFailed, *response.Status)

	// Check that variables are mapped to context
	assert.NotNil(t, response.Context)
	assert.Equal(t, "test_value", (*response.Context)["legacy_var"])
	assert.Equal(t, "legacy", (*response.Context)["env"])

	// Check that error_data is mapped to error
	assert.NotNil(t, response.Error)
	assert.Contains(t, *response.Error, "Legacy error format")

	// Test listing also works with legacy format
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/workflow-runs", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var listResponse WorkflowRunList
	err = json.NewDecoder(w.Body).Decode(&listResponse)
	require.NoError(t, err)

	assert.NotNil(t, listResponse.Runs)
	assert.Len(t, *listResponse.Runs, 1)

	legacyRun := (*listResponse.Runs)[0]
	assert.Equal(t, runID, *legacyRun.Id)
	assert.Equal(t, agentID, *legacyRun.WorkflowId)
}
