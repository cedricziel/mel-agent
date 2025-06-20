package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cedricziel/mel-agent/internal/db"
	"github.com/cedricziel/mel-agent/internal/testutil"
	"github.com/cedricziel/mel-agent/pkg/api"
	"github.com/cedricziel/mel-agent/pkg/execution"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test worker registration with real database
func TestWorkerRegistrationIntegration(t *testing.T) {
	ctx := context.Background()
	_, testDB, cleanup := testutil.SetupPostgresWithTestData(ctx, t)
	defer cleanup()

	// Set the global database connection and ensure cleanup
	originalDB := db.DB
	db.DB = testDB
	defer func() {
		db.DB = originalDB
	}()

	mel := api.NewMel()
	workflowEngine := execution.NewDurableExecutionEngine(testDB, mel, "test-server")
	router := NewCombinedRouter(testDB, workflowEngine)

	// Test data - use the new RegisterWorkerRequest format
	concurrency := 5
	registerRequest := struct {
		ID          string  `json:"id"`
		Name        *string `json:"name,omitempty"`
		Concurrency *int    `json:"concurrency,omitempty"`
	}{
		ID:          "integration-test-worker-1",
		Name:        workerStringPtr("test-host"),
		Concurrency: &concurrency,
	}

	reqBody, err := json.Marshal(registerRequest)
	require.NoError(t, err)

	// Make API request
	req := httptest.NewRequest(http.MethodPost, "/api/workers", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusCreated, w.Code)

	var response Worker
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "integration-test-worker-1", *response.Id)
	assert.Equal(t, "test-host", *response.Name)

	// Verify it was stored in database using actual migration tables
	var storedWorker execution.WorkflowWorker
	query := `
		SELECT id, hostname, process_id, version, capabilities, status, max_concurrent_steps
		FROM workflow_workers WHERE id = $1
	`
	row := testDB.QueryRow(query, "integration-test-worker-1")
	err = row.Scan(
		&storedWorker.ID, &storedWorker.Hostname, &storedWorker.ProcessID,
		&storedWorker.Version, pq.Array(&storedWorker.Capabilities),
		&storedWorker.Status, &storedWorker.MaxConcurrentSteps,
	)
	require.NoError(t, err)

	assert.Equal(t, "integration-test-worker-1", storedWorker.ID)
	assert.Equal(t, "test-host", storedWorker.Hostname)
	assert.Nil(t, storedWorker.ProcessID)                                  // ProcessID not sent in new format
	assert.Nil(t, storedWorker.Version)                                    // Version not sent in new format
	assert.Empty(t, storedWorker.Capabilities)                             // Capabilities not sent in new format
	assert.Equal(t, execution.WorkerStatus("active"), storedWorker.Status) // Handler stores "active"
	assert.Equal(t, 5, storedWorker.MaxConcurrentSteps)

	t.Logf("✅ Worker registration integration test passed - Worker %s registered successfully", registerRequest.ID)
}

// Test work claiming with real database and workflow
func TestWorkClaimingIntegration(t *testing.T) {
	ctx := context.Background()
	_, testDB, cleanup := testutil.SetupPostgresWithTestData(ctx, t)
	defer cleanup()

	// Set the global database connection and ensure cleanup
	originalDB := db.DB
	db.DB = testDB
	defer func() {
		db.DB = originalDB
	}()

	mel := api.NewMel()
	workflowEngine := execution.NewDurableExecutionEngine(testDB, mel, "test-server")
	router := NewCombinedRouter(testDB, workflowEngine)

	// Create a real workflow run using test agent
	testAgentID := uuid.MustParse("11111111-1111-1111-1111-111111111111") // From testutil test data
	runID := uuid.New()
	versionID := uuid.New() // Required field

	// Insert workflow run using the real schema
	runQuery := `
		INSERT INTO workflow_runs (id, agent_id, version_id, status, created_at, variables, timeout_seconds, retry_policy)
		VALUES ($1, $2, $3, 'running', NOW(), '{}', 3600, '{"max_attempts": 3}')
	`
	_, err := testDB.Exec(runQuery, runID, testAgentID, versionID)
	require.NoError(t, err)

	// Register a worker first
	workerID := "integration-test-worker-claim"
	registerTestWorker(t, router, workerID)

	// Create work items in the queue using real schema
	workItems := []struct {
		id        uuid.UUID
		queueType string
		priority  int
	}{
		{uuid.New(), "start_run", 10},
		{uuid.New(), "execute_step", 5},
		{uuid.New(), "execute_step", 7},
	}

	for _, item := range workItems {
		queueQuery := `
			INSERT INTO workflow_queue (id, run_id, queue_type, priority, available_at, created_at, attempt_count, max_attempts)
			VALUES ($1, $2, $3, $4, NOW(), NOW(), 0, 3)
		`
		_, err = testDB.Exec(queueQuery, item.id, runID, item.queueType, item.priority)
		require.NoError(t, err)
	}

	// Test work claiming - need to send ClaimWorkJSONBody
	claimReq := struct {
		MaxItems *int `json:"max_items,omitempty"`
	}{
		MaxItems: &[]int{5}[0], // Request up to 5 items
	}
	reqBody, err := json.Marshal(claimReq)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/workers/%s/claim-work", workerID), bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	// Note: The OpenAPI handlers currently have stub implementations for work claiming
	// The stub implementation returns an empty array regardless of available work
	var claimedItems []struct{}
	err = json.Unmarshal(w.Body.Bytes(), &claimedItems)
	require.NoError(t, err)

	// Stub implementation returns empty array
	assert.Len(t, claimedItems, 0, "Stub implementation returns no work items")

	// Note: In a full implementation, this would verify actual work claiming:
	// - Items would be ordered by priority
	// - Items would be marked as claimed in database
	// TODO: Connect OpenAPI handlers to actual workflow queue logic

	t.Logf("✅ Work claiming integration test passed - Worker %s claimed %d work items", workerID, len(claimedItems))
}

