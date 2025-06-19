package plugin

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/cedricziel/mel-agent/internal/db"
	"github.com/cedricziel/mel-agent/pkg/execution"
	"github.com/google/uuid"
)

// scheduleTriggerPlugin implements the standard MCP "schedule" trigger via cron.
type scheduleTriggerPlugin struct{}

// Meta returns metadata for the schedule trigger plugin.
func (scheduleTriggerPlugin) Meta() PluginMeta {
	return PluginMeta{
		ID:         "schedule",
		Version:    "0.1.0",
		Categories: []string{"trigger"},
		Params: []ParamSpec{
			{Name: "cron", Label: "Cron Expression", Type: "string", Required: true, Group: "Schedule", Description: "Cron schedule to run"},
		},
		UIComponent: "",
	}
}

// OnTrigger is called by the trigger engine when a schedule triggers.
func (scheduleTriggerPlugin) OnTrigger(ctx context.Context, payload interface{}) (interface{}, error) {
	data, ok := payload.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("schedule trigger: invalid payload type")
	}
	triggerID, _ := data["trigger_id"].(string)
	agentID, _ := data["agent_id"].(string)
	nodeID, _ := data["node_id"].(string)

	// Update last_checked timestamp
	if _, err := db.DB.Exec(`UPDATE triggers SET last_checked = now() WHERE id = $1`, triggerID); err != nil {
		log.Printf("schedule trigger update last_checked error: %v", err)
	}

	// Fetch latest agent version
	var versionID sql.NullString
	if err := db.DB.QueryRow(`SELECT latest_version_id FROM agents WHERE id = $1`, agentID).Scan(&versionID); err != nil {
		return nil, err
	}
	if !versionID.Valid {
		return nil, fmt.Errorf("schedule trigger: no version for agent %s", agentID)
	}

	// Parse UUIDs
	agentUUID, err := uuid.Parse(agentID)
	if err != nil {
		return nil, fmt.Errorf("invalid agent_id: %w", err)
	}
	versionUUID, err := uuid.Parse(versionID.String)
	if err != nil {
		return nil, fmt.Errorf("invalid version_id: %w", err)
	}
	triggerUUID, err := uuid.Parse(triggerID)
	if err != nil {
		return nil, fmt.Errorf("invalid trigger_id: %w", err)
	}

	// Build workflow input data
	inputData := map[string]interface{}{
		"triggerId":   triggerID,
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
		"startNodeId": nodeID,
	}

	// Create workflow run using durable execution system
	runID := uuid.New()
	workflowRun := &execution.WorkflowRun{
		ID:        runID,
		AgentID:   agentUUID,
		VersionID: versionUUID,
		TriggerID: &triggerUUID,
		Status:    execution.RunStatusPending,
		InputData: inputData,
		Variables: map[string]interface{}{},
		RetryPolicy: execution.RetryPolicy{
			MaxAttempts:       3,
			InitialDelayMS:    60000, // 1 minute
			BackoffMultiplier: 2.0,
			MaxDelayMS:        3600000, // 1 hour
		},
		TimeoutSeconds: 3600,
	}

	// Insert workflow run
	inputDataJSON, _ := json.Marshal(inputData)
	variablesJSON, _ := json.Marshal(workflowRun.Variables)
	retryPolicyJSON, _ := json.Marshal(workflowRun.RetryPolicy)

	query := `
		INSERT INTO workflow_runs (
			id, agent_id, version_id, trigger_id, status, input_data, 
			variables, timeout_seconds, retry_policy
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)`

	if _, err := db.DB.ExecContext(ctx, query,
		runID, agentUUID, versionUUID, triggerUUID, workflowRun.Status,
		inputDataJSON, variablesJSON, workflowRun.TimeoutSeconds, retryPolicyJSON); err != nil {
		return nil, fmt.Errorf("failed to create workflow run: %w", err)
	}

	// Queue the run for execution
	queueItemID := uuid.New()
	queueQuery := `
		INSERT INTO workflow_queue (
			id, run_id, queue_type, priority, available_at, max_attempts, payload
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)`

	payloadJSON, _ := json.Marshal(map[string]interface{}{})
	if _, err := db.DB.ExecContext(ctx, queueQuery,
		queueItemID, runID, "start_run", 5, time.Now(), 3, payloadJSON); err != nil {
		return nil, fmt.Errorf("failed to queue workflow run: %w", err)
	}

	return runID.String(), nil
}

func init() {
	Register(scheduleTriggerPlugin{})
}
