package execution

import (
	"context"
	"time"

	"github.com/cedricziel/mel-agent/pkg/api"
	"github.com/google/uuid"
)

// WorkflowRunStatus represents the status of a workflow run
type WorkflowRunStatus string

const (
	RunStatusPending   WorkflowRunStatus = "pending"
	RunStatusRunning   WorkflowRunStatus = "running"
	RunStatusPaused    WorkflowRunStatus = "paused"
	RunStatusCompleted WorkflowRunStatus = "completed"
	RunStatusFailed    WorkflowRunStatus = "failed"
	RunStatusCancelled WorkflowRunStatus = "cancelled"
)

// StepStatus represents the status of a workflow step
type StepStatus string

const (
	StepStatusPending   StepStatus = "pending"
	StepStatusRunning   StepStatus = "running"
	StepStatusCompleted StepStatus = "completed"
	StepStatusFailed    StepStatus = "failed"
	StepStatusSkipped   StepStatus = "skipped"
	StepStatusRetrying  StepStatus = "retrying"
)

// WorkerStatus represents the status of a worker
type WorkerStatus string

const (
	WorkerStatusIdle     WorkerStatus = "idle"
	WorkerStatusBusy     WorkerStatus = "busy"
	WorkerStatusDraining WorkerStatus = "draining"
	WorkerStatusOffline  WorkerStatus = "offline"
)

// WorkflowRun represents a workflow execution instance
type WorkflowRun struct {
	ID               uuid.UUID         `json:"id" db:"id"`
	AgentID          uuid.UUID         `json:"agent_id" db:"agent_id"`
	VersionID        uuid.UUID         `json:"version_id" db:"version_id"`
	TriggerID        *uuid.UUID        `json:"trigger_id,omitempty" db:"trigger_id"`
	Status           WorkflowRunStatus `json:"status" db:"status"`
	CreatedAt        time.Time         `json:"created_at" db:"created_at"`
	StartedAt        *time.Time        `json:"started_at,omitempty" db:"started_at"`
	CompletedAt      *time.Time        `json:"completed_at,omitempty" db:"completed_at"`
	InputData        map[string]any    `json:"input_data,omitempty" db:"input_data"`
	OutputData       map[string]any    `json:"output_data,omitempty" db:"output_data"`
	ErrorData        map[string]any    `json:"error_data,omitempty" db:"error_data"`
	Variables        map[string]any    `json:"variables" db:"variables"`
	TimeoutSeconds   int               `json:"timeout_seconds" db:"timeout_seconds"`
	RetryPolicy      RetryPolicy       `json:"retry_policy" db:"retry_policy"`
	AssignedWorkerID *string           `json:"assigned_worker_id,omitempty" db:"assigned_worker_id"`
	WorkerHeartbeat  *time.Time        `json:"worker_heartbeat,omitempty" db:"worker_heartbeat"`
	TotalSteps       int               `json:"total_steps" db:"total_steps"`
	CompletedSteps   int               `json:"completed_steps" db:"completed_steps"`
	FailedSteps      int               `json:"failed_steps" db:"failed_steps"`
}

