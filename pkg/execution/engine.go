package execution

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/cedricziel/mel-agent/pkg/api"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

// DurableExecutionEngine implements durable workflow execution with step-by-step persistence
type DurableExecutionEngine struct {
	db       *sql.DB
	mel      api.Mel
	workerID string
}

// NewDurableExecutionEngine creates a new durable execution engine
func NewDurableExecutionEngine(db *sql.DB, mel api.Mel, workerID string) *DurableExecutionEngine {
	return &DurableExecutionEngine{
		db:       db,
		mel:      mel,
		workerID: workerID,
	}
}

// StartRun initiates a new workflow run
func (e *DurableExecutionEngine) StartRun(ctx context.Context, run *WorkflowRun) error {
	// Insert the workflow run
	query := `
		INSERT INTO workflow_runs (
			id, agent_id, version_id, trigger_id, status, input_data, 
			variables, timeout_seconds, retry_policy
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)`

	inputDataJSON, _ := json.Marshal(run.InputData)
	variablesJSON, _ := json.Marshal(run.Variables)
	retryPolicyJSON, _ := json.Marshal(run.RetryPolicy)

	if _, err := e.db.ExecContext(ctx, query,
		run.ID, run.AgentID, run.VersionID, run.TriggerID, run.Status,
		inputDataJSON, variablesJSON, run.TimeoutSeconds, retryPolicyJSON); err != nil {
		return fmt.Errorf("failed to create workflow run: %w", err)
	}

	// Queue the run for execution
	queueItem := &QueueItem{
		ID:          uuid.New(),
		RunID:       run.ID,
		QueueType:   QueueTypeStartRun,
		Priority:    5,
		AvailableAt: time.Now(),
		MaxAttempts: 3,
		Payload:     map[string]any{},
	}

	return e.enqueueItem(ctx, queueItem)
}

// ExecuteStep executes a single workflow step
func (e *DurableExecutionEngine) ExecuteStep(ctx context.Context, step *WorkflowStep) (*api.Envelope[any], error) {
	// Update step status to running
	if err := e.updateStepStatus(ctx, step.ID, StepStatusRunning, nil); err != nil {
		return nil, fmt.Errorf("failed to update step status: %w", err)
	}

	// Create checkpoint before execution
	if err := e.createCheckpoint(ctx, step.RunID, step.ID, "pre_execution", nil); err != nil {
		log.Printf("Warning: failed to create pre-execution checkpoint: %v", err)
	}

	// Find the node definition
	nodeDef := e.mel.FindDefinition(step.NodeType)
	if nodeDef == nil {
		return nil, fmt.Errorf("node definition not found for type: %s", step.NodeType)
	}

	// Create execution context
	execCtx := api.ExecutionContext{
		AgentID: step.RunID.String(), // Use run ID as agent context
		RunID:   step.RunID.String(),
		Mel:     e.mel,
	}

	// Create node instance from step config
	node := api.Node{
		ID:   step.NodeID,
		Type: step.NodeType,
		Data: step.NodeConfig,
	}

	// Execute the node
	outputEnvelope, err := nodeDef.ExecuteEnvelope(execCtx, node, step.InputEnvelope)
	if err != nil {
		// Handle execution error
		errorDetails := map[string]any{
			"error":     err.Error(),
			"attempt":   step.AttemptCount + 1,
			"timestamp": time.Now(),
		}

		if err := e.updateStepError(ctx, step.ID, errorDetails); err != nil {
			log.Printf("Failed to update step error: %v", err)
		}

		return nil, fmt.Errorf("step execution failed: %w", err)
	}

	// Update step with successful output
	if err := e.updateStepOutput(ctx, step.ID, outputEnvelope); err != nil {
		return nil, fmt.Errorf("failed to update step output: %w", err)
	}

	// Create checkpoint after execution
	if err := e.createCheckpoint(ctx, step.RunID, step.ID, "post_execution", outputEnvelope); err != nil {
		log.Printf("Warning: failed to create post-execution checkpoint: %v", err)
	}

	return outputEnvelope, nil
}

