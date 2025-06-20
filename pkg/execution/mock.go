package execution

import (
	"context"
	"time"

	apiPkg "github.com/cedricziel/mel-agent/pkg/api"
	"github.com/google/uuid"
)

// MockExecutionEngine is a simple mock implementation of ExecutionEngine for testing
type MockExecutionEngine struct{}

func (m *MockExecutionEngine) StartRun(ctx context.Context, run *WorkflowRun) error {
	return nil
}

func (m *MockExecutionEngine) PauseRun(ctx context.Context, runID uuid.UUID) error {
	return nil
}

func (m *MockExecutionEngine) ResumeRun(ctx context.Context, runID uuid.UUID) error {
	return nil
}

func (m *MockExecutionEngine) CancelRun(ctx context.Context, runID uuid.UUID) error {
	return nil
}

func (m *MockExecutionEngine) ExecuteStep(ctx context.Context, step *WorkflowStep) (*apiPkg.Envelope[any], error) {
	return &apiPkg.Envelope[any]{}, nil
}

func (m *MockExecutionEngine) RetryStep(ctx context.Context, stepID uuid.UUID) error {
	return nil
}

func (m *MockExecutionEngine) RegisterWorker(ctx context.Context, worker *WorkflowWorker) error {
	return nil
}

func (m *MockExecutionEngine) UnregisterWorker(ctx context.Context, workerID string) error {
	return nil
}

func (m *MockExecutionEngine) UpdateWorkerHeartbeat(ctx context.Context, workerID string) error {
	return nil
}

func (m *MockExecutionEngine) ClaimWork(ctx context.Context, workerID string, maxItems int) ([]*QueueItem, error) {
	return []*QueueItem{}, nil
}

func (m *MockExecutionEngine) CompleteWork(ctx context.Context, workerID string, itemID uuid.UUID, result *WorkResult) error {
	return nil
}

func (m *MockExecutionEngine) RecoverOrphanedWork(ctx context.Context, workerTimeoutDuration time.Duration) error {
	return nil
}

func (m *MockExecutionEngine) RecoverFailedRuns(ctx context.Context) error {
	return nil
}

// NewMockExecutionEngine creates a new mock execution engine for testing
func NewMockExecutionEngine() ExecutionEngine {
	return &MockExecutionEngine{}
}