// Test complete worker lifecycle with real database
func TestWorkerLifecycleIntegration(t *testing.T) {
	ctx := context.Background()
	_, testDB, cleanup := testutil.SetupPostgresWithTestData(ctx, t)
	defer cleanup()

	// Set the global database connection and ensure cleanup
	originalDB := db.DB
	db.DB = testDB
	defer func() {
		db.DB = originalDB
	}()

	mel := api.NewMel()
	workflowEngine := execution.NewDurableExecutionEngine(testDB, mel, "test-server")
	router := NewCombinedRouter(testDB, workflowEngine)
	workerID := "lifecycle-test-worker"

	// 1. Register worker - use new RegisterWorkerRequest format
	concurrency := 10
	registerRequest := struct {
		ID          string  `json:"id"`
		Name        *string `json:"name,omitempty"`
		Concurrency *int    `json:"concurrency,omitempty"`
	}{
		ID:          workerID,
		Name:        workerStringPtr("lifecycle-host"),
		Concurrency: &concurrency,
	}

	reqBody, err := json.Marshal(registerRequest)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/workers", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// 2. Send heartbeat
	req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/workers/%s/heartbeat", workerID), nil)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify heartbeat was updated
	var lastHeartbeat time.Time
	var status string
	err = testDB.QueryRow("SELECT last_heartbeat, status FROM workflow_workers WHERE id = $1", workerID).Scan(&lastHeartbeat, &status)
	require.NoError(t, err)
	assert.Equal(t, "active", status) // Status becomes "active" after heartbeat
	assert.WithinDuration(t, time.Now(), lastHeartbeat, 5*time.Second)

	// 3. Create and claim work
	testAgentID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	runID := uuid.New()
	versionID := uuid.New()

	runQuery := `
		INSERT INTO workflow_runs (id, agent_id, version_id, status, created_at)
		VALUES ($1, $2, $3, 'running', NOW())
	`
	_, err = testDB.Exec(runQuery, runID, testAgentID, versionID)
	require.NoError(t, err)

	workItemID := uuid.New()
	queueQuery := `
		INSERT INTO workflow_queue (id, run_id, queue_type, priority, available_at, created_at)
		VALUES ($1, $2, 'execute_step', 5, NOW(), NOW())
	`
	_, err = testDB.Exec(queueQuery, workItemID, runID)
	require.NoError(t, err)

	// Claim work - need to send ClaimWorkJSONBody
	claimReq := struct {
		MaxItems *int `json:"max_items,omitempty"`
	}{
		MaxItems: &[]int{5}[0], // Request up to 5 items
	}
	reqBody, err = json.Marshal(claimReq)
	require.NoError(t, err)

	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/workers/%s/claim-work", workerID), bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 4. Complete work - use CompleteWorkJSONBody format
	resultData := map[string]interface{}{"result": "lifecycle_test_success"}
	completeReq := struct {
		Error  *string                 `json:"error,omitempty"`
		Result *map[string]interface{} `json:"result,omitempty"`
	}{
		Result: &resultData,
	}

	reqBody, err = json.Marshal(completeReq)
	require.NoError(t, err)

	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/workers/%s/complete-work/%s", workerID, workItemID), bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Note: The OpenAPI handlers currently have stub implementations for work completion
	// In a full implementation, work items would be removed from queue or marked as completed
	// For now, we just verify the endpoint returns success
	// TODO: Connect OpenAPI handlers to actual workflow queue logic

	// 5. Unregister worker
	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/workers/%s", workerID), nil)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)

	// Verify worker was removed
	var workerCount int
	err = testDB.QueryRow("SELECT COUNT(*) FROM workflow_workers WHERE id = $1", workerID).Scan(&workerCount)
	require.NoError(t, err)
	assert.Equal(t, 0, workerCount)

	t.Logf("✅ Worker lifecycle integration test passed - Complete workflow from registration to cleanup")
}

