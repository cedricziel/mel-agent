package triggers

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cedricziel/mel-agent/internal/db"
	"github.com/cedricziel/mel-agent/internal/testutil"
)

func TestEngine_fireTriggerWithTransaction(t *testing.T) {
	// Setup test database
	ctx := context.Background()
	_, testDB, cleanup := testutil.SetupPostgresWithMigrations(ctx, t)
	defer cleanup()

	// Set global db.DB for trigger engine to use
	originalDB := db.DB
	db.DB = testDB
	defer func() { db.DB = originalDB }()

	engine := NewEngine()

	// Create test data
	agentID := uuid.New().String()
	versionID := uuid.New().String()
	triggerID := uuid.New().String()
	nodeID := "start-node"

	// Insert test agent and version
	_, err := db.DB.Exec(`
		INSERT INTO agents (id, name, description, latest_version_id) 
		VALUES ($1, 'test-agent', 'test description', $2)
	`, agentID, versionID)
	require.NoError(t, err)

	_, err = db.DB.Exec(`
		INSERT INTO agent_versions (id, agent_id, semantic_version, graph) 
		VALUES ($1, $2, 'v1.0.0', '[]')
	`, versionID, agentID)
	require.NoError(t, err)

	// Insert test trigger
	_, err = db.DB.Exec(`
		INSERT INTO triggers (id, user_id, agent_id, provider, config, last_checked) 
		VALUES ($1, '00000000-0000-0000-0000-000000000001', $2, 'schedule', '{}', NOW())
	`, triggerID, agentID)
	require.NoError(t, err)

	t.Run("successful trigger firing creates both workflow run and queue item", func(t *testing.T) {
		// Execute trigger firing
		err := engine.fireTriggerWithTransaction(triggerID, agentID, nodeID)
		require.NoError(t, err)

		// Verify workflow run was created
		var workflowRunCount int
		err = db.DB.QueryRow(`SELECT COUNT(*) FROM workflow_runs WHERE agent_id = $1`, agentID).Scan(&workflowRunCount)
		require.NoError(t, err)
		assert.Equal(t, 1, workflowRunCount)

		// Verify queue item was created
		var queueItemCount int
		err = db.DB.QueryRow(`SELECT COUNT(*) FROM workflow_queue WHERE queue_type = 'start_run'`).Scan(&queueItemCount)
		require.NoError(t, err)
		assert.Equal(t, 1, queueItemCount)

		// Verify workflow run and queue item are linked
		var runID string
		err = db.DB.QueryRow(`SELECT id FROM workflow_runs WHERE agent_id = $1`, agentID).Scan(&runID)
		require.NoError(t, err)

		var queueRunID string
		err = db.DB.QueryRow(`SELECT run_id FROM workflow_queue WHERE queue_type = 'start_run'`).Scan(&queueRunID)
		require.NoError(t, err)

		assert.Equal(t, runID, queueRunID, "workflow run and queue item should be linked")
	})

	t.Run("invalid agent ID returns error", func(t *testing.T) {
		invalidAgentID := "invalid-uuid"
		err := engine.fireTriggerWithTransaction(triggerID, invalidAgentID, nodeID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to query latest_version_id")
	})

	t.Run("non-existent agent returns error", func(t *testing.T) {
		nonExistentAgentID := uuid.New().String()
		err := engine.fireTriggerWithTransaction(triggerID, nonExistentAgentID, nodeID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to query latest_version_id")
	})

	t.Run("agent without version returns error", func(t *testing.T) {
		// Create agent without version
		agentWithoutVersionID := uuid.New().String()
		_, err := db.DB.Exec(`
			INSERT INTO agents (id, name, description, latest_version_id) 
			VALUES ($1, 'test-agent-no-version', 'test description', NULL)
		`, agentWithoutVersionID)
		require.NoError(t, err)

		err = engine.fireTriggerWithTransaction(triggerID, agentWithoutVersionID, nodeID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no version found")
	})

	t.Run("invalid trigger ID returns error", func(t *testing.T) {
		invalidTriggerID := "invalid-uuid"
		err := engine.fireTriggerWithTransaction(invalidTriggerID, agentID, nodeID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid trigger_id")
	})
}

func TestEngine_fireTriggerWithTransaction_Atomicity(t *testing.T) {
	// Setup test database
	ctx := context.Background()
	_, testDB, cleanup := testutil.SetupPostgresWithMigrations(ctx, t)
	defer cleanup()

	// Set global db.DB for trigger engine to use
	originalDB := db.DB
	db.DB = testDB
	defer func() { db.DB = originalDB }()

	engine := NewEngine()

	// Create test data
	agentID := uuid.New().String()
	versionID := uuid.New().String()
	triggerID := uuid.New().String()
	nodeID := "start-node"

	// Insert test agent and version
	_, err := db.DB.Exec(`
		INSERT INTO agents (id, name, description, latest_version_id) 
		VALUES ($1, 'test-agent', 'test description', $2)
	`, agentID, versionID)
	require.NoError(t, err)

	_, err = db.DB.Exec(`
		INSERT INTO agent_versions (id, agent_id, semantic_version, graph) 
		VALUES ($1, $2, 'v1.0.0', '[]')
	`, versionID, agentID)
	require.NoError(t, err)

	// Insert test trigger
	_, err = db.DB.Exec(`
		INSERT INTO triggers (id, user_id, agent_id, provider, config, last_checked) 
		VALUES ($1, '00000000-0000-0000-0000-000000000001', $2, 'schedule', '{}', NOW())
	`, triggerID, agentID)
	require.NoError(t, err)

	t.Run("transaction rollback prevents orphaned workflow runs", func(t *testing.T) {
		// Get initial counts and trigger state
		var initialWorkflowRunCount, initialQueueItemCount int
		var initialLastChecked sql.NullTime

		err = db.DB.QueryRow(`SELECT COUNT(*) FROM workflow_runs`).Scan(&initialWorkflowRunCount)
		require.NoError(t, err)
		err = db.DB.QueryRow(`SELECT COUNT(*) FROM workflow_queue`).Scan(&initialQueueItemCount)
		require.NoError(t, err)
		err = db.DB.QueryRow(`SELECT last_checked FROM triggers WHERE id = $1`, triggerID).Scan(&initialLastChecked)
		require.NoError(t, err)

		// Temporarily corrupt the queue table to force second insert to fail
		// This simulates a scenario where the first insert succeeds but second fails
		_, err = db.DB.Exec(`ALTER TABLE workflow_queue DROP COLUMN id`)
		require.NoError(t, err)

		// Execute trigger firing - should fail and rollback
		err = engine.fireTriggerWithTransaction(triggerID, agentID, nodeID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to queue workflow run")

		// Restore the table for verification
		_, err = db.DB.Exec(`ALTER TABLE workflow_queue ADD COLUMN id UUID`)
		require.NoError(t, err)

		// Verify complete rollback - no changes to any table
		var finalWorkflowRunCount, finalQueueItemCount int
		var finalLastChecked sql.NullTime

		err = db.DB.QueryRow(`SELECT COUNT(*) FROM workflow_runs`).Scan(&finalWorkflowRunCount)
		require.NoError(t, err)
		err = db.DB.QueryRow(`SELECT COUNT(*) FROM workflow_queue`).Scan(&finalQueueItemCount)
		require.NoError(t, err)
		err = db.DB.QueryRow(`SELECT last_checked FROM triggers WHERE id = $1`, triggerID).Scan(&finalLastChecked)
		require.NoError(t, err)

		assert.Equal(t, initialWorkflowRunCount, finalWorkflowRunCount, "no workflow runs should be created due to rollback")
		assert.Equal(t, initialQueueItemCount, finalQueueItemCount, "no queue items should be created due to rollback")
		assert.Equal(t, initialLastChecked, finalLastChecked, "trigger last_checked should not be updated due to rollback")
	})
}

func TestEngine_fireTrigger_ErrorHandling(t *testing.T) {
	// Setup test database
	ctx := context.Background()
	_, testDB, cleanup := testutil.SetupPostgresWithMigrations(ctx, t)
	defer cleanup()

	// Set global db.DB for trigger engine to use
	originalDB := db.DB
	db.DB = testDB
	defer func() { db.DB = originalDB }()

	engine := NewEngine()

	t.Run("fireTrigger logs error and continues", func(t *testing.T) {
		// This should fail due to invalid UUID
		// The main fireTrigger method should log the error but not panic
		invalidAgentID := "invalid-uuid"
		triggerID := uuid.New().String()
		nodeID := "start-node"

		// This should not panic and should handle the error gracefully
		engine.fireTrigger(triggerID, invalidAgentID, nodeID)

		// No assertion needed - the test passes if it doesn't panic
		// The error is logged in the fireTrigger method
	})
}

func TestEngine_fireTriggerWithTransaction_JSONMarshaling(t *testing.T) {
	// Setup test database
	ctx := context.Background()
	_, testDB, cleanup := testutil.SetupPostgresWithMigrations(ctx, t)
	defer cleanup()

	// Set global db.DB for trigger engine to use
	originalDB := db.DB
	db.DB = testDB
	defer func() { db.DB = originalDB }()

	engine := NewEngine()

	// Create test data
	agentID := uuid.New().String()
	versionID := uuid.New().String()
	triggerID := uuid.New().String()
	nodeID := "start-node"

	// Insert test agent and version
	_, err := db.DB.Exec(`
		INSERT INTO agents (id, name, description, latest_version_id) 
		VALUES ($1, 'test-agent', 'test description', $2)
	`, agentID, versionID)
	require.NoError(t, err)

	_, err = db.DB.Exec(`
		INSERT INTO agent_versions (id, agent_id, semantic_version, graph) 
		VALUES ($1, $2, 'v1.0.0', '[]')
	`, versionID, agentID)
	require.NoError(t, err)

	// Insert test trigger
	_, err = db.DB.Exec(`
		INSERT INTO triggers (id, user_id, agent_id, provider, config, last_checked) 
		VALUES ($1, '00000000-0000-0000-0000-000000000001', $2, 'schedule', '{}', NOW())
	`, triggerID, agentID)
	require.NoError(t, err)

	t.Run("successful JSON marshaling and storage", func(t *testing.T) {
		err := engine.fireTriggerWithTransaction(triggerID, agentID, nodeID)
		require.NoError(t, err)

		// Verify the JSON data was stored correctly
		var inputDataJSON, variablesJSON, retryPolicyJSON string
		err = db.DB.QueryRow(`
			SELECT input_data, variables, retry_policy 
			FROM workflow_runs WHERE agent_id = $1
		`, agentID).Scan(&inputDataJSON, &variablesJSON, &retryPolicyJSON)
		require.NoError(t, err)

		// Verify JSON is valid
		assert.NotEmpty(t, inputDataJSON)
		assert.NotEmpty(t, variablesJSON)
		assert.NotEmpty(t, retryPolicyJSON)

		// Verify input data contains expected fields
		assert.Contains(t, inputDataJSON, triggerID)
		assert.Contains(t, inputDataJSON, nodeID)
		assert.Contains(t, inputDataJSON, "timestamp")
	})
}
