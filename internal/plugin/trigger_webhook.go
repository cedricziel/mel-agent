package plugin

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/cedricziel/mel-agent/internal/db"
	"github.com/cedricziel/mel-agent/pkg/execution"
)

// webhookTriggerPlugin handles HTTP webhook triggers.
type webhookTriggerPlugin struct{}

// Meta describes the webhook trigger configuration schema.
func (webhookTriggerPlugin) Meta() PluginMeta {
	return PluginMeta{
		ID:         "webhook",
		Version:    "0.1.0",
		Categories: []string{"trigger"},
		Params: []ParamSpec{
			{Name: "method", Label: "HTTP Method", Type: "enum", Required: true, Default: "POST", Options: []string{"ANY", "GET", "POST", "PUT", "PATCH", "DELETE"}, Group: "HTTP", Description: "Allowed HTTP method"},
			{Name: "secret", Label: "Secret", Type: "string", Required: false, Group: "Security", Description: "HMAC or token for request validation"},
			{Name: "mode", Label: "Mode", Type: "enum", Required: true, Default: "async", Options: []string{"async", "sync"}, Group: "Execution", Description: "Async enqueue or Sync inline"},
			{Name: "statusCode", Label: "Response Status", Type: "number", Required: false, Default: 200, Group: "Response", Description: "HTTP status code (sync)"},
			{Name: "responseBody", Label: "Response Body", Type: "string", Required: false, Default: "", Group: "Response", Description: "HTTP body (sync)"},
		},
	}
}

// OnTrigger fires when a webhook event arrives.
// payload must include: trigger_id, agent_id, node_id, http_method, headers, body
func (webhookTriggerPlugin) OnTrigger(ctx context.Context, payload interface{}) (interface{}, error) {
	data, ok := payload.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("webhook trigger: invalid payload")
	}
	triggerID, _ := data["trigger_id"].(string)
	agentID, _ := data["agent_id"].(string)
	nodeID, _ := data["node_id"].(string)
	httpMethod, _ := data["http_method"].(string)
	headers, _ := data["headers"].(map[string][]string)
	bodyRaw, _ := data["body_raw"].([]byte)
	
	// Load trigger config for validation
	var cfgRaw []byte
	row := db.DB.QueryRow(`SELECT config FROM triggers WHERE id=$1`, triggerID)
	if err := row.Scan(&cfgRaw); err != nil {
		return nil, err
	}
	var cfg map[string]interface{}
	if err := json.Unmarshal(cfgRaw, &cfg); err != nil {
		cfg = map[string]interface{}{}
	}
	
	// Validate HTTP method
	methodAllowed, _ := cfg["method"].(string)
	if methodAllowed == "" {
		methodAllowed = "ANY"
	}
	if methodAllowed != "ANY" && !strings.EqualFold(httpMethod, methodAllowed) {
		return nil, fmt.Errorf("method not allowed")
	}
	
	// Optionally validate secret
	if secret, ok := cfg["secret"].(string); ok && secret != "" {
		token, _ := data["secret"].(string)
		if token != secret {
			return nil, fmt.Errorf("invalid secret")
		}
	}
	
	// Get latest agent version
	var versionID sql.NullString
	if err := db.DB.QueryRow(`SELECT latest_version_id FROM agents WHERE id=$1`, agentID).Scan(&versionID); err != nil {
		return nil, err
	}
	if !versionID.Valid {
		return nil, fmt.Errorf("no version for agent %s", agentID)
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
		"http_method": httpMethod,
		"headers":     headers,
		"body":        string(bodyRaw),
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
	
	// Use execution engine to start the run
	// Note: This requires access to the execution engine instance
	// For now, we'll insert directly to workflow_runs and queue tables
	
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
	Register(webhookTriggerPlugin{})
}
