package plugin

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cedricziel/mel-agent/internal/db"
	"github.com/cedricziel/mel-agent/internal/testutil"
)

func TestWebhookTriggerPlugin_Meta(t *testing.T) {
	plugin := webhookTriggerPlugin{}
	meta := plugin.Meta()

	assert.Equal(t, "webhook", meta.ID)
	assert.Equal(t, "0.1.0", meta.Version)
	assert.Contains(t, meta.Categories, "trigger")

	// Check that expected parameters exist
	paramNames := make(map[string]bool)
	for _, param := range meta.Params {
		paramNames[param.Name] = true
	}

	assert.True(t, paramNames["method"])
	assert.True(t, paramNames["secret"])
	assert.True(t, paramNames["mode"])
}

func TestWebhookTriggerPlugin_OnTrigger_Integration(t *testing.T) {
	ctx := context.Background()
	_, testDB, cleanup := testutil.SetupPostgresWithMigrations(ctx, t)
	defer cleanup()

	oldDB := db.DB
	db.DB = testDB
	defer func() { db.DB = oldDB }()

	plugin := webhookTriggerPlugin{}

	t.Run("creates workflow run and queue item", func(t *testing.T) {
		// Setup test data
		userID := uuid.New()
		agentID := uuid.New()
		versionID := uuid.New()
		triggerID := uuid.New()
		nodeID := "test-node-1"

		// Insert test user first
		_, err := testDB.Exec(`
			INSERT INTO users (id, email, created_at) 
			VALUES ($1, $2, NOW())
		`, userID, "test@example.com")
		require.NoError(t, err)

		// Insert test agent
		_, err = testDB.Exec(`
			INSERT INTO agents (id, user_id, name, latest_version_id) 
			VALUES ($1, $2, $3, $4)
		`, agentID, userID, "Test Agent", versionID)
		require.NoError(t, err)

		// Insert test trigger with config
		triggerConfig := map[string]interface{}{
			"method": "POST",
			"secret": "test-secret",
		}
		configJSON, _ := json.Marshal(triggerConfig)
		_, err = testDB.Exec(`
			INSERT INTO triggers (id, user_id, provider, agent_id, node_id, config, enabled) 
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, triggerID, userID, "webhook", agentID, nodeID, configJSON, true)
		require.NoError(t, err)

		// Prepare trigger payload
		payload := map[string]interface{}{
			"trigger_id":  triggerID.String(),
			"agent_id":    agentID.String(),
			"node_id":     nodeID,
			"http_method": "POST",
			"headers":     map[string][]string{"Content-Type": {"application/json"}},
			"body_raw":    []byte(`{"test":"data"}`),
			"secret":      "test-secret",
		}

		// Execute trigger
		ctx := context.Background()
		result, err := plugin.OnTrigger(ctx, payload)
		require.NoError(t, err)
		require.NotNil(t, result)

		runID := result.(string)
		runUUID, err := uuid.Parse(runID)
		require.NoError(t, err)

		// Verify workflow run was created
		var count int
		err = testDB.QueryRow(`SELECT COUNT(*) FROM workflow_runs WHERE id = $1`, runUUID).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// Verify queue item was created
		err = testDB.QueryRow(`SELECT COUNT(*) FROM workflow_queue WHERE run_id = $1`, runUUID).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// Verify queue item has correct type
		var queueType string
		err = testDB.QueryRow(`SELECT queue_type FROM workflow_queue WHERE run_id = $1`, runUUID).Scan(&queueType)
		require.NoError(t, err)
		assert.Equal(t, "start_run", queueType)
	})
}

func TestScheduleTriggerPlugin_Meta(t *testing.T) {
	plugin := scheduleTriggerPlugin{}
	meta := plugin.Meta()

	assert.Equal(t, "schedule", meta.ID)
	assert.Equal(t, "0.1.0", meta.Version)
	assert.Contains(t, meta.Categories, "trigger")

	// Check for cron parameter
	found := false
	for _, param := range meta.Params {
		if param.Name == "cron" {
			found = true
			assert.True(t, param.Required)
			assert.Equal(t, "Cron Expression", param.Label)
			break
		}
	}
	assert.True(t, found, "cron parameter should exist")
}

func TestScheduleTriggerPlugin_OnTrigger_Integration(t *testing.T) {
	ctx := context.Background()
	_, testDB, cleanup := testutil.SetupPostgresWithMigrations(ctx, t)
	defer cleanup()

	oldDB := db.DB
	db.DB = testDB
	defer func() { db.DB = oldDB }()

	plugin := scheduleTriggerPlugin{}

	t.Run("creates workflow run and updates trigger timestamp", func(t *testing.T) {
		// Setup test data
		userID := uuid.New()
		agentID := uuid.New()
		versionID := uuid.New()
		triggerID := uuid.New()
		nodeID := "schedule-node-1"

		// Insert test user first
		_, err := testDB.Exec(`
			INSERT INTO users (id, email, created_at) 
			VALUES ($1, $2, NOW())
		`, userID, "test2@example.com")
		require.NoError(t, err)

		// Insert test agent
		_, err = testDB.Exec(`
			INSERT INTO agents (id, user_id, name, latest_version_id) 
			VALUES ($1, $2, $3, $4)
		`, agentID, userID, "Test Agent", versionID)
		require.NoError(t, err)

		// Insert test trigger
		triggerConfig := map[string]interface{}{
			"cron": "0 */5 * * * *", // Every 5 minutes
		}
		configJSON, _ := json.Marshal(triggerConfig)
		_, err = testDB.Exec(`
			INSERT INTO triggers (id, user_id, provider, agent_id, node_id, config, enabled) 
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, triggerID, userID, "schedule", agentID, nodeID, configJSON, true)
		require.NoError(t, err)

		// Prepare trigger payload
		payload := map[string]interface{}{
			"trigger_id": triggerID.String(),
			"agent_id":   agentID.String(),
			"node_id":    nodeID,
		}

		// Execute trigger
		ctx := context.Background()
		result, err := plugin.OnTrigger(ctx, payload)
		require.NoError(t, err)
		require.NotNil(t, result)

		runID := result.(string)
		runUUID, err := uuid.Parse(runID)
		require.NoError(t, err)

		// Verify workflow run was created
		var count int
		err = testDB.QueryRow(`SELECT COUNT(*) FROM workflow_runs WHERE id = $1`, runUUID).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// Verify trigger was updated with last_checked timestamp
		var lastChecked *string
		err = testDB.QueryRow(`SELECT last_checked FROM triggers WHERE id = $1`, triggerID).Scan(&lastChecked)
		require.NoError(t, err)
		assert.NotNil(t, lastChecked, "last_checked should be updated")

		// Verify queue item was created
		err = testDB.QueryRow(`SELECT COUNT(*) FROM workflow_queue WHERE run_id = $1`, runUUID).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})
}