// ClaimWork claims available work items for a worker
func (e *DurableExecutionEngine) ClaimWork(ctx context.Context, workerID string, maxItems int) ([]*QueueItem, error) {
	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Find available work items
	query := `
		SELECT id, run_id, step_id, queue_type, priority, available_at, 
		       created_at, attempt_count, max_attempts, payload
		FROM workflow_queue 
		WHERE claimed_by IS NULL 
		  AND available_at <= NOW()
		ORDER BY priority ASC, created_at ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED`

	rows, err := tx.QueryContext(ctx, query, maxItems)
	if err != nil {
		return nil, fmt.Errorf("failed to find work items: %w", err)
	}
	defer rows.Close()

	var items []*QueueItem
	for rows.Next() {
		var item QueueItem
		var payloadJSON []byte

		err := rows.Scan(
			&item.ID, &item.RunID, &item.StepID, &item.QueueType, &item.Priority,
			&item.AvailableAt, &item.CreatedAt, &item.AttemptCount, &item.MaxAttempts,
			&payloadJSON,
		)
		if err != nil {
			continue
		}

		if len(payloadJSON) > 0 {
			json.Unmarshal(payloadJSON, &item.Payload)
		}

		items = append(items, &item)
	}

	if len(items) == 0 {
		return items, nil
	}

	// Claim the items
	itemIDs := make([]uuid.UUID, len(items))
	for i, item := range items {
		itemIDs[i] = item.ID
		item.ClaimedBy = &workerID
		now := time.Now()
		item.ClaimedAt = &now
	}

	// Update claimed items
	claimQuery := `
		UPDATE workflow_queue 
		SET claimed_by = $1, claimed_at = NOW()
		WHERE id = ANY($2)`

	if _, err := tx.ExecContext(ctx, claimQuery, workerID, pq.Array(itemIDs)); err != nil {
		return nil, fmt.Errorf("failed to claim work items: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit claim transaction: %w", err)
	}

	return items, nil
}

// CompleteWork marks a work item as completed
func (e *DurableExecutionEngine) CompleteWork(ctx context.Context, workerID string, itemID uuid.UUID, result *WorkResult) error {
	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get the run_id before deleting the item
	var originalRunID uuid.UUID
	err = tx.QueryRowContext(ctx, "SELECT run_id FROM workflow_queue WHERE id = $1", itemID).Scan(&originalRunID)
	if err != nil {
		return fmt.Errorf("failed to get original run_id: %w", err)
	}

	// Remove completed item from queue
	deleteQuery := `DELETE FROM workflow_queue WHERE id = $1 AND claimed_by = $2`
	if _, err := tx.ExecContext(ctx, deleteQuery, itemID, workerID); err != nil {
		return fmt.Errorf("failed to remove completed item: %w", err)
	}

	// If there was an error and should retry, requeue the item
	if !result.Success && result.ShouldRetry {
		var retryAt time.Time
		if result.RetryDelay != nil {
			retryAt = time.Now().Add(*result.RetryDelay)
		} else {
			retryAt = time.Now().Add(5 * time.Minute) // Default 5 min delay
		}

		retryItem := &QueueItem{
			ID:          uuid.New(),
			RunID:       uuid.New(), // This should be populated from the original item
			QueueType:   QueueTypeRetryStep,
			Priority:    8, // Higher priority for retries
			AvailableAt: retryAt,
			MaxAttempts: 3,
		}

		if err := e.enqueueItemTx(ctx, tx, retryItem); err != nil {
			return fmt.Errorf("failed to requeue retry item: %w", err)
		}
	}

	// Queue next steps if provided
	for _, nextStepID := range result.NextSteps {
		nextItem := &QueueItem{
			ID:          uuid.New(),
			RunID:       originalRunID,
			StepID:      &nextStepID,
			QueueType:   QueueTypeExecuteStep,
			Priority:    5,
			AvailableAt: time.Now(),
			MaxAttempts: 3,
		}

		if err := e.enqueueItemTx(ctx, tx, nextItem); err != nil {
			return fmt.Errorf("failed to queue next step: %w", err)
		}
	}

	return tx.Commit()
}

