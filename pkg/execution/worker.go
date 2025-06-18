package execution

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/cedricziel/mel-agent/pkg/api"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Worker represents a workflow execution worker
type Worker struct {
	id                 string
	hostname           string
	version            string
	capabilities       []string
	maxConcurrentSteps int

	engine ExecutionEngine
	db     *sql.DB
	mel    api.Mel

	// Runtime state
	mu           sync.RWMutex
	running      bool
	currentSteps map[uuid.UUID]*WorkflowStep
	ctx          context.Context
	cancel       context.CancelFunc

	// Configuration
	heartbeatInterval time.Duration
	pollInterval      time.Duration
	workerTimeout     time.Duration
}

// NewWorker creates a new workflow worker
func NewWorker(db *sql.DB, mel api.Mel, config WorkerConfig) *Worker {
	hostname, _ := os.Hostname()

	worker := &Worker{
		id:                 fmt.Sprintf("%s-%d-%s", hostname, os.Getpid(), uuid.New().String()[:8]),
		hostname:           hostname,
		version:            config.Version,
		capabilities:       config.Capabilities,
		maxConcurrentSteps: config.MaxConcurrentSteps,
		db:                 db,
		mel:                mel,
		currentSteps:       make(map[uuid.UUID]*WorkflowStep),
		heartbeatInterval:  config.HeartbeatInterval,
		pollInterval:       config.PollInterval,
		workerTimeout:      config.WorkerTimeout,
	}

	worker.engine = NewDurableExecutionEngine(db, mel, worker.id)
	return worker
}

// WorkerConfig holds worker configuration
type WorkerConfig struct {
	Version            string
	Capabilities       []string
	MaxConcurrentSteps int
	HeartbeatInterval  time.Duration
	PollInterval       time.Duration
	WorkerTimeout      time.Duration
}

// DefaultWorkerConfig returns sensible defaults
func DefaultWorkerConfig() WorkerConfig {
	return WorkerConfig{
		Version:            "1.0.0",
		Capabilities:       []string{"*"}, // Support all node types
		MaxConcurrentSteps: 10,
		HeartbeatInterval:  30 * time.Second,
		PollInterval:       5 * time.Second,
		WorkerTimeout:      5 * time.Minute,
	}
}

// Start begins the worker's execution loop
func (w *Worker) Start(ctx context.Context) error {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return fmt.Errorf("worker is already running")
	}

	w.ctx, w.cancel = context.WithCancel(ctx)
	w.running = true
	w.mu.Unlock()

	// Register the worker
	workerRecord := &WorkflowWorker{
		ID:                 w.id,
		Hostname:           w.hostname,
		Version:            &w.version,
		Capabilities:       w.capabilities,
		Status:             WorkerStatusIdle,
		MaxConcurrentSteps: w.maxConcurrentSteps,
		CurrentStepCount:   0,
	}

	if err := w.engine.RegisterWorker(w.ctx, workerRecord); err != nil {
		return fmt.Errorf("failed to register worker: %w", err)
	}

	log.Printf("Worker %s started with capabilities: %v", w.id, w.capabilities)

	// Start background goroutines
	var wg sync.WaitGroup

	// Heartbeat goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		w.heartbeatLoop()
	}()

	// Work polling goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		w.workLoop()
	}()

	// Recovery goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		w.recoveryLoop()
	}()

	// Wait for context cancellation
	<-w.ctx.Done()

	// Graceful shutdown
	log.Printf("Worker %s shutting down...", w.id)

	// Wait for all goroutines to finish
	wg.Wait()

	// Unregister the worker
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := w.engine.UnregisterWorker(shutdownCtx, w.id); err != nil {
		log.Printf("Failed to unregister worker: %v", err)
	}

	w.mu.Lock()
	w.running = false
	w.mu.Unlock()

	log.Printf("Worker %s stopped", w.id)
	return nil
}

// Stop gracefully stops the worker
func (w *Worker) Stop() error {
	w.mu.RLock()
	if !w.running {
		w.mu.RUnlock()
		return fmt.Errorf("worker is not running")
	}
	cancel := w.cancel
	w.mu.RUnlock()

	if cancel != nil {
		cancel()
	}

	return nil
}

// heartbeatLoop maintains the worker's heartbeat
func (w *Worker) heartbeatLoop() {
	ticker := time.NewTicker(w.heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-ticker.C:
			if err := w.engine.UpdateWorkerHeartbeat(w.ctx, w.id); err != nil {
				log.Printf("Failed to update heartbeat: %v", err)
			}
		}
	}
}

