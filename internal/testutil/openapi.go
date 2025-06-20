package testutil

import (
	"context"
	"database/sql"
	"testing"
	"time"

	apiPkg "github.com/cedricziel/mel-agent/pkg/api"
	"github.com/cedricziel/mel-agent/pkg/execution"
	_ "github.com/cedricziel/mel-agent/pkg/nodes"
	"github.com/google/uuid"
)

// MockExecutionEngine is a simple mock implementation of execution.Engine for testing
type MockExecutionEngine struct{}

func (m *MockExecutionEngine) StartRun(ctx context.Context, run *execution.WorkflowRun) error {
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

func (m *MockExecutionEngine) ExecuteStep(ctx context.Context, step *execution.WorkflowStep) (*apiPkg.Envelope[any], error) {
	return &apiPkg.Envelope[any]{}, nil
}

func (m *MockExecutionEngine) RetryStep(ctx context.Context, stepID uuid.UUID) error {
	return nil
}

func (m *MockExecutionEngine) RegisterWorker(ctx context.Context, worker *execution.WorkflowWorker) error {
	return nil
}

func (m *MockExecutionEngine) UnregisterWorker(ctx context.Context, workerID string) error {
	return nil
}

func (m *MockExecutionEngine) UpdateWorkerHeartbeat(ctx context.Context, workerID string) error {
	return nil
}

func (m *MockExecutionEngine) ClaimWork(ctx context.Context, workerID string, maxItems int) ([]*execution.QueueItem, error) {
	return []*execution.QueueItem{}, nil
}

func (m *MockExecutionEngine) CompleteWork(ctx context.Context, workerID string, itemID uuid.UUID, result *execution.WorkResult) error {
	return nil
}

func (m *MockExecutionEngine) RecoverOrphanedWork(ctx context.Context, workerTimeoutDuration time.Duration) error {
	return nil
}

func (m *MockExecutionEngine) RecoverFailedRuns(ctx context.Context) error {
	return nil
}

// SetupOpenAPITestDB sets up a test database and returns a mock execution engine.
// The caller needs to create the router to avoid import cycles.
// Uses clean database without pre-inserted test data for OpenAPI tests.
func SetupOpenAPITestDB(t *testing.T) (*sql.DB, execution.ExecutionEngine, func()) {
	ctx := context.Background()
	_, db, cleanup := SetupPostgresWithMigrations(ctx, t)

	// Create mock execution engine
	mockEngine := &MockExecutionEngine{}

	return db, mockEngine, cleanup
}

// Helper functions for tests

// StringPtr returns a pointer to the given string
func StringPtr(s string) *string {
	return &s
}

// IntPtr returns a pointer to the given int
func IntPtr(i int) *int {
	return &i
}

// BoolPtr returns a pointer to the given bool
func BoolPtr(b bool) *bool {
	return &b
}