// RegisterWorker registers a new worker in the pool
func (e *DurableExecutionEngine) RegisterWorker(ctx context.Context, worker *WorkflowWorker) error {
	query := `
		INSERT INTO workflow_workers (
			id, hostname, process_id, version, capabilities, status,
			max_concurrent_steps
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		) ON CONFLICT (id) DO UPDATE SET
			hostname = EXCLUDED.hostname,
			process_id = EXCLUDED.process_id,
			version = EXCLUDED.version,
			capabilities = EXCLUDED.capabilities,
			last_heartbeat = NOW(),
			status = EXCLUDED.status`

	_, err := e.db.ExecContext(ctx, query,
		worker.ID, worker.Hostname, worker.ProcessID, worker.Version,
		pq.Array(worker.Capabilities), worker.Status, worker.MaxConcurrentSteps)
	return err
}

// UnregisterWorker removes a worker from the pool
func (e *DurableExecutionEngine) UnregisterWorker(ctx context.Context, workerID string) error {
	query := `UPDATE workflow_workers SET status = 'offline' WHERE id = $1`
	_, err := e.db.ExecContext(ctx, query, workerID)
	return err
}

// UpdateWorkerHeartbeat updates the worker's last heartbeat
func (e *DurableExecutionEngine) UpdateWorkerHeartbeat(ctx context.Context, workerID string) error {
	query := `UPDATE workflow_workers SET last_heartbeat = NOW() WHERE id = $1`
	_, err := e.db.ExecContext(ctx, query, workerID)
	return err
}

// RecoverOrphanedWork recovers work from workers that have timed out
func (e *DurableExecutionEngine) RecoverOrphanedWork(ctx context.Context, workerTimeoutDuration time.Duration) error {
	// Find orphaned queue items
	orphanedQuery := `
		UPDATE workflow_queue 
		SET claimed_by = NULL, claimed_at = NULL, attempt_count = attempt_count + 1
		WHERE claimed_by IS NOT NULL 
		  AND claimed_at < NOW() - INTERVAL '%d seconds'
		  AND attempt_count < max_attempts`

	_, err := e.db.ExecContext(ctx, fmt.Sprintf(orphanedQuery, int(workerTimeoutDuration.Seconds())))
	if err != nil {
		return fmt.Errorf("failed to recover orphaned queue items: %w", err)
	}

	// Find orphaned workflow steps
	stepQuery := `
		UPDATE workflow_steps 
		SET assigned_worker_id = NULL, status = 'pending'
		WHERE assigned_worker_id IS NOT NULL 
		  AND worker_heartbeat < NOW() - INTERVAL '%d seconds'
		  AND status = 'running'`

	_, err = e.db.ExecContext(ctx, fmt.Sprintf(stepQuery, int(workerTimeoutDuration.Seconds())))
	if err != nil {
		return fmt.Errorf("failed to recover orphaned steps: %w", err)
	}

	return nil
}

// Helper methods

func (e *DurableExecutionEngine) enqueueItem(ctx context.Context, item *QueueItem) error {
	query := `
		INSERT INTO workflow_queue (
			id, run_id, step_id, queue_type, priority, available_at,
			max_attempts, payload
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)`

	payloadJSON, _ := json.Marshal(item.Payload)

	_, err := e.db.ExecContext(ctx, query,
		item.ID, item.RunID, item.StepID, item.QueueType, item.Priority,
		item.AvailableAt, item.MaxAttempts, payloadJSON)
	return err
}