// workLoop continuously polls for and processes work
func (w *Worker) workLoop() {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-ticker.C:
			w.processWork()
		}
	}
}

// recoveryLoop periodically recovers orphaned work
func (w *Worker) recoveryLoop() {
	ticker := time.NewTicker(w.workerTimeout)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-ticker.C:
			if err := w.engine.RecoverOrphanedWork(w.ctx, w.workerTimeout); err != nil {
				log.Printf("Failed to recover orphaned work: %v", err)
			}
		}
	}
}

// processWork claims and processes available work items
func (w *Worker) processWork() {
	w.mu.RLock()
	availableCapacity := w.maxConcurrentSteps - len(w.currentSteps)
	w.mu.RUnlock()

	if availableCapacity <= 0 {
		return // Worker is at capacity
	}

	// Claim work items
	items, err := w.engine.ClaimWork(w.ctx, w.id, availableCapacity)
	if err != nil {
		log.Printf("Failed to claim work: %v", err)
		return
	}

	if len(items) == 0 {
		return // No work available
	}

	log.Printf("Worker %s claimed %d work items", w.id, len(items))

	// Process each item
	for _, item := range items {
		go w.processQueueItem(item)
	}
}

// processQueueItem processes a single queue item
func (w *Worker) processQueueItem(item *QueueItem) {
	start := time.Now()
	result := &WorkResult{Success: false}

	defer func() {
		duration := time.Since(start)

		// Complete the work item
		if err := w.engine.CompleteWork(w.ctx, w.id, item.ID, result); err != nil {
			log.Printf("Failed to complete work item %s: %v", item.ID, err)
		}

		log.Printf("Worker %s completed item %s in %v (success: %t)",
			w.id, item.ID, duration, result.Success)
	}()

	switch item.QueueType {
	case QueueTypeStartRun:
		result = w.processStartRun(item)
	case QueueTypeExecuteStep:
		result = w.processExecuteStep(item)
	case QueueTypeRetryStep:
		result = w.processRetryStep(item)
	case QueueTypeCompleteRun:
		result = w.processCompleteRun(item)
	default:
		result.Error = stringPtr(fmt.Sprintf("unknown queue type: %s", item.QueueType))
	}
}

// processStartRun processes a workflow run start
func (w *Worker) processStartRun(item *QueueItem) *WorkResult {
	// Load the workflow run
	run, err := w.loadWorkflowRun(item.RunID)
	if err != nil {
		return &WorkResult{
			Success: false,
			Error:   stringPtr(fmt.Sprintf("failed to load run: %v", err)),
		}
	}

	// Load the workflow graph
	graph, err := w.loadWorkflowGraph(run.AgentID, run.VersionID)
	if err != nil {
		return &WorkResult{
			Success: false,
			Error:   stringPtr(fmt.Sprintf("failed to load workflow graph: %v", err)),
		}
	}

	// Create steps for the workflow
	steps, err := w.createWorkflowSteps(run, graph)
	if err != nil {
		return &WorkResult{
			Success: false,
			Error:   stringPtr(fmt.Sprintf("failed to create workflow steps: %v", err)),
		}
	}

	// Queue the entry point steps
	var nextSteps []uuid.UUID
	for _, step := range steps {
		if len(step.DependsOn) == 0 { // Entry point steps have no dependencies
			nextSteps = append(nextSteps, step.ID)
		}
	}

	return &WorkResult{
		Success:   true,
		NextSteps: nextSteps,
	}
}

// processExecuteStep processes a step execution
func (w *Worker) processExecuteStep(item *QueueItem) *WorkResult {
	if item.StepID == nil {
		return &WorkResult{
			Success: false,
			Error:   stringPtr("step ID is required for execute_step"),
		}
	}

	// Load the step
	step, err := w.loadWorkflowStep(*item.StepID)
	if err != nil {
		return &WorkResult{
			Success: false,
			Error:   stringPtr(fmt.Sprintf("failed to load step: %v", err)),
		}
	}

	// Check dependencies
	ready, err := w.areStepDependenciesReady(step)
	if err != nil {
		return &WorkResult{
			Success: false,
			Error:   stringPtr(fmt.Sprintf("failed to check dependencies: %v", err)),
		}
	}

	if !ready {
		// Dependencies not ready, requeue for later
		return &WorkResult{
			Success:     false,
			ShouldRetry: true,
			RetryDelay:  durationPtr(30 * time.Second),
		}
	}

	// Add step to current steps
	w.mu.Lock()
	w.currentSteps[step.ID] = step
	w.mu.Unlock()

	defer func() {
		w.mu.Lock()
		delete(w.currentSteps, step.ID)
		w.mu.Unlock()
	}()

	// Execute the step
	output, err := w.engine.ExecuteStep(w.ctx, step)
	if err != nil {
		return &WorkResult{
			Success:     false,
			Error:       stringPtr(fmt.Sprintf("step execution failed: %v", err)),
			ShouldRetry: true,
		}
	}

	// Find next steps
	nextSteps, err := w.findNextSteps(step.RunID, step.ID)
	if err != nil {
		log.Printf("Failed to find next steps: %v", err)
	}

	return &WorkResult{
		Success:    true,
		OutputData: map[string]any{"envelope": output},
		NextSteps:  nextSteps,
	}
}

