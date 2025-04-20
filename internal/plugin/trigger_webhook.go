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
	// Build run payload
	runPayload := map[string]interface{}{
		"versionId":   nil, // filled below
		"startNodeId": nodeID,
		"input": map[string]interface{}{
			"triggerId":   triggerID,
			"timestamp":   time.Now().UTC().Format(time.RFC3339),
			"http_method": httpMethod,
			"headers":     headers,
			"body":        string(bodyRaw),
		},
	}
	// Get latest agent version
	var versionID sql.NullString
	if err := db.DB.QueryRow(`SELECT latest_version_id FROM agents WHERE id=$1`, agentID).Scan(&versionID); err != nil {
		return nil, err
	}
	if !versionID.Valid {
		return nil, fmt.Errorf("no version for agent %s", agentID)
	}
	runPayload["versionId"] = versionID.String
	// Persist run
	raw, err := json.Marshal(runPayload)
	if err != nil {
		return nil, err
	}
	runID := uuid.New().String()
	if _, err := db.DB.Exec(`INSERT INTO agent_runs (id, agent_id, payload) VALUES ($1,$2,$3::jsonb)`, runID, agentID, string(raw)); err != nil {
		return nil, err
	}
	return runID, nil
}

func init() {
	Register(webhookTriggerPlugin{})
}