func (e *DurableExecutionEngine) enqueueItemTx(ctx context.Context, tx *sql.Tx, item *QueueItem) error {
	query := `
		INSERT INTO workflow_queue (
			id, run_id, step_id, queue_type, priority, available_at,
			max_attempts, payload
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)`

	payloadJSON, _ := json.Marshal(item.Payload)

	_, err := tx.ExecContext(ctx, query,
		item.ID, item.RunID, item.StepID, item.QueueType, item.Priority,
		item.AvailableAt, item.MaxAttempts, payloadJSON)
	return err
}

func (e *DurableExecutionEngine) updateStepStatus(ctx context.Context, stepID uuid.UUID, status StepStatus, workerID *string) error {
	query := `
		UPDATE workflow_steps 
		SET status = $1, assigned_worker_id = $2, worker_heartbeat = NOW()
		WHERE id = $3`

	_, err := e.db.ExecContext(ctx, query, status, workerID, stepID)
	return err
}

func (e *DurableExecutionEngine) updateStepOutput(ctx context.Context, stepID uuid.UUID, output *api.Envelope[any]) error {
	outputJSON, err := json.Marshal(output)
	if err != nil {
		return fmt.Errorf("failed to marshal output: %w", err)
	}

	query := `
		UPDATE workflow_steps 
		SET status = 'completed', output_envelope = $1, completed_at = NOW()
		WHERE id = $2`

	_, err = e.db.ExecContext(ctx, query, outputJSON, stepID)
	return err
}

func (e *DurableExecutionEngine) updateStepError(ctx context.Context, stepID uuid.UUID, errorDetails map[string]any) error {
	errorJSON, err := json.Marshal(errorDetails)
	if err != nil {
		return fmt.Errorf("failed to marshal error: %w", err)
	}

	query := `
		UPDATE workflow_steps 
		SET status = 'failed', error_details = $1, attempt_count = attempt_count + 1
		WHERE id = $2`

	_, err = e.db.ExecContext(ctx, query, errorJSON, stepID)
	return err
}

func (e *DurableExecutionEngine) createCheckpoint(ctx context.Context, runID, stepID uuid.UUID, checkpointType string, data any) error {
	var dataJSON []byte
	var err error

	if data != nil {
		dataJSON, err = json.Marshal(data)
		if err != nil {
			return fmt.Errorf("failed to marshal checkpoint data: %w", err)
		}
	}

	query := `
		INSERT INTO workflow_checkpoints (
			id, run_id, step_id, checkpoint_type, execution_context
		) VALUES (
			$1, $2, $3, $4, $5
		)`

	_, err = e.db.ExecContext(ctx, query, uuid.New(), runID, stepID, checkpointType, dataJSON)
	return err
}

// PauseRun pauses a running workflow
func (e *DurableExecutionEngine) PauseRun(ctx context.Context, runID uuid.UUID) error {
	query := `UPDATE workflow_runs SET status = 'paused' WHERE id = $1 AND status = 'running'`
	_, err := e.db.ExecContext(ctx, query, runID)
	return err
}

// ResumeRun resumes a paused workflow
func (e *DurableExecutionEngine) ResumeRun(ctx context.Context, runID uuid.UUID) error {
	// Update run status
	query := `UPDATE workflow_runs SET status = 'running' WHERE id = $1 AND status = 'paused'`
	if _, err := e.db.ExecContext(ctx, query, runID); err != nil {
		return err
	}

	// Queue pending steps for execution
	stepsQuery := `
		SELECT id FROM workflow_steps 
		WHERE run_id = $1 AND status = 'pending'
		ORDER BY step_number`

	rows, err := e.db.QueryContext(ctx, stepsQuery, runID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var stepID uuid.UUID
		if err := rows.Scan(&stepID); err != nil {
			continue
		}

		queueItem := &QueueItem{
			ID:          uuid.New(),
			RunID:       runID,
			StepID:      &stepID,
			QueueType:   QueueTypeExecuteStep,
			Priority:    5,
			AvailableAt: time.Now(),
			MaxAttempts: 3,
		}

		if err := e.enqueueItem(ctx, queueItem); err != nil {
			log.Printf("Failed to queue step %s: %v", stepID, err)
		}
	}

	return nil
}

