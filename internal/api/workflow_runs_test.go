package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cedricziel/mel-agent/pkg/api"
	"github.com/cedricziel/mel-agent/pkg/execution"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test database setup for API tests
var apiTestDB *sql.DB

func setupAPITestDB(t *testing.T) *sql.DB {
	if apiTestDB != nil {
		return apiTestDB
	}

	// Use a test database connection
	dbURL := "postgres://postgres:postgres@localhost:5432/agentsaas_test?sslmode=disable"
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Skipf("Skipping API tests: %v", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		t.Skipf("Skipping API tests: database not available: %v", err)
	}

	// Create test schema for API tests
	createAPITestSchema(t, db)
	apiTestDB = db
	return db
}

func createAPITestSchema(t *testing.T, db *sql.DB) {
	// Create a minimal schema for API testing
	schema := `
	DROP TABLE IF EXISTS api_test_workflow_events CASCADE;
	DROP TABLE IF EXISTS api_test_workflow_checkpoints CASCADE;
	DROP TABLE IF EXISTS api_test_workflow_queue CASCADE;
	DROP TABLE IF EXISTS api_test_workflow_steps CASCADE;
	DROP TABLE IF EXISTS api_test_workflow_runs CASCADE;
	DROP TABLE IF EXISTS api_test_workflow_workers CASCADE;
	DROP TABLE IF EXISTS api_test_agents CASCADE;

	CREATE TABLE api_test_agents (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		name TEXT NOT NULL
	);

	CREATE TABLE api_test_workflow_runs (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		agent_id UUID NOT NULL REFERENCES api_test_agents(id),
		version_id UUID,
		trigger_id UUID,
		status TEXT NOT NULL DEFAULT 'pending',
		created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
		started_at TIMESTAMP WITH TIME ZONE,
		completed_at TIMESTAMP WITH TIME ZONE,
		input_data JSONB,
		output_data JSONB,
		error_data JSONB,
		variables JSONB DEFAULT '{}',
		timeout_seconds INTEGER DEFAULT 3600,
		retry_policy JSONB DEFAULT '{}',
		assigned_worker_id TEXT,
		worker_heartbeat TIMESTAMP WITH TIME ZONE,
		total_steps INTEGER DEFAULT 0,
		completed_steps INTEGER DEFAULT 0,
		failed_steps INTEGER DEFAULT 0
	);

	CREATE TABLE api_test_workflow_steps (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		run_id UUID NOT NULL REFERENCES api_test_workflow_runs(id) ON DELETE CASCADE,
		node_id TEXT NOT NULL,
		node_type TEXT NOT NULL,
		step_number INTEGER NOT NULL,
		status TEXT NOT NULL DEFAULT 'pending',
		attempt_count INTEGER DEFAULT 0,
		max_attempts INTEGER DEFAULT 3,
		created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
		started_at TIMESTAMP WITH TIME ZONE,
		completed_at TIMESTAMP WITH TIME ZONE,
		next_retry_at TIMESTAMP WITH TIME ZONE,
		input_envelope JSONB,
		output_envelope JSONB,
		node_config JSONB,
		error_details JSONB,
		assigned_worker_id TEXT,
		worker_heartbeat TIMESTAMP WITH TIME ZONE,
		depends_on UUID[] DEFAULT '{}',
		UNIQUE(run_id, node_id)
	);

	-- Insert test agent
	INSERT INTO api_test_agents (id, name) VALUES 
	('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'API Test Agent');
	`

	_, err := db.Exec(schema)
	require.NoError(t, err, "Failed to create API test schema")
}

func cleanupAPITestData(t *testing.T, db *sql.DB) {
	tables := []string{
		"api_test_workflow_steps",
		"api_test_workflow_runs",
	}

	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			t.Logf("Warning: failed to clean table %s: %v", table, err)
		}
	}
}

// MockAPIEngine is a test implementation that uses test tables
type MockAPIEngine struct {
	db *sql.DB
}

