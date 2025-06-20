package execution

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/cedricziel/mel-agent/pkg/api"
	"github.com/cedricziel/mel-agent/pkg/client"
	"github.com/google/uuid"
)

// RemoteWorker represents a worker that connects to a remote API server
type RemoteWorker struct {
	serverURL   string
	token       string
	workerID    string
	mel         api.Mel
	concurrency int
	apiClient   client.ClientWithResponsesInterface
	workerInfo  *WorkflowWorker
}

// NewRemoteWorker creates a new remote worker instance
func NewRemoteWorker(serverURL, token, workerID string, mel api.Mel, concurrency int) (*RemoteWorker, error) {
	// Create the API client
	apiClient, err := client.NewClientWithResponses(serverURL, client.WithRequestEditorFn(client.WithBearerToken(token)))
	if err != nil {
		return nil, fmt.Errorf("failed to create API client: %w", err)
	}

	return &RemoteWorker{
		serverURL:   serverURL,
		token:       token,
		workerID:    workerID,
		mel:         mel,
		concurrency: concurrency,
		apiClient:   apiClient,
	}, nil
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
	// Convert internal WorkflowWorker to client RegisterWorkerRequest
	req := client.RegisterWorkerRequest{
		Id:          rw.workerID,
		Name:        &rw.workerInfo.Hostname,
		Concurrency: &rw.concurrency,
	}

	resp, err := rw.apiClient.RegisterWorkerWithResponse(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to register worker: %w", err)
	}

	if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusCreated {
		var errMsg string
		if resp.JSON400 != nil && resp.JSON400.Message != nil {
			errMsg = *resp.JSON400.Message
		} else if resp.JSON500 != nil && resp.JSON500.Message != nil {
			errMsg = *resp.JSON500.Message
		} else {
			errMsg = string(resp.Body)
		}
		return fmt.Errorf("worker registration failed with status %d: %s", resp.StatusCode(), errMsg)
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
	resp, err := rw.apiClient.UpdateWorkerHeartbeatWithResponse(ctx, rw.workerID)
	if err != nil {
		return fmt.Errorf("failed to send heartbeat: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		var errMsg string
		if resp.JSON404 != nil && resp.JSON404.Message != nil {
			errMsg = *resp.JSON404.Message
		} else if resp.JSON500 != nil && resp.JSON500.Message != nil {
			errMsg = *resp.JSON500.Message
		} else {
			errMsg = string(resp.Body)
		}
		return fmt.Errorf("heartbeat failed with status %d: %s", resp.StatusCode(), errMsg)
	}

	return nil
}

// unregisterWorker unregisters this worker from the API server
func (rw *RemoteWorker) unregisterWorker(ctx context.Context) error {
	resp, err := rw.apiClient.UnregisterWorkerWithResponse(ctx, rw.workerID)
	if err != nil {
		return fmt.Errorf("failed to unregister worker: %w", err)
	}

	if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusNoContent {
		var errMsg string
		if resp.JSON404 != nil && resp.JSON404.Message != nil {
			errMsg = *resp.JSON404.Message
		} else if resp.JSON500 != nil && resp.JSON500.Message != nil {
			errMsg = *resp.JSON500.Message
		} else {
			errMsg = string(resp.Body)
		}
		return fmt.Errorf("unregister failed with status %d: %s", resp.StatusCode(), errMsg)
	}

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
	// Create the claim work request body
	maxItems := rw.concurrency
	reqBody := client.ClaimWorkJSONBody{
		MaxItems: &maxItems,
	}

	resp, err := rw.apiClient.ClaimWorkWithResponse(ctx, rw.workerID, client.ClaimWorkJSONRequestBody(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to claim work: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("claim work failed with status %d: %s", resp.StatusCode(), string(resp.Body))
	}

	// Convert client WorkItem to internal QueueItem
	var workItems []*QueueItem
	if resp.JSON200 != nil {
		for _, item := range *resp.JSON200 {
			// Convert from client.WorkItem to internal QueueItem
			queueItem, err := convertWorkItemToQueueItem(&item)
			if err != nil {
				return nil, fmt.Errorf("failed to convert work item: %w", err)
			}
			workItems = append(workItems, queueItem)
		}
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
	_ = ctx
	// Implementation would depend on how the payload is structured
	// For now, return success
	log.Printf("Started run for item %s", item.ID)
	return &WorkResult{
		Success:    true,
		OutputData: map[string]any{"action": "start_run", "item_id": item.ID.String()},
	}
}

// processExecuteStep handles executing a workflow step
func (rw *RemoteWorker) processExecuteStep(ctx context.Context, item *QueueItem) *WorkResult {
	_ = ctx
	// This would load the step details and execute the node
	// For now, return success
	log.Printf("Executed step for item %s", item.ID)
	return &WorkResult{
		Success:    true,
		OutputData: map[string]any{"action": "execute_step", "item_id": item.ID.String()},
	}
}

// processRetryStep handles retrying a failed workflow step
func (rw *RemoteWorker) processRetryStep(ctx context.Context, item *QueueItem) *WorkResult {
	_ = ctx
	// Implementation would retry the step
	log.Printf("Retried step for item %s", item.ID)
	return &WorkResult{
		Success:    true,
		OutputData: map[string]any{"action": "retry_step", "item_id": item.ID.String()},
	}
}

// processCompleteRun handles completing a workflow run
func (rw *RemoteWorker) processCompleteRun(ctx context.Context, item *QueueItem) *WorkResult {
	_ = ctx
	// Implementation would finalize the run
	log.Printf("Completed run for item %s", item.ID)
	return &WorkResult{
		Success:    true,
		OutputData: map[string]any{"action": "complete_run", "item_id": item.ID.String()},
	}
}

// completeWork reports the work result back to the API server
func (rw *RemoteWorker) completeWork(ctx context.Context, itemID uuid.UUID, result *WorkResult) error {
	// Convert internal WorkResult to client CompleteWorkJSONBody
	reqBody := client.CompleteWorkJSONBody{}

	if result.Success {
		if result.OutputData != nil {
			reqBody.Result = &result.OutputData
		}
	} else {
		if result.Error != nil {
			reqBody.Error = result.Error
		}
	}

	resp, err := rw.apiClient.CompleteWorkWithResponse(ctx, rw.workerID, itemID.String(), client.CompleteWorkJSONRequestBody(reqBody))
	if err != nil {
		return fmt.Errorf("failed to complete work: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("complete work failed with status %d: %s", resp.StatusCode(), string(resp.Body))
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

// convertWorkItemToQueueItem converts a client WorkItem to internal QueueItem
func convertWorkItemToQueueItem(item *client.WorkItem) (*QueueItem, error) {
	if item.Id == nil || item.Type == nil {
		return nil, fmt.Errorf("invalid work item: missing required fields")
	}

	// Parse the ID
	id, err := uuid.Parse(*item.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid work item ID: %w", err)
	}

	// Create the queue item
	queueItem := &QueueItem{
		ID:        id,
		QueueType: QueueType(*item.Type),
	}

	// Extract additional fields from payload if available
	if item.Payload != nil {
		payload := *item.Payload

		// Try to extract run_id
		if runIDStr, ok := payload["run_id"].(string); ok {
			if runID, err := uuid.Parse(runIDStr); err == nil {
				queueItem.RunID = runID
			}
		}

		// Try to extract step_id
		if stepIDStr, ok := payload["step_id"].(string); ok {
			if stepID, err := uuid.Parse(stepIDStr); err == nil {
				queueItem.StepID = &stepID
			}
		}

		// Try to extract priority
		if priority, ok := payload["priority"].(float64); ok {
			queueItem.Priority = int(priority)
		}

		// Store the entire payload
		queueItem.Payload = payload
	}

	// Set created_at if available
	if item.CreatedAt != nil {
		queueItem.CreatedAt = *item.CreatedAt
	}

	return queueItem, nil
}