// CancelRun cancels a workflow run
func (e *DurableExecutionEngine) CancelRun(ctx context.Context, runID uuid.UUID) error {
	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Cancel the run
	runQuery := `UPDATE workflow_runs SET status = 'cancelled', completed_at = NOW() WHERE id = $1`
	if _, err := tx.ExecContext(ctx, runQuery, runID); err != nil {
		return fmt.Errorf("failed to cancel run: %w", err)
	}

	// Cancel pending steps
	stepsQuery := `UPDATE workflow_steps SET status = 'skipped' WHERE run_id = $1 AND status IN ('pending', 'retrying')`
	if _, err := tx.ExecContext(ctx, stepsQuery, runID); err != nil {
		return fmt.Errorf("failed to cancel steps: %w", err)
	}

	// Remove queue items
	queueQuery := `DELETE FROM workflow_queue WHERE run_id = $1 AND claimed_by IS NULL`
	if _, err := tx.ExecContext(ctx, queueQuery, runID); err != nil {
		return fmt.Errorf("failed to clear queue: %w", err)
	}

	return tx.Commit()
}

// RetryStep retries a failed step
func (e *DurableExecutionEngine) RetryStep(ctx context.Context, stepID uuid.UUID) error {
	// Load the step
	var step WorkflowStep
	query := `SELECT run_id, retry_policy FROM workflow_steps s
			  JOIN workflow_runs r ON s.run_id = r.id 
			  WHERE s.id = $1`

	var retryPolicyJSON []byte
	if err := e.db.QueryRowContext(ctx, query, stepID).Scan(&step.RunID, &retryPolicyJSON); err != nil {
		return fmt.Errorf("failed to load step: %w", err)
	}

	// Parse retry policy
	var retryPolicy RetryPolicy
	if err := json.Unmarshal(retryPolicyJSON, &retryPolicy); err != nil {
		retryPolicy = DefaultRetryPolicy()
	}

	// Calculate retry delay
	var attemptCount int
	countQuery := `SELECT attempt_count FROM workflow_steps WHERE id = $1`
	if err := e.db.QueryRowContext(ctx, countQuery, stepID).Scan(&attemptCount); err != nil {
		return fmt.Errorf("failed to get attempt count: %w", err)
	}

	if attemptCount >= retryPolicy.MaxAttempts {
		return fmt.Errorf("step has exceeded max retry attempts")
	}

	retryDelay := retryPolicy.CalculateRetryDelay(attemptCount)

	// Queue retry
	queueItem := &QueueItem{
		ID:          uuid.New(),
		RunID:       step.RunID,
		StepID:      &stepID,
		QueueType:   QueueTypeRetryStep,
		Priority:    8, // Higher priority for retries
		AvailableAt: time.Now().Add(retryDelay),
		MaxAttempts: 3,
	}

	return e.enqueueItem(ctx, queueItem)
}

// RecoverFailedRuns recovers runs that failed due to system issues
func (e *DurableExecutionEngine) RecoverFailedRuns(ctx context.Context) error {
	// Find runs that were running but have no recent heartbeat
	query := `
		UPDATE workflow_runs 
		SET status = 'pending', assigned_worker_id = NULL
		WHERE status = 'running' 
		  AND (worker_heartbeat IS NULL OR worker_heartbeat < NOW() - INTERVAL '5 minutes')
		RETURNING id`

	rows, err := e.db.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to find failed runs: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var runID uuid.UUID
		if err := rows.Scan(&runID); err != nil {
			continue
		}

		// Queue the run for restart
		queueItem := &QueueItem{
			ID:          uuid.New(),
			RunID:       runID,
			QueueType:   QueueTypeStartRun,
			Priority:    7, // High priority for recovery
			AvailableAt: time.Now(),
			MaxAttempts: 3,
		}

		if err := e.enqueueItem(ctx, queueItem); err != nil {
			log.Printf("Failed to queue recovered run %s: %v", runID, err)
		}
	}

	return nil
}