func (m *MockAPIEngine) StartRun(ctx context.Context, run *execution.WorkflowRun) error {
	query := `
		INSERT INTO api_test_workflow_runs (
			id, agent_id, version_id, trigger_id, status, input_data, 
			variables, timeout_seconds, retry_policy
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)`

	inputDataJSON, _ := json.Marshal(run.InputData)
	variablesJSON, _ := json.Marshal(run.Variables)
	retryPolicyJSON, _ := json.Marshal(run.RetryPolicy)

	_, err := m.db.ExecContext(ctx, query,
		run.ID, run.AgentID, run.VersionID, run.TriggerID, run.Status,
		inputDataJSON, variablesJSON, run.TimeoutSeconds, retryPolicyJSON)
	return err
}

func (m *MockAPIEngine) PauseRun(ctx context.Context, runID uuid.UUID) error {
	query := `UPDATE api_test_workflow_runs SET status = 'paused' WHERE id = $1 AND status = 'running'`
	_, err := m.db.ExecContext(ctx, query, runID)
	return err
}

func (m *MockAPIEngine) ResumeRun(ctx context.Context, runID uuid.UUID) error {
	query := `UPDATE api_test_workflow_runs SET status = 'running' WHERE id = $1 AND status = 'paused'`
	_, err := m.db.ExecContext(ctx, query, runID)
	return err
}

func (m *MockAPIEngine) CancelRun(ctx context.Context, runID uuid.UUID) error {
	query := `UPDATE api_test_workflow_runs SET status = 'cancelled', completed_at = NOW() WHERE id = $1`
	_, err := m.db.ExecContext(ctx, query, runID)
	return err
}

func (m *MockAPIEngine) RetryStep(ctx context.Context, stepID uuid.UUID) error {
	// For this test, just return success
	return nil
}

// Implement remaining ExecutionEngine interface methods for testing
func (m *MockAPIEngine) ExecuteStep(ctx context.Context, step *execution.WorkflowStep) (*api.Envelope[any], error) {
	return nil, nil
}

func (m *MockAPIEngine) ClaimWork(ctx context.Context, workerID string, maxItems int) ([]*execution.QueueItem, error) {
	return nil, nil
}

func (m *MockAPIEngine) CompleteWork(ctx context.Context, workerID string, itemID uuid.UUID, result *execution.WorkResult) error {
	return nil
}

func (m *MockAPIEngine) RegisterWorker(ctx context.Context, worker *execution.WorkflowWorker) error {
	return nil
}

func (m *MockAPIEngine) UnregisterWorker(ctx context.Context, workerID string) error {
	return nil
}

func (m *MockAPIEngine) UpdateWorkerHeartbeat(ctx context.Context, workerID string) error {
	return nil
}

func (m *MockAPIEngine) RecoverOrphanedWork(ctx context.Context, workerTimeoutDuration time.Duration) error {
	return nil
}

func (m *MockAPIEngine) RecoverFailedRuns(ctx context.Context) error {
	return nil
}

// Helper to create test router with mock engine
func createTestRouter(db *sql.DB) http.Handler {
	r := chi.NewRouter()

	// Initialize mock workflow engine
	mockEngine := &MockAPIEngine{db: db}
	handler := NewWorkflowRunsHandler(db, mockEngine)

	// Add the workflow runs routes
	r.Get("/workflow-runs", handler.listWorkflowRuns)
	r.Post("/workflow-runs", handler.createWorkflowRun)
	r.Get("/workflow-runs/{runID}", handler.getWorkflowRun)
	r.Post("/workflow-runs/{runID}/control", handler.controlWorkflowRun)
	r.Post("/workflow-runs/{runID}/steps/{stepID}/retry", handler.retryWorkflowStep)

	return r
}

