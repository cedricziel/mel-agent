package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cedricziel/mel-agent/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOpenAPIRegisterWorker tests registering a new worker
func TestOpenAPIRegisterWorker(t *testing.T) {
	db, mockEngine, cleanup := testutil.SetupOpenAPITestDB(t)
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	registerReq := RegisterWorkerRequest{
		Id:          "worker-test-001",
		Name:        testutil.StringPtr("Test Worker"),
		Concurrency: testutil.IntPtr(10),
	}

	reqBody, err := json.Marshal(registerReq)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/workers", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Logf("Response body: %s", w.Body.String())
	}
	assert.Equal(t, http.StatusCreated, w.Code)

	var response Worker
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "worker-test-001", *response.Id)
	assert.Equal(t, "Test Worker", *response.Name)
	assert.Equal(t, Active, *response.Status)
	assert.Equal(t, 10, *response.Concurrency)
	assert.NotNil(t, response.LastHeartbeat)
	assert.NotNil(t, response.RegisteredAt)

	// Verify the worker was actually created in the database
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM workflow_workers WHERE id = $1", response.Id).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

// TestOpenAPIRegisterWorkerMinimal tests registering a worker with minimal data
func TestOpenAPIRegisterWorkerMinimal(t *testing.T) {
	db, mockEngine, cleanup := testutil.SetupOpenAPITestDB(t)
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	registerReq := RegisterWorkerRequest{
		Id: "worker-minimal-001",
	}

	reqBody, err := json.Marshal(registerReq)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/workers", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Logf("Response body: %s", w.Body.String())
	}
	assert.Equal(t, http.StatusCreated, w.Code)

	var response Worker
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "worker-minimal-001", *response.Id)
	assert.Equal(t, "worker-minimal-001", *response.Name) // Should default to ID
	assert.Equal(t, Active, *response.Status)
	assert.Equal(t, 5, *response.Concurrency) // Default concurrency
}

// TestOpenAPIRegisterWorkerUpdate tests updating an existing worker via re-registration
func TestOpenAPIRegisterWorkerUpdate(t *testing.T) {
	db, mockEngine, cleanup := testutil.SetupOpenAPITestDB(t)
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	// First registration
	registerReq := RegisterWorkerRequest{
		Id:          "worker-update-001",
		Name:        testutil.StringPtr("Original Worker"),
		Concurrency: testutil.IntPtr(5),
	}

	reqBody, _ := json.Marshal(registerReq)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/workers", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	// Second registration (update)
	updateReq := RegisterWorkerRequest{
		Id:          "worker-update-001",
		Name:        testutil.StringPtr("Updated Worker"),
		Concurrency: testutil.IntPtr(15),
	}

	reqBody, _ = json.Marshal(updateReq)
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/workers", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response Worker
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "worker-update-001", *response.Id)
	assert.Equal(t, "Updated Worker", *response.Name)
	assert.Equal(t, 15, *response.Concurrency)

	// Verify only one worker exists in database
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM workflow_workers WHERE id = $1", response.Id).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

// TestOpenAPIListWorkers tests listing workers
func TestOpenAPIListWorkers(t *testing.T) {
	db, mockEngine, cleanup := testutil.SetupOpenAPITestDB(t)
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	// Create test workers
	testWorkers := []string{"worker-list-001", "worker-list-002", "worker-list-003"}
	for i, workerId := range testWorkers {
		registerReq := RegisterWorkerRequest{
			Id:          workerId,
			Name:        testutil.StringPtr(fmt.Sprintf("Test Worker %d", i+1)),
			Concurrency: testutil.IntPtr((i + 1) * 5),
		}
		reqBody, _ := json.Marshal(registerReq)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/workers", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)
	}

	// List workers
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/workers", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []Worker
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Len(t, response, 3)

	// Check that we have all the expected workers
	workerIds := make(map[string]bool)
	for _, worker := range response {
		workerIds[*worker.Id] = true
	}

	for _, expectedId := range testWorkers {
		assert.True(t, workerIds[expectedId], "Should find worker %s", expectedId)
	}
}

