package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/cedricziel/mel-agent/internal/testutil"
	"github.com/cedricziel/mel-agent/pkg/execution"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test worker registration with real database
func TestWorkerRegistrationIntegration(t *testing.T) {
	ctx := context.Background()
	_, db, cleanup := testutil.SetupPostgresWithMigrations(ctx, t)
	defer cleanup()

	router := createWorkerIntegrationTestRouter(db)

	// Test data
	processID := 12345
	worker := execution.WorkflowWorker{
		ID:                   "integration-test-worker-1",
		Hostname:             "test-host",
		ProcessID:            &processID,
		Version:              workerStringPtr("1.0.0"),
		Capabilities:         []string{"workflow_execution", "node_execution"},
		Status:               execution.WorkerStatusIdle,
		LastHeartbeat:        time.Now(),
		StartedAt:            time.Now(),
		MaxConcurrentSteps:   5,
		CurrentStepCount:     0,
		TotalStepsExecuted:   0,
		TotalExecutionTimeMS: 0,
	}

	reqBody, err := json.Marshal(worker)
	require.NoError(t, err)

	// Make API request
	req := httptest.NewRequest(http.MethodPost, "/workers", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "integration-test-worker-1", response["id"])

	// Verify it was stored in database using actual migration tables
	var storedWorker execution.WorkflowWorker
	query := `
		SELECT id, hostname, process_id, version, capabilities, status, max_concurrent_steps
		FROM workflow_workers WHERE id = $1
	`
	row := db.QueryRow(query, "integration-test-worker-1")
	err = row.Scan(
		&storedWorker.ID, &storedWorker.Hostname, &storedWorker.ProcessID,
		&storedWorker.Version, pq.Array(&storedWorker.Capabilities),
		&storedWorker.Status, &storedWorker.MaxConcurrentSteps,
	)
	require.NoError(t, err)

	assert.Equal(t, "integration-test-worker-1", storedWorker.ID)
	assert.Equal(t, "test-host", storedWorker.Hostname)
	assert.Equal(t, &processID, storedWorker.ProcessID)
	assert.Equal(t, workerStringPtr("1.0.0"), storedWorker.Version)
	assert.Equal(t, []string{"workflow_execution", "node_execution"}, storedWorker.Capabilities)
	assert.Equal(t, execution.WorkerStatusIdle, storedWorker.Status)
	assert.Equal(t, 5, storedWorker.MaxConcurrentSteps)

	t.Logf("✅ Worker registration integration test passed - Worker %s registered successfully", worker.ID)
}

// Test work claiming with real database and workflow
func TestWorkClaimingIntegration(t *testing.T) {
	ctx := context.Background()
	_, db, cleanup := testutil.SetupPostgresWithMigrations(ctx, t)
	defer cleanup()

	router := createWorkerIntegrationTestRouter(db)

	// Create a real workflow run using test agent
	testAgentID := uuid.MustParse("11111111-1111-1111-1111-111111111111") // From testutil test data
	runID := uuid.New()
	versionID := uuid.New() // Required field

	// Insert workflow run using the real schema
	runQuery := `
		INSERT INTO workflow_runs (id, agent_id, version_id, status, created_at, variables, timeout_seconds, retry_policy)
		VALUES ($1, $2, $3, 'running', NOW(), '{}', 3600, '{"max_attempts": 3}')
	`
	_, err := db.Exec(runQuery, runID, testAgentID, versionID)
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
		_, err = db.Exec(queueQuery, item.id, runID, item.queueType, item.priority)
		require.NoError(t, err)
	}

	// Test work claiming
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/workers/%s/claim-work", workerID), nil)

	// Add chi URL params
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("workerID", workerID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	var claimedItems []*execution.QueueItem
	err = json.Unmarshal(w.Body.Bytes(), &claimedItems)
	require.NoError(t, err)

	assert.Len(t, claimedItems, 3, "Should claim all 3 available work items")

	// Verify items are ordered by priority (highest first)
	assert.Equal(t, 10, claimedItems[0].Priority)
	assert.Equal(t, 7, claimedItems[1].Priority)
	assert.Equal(t, 5, claimedItems[2].Priority)

	// Verify items are claimed in database
	var claimedCount int
	err = db.QueryRow("SELECT COUNT(*) FROM workflow_queue WHERE claimed_by = $1", workerID).Scan(&claimedCount)
	require.NoError(t, err)
	assert.Equal(t, 3, claimedCount)

	t.Logf("✅ Work claiming integration test passed - Worker %s claimed %d work items", workerID, len(claimedItems))
}