// Test creating a workflow run via API
func TestCreateWorkflowRunAPI(t *testing.T) {
	db := setupAPITestDB(t)
	defer cleanupAPITestData(t, db)

	router := createTestRouter(db)

	// Test data
	agentID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")

	createReq := CreateWorkflowRunRequest{
		AgentID: agentID,
		InputData: map[string]any{
			"test_input": "api_test_value",
			"number":     42,
		},
		Variables: map[string]any{
			"env": "test",
		},
		TimeoutSeconds: 1800,
	}

	reqBody, err := json.Marshal(createReq)
	require.NoError(t, err)

	// Make API request
	req := httptest.NewRequest(http.MethodPost, "/workflow-runs", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusCreated, w.Code)

	var response execution.WorkflowRun
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, agentID, response.AgentID)
	assert.Equal(t, execution.RunStatusPending, response.Status)
	assert.Equal(t, 1800, response.TimeoutSeconds)
	assert.NotEqual(t, uuid.Nil, response.ID)

	// Verify it was actually stored in database
	var storedCount int
	err = db.QueryRow("SELECT COUNT(*) FROM api_test_workflow_runs WHERE id = $1", response.ID).Scan(&storedCount)
	require.NoError(t, err)
	assert.Equal(t, 1, storedCount, "Workflow run should be stored in database")

	t.Logf("✅ Create workflow run API test passed - Created run %s", response.ID)
}