// WorkflowStep represents a single node execution within a workflow run
type WorkflowStep struct {
	ID               uuid.UUID          `json:"id" db:"id"`
	RunID            uuid.UUID          `json:"run_id" db:"run_id"`
	NodeID           string             `json:"node_id" db:"node_id"`
	NodeType         string             `json:"node_type" db:"node_type"`
	StepNumber       int                `json:"step_number" db:"step_number"`
	Status           StepStatus         `json:"status" db:"status"`
	AttemptCount     int                `json:"attempt_count" db:"attempt_count"`
	MaxAttempts      int                `json:"max_attempts" db:"max_attempts"`
	CreatedAt        time.Time          `json:"created_at" db:"created_at"`
	StartedAt        *time.Time         `json:"started_at,omitempty" db:"started_at"`
	CompletedAt      *time.Time         `json:"completed_at,omitempty" db:"completed_at"`
	NextRetryAt      *time.Time         `json:"next_retry_at,omitempty" db:"next_retry_at"`
	InputEnvelope    *api.Envelope[any] `json:"input_envelope,omitempty" db:"input_envelope"`
	OutputEnvelope   *api.Envelope[any] `json:"output_envelope,omitempty" db:"output_envelope"`
	NodeConfig       map[string]any     `json:"node_config,omitempty" db:"node_config"`
	ErrorDetails     map[string]any     `json:"error_details,omitempty" db:"error_details"`
	AssignedWorkerID *string            `json:"assigned_worker_id,omitempty" db:"assigned_worker_id"`
	WorkerHeartbeat  *time.Time         `json:"worker_heartbeat,omitempty" db:"worker_heartbeat"`
	DependsOn        []uuid.UUID        `json:"depends_on" db:"depends_on"`
}

// WorkflowWorker represents a worker instance in the pool
type WorkflowWorker struct {
	ID                   string       `json:"id" db:"id"`
	Hostname             string       `json:"hostname" db:"hostname"`
	ProcessID            *int         `json:"process_id,omitempty" db:"process_id"`
	Version              *string      `json:"version,omitempty" db:"version"`
	Capabilities         []string     `json:"capabilities" db:"capabilities"`
	Status               WorkerStatus `json:"status" db:"status"`
	LastHeartbeat        time.Time    `json:"last_heartbeat" db:"last_heartbeat"`
	StartedAt            time.Time    `json:"started_at" db:"started_at"`
	MaxConcurrentSteps   int          `json:"max_concurrent_steps" db:"max_concurrent_steps"`
	CurrentStepCount     int          `json:"current_step_count" db:"current_step_count"`
	TotalStepsExecuted   int          `json:"total_steps_executed" db:"total_steps_executed"`
	TotalExecutionTimeMS int64        `json:"total_execution_time_ms" db:"total_execution_time_ms"`
}

// RetryPolicy defines how failed steps should be retried
type RetryPolicy struct {
	MaxAttempts        int      `json:"max_attempts"`
	BackoffMultiplier  float64  `json:"backoff_multiplier"`
	InitialDelayMS     int64    `json:"initial_delay_ms"`
	MaxDelayMS         int64    `json:"max_delay_ms"`
	RetryableErrors    []string `json:"retryable_errors,omitempty"`
	NonRetryableErrors []string `json:"non_retryable_errors,omitempty"`
}

// DefaultRetryPolicy returns a sensible default retry policy
func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxAttempts:       3,
		BackoffMultiplier: 2.0,
		InitialDelayMS:    1000,   // 1 second
		MaxDelayMS:        300000, // 5 minutes
	}
}

// ExecutionEngine defines the interface for workflow execution
type ExecutionEngine interface {
	// Run management
	StartRun(ctx context.Context, run *WorkflowRun) error
	PauseRun(ctx context.Context, runID uuid.UUID) error
	ResumeRun(ctx context.Context, runID uuid.UUID) error
	CancelRun(ctx context.Context, runID uuid.UUID) error

	// Step execution
	ExecuteStep(ctx context.Context, step *WorkflowStep) (*api.Envelope[any], error)
	RetryStep(ctx context.Context, stepID uuid.UUID) error

	// Worker management
	RegisterWorker(ctx context.Context, worker *WorkflowWorker) error
	UnregisterWorker(ctx context.Context, workerID string) error
	UpdateWorkerHeartbeat(ctx context.Context, workerID string) error

	// Queue management
	ClaimWork(ctx context.Context, workerID string, maxItems int) ([]*QueueItem, error)
	CompleteWork(ctx context.Context, workerID string, itemID uuid.UUID, result *WorkResult) error

	// Recovery
	RecoverOrphanedWork(ctx context.Context, workerTimeoutDuration time.Duration) error
	RecoverFailedRuns(ctx context.Context) error
}