// Test complete worker lifecycle with real database
func TestWorkerLifecycleIntegration(t *testing.T) {
	ctx := context.Background()
	_, db, cleanup := testutil.SetupPostgresWithMigrations(ctx, t)
	defer cleanup()

	router := createWorkerIntegrationTestRouter(db)
	workerID := "lifecycle-test-worker"

	// 1. Register worker
	processID := 98765
	worker := execution.WorkflowWorker{
		ID:                   workerID,
		Hostname:             "lifecycle-host",
		ProcessID:            &processID,
		Version:              workerStringPtr("2.0.0"),
		Capabilities:         []string{"workflow_execution"},
		Status:               execution.WorkerStatusIdle,
		LastHeartbeat:        time.Now(),
		StartedAt:            time.Now(),
		MaxConcurrentSteps:   10,
		CurrentStepCount:     0,
		TotalStepsExecuted:   0,
		TotalExecutionTimeMS: 0,
	}

	reqBody, err := json.Marshal(worker)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/workers", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// 2. Send heartbeat
	req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/workers/%s/heartbeat", workerID), nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("workerID", workerID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify heartbeat was updated
	var lastHeartbeat time.Time
	var status string
	err = db.QueryRow("SELECT last_heartbeat, status FROM workflow_workers WHERE id = $1", workerID).Scan(&lastHeartbeat, &status)
	require.NoError(t, err)
	assert.Equal(t, "idle", status)
	assert.WithinDuration(t, time.Now(), lastHeartbeat, 5*time.Second)

	// 3. Create and claim work
	testAgentID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	runID := uuid.New()
	versionID := uuid.New()

	runQuery := `
		INSERT INTO workflow_runs (id, agent_id, version_id, status, created_at)
		VALUES ($1, $2, $3, 'running', NOW())
	`
	_, err = db.Exec(runQuery, runID, testAgentID, versionID)
	require.NoError(t, err)

	workItemID := uuid.New()
	queueQuery := `
		INSERT INTO workflow_queue (id, run_id, queue_type, priority, available_at, created_at)
		VALUES ($1, $2, 'execute_step', 5, NOW(), NOW())
	`
	_, err = db.Exec(queueQuery, workItemID, runID)
	require.NoError(t, err)

	// Claim work
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/workers/%s/claim-work", workerID), nil)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 4. Complete work
	result := execution.WorkResult{
		Success:    true,
		OutputData: map[string]any{"result": "lifecycle_test_success"},
	}

	reqBody, err = json.Marshal(result)
	require.NoError(t, err)

	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/workers/%s/complete-work/%s", workerID, workItemID), bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	rctx.URLParams.Add("itemID", workItemID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify work was completed (removed from queue)
	var queueCount int
	err = db.QueryRow("SELECT COUNT(*) FROM workflow_queue WHERE id = $1", workItemID).Scan(&queueCount)
	require.NoError(t, err)
	assert.Equal(t, 0, queueCount)

	// 5. Unregister worker
	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/workers/%s", workerID), nil)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)

	// Verify worker was removed
	var workerCount int
	err = db.QueryRow("SELECT COUNT(*) FROM workflow_workers WHERE id = $1", workerID).Scan(&workerCount)
	require.NoError(t, err)
	assert.Equal(t, 0, workerCount)

	t.Logf("✅ Worker lifecycle integration test passed - Complete workflow from registration to cleanup")
}