// processRetryStep processes a step retry
func (w *Worker) processRetryStep(item *QueueItem) *WorkResult {
	// Similar to processExecuteStep but with retry logic
	return w.processExecuteStep(item)
}

// processCompleteRun processes workflow run completion
func (w *Worker) processCompleteRun(item *QueueItem) *WorkResult {
	// Update run status to completed
	query := `UPDATE workflow_runs SET status = 'completed', completed_at = NOW() WHERE id = $1`
	if _, err := w.db.ExecContext(w.ctx, query, item.RunID); err != nil {
		return &WorkResult{
			Success: false,
			Error:   stringPtr(fmt.Sprintf("failed to complete run: %v", err)),
		}
	}

	return &WorkResult{Success: true}
}

// Helper methods

func (w *Worker) loadWorkflowRun(runID uuid.UUID) (*WorkflowRun, error) {
	query := `SELECT id, agent_id, version_id, status FROM workflow_runs WHERE id = $1`
	row := w.db.QueryRowContext(w.ctx, query, runID)

	var run WorkflowRun
	if err := row.Scan(&run.ID, &run.AgentID, &run.VersionID, &run.Status); err != nil {
		return nil, err
	}
	return &run, nil
}

func (w *Worker) loadWorkflowStep(stepID uuid.UUID) (*WorkflowStep, error) {
	query := `SELECT id, run_id, node_id, node_type, status FROM workflow_steps WHERE id = $1`
	row := w.db.QueryRowContext(w.ctx, query, stepID)

	var step WorkflowStep
	if err := row.Scan(&step.ID, &step.RunID, &step.NodeID, &step.NodeType, &step.Status); err != nil {
		return nil, err
	}
	return &step, nil
}

func (w *Worker) loadWorkflowGraph(agentID, versionID uuid.UUID) (*WorkflowGraph, error) {
	// This would load the workflow definition and convert it to a WorkflowGraph
	// Implementation depends on how workflows are stored
	return &WorkflowGraph{}, nil
}

func (w *Worker) createWorkflowSteps(run *WorkflowRun, graph *WorkflowGraph) ([]*WorkflowStep, error) {
	// This would create WorkflowStep records for each node in the graph
	// Implementation depends on workflow graph structure
	return []*WorkflowStep{}, nil
}

func (w *Worker) areStepDependenciesReady(step *WorkflowStep) (bool, error) {
	if len(step.DependsOn) == 0 {
		return true, nil
	}

	// Check if all dependencies are completed
	query := `
		SELECT COUNT(*) FROM workflow_steps 
		WHERE id = ANY($1) AND status = 'completed'`

	var completedCount int
	row := w.db.QueryRowContext(w.ctx, query, pq.Array(step.DependsOn))
	if err := row.Scan(&completedCount); err != nil {
		return false, err
	}

	return completedCount == len(step.DependsOn), nil
}

func (w *Worker) findNextSteps(runID, completedStepID uuid.UUID) ([]uuid.UUID, error) {
	// Find steps that depend on the completed step
	query := `
		SELECT id FROM workflow_steps 
		WHERE run_id = $1 AND $2 = ANY(depends_on) AND status = 'pending'`

	rows, err := w.db.QueryContext(w.ctx, query, runID, completedStepID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nextSteps []uuid.UUID
	for rows.Next() {
		var stepID uuid.UUID
		if err := rows.Scan(&stepID); err != nil {
			continue
		}
		nextSteps = append(nextSteps, stepID)
	}

	return nextSteps, nil
}

// Helper functions
func stringPtr(s string) *string                 { return &s }
func durationPtr(d time.Duration) *time.Duration { return &d }