// Test listing workflow runs via API
func TestListWorkflowRunsAPI(t *testing.T) {
	db := setupAPITestDB(t)
	defer cleanupAPITestData(t, db)

	router := createTestRouter(db)

	// Create test data - insert some workflow runs directly
	agentID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")

	testRuns := []struct {
		id     uuid.UUID
		status string
	}{
		{uuid.New(), "completed"},
		{uuid.New(), "running"},
		{uuid.New(), "failed"},
	}

	for _, run := range testRuns {
		query := `
			INSERT INTO api_test_workflow_runs (id, agent_id, status, total_steps, completed_steps, failed_steps)
			VALUES ($1, $2, $3, $4, $5, $6)`

		totalSteps := 5
		completedSteps := 0
		failedSteps := 0

		if run.status == "completed" {
			completedSteps = 5
		} else if run.status == "failed" {
			completedSteps = 2
			failedSteps = 1
		} else if run.status == "running" {
			completedSteps = 3
		}

		_, err := db.Exec(query, run.id, agentID, run.status, totalSteps, completedSteps, failedSteps)
		require.NoError(t, err)
	}

	// Test 1: List all runs
	req := httptest.NewRequest(http.MethodGet, "/workflow-runs", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response PaginatedResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	runs, ok := response.Data.([]interface{})
	require.True(t, ok, "Response data should be an array")
	assert.Len(t, runs, 3, "Should return 3 workflow runs")

	// Test 2: Filter by agent ID
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/workflow-runs?agent_id=%s", agentID), nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	filteredRuns, ok := response.Data.([]interface{})
	require.True(t, ok)
	assert.Len(t, filteredRuns, 3, "Should return all 3 runs for this agent")

	// Test 3: Filter by status
	req = httptest.NewRequest(http.MethodGet, "/workflow-runs?status=completed", nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	completedRuns, ok := response.Data.([]interface{})
	require.True(t, ok)
	assert.Len(t, completedRuns, 1, "Should return 1 completed run")

	// Test 4: Pagination
	req = httptest.NewRequest(http.MethodGet, "/workflow-runs?limit=2&offset=0", nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	paginatedRuns, ok := response.Data.([]interface{})
	require.True(t, ok)
	assert.Len(t, paginatedRuns, 2, "Should return 2 runs with limit=2")

	assert.Equal(t, 2, response.Pagination.Limit)
	assert.Equal(t, 0, response.Pagination.Offset)
	assert.Equal(t, 3, response.Pagination.Total)
	assert.True(t, response.Pagination.HasMore)

	t.Logf("✅ List workflow runs API test passed - Tested filtering and pagination")
}

// Test workflow run control operations (pause/resume/cancel)
func TestWorkflowRunControlAPI(t *testing.T) {
	db := setupAPITestDB(t)
	defer cleanupAPITestData(t, db)

	router := createTestRouter(db)

	// Create a test workflow run
	agentID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	runID := uuid.New()

	query := `
		INSERT INTO api_test_workflow_runs (id, agent_id, status)
		VALUES ($1, $2, 'running')`

	_, err := db.Exec(query, runID, agentID)
	require.NoError(t, err)

	// Test 1: Pause workflow
	pauseReq := ControlWorkflowRunRequest{Action: "pause"}
	reqBody, err := json.Marshal(pauseReq)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/workflow-runs/%s/control", runID), bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	// Add chi URL params manually for testing
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("runID", runID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "success", response["status"])
	assert.Equal(t, "pause", response["action"])

	// Verify the run was paused in the database
	var status string
	err = db.QueryRow("SELECT status FROM api_test_workflow_runs WHERE id = $1", runID).Scan(&status)
	require.NoError(t, err)
	assert.Equal(t, "paused", status)

	// Test 2: Resume workflow
	resumeReq := ControlWorkflowRunRequest{Action: "resume"}
	reqBody, err = json.Marshal(resumeReq)
	require.NoError(t, err)

	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/workflow-runs/%s/control", runID), bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "resume", response["action"])

	// Test 3: Cancel workflow
	cancelReq := ControlWorkflowRunRequest{Action: "cancel"}
	reqBody, err = json.Marshal(cancelReq)
	require.NoError(t, err)

	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/workflow-runs/%s/control", runID), bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "cancel", response["action"])

	// Verify the run was cancelled in the database
	err = db.QueryRow("SELECT status FROM api_test_workflow_runs WHERE id = $1", runID).Scan(&status)
	require.NoError(t, err)
	assert.Equal(t, "cancelled", status)

	// Test 4: Invalid action
	invalidReq := ControlWorkflowRunRequest{Action: "invalid_action"}
	reqBody, err = json.Marshal(invalidReq)
	require.NoError(t, err)

	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/workflow-runs/%s/control", runID), bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	t.Logf("✅ Workflow run control API test passed - Tested pause/resume/cancel operations")
}

// Test getting a specific workflow run
func TestGetWorkflowRunAPI(t *testing.T) {
	db := setupAPITestDB(t)
	defer cleanupAPITestData(t, db)

	router := createTestRouter(db)

	// Create a test workflow run with some steps
	agentID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	runID := uuid.New()

	// Insert workflow run
	runQuery := `
		INSERT INTO api_test_workflow_runs (id, agent_id, status, total_steps, completed_steps, input_data)
		VALUES ($1, $2, 'running', 3, 1, $3)`

	inputData := map[string]any{"test": "get_workflow_run"}
	inputJSON, _ := json.Marshal(inputData)

	_, err := db.Exec(runQuery, runID, agentID, inputJSON)
	require.NoError(t, err)

	// Insert some workflow steps
	step1ID := uuid.New()
	step2ID := uuid.New()

	stepQuery := `
		INSERT INTO api_test_workflow_steps (id, run_id, node_id, node_type, step_number, status)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err = db.Exec(stepQuery, step1ID, runID, "trigger", "trigger", 1, "completed")
	require.NoError(t, err)

	_, err = db.Exec(stepQuery, step2ID, runID, "action1", "action", 2, "running")
	require.NoError(t, err)

	// Create the get_workflow_run_details function for testing
	functionSQL := `
	CREATE OR REPLACE FUNCTION get_workflow_run_details(run_uuid UUID)
	RETURNS TABLE(
		run_id UUID,
		agent_id UUID,
		agent_name TEXT,
		status TEXT,
		created_at TIMESTAMPTZ,
		started_at TIMESTAMPTZ,
		completed_at TIMESTAMPTZ,
		duration_seconds NUMERIC,
		progress_percentage NUMERIC,
		total_steps INTEGER,
		completed_steps INTEGER,
		failed_steps INTEGER,
		input_data JSONB,
		output_data JSONB,
		error_data JSONB,
		steps JSONB,
		checkpoints JSONB
	) AS $$
	BEGIN
		RETURN QUERY
		SELECT 
			wr.id as run_id,
			wr.agent_id,
			a.name as agent_name,
			wr.status::TEXT,
			wr.created_at,
			wr.started_at,
			wr.completed_at,
			NULL::NUMERIC as duration_seconds,
			CASE 
				WHEN wr.total_steps > 0 THEN 
					ROUND((wr.completed_steps::decimal / wr.total_steps::decimal) * 100, 1)
				ELSE 0 
			END as progress_percentage,
			wr.total_steps,
			wr.completed_steps,
			wr.failed_steps,
			wr.input_data,
			wr.output_data,
			wr.error_data,
			'[]'::JSONB as steps,
			'[]'::JSONB as checkpoints
		FROM api_test_workflow_runs wr
		LEFT JOIN api_test_agents a ON wr.agent_id = a.id
		WHERE wr.id = run_uuid;
	END;
	$$ LANGUAGE plpgsql;`

	_, err = db.Exec(functionSQL)
	require.NoError(t, err)

	// Make API request
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/workflow-runs/%s", runID), nil)

	// Add chi URL params manually for testing
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("runID", runID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response WorkflowRunDetails
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, runID, response.ID)
	assert.Equal(t, agentID, response.AgentID)
	assert.Equal(t, "running", response.Status)
	assert.Equal(t, 3, response.TotalSteps)
	assert.Equal(t, 1, response.CompletedSteps)
	assert.Equal(t, 0, response.FailedSteps)
	assert.Equal(t, 33.3, response.ProgressPercentage) // 1/3 * 100 = 33.3
	assert.Equal(t, "API Test Agent", *response.AgentName)

	t.Logf("✅ Get workflow run API test passed - Retrieved run %s with correct details", runID)
}

// Test step retry functionality
func TestRetryWorkflowStepAPI(t *testing.T) {
	db := setupAPITestDB(t)
	defer cleanupAPITestData(t, db)

	router := createTestRouter(db)

	// Create test data
	agentID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	runID := uuid.New()
	stepID := uuid.New()

	// Insert workflow run
	runQuery := `INSERT INTO api_test_workflow_runs (id, agent_id, status) VALUES ($1, $2, 'running')`
	_, err := db.Exec(runQuery, runID, agentID)
	require.NoError(t, err)

	// Insert failed step
	stepQuery := `INSERT INTO api_test_workflow_steps (id, run_id, node_id, node_type, step_number, status) VALUES ($1, $2, 'failed_node', 'action', 1, 'failed')`
	_, err = db.Exec(stepQuery, stepID, runID)
	require.NoError(t, err)

	// Make retry request
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/workflow-runs/%s/steps/%s/retry", runID, stepID), nil)

	// Add chi URL params manually
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("runID", runID.String())
	rctx.URLParams.Add("stepID", stepID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "success", response["status"])
	assert.Equal(t, "retry", response["action"])

	t.Logf("✅ Retry workflow step API test passed - Step retry successful")
}

// Test error handling
func TestWorkflowRunAPIErrorHandling(t *testing.T) {
	db := setupAPITestDB(t)
	defer cleanupAPITestData(t, db)

	router := createTestRouter(db)

	// Test 1: Invalid JSON in create request
	req := httptest.NewRequest(http.MethodPost, "/workflow-runs", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test 2: Missing agent_id in create request
	invalidReq := map[string]any{"input_data": map[string]any{"test": "value"}}
	reqBody, _ := json.Marshal(invalidReq)

	req = httptest.NewRequest(http.MethodPost, "/workflow-runs", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test 3: Invalid UUID in get request
	req = httptest.NewRequest(http.MethodGet, "/workflow-runs/invalid-uuid", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("runID", "invalid-uuid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test 4: Non-existent workflow run
	nonExistentID := uuid.New()
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/workflow-runs/%s", nonExistentID), nil)

	rctx = chi.NewRouteContext()
	rctx.URLParams.Add("runID", nonExistentID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)

	t.Logf("✅ Workflow run API error handling test passed - All error cases handled correctly")
}