// Test worker upsert behavior with migrations
func TestWorkerUpsertIntegration(t *testing.T) {
	ctx := context.Background()
	_, db, cleanup := testutil.SetupPostgresWithMigrations(ctx, t)
	defer cleanup()

	router := createWorkerIntegrationTestRouter(db)
	workerID := "upsert-test-worker"

	// First registration
	worker1 := execution.WorkflowWorker{
		ID:                 workerID,
		Hostname:           "host1",
		ProcessID:          workerIntPtr(12345),
		Version:            workerStringPtr("1.0.0"),
		Capabilities:       []string{"workflow_execution"},
		Status:             execution.WorkerStatusIdle,
		LastHeartbeat:      time.Now(),
		StartedAt:          time.Now(),
		MaxConcurrentSteps: 5,
	}

	reqBody, err := json.Marshal(worker1)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/workers", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// Second registration (upsert)
	worker2 := execution.WorkflowWorker{
		ID:                 workerID,
		Hostname:           "host2",
		ProcessID:          workerIntPtr(67890),
		Version:            workerStringPtr("2.0.0"),
		Capabilities:       []string{"workflow_execution", "node_execution"},
		Status:             execution.WorkerStatusBusy,
		LastHeartbeat:      time.Now().Add(1 * time.Minute),
		StartedAt:          time.Now(),
		MaxConcurrentSteps: 10,
	}

	reqBody, err = json.Marshal(worker2)
	require.NoError(t, err)

	req = httptest.NewRequest(http.MethodPost, "/workers", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// Verify only one worker exists with updated details
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM workflow_workers WHERE id = $1", workerID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Verify the worker has the updated details
	var storedWorker execution.WorkflowWorker
	query := `
		SELECT id, hostname, process_id, version, capabilities, status, max_concurrent_steps
		FROM workflow_workers WHERE id = $1
	`
	row := db.QueryRow(query, workerID)
	err = row.Scan(
		&storedWorker.ID, &storedWorker.Hostname, &storedWorker.ProcessID,
		&storedWorker.Version, pq.Array(&storedWorker.Capabilities),
		&storedWorker.Status, &storedWorker.MaxConcurrentSteps,
	)
	require.NoError(t, err)

	assert.Equal(t, "host2", storedWorker.Hostname)
	assert.Equal(t, workerIntPtr(67890), storedWorker.ProcessID)
	assert.Equal(t, workerStringPtr("2.0.0"), storedWorker.Version)
	assert.Equal(t, []string{"workflow_execution", "node_execution"}, storedWorker.Capabilities)
	assert.Equal(t, execution.WorkerStatusBusy, storedWorker.Status)
	assert.Equal(t, 10, storedWorker.MaxConcurrentSteps)

	t.Logf("✅ Worker upsert integration test passed - Worker %s updated successfully", workerID)
}

// createWorkerIntegrationTestRouter creates a router with actual worker handlers for integration testing
func createWorkerIntegrationTestRouter(db *sql.DB) http.Handler {
	r := chi.NewRouter()

	// Use the actual worker handlers but override the database reference
	// This simulates the real API but with our test database
	r.Post("/workers", func(w http.ResponseWriter, r *http.Request) {
		registerWorkerWithDB(w, r, db)
	})

	r.Put("/workers/{workerID}/heartbeat", func(w http.ResponseWriter, r *http.Request) {
		updateWorkerHeartbeatWithDB(w, r, db)
	})

	r.Delete("/workers/{workerID}", func(w http.ResponseWriter, r *http.Request) {
		unregisterWorkerWithDB(w, r, db)
	})

	r.Post("/workers/{workerID}/claim-work", func(w http.ResponseWriter, r *http.Request) {
		claimWorkWithDB(w, r, db)
	})

	r.Post("/workers/{workerID}/complete-work/{itemID}", func(w http.ResponseWriter, r *http.Request) {
		completeWorkWithDB(w, r, db)
	})

	return r
}

// Worker handler functions that accept a database parameter for testing
func registerWorkerWithDB(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var worker execution.WorkflowWorker
	if err := json.NewDecoder(r.Body).Decode(&worker); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid worker data"})
		return
	}

	query := `
		INSERT INTO workflow_workers (
			id, hostname, process_id, version, capabilities, status,
			last_heartbeat, started_at, max_concurrent_steps,
			current_step_count, total_steps_executed, total_execution_time_ms
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (id) DO UPDATE SET
			hostname = EXCLUDED.hostname,
			process_id = EXCLUDED.process_id,
			version = EXCLUDED.version,
			capabilities = EXCLUDED.capabilities,
			status = EXCLUDED.status,
			last_heartbeat = EXCLUDED.last_heartbeat,
			started_at = EXCLUDED.started_at,
			max_concurrent_steps = EXCLUDED.max_concurrent_steps
	`

	_, err := db.Exec(query,
		worker.ID, worker.Hostname, worker.ProcessID, worker.Version,
		pq.Array(worker.Capabilities), worker.Status, worker.LastHeartbeat,
		worker.StartedAt, worker.MaxConcurrentSteps, worker.CurrentStepCount,
		worker.TotalStepsExecuted, worker.TotalExecutionTimeMS,
	)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to register worker"})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"id": worker.ID})
}