// QueueItem represents a work item in the execution queue
type QueueItem struct {
	ID           uuid.UUID      `json:"id" db:"id"`
	RunID        uuid.UUID      `json:"run_id" db:"run_id"`
	StepID       *uuid.UUID     `json:"step_id,omitempty" db:"step_id"`
	QueueType    QueueType      `json:"queue_type" db:"queue_type"`
	Priority     int            `json:"priority" db:"priority"`
	AvailableAt  time.Time      `json:"available_at" db:"available_at"`
	CreatedAt    time.Time      `json:"created_at" db:"created_at"`
	ClaimedAt    *time.Time     `json:"claimed_at,omitempty" db:"claimed_at"`
	ClaimedBy    *string        `json:"claimed_by,omitempty" db:"claimed_by"`
	AttemptCount int            `json:"attempt_count" db:"attempt_count"`
	MaxAttempts  int            `json:"max_attempts" db:"max_attempts"`
	Payload      map[string]any `json:"payload,omitempty" db:"payload"`
}

// QueueType represents different types of queue items
type QueueType string

const (
	QueueTypeStartRun    QueueType = "start_run"
	QueueTypeExecuteStep QueueType = "execute_step"
	QueueTypeRetryStep   QueueType = "retry_step"
	QueueTypeCompleteRun QueueType = "complete_run"
)

// WorkResult represents the result of processing a queue item
type WorkResult struct {
	Success     bool           `json:"success"`
	Error       *string        `json:"error,omitempty"`
	OutputData  map[string]any `json:"output_data,omitempty"`
	NextSteps   []uuid.UUID    `json:"next_steps,omitempty"`
	ShouldRetry bool           `json:"should_retry"`
	RetryDelay  *time.Duration `json:"retry_delay,omitempty"`
}

// ExecutionContext provides context for step execution
type ExecutionContext struct {
	api.ExecutionContext
	Run    *WorkflowRun    `json:"run"`
	Step   *WorkflowStep   `json:"step"`
	Worker *WorkflowWorker `json:"worker"`
}

// StepDependency represents a dependency between workflow steps
type StepDependency struct {
	StepID    uuid.UUID `json:"step_id"`
	DependsOn uuid.UUID `json:"depends_on"`
	Condition string    `json:"condition,omitempty"` // Optional condition for dependency
}

// WorkflowGraph represents the execution graph of a workflow
type WorkflowGraph struct {
	Steps        []*WorkflowStep  `json:"steps"`
	Dependencies []StepDependency `json:"dependencies"`
	EntryPoints  []uuid.UUID      `json:"entry_points"`
}

// IsRetryable determines if an error should be retried based on the retry policy
func (rp *RetryPolicy) IsRetryable(err error, attemptCount int) bool {
	// Check attempt count
	if attemptCount >= rp.MaxAttempts {
		return false
	}

	errMsg := err.Error()

	// Check non-retryable errors first
	for _, nonRetryable := range rp.NonRetryableErrors {
		if nonRetryable == errMsg {
			return false
		}
	}

	// If retryable errors are specified, only retry those
	if len(rp.RetryableErrors) > 0 {
		for _, retryable := range rp.RetryableErrors {
			if retryable == errMsg {
				return true
			}
		}
		return false
	}

	// Default: retry most errors except for specific non-retryable ones
	return true
}

// CalculateRetryDelay calculates the delay before the next retry attempt
func (rp *RetryPolicy) CalculateRetryDelay(attemptCount int) time.Duration {
	if attemptCount <= 0 {
		return time.Duration(rp.InitialDelayMS) * time.Millisecond
	}

	// Exponential backoff
	delayMS := float64(rp.InitialDelayMS)
	for i := 0; i < attemptCount; i++ {
		delayMS *= rp.BackoffMultiplier
	}

	// Cap at max delay
	if delayMS > float64(rp.MaxDelayMS) {
		delayMS = float64(rp.MaxDelayMS)
	}

	return time.Duration(delayMS) * time.Millisecond
}
