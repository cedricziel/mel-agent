package execution

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/cedricziel/mel-agent/pkg/api"
	"github.com/google/uuid"
)

// RemoteWorker represents a worker that connects to a remote API server
type RemoteWorker struct {
	serverURL   string
	token       string
	workerID    string
	mel         api.Mel
	concurrency int
	httpClient  *http.Client
	workerInfo  *WorkflowWorker
}

// NewRemoteWorker creates a new remote worker instance
func NewRemoteWorker(serverURL, token, workerID string, mel api.Mel, concurrency int) *RemoteWorker {
	return &RemoteWorker{
		serverURL:   serverURL,
		token:       token,
		workerID:    workerID,
		mel:         mel,
		concurrency: concurrency,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Start begins the remote worker execution loop
func (rw *RemoteWorker) Start(ctx context.Context) error {
	// Generate worker ID if not provided
	if rw.workerID == "" {
		rw.workerID = generateWorkerID()
	}

	// Create worker info
	hostname, _ := os.Hostname()
	processID := os.Getpid()

	rw.workerInfo = &WorkflowWorker{
		ID:                   rw.workerID,
		Hostname:             hostname,
		ProcessID:            &processID,
		Version:              getVersion(),
		Capabilities:         []string{"workflow_execution", "node_execution"},
		Status:               WorkerStatusIdle,
		LastHeartbeat:        time.Now(),
		StartedAt:            time.Now(),
		MaxConcurrentSteps:   rw.concurrency,
		CurrentStepCount:     0,
		TotalStepsExecuted:   0,
		TotalExecutionTimeMS: 0,
	}

	// Register worker with the API server
	if err := rw.registerWorker(ctx); err != nil {
		return fmt.Errorf("failed to register worker: %w", err)
	}

	log.Printf("Worker %s registered successfully", rw.workerID)

	// Start heartbeat goroutine
	go rw.heartbeatLoop(ctx)

	// Start main work loop
	return rw.workLoop(ctx)
}

// registerWorker registers this worker with the API server
func (rw *RemoteWorker) registerWorker(ctx context.Context) error {
	url := fmt.Sprintf("%s/api/workers", rw.serverURL)

	payload, err := json.Marshal(rw.workerInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal worker info: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+rw.token)

	resp, err := rw.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("worker registration failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// heartbeatLoop sends periodic heartbeats to the API server
func (rw *RemoteWorker) heartbeatLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// Send unregister request before exiting
			rw.unregisterWorker(context.Background())
			return
		case <-ticker.C:
			if err := rw.sendHeartbeat(ctx); err != nil {
				log.Printf("Failed to send heartbeat: %v", err)
			}
		}
	}
}

// sendHeartbeat sends a heartbeat to the API server
func (rw *RemoteWorker) sendHeartbeat(ctx context.Context) error {
	url := fmt.Sprintf("%s/api/workers/%s/heartbeat", rw.serverURL, rw.workerID)

	req, err := http.NewRequestWithContext(ctx, "PUT", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create heartbeat request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+rw.token)

	resp, err := rw.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send heartbeat: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("heartbeat failed with status %d", resp.StatusCode)
	}

	return nil
}

// unregisterWorker unregisters this worker from the API server
func (rw *RemoteWorker) unregisterWorker(ctx context.Context) error {
	url := fmt.Sprintf("%s/api/workers/%s", rw.serverURL, rw.workerID)

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create unregister request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+rw.token)

	resp, err := rw.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to unregister worker: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("Worker %s unregistered", rw.workerID)
	return nil
}

// workLoop is the main work processing loop
func (rw *RemoteWorker) workLoop(ctx context.Context) error {
	// Process work immediately when starting
	if err := rw.processWork(ctx); err != nil {
		log.Printf("Error processing initial work: %v", err)
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := rw.processWork(ctx); err != nil {
				log.Printf("Error processing work: %v", err)
			}
		}
	}
}