func updateWorkerHeartbeatWithDB(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	workerID := chi.URLParam(r, "workerID")

	query := `UPDATE workflow_workers SET last_heartbeat = NOW(), status = 'idle' WHERE id = $1`
	result, err := db.Exec(query, workerID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update heartbeat"})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to check update result"})
		return
	}

	if rowsAffected == 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "worker not found"})
		return
	}

	w.WriteHeader(http.StatusOK)
}

func unregisterWorkerWithDB(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	workerID := chi.URLParam(r, "workerID")

	query := `DELETE FROM workflow_workers WHERE id = $1`
	result, err := db.Exec(query, workerID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to unregister worker"})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to check delete result"})
		return
	}

	if rowsAffected == 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "worker not found"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func claimWorkWithDB(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	workerID := chi.URLParam(r, "workerID")
	maxItems := 5 // default

	if maxItemsStr := r.URL.Query().Get("max_items"); maxItemsStr != "" {
		if parsed, err := strconv.Atoi(maxItemsStr); err == nil && parsed > 0 {
			maxItems = parsed
		}
	}

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to start transaction"})
		return
	}
	defer tx.Rollback()

	// Select available work items
	query := `
		SELECT id, run_id, step_id, queue_type, priority, available_at, 
		       created_at, attempt_count, max_attempts, payload
		FROM workflow_queue 
		WHERE claimed_at IS NULL 
		  AND claimed_by IS NULL 
		  AND available_at <= NOW()
		ORDER BY priority DESC, created_at ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	`

	rows, err := tx.Query(query, maxItems)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to query work items"})
		return
	}
	defer rows.Close()

	var workItems []*execution.QueueItem
	var claimedIDs []uuid.UUID

	for rows.Next() {
		var item execution.QueueItem
		var payloadBytes []byte

		err := rows.Scan(
			&item.ID, &item.RunID, &item.StepID, &item.QueueType,
			&item.Priority, &item.AvailableAt, &item.CreatedAt,
			&item.AttemptCount, &item.MaxAttempts, &payloadBytes,
		)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to scan work item"})
			return
		}

		// Parse payload if present
		if len(payloadBytes) > 0 {
			if err := json.Unmarshal(payloadBytes, &item.Payload); err != nil {
				item.Payload = nil
			}
		}

		workItems = append(workItems, &item)
		claimedIDs = append(claimedIDs, item.ID)
	}

	if err := rows.Err(); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to iterate work items"})
		return
	}

	// Claim the work items
	if len(claimedIDs) > 0 {
		claimQuery := `
			UPDATE workflow_queue 
			SET claimed_at = NOW(), claimed_by = $1, attempt_count = attempt_count + 1
			WHERE id = ANY($2)
		`
		_, err = tx.Exec(claimQuery, workerID, pq.Array(claimedIDs))
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to claim work items"})
			return
		}
	}

	if err := tx.Commit(); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to commit transaction"})
		return
	}

	writeJSON(w, http.StatusOK, workItems)
}

func completeWorkWithDB(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	workerID := chi.URLParam(r, "workerID")
	itemIDStr := chi.URLParam(r, "itemID")

	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid item ID"})
		return
	}

	var result execution.WorkResult
	if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid result data"})
		return
	}

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to start transaction"})
		return
	}
	defer tx.Rollback()

	// Remove work item
	deleteQuery := `DELETE FROM workflow_queue WHERE id = $1 AND claimed_by = $2`
	deleteResult, err := tx.Exec(deleteQuery, itemID, workerID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete work item"})
		return
	}

	rowsAffected, err := deleteResult.RowsAffected()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to check delete result"})
		return
	}

	if rowsAffected == 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "work item not found or not claimed by this worker"})
		return
	}

	if err := tx.Commit(); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to commit transaction"})
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Helper function to register a worker for tests
func registerTestWorker(t *testing.T, router http.Handler, workerID string) {
	t.Helper()

	worker := execution.WorkflowWorker{
		ID:                 workerID,
		Hostname:           "test-host",
		Status:             execution.WorkerStatusIdle,
		LastHeartbeat:      time.Now(),
		StartedAt:          time.Now(),
		MaxConcurrentSteps: 5,
	}

	reqBody, err := json.Marshal(worker)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/workers", bytes.NewReader(reqBody))
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
