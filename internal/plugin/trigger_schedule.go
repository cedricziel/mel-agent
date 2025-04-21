package plugin

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/cedricziel/mel-agent/internal/db"
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
	// Build run payload
	runPayload := map[string]interface{}{
		"versionId":   versionID.String,
		"startNodeId": nodeID,
		"input": map[string]interface{}{
			"triggerId": triggerID,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		},
	}
	raw, err := json.Marshal(runPayload)
	if err != nil {
		return nil, err
	}
	runID := uuid.New().String()
	if _, err := db.DB.Exec(`INSERT INTO agent_runs (id, agent_id, payload) VALUES ($1, $2, $3::jsonb)`, runID, agentID, string(raw)); err != nil {
		return nil, err
	}
	return runID, nil
}

func init() {
	Register(scheduleTriggerPlugin{})
}
