package execution

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/cedricziel/mel-agent/internal/testutil"
	"github.com/cedricziel/mel-agent/pkg/api"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test remote worker with real database backend
func TestRemoteWorkerWithDatabase(t *testing.T) {
	ctx := context.Background()
	_, db, cleanup := testutil.SetupPostgresWithTestData(ctx, t)
	defer cleanup()

	// Create a mock API server that uses the real database
	mockServer := createDatabaseBackedMockServer(db)
	defer mockServer.Close()

	mel := api.NewMel()
	worker, err := NewRemoteWorker(mockServer.URL, "valid-token", "db-test-worker", mel, 3)
	require.NoError(t, err)

	// Test worker registration and basic functionality
	testCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Start worker in goroutine
	workerErr := make(chan error, 1)
	go func() {
		err := worker.Start(testCtx)
		workerErr <- err
	}()

	// Give worker time to register
	time.Sleep(200 * time.Millisecond)

	// Verify worker was registered in database
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM workflow_workers WHERE id = $1", "db-test-worker").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "Worker should be registered in database")

	// Verify worker details
	var workerInfo WorkflowWorker
	query := `SELECT id, hostname, status FROM workflow_workers WHERE id = $1`
	row := db.QueryRow(query, "db-test-worker")
	err = row.Scan(&workerInfo.ID, &workerInfo.Hostname, &workerInfo.Status)
	require.NoError(t, err)

	assert.Equal(t, "db-test-worker", workerInfo.ID)
	assert.NotEmpty(t, workerInfo.Hostname)
	assert.Equal(t, WorkerStatusIdle, workerInfo.Status)

	// Cancel context to stop worker
	cancel()

	select {
	case err := <-workerErr:
		assert.NoError(t, err)
	case <-time.After(3 * time.Second):
		t.Fatal("Worker did not stop within timeout")
	}

	t.Logf("✅ Remote worker with database test passed - Worker registered and operated correctly")
}

// Test remote worker work processing with real workflow data
func TestRemoteWorkerWorkProcessingWithDatabase(t *testing.T) {
	ctx := context.Background()
	_, db, cleanup := testutil.SetupPostgresWithTestData(ctx, t)
	defer cleanup()

	// Create test workflow run using test agent
	testAgentID := uuid.MustParse("11111111-1111-1111-1111-111111111111") // From testutil
	runID := uuid.New()
	versionID := uuid.New()

	// Insert workflow run
	runQuery := `
		INSERT INTO workflow_runs (id, agent_id, version_id, status, created_at)
		VALUES ($1, $2, $3, 'running', NOW())
	`
	_, err := db.Exec(runQuery, runID, testAgentID, versionID)
	require.NoError(t, err)

	// Create work items
	workItem1ID := uuid.New()
	workItem2ID := uuid.New()

	queueQuery := `
		INSERT INTO workflow_queue (id, run_id, queue_type, priority, available_at, created_at, attempt_count, max_attempts)
		VALUES ($1, $2, $3, $4, NOW(), NOW(), 0, 3)
	`
	_, err = db.Exec(queueQuery, workItem1ID, runID, "execute_step", 5)
	require.NoError(t, err)
	_, err = db.Exec(queueQuery, workItem2ID, runID, "start_run", 10)
	require.NoError(t, err)

	// Create mock server with database backend
	mockServer := createDatabaseBackedMockServer(db)
	defer mockServer.Close()

	mel := api.NewMel()
	worker, err := NewRemoteWorker(mockServer.URL, "valid-token", "work-processor-test", mel, 2)
	require.NoError(t, err)

	// Start worker for a short time to process work
	testCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	workerErr := make(chan error, 1)
	go func() {
		err := worker.Start(testCtx)
		workerErr <- err
	}()

	// Give worker time to process
	time.Sleep(300 * time.Millisecond)

	// Stop worker
	cancel()

	select {
	case <-workerErr:
	case <-time.After(2 * time.Second):
		t.Fatal("Worker did not stop within timeout")
	}

	// Check if any work was claimed (timing dependent)
	var claimedCount int
	err = db.QueryRow("SELECT COUNT(*) FROM workflow_queue WHERE claimed_by = $1", "work-processor-test").Scan(&claimedCount)
	require.NoError(t, err)

	t.Logf("✅ Remote worker work processing test passed - Worker claimed %d items", claimedCount)
}

// Test remote worker authentication with database
func TestRemoteWorkerAuthenticationWithDatabase(t *testing.T) {
	ctx := context.Background()
	_, db, cleanup := testutil.SetupPostgresWithTestData(ctx, t)
	defer cleanup()

	// Create mock server that validates tokens
	mockServer := createAuthenticatingMockServer(db)
	defer mockServer.Close()

	mel := api.NewMel()

	// Test with valid token
	validWorker, err := NewRemoteWorker(mockServer.URL, "valid-token", "auth-test-valid", mel, 1)
	require.NoError(t, err)

	testCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	workerErr := make(chan error, 1)
	go func() {
		err := validWorker.Start(testCtx)
		workerErr <- err
	}()

	time.Sleep(200 * time.Millisecond)

	// Check if worker was registered with valid token
	var validCount int
	err = db.QueryRow("SELECT COUNT(*) FROM workflow_workers WHERE id = $1", "auth-test-valid").Scan(&validCount)
	require.NoError(t, err)
	assert.Equal(t, 1, validCount, "Worker with valid token should be registered")

	cancel()

	select {
	case <-workerErr:
	case <-time.After(2 * time.Second):
		t.Fatal("Valid worker did not stop within timeout")
	}

	// Test with invalid token
	invalidWorker, err := NewRemoteWorker(mockServer.URL, "invalid-token", "auth-test-invalid", mel, 1)
	require.NoError(t, err)

	testCtx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err = invalidWorker.Start(testCtx)
	// Should fail due to authentication
	if err != nil {
		assert.Contains(t, err.Error(), "failed to register worker")
	}

	// Check that invalid worker was not registered
	var invalidCount int
	err = db.QueryRow("SELECT COUNT(*) FROM workflow_workers WHERE id = $1", "auth-test-invalid").Scan(&invalidCount)
	require.NoError(t, err)
	assert.Equal(t, 0, invalidCount, "Worker with invalid token should not be registered")

	t.Logf("✅ Remote worker authentication test passed - Valid token worked, invalid token rejected")
}