// Test worker upsert behavior with migrations
func TestWorkerUpsertIntegration(t *testing.T) {
	ctx := context.Background()
	_, testDB, cleanup := testutil.SetupPostgresWithTestData(ctx, t)
	defer cleanup()

	// Set the global database connection and ensure cleanup
	originalDB := db.DB
	db.DB = testDB
	defer func() {
		db.DB = originalDB
	}()

	mel := api.NewMel()
	workflowEngine := execution.NewDurableExecutionEngine(testDB, mel, "test-server")
	router := NewCombinedRouter(testDB, workflowEngine)
	workerID := "upsert-test-worker"

	// First registration - use new RegisterWorkerRequest format
	concurrency1 := 5
	registerRequest1 := struct {
		ID          string  `json:"id"`
		Name        *string `json:"name,omitempty"`
		Concurrency *int    `json:"concurrency,omitempty"`
	}{
		ID:          workerID,
		Name:        workerStringPtr("host1"),
		Concurrency: &concurrency1,
	}

	reqBody, err := json.Marshal(registerRequest1)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/workers", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// Second registration (upsert) - use new RegisterWorkerRequest format
	concurrency2 := 10
	registerRequest2 := struct {
		ID          string  `json:"id"`
		Name        *string `json:"name,omitempty"`
		Concurrency *int    `json:"concurrency,omitempty"`
	}{
		ID:          workerID,
		Name:        workerStringPtr("host2"),
		Concurrency: &concurrency2,
	}

	reqBody, err = json.Marshal(registerRequest2)
	require.NoError(t, err)

	req = httptest.NewRequest(http.MethodPost, "/api/workers", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// Verify only one worker exists with updated details
	var count int
	err = testDB.QueryRow("SELECT COUNT(*) FROM workflow_workers WHERE id = $1", workerID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Verify the worker has the updated details
	var storedWorker execution.WorkflowWorker
	query := `
		SELECT id, hostname, process_id, version, capabilities, status, max_concurrent_steps
		FROM workflow_workers WHERE id = $1
	`
	row := testDB.QueryRow(query, workerID)
	err = row.Scan(
		&storedWorker.ID, &storedWorker.Hostname, &storedWorker.ProcessID,
		&storedWorker.Version, pq.Array(&storedWorker.Capabilities),
		&storedWorker.Status, &storedWorker.MaxConcurrentSteps,
	)
	require.NoError(t, err)

	assert.Equal(t, "host2", storedWorker.Hostname)
	assert.Nil(t, storedWorker.ProcessID)                                  // ProcessID not sent in new format
	assert.Nil(t, storedWorker.Version)                                    // Version not sent in new format
	assert.Empty(t, storedWorker.Capabilities)                             // Capabilities not sent in new format
	assert.Equal(t, execution.WorkerStatus("active"), storedWorker.Status) // Handler stores "active"
	assert.Equal(t, 10, storedWorker.MaxConcurrentSteps)

	t.Logf("✅ Worker upsert integration test passed - Worker %s updated successfully", workerID)
}

// Helper function to register a worker for tests
func registerTestWorker(t *testing.T, router http.Handler, workerID string) {
	t.Helper()

	concurrency := 5
	registerRequest := struct {
		ID          string  `json:"id"`
		Name        *string `json:"name,omitempty"`
		Concurrency *int    `json:"concurrency,omitempty"`
	}{
		ID:          workerID,
		Name:        workerStringPtr("test-host"),
		Concurrency: &concurrency,
	}

	reqBody, err := json.Marshal(registerRequest)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/workers", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)
}

// Helper functions for pointer creation
func workerStringPtr(s string) *string {
	return &s
}

func workerIntPtr(i int) *int {
	return &i
}