func TestTriggerPlugins_ErrorHandling(t *testing.T) {
	ctx := context.Background()
	_, testDB, cleanup := testutil.SetupPostgresWithMigrations(ctx, t)
	defer cleanup()

	oldDB := db.DB
	db.DB = testDB
	defer func() { db.DB = oldDB }()

	t.Run("webhook plugin handles invalid payload", func(t *testing.T) {
		plugin := webhookTriggerPlugin{}
		ctx := context.Background()

		_, err := plugin.OnTrigger(ctx, "invalid-payload")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid payload")
	})

	t.Run("schedule plugin handles invalid payload", func(t *testing.T) {
		plugin := scheduleTriggerPlugin{}
		ctx := context.Background()

		_, err := plugin.OnTrigger(ctx, "invalid-payload")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid payload")
	})

	t.Run("webhook plugin handles agent without version", func(t *testing.T) {
		plugin := webhookTriggerPlugin{}
		ctx := context.Background()

		userID := uuid.New()
		triggerID := uuid.New()
		agentID := uuid.New()

		// Insert test user first
		_, err := testDB.Exec(`
			INSERT INTO users (id, email, created_at) 
			VALUES ($1, $2, NOW())
		`, userID, "test3@example.com")
		require.NoError(t, err)

		// Insert agent WITHOUT latest_version_id (NULL)
		_, err = testDB.Exec(`
			INSERT INTO agents (id, user_id, name) 
			VALUES ($1, $2, $3)
		`, agentID, userID, "Test Agent Without Version")
		require.NoError(t, err)

		// Insert trigger with agent that has no version
		triggerConfig := map[string]interface{}{"method": "POST"}
		configJSON, _ := json.Marshal(triggerConfig)
		_, err = testDB.Exec(`
			INSERT INTO triggers (id, user_id, provider, agent_id, node_id, config, enabled) 
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, triggerID, userID, "webhook", agentID, "node-1", configJSON, true)
		require.NoError(t, err)

		payload := map[string]interface{}{
			"trigger_id":  triggerID.String(),
			"agent_id":    agentID.String(),
			"node_id":     "node-1",
			"http_method": "POST",
			"headers":     map[string][]string{},
			"body_raw":    []byte(``),
		}

		_, err = plugin.OnTrigger(ctx, payload)
		assert.Error(t, err) // Should fail because agent has no version
		assert.Contains(t, err.Error(), "no version for agent")
	})
}