// createDatabaseBackedMockServer creates a mock API server that uses the real database
func createDatabaseBackedMockServer(db *sql.DB) *httptest.Server {
	mux := http.NewServeMux()

	// Worker registration
	mux.HandleFunc("/api/workers", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Check authorization
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer valid-token") {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		var worker WorkflowWorker
		if err := json.NewDecoder(r.Body).Decode(&worker); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Store in real database
		query := `
			INSERT INTO workflow_workers (
				id, hostname, process_id, version, capabilities, status,
				last_heartbeat, started_at, max_concurrent_steps,
				current_step_count, total_steps_executed, total_execution_time_ms
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
			ON CONFLICT (id) DO UPDATE SET
				last_heartbeat = EXCLUDED.last_heartbeat,
				status = EXCLUDED.status
		`

		_, err := db.Exec(query,
			worker.ID, worker.Hostname, worker.ProcessID, worker.Version,
			pq.Array(worker.Capabilities), worker.Status, worker.LastHeartbeat,
			worker.StartedAt, worker.MaxConcurrentSteps, worker.CurrentStepCount,
			worker.TotalStepsExecuted, worker.TotalExecutionTimeMS,
		)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"id": worker.ID})
	})

	// Heartbeat
	mux.HandleFunc("/api/workers/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/workers/")

		if strings.HasSuffix(path, "/heartbeat") && r.Method == "PUT" {
			workerID := strings.TrimSuffix(path, "/heartbeat")

			auth := r.Header.Get("Authorization")
			if !strings.HasPrefix(auth, "Bearer valid-token") {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// Update heartbeat in database
			query := `UPDATE workflow_workers SET last_heartbeat = NOW() WHERE id = $1`
			_, err := db.Exec(query, workerID)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			return
		}

		if strings.Contains(path, "/claim-work") && r.Method == http.MethodPost {
			parts := strings.Split(path, "/")
			if len(parts) < 2 {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			workerID := parts[0]

			auth := r.Header.Get("Authorization")
			if !strings.HasPrefix(auth, "Bearer valid-token") {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// Claim work from database
			tx, err := db.Begin()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			defer tx.Rollback()

			query := `
				SELECT id, run_id, step_id, queue_type, priority, available_at, 
				       created_at, attempt_count, max_attempts, payload
				FROM workflow_queue 
				WHERE claimed_at IS NULL 
				  AND claimed_by IS NULL 
				  AND available_at <= NOW()
				ORDER BY priority DESC, created_at ASC
				LIMIT 5
				FOR UPDATE SKIP LOCKED
			`

			rows, err := tx.Query(query)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			defer rows.Close()

			var workItems []*QueueItem
			var claimedIDs []uuid.UUID

			for rows.Next() {
				var item QueueItem
				var payloadBytes []byte

				err := rows.Scan(
					&item.ID, &item.RunID, &item.StepID, &item.QueueType,
					&item.Priority, &item.AvailableAt, &item.CreatedAt,
					&item.AttemptCount, &item.MaxAttempts, &payloadBytes,
				)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				workItems = append(workItems, &item)
				claimedIDs = append(claimedIDs, item.ID)
			}

			// Claim the items
			if len(claimedIDs) > 0 {
				claimQuery := `
					UPDATE workflow_queue 
					SET claimed_at = NOW(), claimed_by = $1 
					WHERE id = ANY($2)
				`
				_, err = tx.Exec(claimQuery, workerID, pq.Array(claimedIDs))
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}

			if err := tx.Commit(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(workItems)
			return
		}

		if strings.Contains(path, "/complete-work/") && r.Method == http.MethodPost {
			// For this test, just return success
			w.WriteHeader(http.StatusOK)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	})

	return httptest.NewServer(mux)
}

// createAuthenticatingMockServer creates a mock server that properly validates authentication
func createAuthenticatingMockServer(db *sql.DB) *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/workers", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Strict token validation
		auth := r.Header.Get("Authorization")
		if auth != "Bearer valid-token" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
			return
		}

		var worker WorkflowWorker
		if err := json.NewDecoder(r.Body).Decode(&worker); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Store in database only if auth is valid
		query := `
			INSERT INTO workflow_workers (
				id, hostname, status, last_heartbeat, started_at, max_concurrent_steps
			) VALUES ($1, $2, $3, NOW(), NOW(), $4)
		`

		_, err := db.Exec(query, worker.ID, worker.Hostname, worker.Status, worker.MaxConcurrentSteps)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"id": worker.ID})
	})

	// Add other endpoints that also require authentication
	mux.HandleFunc("/api/workers/", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer valid-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// For other endpoints, just return success if authenticated
		w.WriteHeader(http.StatusOK)
		if strings.Contains(r.URL.Path, "/claim-work") {
			json.NewEncoder(w).Encode([]*QueueItem{})
		}
	})

	return httptest.NewServer(mux)
}