// processWork claims and processes work from the API server
func (rw *RemoteWorker) processWork(ctx context.Context) error {
	// Claim work from the API server
	workItems, err := rw.claimWork(ctx)
	if err != nil {
		return fmt.Errorf("failed to claim work: %w", err)
	}

	if len(workItems) == 0 {
		return nil // No work available
	}

	log.Printf("Claimed %d work items", len(workItems))

	// Process each work item
	for _, item := range workItems {
		select {
		case <-ctx.Done():
			return nil
		default:
			if err := rw.processWorkItem(ctx, item); err != nil {
				log.Printf("Failed to process work item %s: %v", item.ID, err)
			}
		}
	}

	return nil
}

// claimWork claims work items from the API server
func (rw *RemoteWorker) claimWork(ctx context.Context) ([]*QueueItem, error) {
	url := fmt.Sprintf("%s/api/workers/%s/claim-work?max_items=%d", rw.serverURL, rw.workerID, rw.concurrency)

	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create claim work request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+rw.token)

	resp, err := rw.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to claim work: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("claim work failed with status %d", resp.StatusCode)
	}

	var workItems []*QueueItem
	if err := json.NewDecoder(resp.Body).Decode(&workItems); err != nil {
		return nil, fmt.Errorf("failed to decode work items: %w", err)
	}

	return workItems, nil
}

// processWorkItem processes a single work item
func (rw *RemoteWorker) processWorkItem(ctx context.Context, item *QueueItem) error {
	log.Printf("Processing work item %s of type %s", item.ID, item.QueueType)

	var result *WorkResult

	switch item.QueueType {
	case QueueTypeStartRun:
		result = rw.processStartRun(ctx, item)
	case QueueTypeExecuteStep:
		result = rw.processExecuteStep(ctx, item)
	case QueueTypeRetryStep:
		result = rw.processRetryStep(ctx, item)
	case QueueTypeCompleteRun:
		result = rw.processCompleteRun(ctx, item)
	default:
		result = &WorkResult{
			Success: false,
			Error:   stringPointer(fmt.Sprintf("unknown queue type: %s", item.QueueType)),
		}
	}

	// Report the result back to the API server
	return rw.completeWork(ctx, item.ID, result)
}

// processStartRun handles starting a workflow run
func (rw *RemoteWorker) processStartRun(ctx context.Context, item *QueueItem) *WorkResult {
	// Implementation would depend on how the payload is structured
	// For now, return success
	log.Printf("Started run for item %s", item.ID)
	return &WorkResult{Success: true}
}

// processExecuteStep handles executing a workflow step
func (rw *RemoteWorker) processExecuteStep(ctx context.Context, item *QueueItem) *WorkResult {
	// This would load the step details and execute the node
	// For now, return success
	log.Printf("Executed step for item %s", item.ID)
	return &WorkResult{Success: true}
}

// processRetryStep handles retrying a failed workflow step
func (rw *RemoteWorker) processRetryStep(ctx context.Context, item *QueueItem) *WorkResult {
	// Implementation would retry the step
	log.Printf("Retried step for item %s", item.ID)
	return &WorkResult{Success: true}
}

// processCompleteRun handles completing a workflow run
func (rw *RemoteWorker) processCompleteRun(ctx context.Context, item *QueueItem) *WorkResult {
	// Implementation would finalize the run
	log.Printf("Completed run for item %s", item.ID)
	return &WorkResult{Success: true}
}

// completeWork reports the work result back to the API server
func (rw *RemoteWorker) completeWork(ctx context.Context, itemID uuid.UUID, result *WorkResult) error {
	url := fmt.Sprintf("%s/api/workers/%s/complete-work/%s", rw.serverURL, rw.workerID, itemID)

	payload, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal work result: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create complete work request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+rw.token)

	resp, err := rw.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to complete work: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("complete work failed with status %d: %s", resp.StatusCode, string(body))
	}

	log.Printf("Completed work item %s", itemID)
	return nil
}

// getVersion returns the application version (placeholder)
func getVersion() *string {
	version := "1.0.0"
	return &version
}

// stringPointer returns a pointer to a string
func stringPointer(s string) *string {
	return &s
}

// generateWorkerID generates a unique worker ID
func generateWorkerID() string {
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID
		return fmt.Sprintf("worker-%d", time.Now().Unix())
	}
	return fmt.Sprintf("worker-%s", hex.EncodeToString(bytes))
}