// TestOpenAPIUpdateWorkerHeartbeat tests updating a worker's heartbeat
func TestOpenAPIUpdateWorkerHeartbeat(t *testing.T) {
	db, mockEngine, cleanup := testutil.SetupOpenAPITestDB(t)
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	// Register a worker first
	registerReq := RegisterWorkerRequest{
		Id:          "worker-heartbeat-001",
		Name:        testutil.StringPtr("Heartbeat Worker"),
		Concurrency: testutil.IntPtr(8),
	}
	reqBody, _ := json.Marshal(registerReq)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/workers", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	// Update heartbeat
	w = httptest.NewRecorder()
	req = httptest.NewRequest("PUT", "/api/workers/worker-heartbeat-001/heartbeat", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify the worker's status is active in the database
	var status string
	err := db.QueryRow("SELECT status FROM workflow_workers WHERE id = $1", "worker-heartbeat-001").Scan(&status)
	require.NoError(t, err)
	assert.Equal(t, "active", status)
}

// TestOpenAPIUpdateWorkerHeartbeatNotFound tests updating heartbeat for non-existent worker
func TestOpenAPIUpdateWorkerHeartbeatNotFound(t *testing.T) {
	db, mockEngine, cleanup := testutil.SetupOpenAPITestDB(t)
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/api/workers/non-existent-worker/heartbeat", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response Error
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "not found", *response.Error)
}

// TestOpenAPIUnregisterWorker tests unregistering a worker
func TestOpenAPIUnregisterWorker(t *testing.T) {
	db, mockEngine, cleanup := testutil.SetupOpenAPITestDB(t)
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	// Register a worker first
	registerReq := RegisterWorkerRequest{
		Id:          "worker-unregister-001",
		Name:        testutil.StringPtr("Unregister Worker"),
		Concurrency: testutil.IntPtr(6),
	}
	reqBody, _ := json.Marshal(registerReq)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/workers", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	// Unregister the worker
	w = httptest.NewRecorder()
	req = httptest.NewRequest("DELETE", "/api/workers/worker-unregister-001", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	// Verify the worker was deleted from the database
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM workflow_workers WHERE id = $1", "worker-unregister-001").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

// TestOpenAPIUnregisterWorkerNotFound tests unregistering a non-existent worker
func TestOpenAPIUnregisterWorkerNotFound(t *testing.T) {
	db, mockEngine, cleanup := testutil.SetupOpenAPITestDB(t)
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/api/workers/non-existent-worker", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response Error
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "not found", *response.Error)
}

// TestOpenAPIClaimWork tests claiming work items
func TestOpenAPIClaimWork(t *testing.T) {
	db, mockEngine, cleanup := testutil.SetupOpenAPITestDB(t)
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	// Register a worker first
	registerReq := RegisterWorkerRequest{
		Id:          "worker-claim-001",
		Name:        testutil.StringPtr("Claim Worker"),
		Concurrency: testutil.IntPtr(4),
	}
	reqBody, _ := json.Marshal(registerReq)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/workers", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	// Claim work
	claimReq := ClaimWorkJSONBody{
		MaxItems: testutil.IntPtr(5),
	}
	reqBody, _ = json.Marshal(claimReq)

	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/workers/worker-claim-001/claim-work", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []WorkItem
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	// For now, should return empty array since work queue is not implemented
	assert.Len(t, response, 0)
}

// TestOpenAPICompleteWork tests completing work items
func TestOpenAPICompleteWork(t *testing.T) {
	db, mockEngine, cleanup := testutil.SetupOpenAPITestDB(t)
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	// Register a worker first
	registerReq := RegisterWorkerRequest{
		Id:          "worker-complete-001",
		Name:        testutil.StringPtr("Complete Worker"),
		Concurrency: testutil.IntPtr(3),
	}
	reqBody, _ := json.Marshal(registerReq)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/workers", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	// Complete work
	completeReq := CompleteWorkJSONBody{
		Result: &map[string]interface{}{
			"status": "success",
			"output": "Work completed successfully",
		},
	}
	reqBody, _ = json.Marshal(completeReq)

	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/workers/worker-complete-001/complete-work/work-item-123", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
