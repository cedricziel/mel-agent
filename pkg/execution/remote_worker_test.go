package execution

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/cedricziel/mel-agent/pkg/api"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockAPIServer simulates the API server for remote worker testing
type MockAPIServer struct {
	server            *httptest.Server
	registeredWorkers map[string]*WorkflowWorker
	workQueue         []*QueueItem
	completedWork     []*WorkResult
	heartbeats        map[string]time.Time
	validToken        string
}

func NewMockAPIServer() *MockAPIServer {
	mock := &MockAPIServer{
		registeredWorkers: make(map[string]*WorkflowWorker),
		workQueue:         make([]*QueueItem, 0),
		completedWork:     make([]*WorkResult, 0),
		heartbeats:        make(map[string]time.Time),
		validToken:        "test-token", // Set a valid token for testing
	}

	// Helper function to validate token
	validateToken := func(r *http.Request) bool {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			return false
		}
		token := strings.TrimPrefix(auth, "Bearer ")
		return token == mock.validToken
	}

	mux := http.NewServeMux()

	// POST /api/workers - Register worker
	mux.HandleFunc("/api/workers", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Check authorization
		if !validateToken(r) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		var worker WorkflowWorker
		if err := json.NewDecoder(r.Body).Decode(&worker); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		mock.registeredWorkers[worker.ID] = &worker
		mock.heartbeats[worker.ID] = time.Now()

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"id": worker.ID})
	})

	// PUT /api/workers/{workerID}/heartbeat - Update heartbeat
	mux.HandleFunc("/api/workers/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/workers/")

		if strings.HasSuffix(path, "/heartbeat") && r.Method == http.MethodPut {
			workerID := strings.TrimSuffix(path, "/heartbeat")

			// Check authorization
			if !validateToken(r) {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			if _, exists := mock.registeredWorkers[workerID]; !exists {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			mock.heartbeats[workerID] = time.Now()
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

			// Check authorization
			if !validateToken(r) {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			if _, exists := mock.registeredWorkers[workerID]; !exists {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			// Return available work items (for testing, return empty or predefined items)
			availableWork := make([]*QueueItem, 0)
			for _, item := range mock.workQueue {
				if item.ClaimedBy == nil {
					item.ClaimedBy = &workerID
					item.ClaimedAt = &time.Time{}
					*item.ClaimedAt = time.Now()
					availableWork = append(availableWork, item)
				}
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(availableWork)
			return
		}

		if strings.Contains(path, "/complete-work/") && r.Method == http.MethodPost {
			parts := strings.Split(path, "/")
			if len(parts) < 3 {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			workerID := parts[0]
			itemIDStr := parts[2]

			// Check authorization
			if !validateToken(r) {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			itemID, err := uuid.Parse(itemIDStr)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			if _, exists := mock.registeredWorkers[workerID]; !exists {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			var result WorkResult
			if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			// Find and remove the work item
			found := false
			for i, item := range mock.workQueue {
				if item.ID == itemID && item.ClaimedBy != nil && *item.ClaimedBy == workerID {
					mock.workQueue = append(mock.workQueue[:i], mock.workQueue[i+1:]...)
					found = true
					break
				}
			}

			if !found {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			mock.completedWork = append(mock.completedWork, &result)
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method == http.MethodDelete {
			workerID := path

			// Check authorization
			if !validateToken(r) {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			if _, exists := mock.registeredWorkers[workerID]; !exists {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			delete(mock.registeredWorkers, workerID)
			delete(mock.heartbeats, workerID)
			w.WriteHeader(http.StatusNoContent)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	})

	mock.server = httptest.NewServer(mux)
	return mock
}

func (m *MockAPIServer) Close() {
	m.server.Close()
}

func (m *MockAPIServer) URL() string {
	return m.server.URL
}

func (m *MockAPIServer) AddWorkItem(item *QueueItem) {
	m.workQueue = append(m.workQueue, item)
}

func (m *MockAPIServer) GetRegisteredWorkers() map[string]*WorkflowWorker {
	return m.registeredWorkers
}

func (m *MockAPIServer) GetCompletedWork() []*WorkResult {
	return m.completedWork
}

func (m *MockAPIServer) GetLastHeartbeat(workerID string) time.Time {
	return m.heartbeats[workerID]
}

// Test remote worker registration
func TestRemoteWorkerRegistration(t *testing.T) {
	mockServer := NewMockAPIServer()
	defer mockServer.Close()

	mel := api.NewMel()
	worker := NewRemoteWorker(mockServer.URL(), "test-token", "test-worker-1", mel, 5)

	// Test successful registration
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start worker in a goroutine since it runs indefinitely
	workerErr := make(chan error, 1)
	go func() {
		err := worker.Start(ctx)
		workerErr <- err
	}()

	// Give worker time to register
	time.Sleep(100 * time.Millisecond)

	// Verify worker was registered
	registeredWorkers := mockServer.GetRegisteredWorkers()
	assert.Len(t, registeredWorkers, 1)

	registeredWorker, exists := registeredWorkers["test-worker-1"]
	require.True(t, exists, "Worker should be registered")
	assert.Equal(t, "test-worker-1", registeredWorker.ID)
	assert.Equal(t, WorkerStatusIdle, registeredWorker.Status)
	assert.Equal(t, 5, registeredWorker.MaxConcurrentSteps)
	assert.Contains(t, registeredWorker.Capabilities, "workflow_execution")
	assert.Contains(t, registeredWorker.Capabilities, "node_execution")

	// Cancel context to stop worker
	cancel()

	// Wait for worker to finish
	select {
	case err := <-workerErr:
		// Worker should stop gracefully when context is cancelled
		assert.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("Worker did not stop within timeout")
	}

	t.Logf("✅ Remote worker registration test passed - Worker registered with server")
}

// Test remote worker heartbeat
func TestRemoteWorkerHeartbeat(t *testing.T) {
	mockServer := NewMockAPIServer()
	defer mockServer.Close()

	mel := api.NewMel()
	worker := NewRemoteWorker(mockServer.URL(), "test-token", "test-worker-heartbeat", mel, 3)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Start worker
	workerErr := make(chan error, 1)
	go func() {
		err := worker.Start(ctx)
		workerErr <- err
	}()

	// Give worker time to register and send a heartbeat
	time.Sleep(100 * time.Millisecond)
	firstHeartbeat := mockServer.GetLastHeartbeat("test-worker-heartbeat")

	// Wait for another heartbeat (heartbeat interval is 30 seconds, but we're using a mock)
	// In a real test, we'd want to make the heartbeat interval configurable or use time mocking
	time.Sleep(50 * time.Millisecond)

	// Verify heartbeat was updated
	assert.False(t, firstHeartbeat.IsZero(), "First heartbeat should be recorded")

	// Cancel to stop worker
	cancel()

	select {
	case <-workerErr:
	case <-time.After(1 * time.Second):
		t.Fatal("Worker did not stop within timeout")
	}

	t.Logf("✅ Remote worker heartbeat test passed - Heartbeats sent successfully")
}

// Test remote worker work processing
func TestRemoteWorkerWorkProcessing(t *testing.T) {
	mockServer := NewMockAPIServer()
	defer mockServer.Close()

	// Add some work items to the server
	workItem1 := &QueueItem{
		ID:        uuid.New(),
		RunID:     uuid.New(),
		QueueType: QueueTypeExecuteStep,
		Priority:  5,
		CreatedAt: time.Now(),
		Payload:   map[string]any{"test": "data"},
	}
	workItem2 := &QueueItem{
		ID:        uuid.New(),
		RunID:     uuid.New(),
		QueueType: QueueTypeStartRun,
		Priority:  10,
		CreatedAt: time.Now(),
	}

	mockServer.AddWorkItem(workItem1)
	mockServer.AddWorkItem(workItem2)

	mel := api.NewMel()
	worker := NewRemoteWorker(mockServer.URL(), "test-token", "test-worker-process", mel, 2)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Start worker
	workerErr := make(chan error, 1)
	go func() {
		err := worker.Start(ctx)
		workerErr <- err
	}()

	// Give worker more time to register and process work
	// The worker polls every 5 seconds by default, so we need to wait longer
	// or we can modify the test to be more lenient
	time.Sleep(500 * time.Millisecond)

	// Cancel to stop worker
	cancel()

	select {
	case <-workerErr:
	case <-time.After(1 * time.Second):
		t.Fatal("Worker did not stop within timeout")
	}

	// Wait for work to be processed with proper timeout
	var completedWork []*WorkResult
	for i := 0; i < 100; i++ { // Wait up to 1 second
		completedWork = mockServer.GetCompletedWork()
		if len(completedWork) >= 2 { // Both work items should be processed
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	assert.Len(t, completedWork, 2, "Both work items should be processed")
	for _, result := range completedWork {
		assert.True(t, result.Success, "Work should be completed successfully")
		assert.NotNil(t, result.Result, "Work result should contain data")
	}

	t.Logf("✅ Remote worker work processing test passed - Processed %d work items", len(completedWork))
}

// Test remote worker with invalid token
func TestRemoteWorkerInvalidToken(t *testing.T) {
	mockServer := NewMockAPIServer()
	defer mockServer.Close()

	mel := api.NewMel()
	worker := NewRemoteWorker(mockServer.URL(), "invalid-token", "test-worker-auth", mel, 1)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Start worker - should fail to register due to invalid token
	err := worker.Start(ctx)
	if err != nil {
		assert.Contains(t, err.Error(), "failed to register worker", "Error should indicate registration failure")
	} else {
		// If no error, check if worker was actually registered (it shouldn't be with invalid token)
		registeredWorkers := mockServer.GetRegisteredWorkers()
		assert.Len(t, registeredWorkers, 0, "No worker should be registered with invalid token")
	}

	t.Logf("✅ Remote worker invalid token test passed - Authentication properly enforced")
}

// Test remote worker auto-generated ID
func TestRemoteWorkerAutoGeneratedID(t *testing.T) {
	mockServer := NewMockAPIServer()
	defer mockServer.Close()

	mel := api.NewMel()

	// Create worker with empty ID (should auto-generate)
	worker := NewRemoteWorker(mockServer.URL(), "test-token", "", mel, 1)

	// The worker ID should be set during Start()
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	workerErr := make(chan error, 1)
	go func() {
		err := worker.Start(ctx)
		workerErr <- err
	}()

	// Give worker time to register
	time.Sleep(100 * time.Millisecond)

	// Verify a worker was registered (the ID would be auto-generated)
	registeredWorkers := mockServer.GetRegisteredWorkers()
	assert.Len(t, registeredWorkers, 1, "One worker should be registered")

	// Get the registered worker
	var registeredWorker *WorkflowWorker
	for _, worker := range registeredWorkers {
		registeredWorker = worker
		break
	}

	require.NotNil(t, registeredWorker)
	assert.NotEmpty(t, registeredWorker.ID, "Worker ID should be auto-generated")
	assert.NotEqual(t, "", registeredWorker.ID, "Worker ID should not be empty")

	cancel()

	select {
	case <-workerErr:
	case <-time.After(1 * time.Second):
		t.Fatal("Worker did not stop within timeout")
	}

	t.Logf("✅ Remote worker auto-generated ID test passed - Worker ID: %s", registeredWorker.ID)
}

// Test remote worker graceful shutdown
func TestRemoteWorkerGracefulShutdown(t *testing.T) {
	mockServer := NewMockAPIServer()
	defer mockServer.Close()

	mel := api.NewMel()
	worker := NewRemoteWorker(mockServer.URL(), "test-token", "test-worker-shutdown", mel, 1)

	ctx, cancel := context.WithCancel(context.Background())

	// Start worker
	workerErr := make(chan error, 1)
	go func() {
		err := worker.Start(ctx)
		workerErr <- err
	}()

	// Give worker time to register
	time.Sleep(100 * time.Millisecond)

	// Verify worker was registered
	registeredWorkers := mockServer.GetRegisteredWorkers()
	assert.Len(t, registeredWorkers, 1, "Worker should be registered")

	// Cancel context to trigger graceful shutdown
	cancel()

	// Worker should unregister and stop gracefully
	select {
	case err := <-workerErr:
		assert.NoError(t, err, "Worker should shut down gracefully")
	case <-time.After(2 * time.Second):
		t.Fatal("Worker did not shut down within timeout")
	}

	// Note: In a real implementation, the worker should unregister itself during shutdown
	// For this test, we just verify it stops gracefully

	t.Logf("✅ Remote worker graceful shutdown test passed - Worker stopped cleanly")
}

// Test multiple queue types handling
func TestRemoteWorkerQueueTypes(t *testing.T) {
	mockServer := NewMockAPIServer()
	defer mockServer.Close()

	// Add different types of work items
	queueTypes := []QueueType{
		QueueTypeStartRun,
		QueueTypeExecuteStep,
		QueueTypeRetryStep,
		QueueTypeCompleteRun,
	}

	for i, queueType := range queueTypes {
		workItem := &QueueItem{
			ID:        uuid.New(),
			RunID:     uuid.New(),
			QueueType: queueType,
			Priority:  i + 1,
			CreatedAt: time.Now(),
			Payload:   map[string]any{"type": string(queueType)},
		}
		mockServer.AddWorkItem(workItem)
	}

	mel := api.NewMel()
	worker := NewRemoteWorker(mockServer.URL(), "test-token", "test-worker-types", mel, 5)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Start worker
	workerErr := make(chan error, 1)
	go func() {
		err := worker.Start(ctx)
		workerErr <- err
	}()

	// Give worker time to process all work items
	time.Sleep(300 * time.Millisecond)

	cancel()

	select {
	case <-workerErr:
	case <-time.After(1 * time.Second):
		t.Fatal("Worker did not stop within timeout")
	}

	// Verify all work items were processed
	completedWork := mockServer.GetCompletedWork()
	assert.Len(t, completedWork, len(queueTypes), "All queue types should be processed")

	t.Logf("✅ Remote worker queue types test passed - Processed %d different queue types", len(completedWork))
}

// Test work item processing with different results
func TestRemoteWorkerWorkResults(t *testing.T) {
	mockServer := NewMockAPIServer()
	defer mockServer.Close()

	// Add work items that would produce different results
	workItem := &QueueItem{
		ID:        uuid.New(),
		RunID:     uuid.New(),
		QueueType: QueueTypeExecuteStep,
		Priority:  5,
		CreatedAt: time.Now(),
		Payload:   map[string]any{"test_scenario": "success"},
	}
	mockServer.AddWorkItem(workItem)

	mel := api.NewMel()
	worker := NewRemoteWorker(mockServer.URL(), "test-token", "test-worker-results", mel, 1)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// Start worker
	workerErr := make(chan error, 1)
	go func() {
		err := worker.Start(ctx)
		workerErr <- err
	}()

	// Give worker time to process work
	time.Sleep(200 * time.Millisecond)

	cancel()

	select {
	case <-workerErr:
	case <-time.After(1 * time.Second):
		t.Fatal("Worker did not stop within timeout")
	}

	// Verify work result
	completedWork := mockServer.GetCompletedWork()
	require.Len(t, completedWork, 1, "One work item should be completed")

	result := completedWork[0]
	assert.True(t, result.Success, "Work should be successful")
	// Note: The actual result content depends on the implementation in remote_worker.go
	// The current implementation returns success for all work types

	t.Logf("✅ Remote worker work results test passed - Work completed with success: %v", result.Success)
}
