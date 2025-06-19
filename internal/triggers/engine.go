package triggers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"

	"github.com/cedricziel/mel-agent/internal/db"
	"github.com/cedricziel/mel-agent/internal/plugin"
	"github.com/cedricziel/mel-agent/pkg/execution"
)

// Engine schedules and fires trigger providers based on persisted trigger instances.
type Engine struct {
	scheduler *cron.Cron
	mu        sync.Mutex
	jobs      map[string]cron.EntryID
}

// NewEngine creates a new trigger Engine.
func NewEngine() *Engine {
	return &Engine{
		scheduler: cron.New(),
		jobs:      make(map[string]cron.EntryID),
	}
}

// Start begins the scheduler and watches for trigger changes.
func (e *Engine) Start(ctx context.Context) {
	e.scheduler.Start()
	go e.watch(ctx)
}

// watch polls the triggers table periodically to sync jobs.
func (e *Engine) watch(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	e.sync()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			e.sync()
		}
	}
}

// sync loads schedule triggers and updates scheduler jobs.
func (e *Engine) sync() {
	// Load all triggers (any provider)
	rows, err := db.DB.Query(
		`SELECT id, provider, agent_id, node_id, config, enabled FROM triggers`,
	)
	if err != nil {
		log.Printf("trigger engine sync error: %v", err)
		return
	}
	defer rows.Close()
	current := map[string]struct{}{}
	for rows.Next() {
		var id, provider, agentID, nodeID string
		var configRaw []byte
		var enabled bool
		if err := rows.Scan(&id, &provider, &agentID, &nodeID, &configRaw, &enabled); err != nil {
			log.Printf("trigger engine scan error: %v", err)
			continue
		}
		// Remove disabled triggers
		if !enabled {
			e.removeJob(id)
			continue
		}
		// Lookup trigger plugin for this provider
		p, ok := plugin.GetTriggerPlugin(provider)
		if !ok {
			// No plugin registered for this provider
			continue
		}
		// Only schedule plugins that declare a "cron" parameter
		meta := p.Meta()
		hasCron := false
		for _, ps := range meta.Params {
			if ps.Name == "cron" {
				hasCron = true
				break
			}
		}
		if !hasCron {
			// Not a schedule-type trigger; skip scheduling here
			continue
		}
		// Parse trigger config
		var cfg map[string]interface{}
		if err := json.Unmarshal(configRaw, &cfg); err != nil {
			log.Printf("trigger engine unmarshal config error for %s: %v", id, err)
			continue
		}
		// Extract cron schedule spec
		cronSpec, _ := cfg["cron"].(string)
		if cronSpec == "" {
			log.Printf("trigger engine missing cron spec for %s", id)
			continue
		}
		current[id] = struct{}{}
		// Build payload for plugin
		payload := map[string]interface{}{
			"trigger_id": id,
			"agent_id":   agentID,
			"node_id":    nodeID,
			"config":     cfg,
		}
		// Schedule the trigger
		e.mu.Lock()
		if _, exists := e.jobs[id]; !exists {
			entryID, err := e.scheduler.AddFunc(cronSpec, func() {
				if _, err := p.OnTrigger(context.Background(), payload); err != nil {
					log.Printf("trigger plugin OnTrigger error for %s: %v", id, err)
				}
			})
			if err != nil {
				log.Printf("trigger engine add job error for %s: %v", id, err)
			} else {
				e.jobs[id] = entryID
				log.Printf("trigger engine scheduled %s with cron %s", id, cronSpec)
			}
		}
		e.mu.Unlock()
	}
	e.mu.Lock()
	for id := range e.jobs {
		if _, ok := current[id]; !ok {
			e.removeJob(id)
		}
	}
	e.mu.Unlock()
}

// addJob schedules a new cron job for the given trigger.
func (e *Engine) addJob(id, agentID, nodeID, cronSpec string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if _, exists := e.jobs[id]; exists {
		return
	}
	entryID, err := e.scheduler.AddFunc(cronSpec, func() { e.fireTrigger(id, agentID, nodeID) })
	if err != nil {
		log.Printf("trigger engine add job error for %s: %v", id, err)
		return
	}
	e.jobs[id] = entryID
	log.Printf("trigger engine scheduled %s with cron %s", id, cronSpec)
}

// removeJob stops the cron job for the given trigger.
func (e *Engine) removeJob(id string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if entryID, exists := e.jobs[id]; exists {
		e.scheduler.Remove(entryID)
		delete(e.jobs, id)
		log.Printf("trigger engine removed %s", id)
	}
}

// fireTrigger handles the trigger firing: records check and creates a run.
func (e *Engine) fireTrigger(triggerID, agentID, nodeID string) {
	if err := e.fireTriggerWithTransaction(triggerID, agentID, nodeID); err != nil {
		log.Printf("trigger engine failed to fire trigger %s: %v", triggerID, err)
	}
}

// fireTriggerWithTransaction handles the trigger firing with proper transaction management.
func (e *Engine) fireTriggerWithTransaction(triggerID, agentID, nodeID string) error {
	// get latest version for agent (read-only operation, can be done outside transaction)
	var versionID sql.NullString
	if err := db.DB.QueryRow(`SELECT latest_version_id FROM agents WHERE id = $1`, agentID).Scan(&versionID); err != nil {
		return fmt.Errorf("failed to query latest_version_id for agent %s: %w", agentID, err)
	}
	if !versionID.Valid {
		return fmt.Errorf("no version found for agent %s", agentID)
	}

	// Parse UUIDs (validation can be done outside transaction)
	agentUUID, err := uuid.Parse(agentID)
	if err != nil {
		return fmt.Errorf("invalid agent_id %s: %w", agentID, err)
	}
	versionUUID, err := uuid.Parse(versionID.String)
	if err != nil {
		return fmt.Errorf("invalid version_id %s: %w", versionID.String, err)
	}
	triggerUUID, err := uuid.Parse(triggerID)
	if err != nil {
		return fmt.Errorf("invalid trigger_id %s: %w", triggerID, err)
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

	// Marshal JSON data
	inputDataJSON, err := json.Marshal(inputData)
	if err != nil {
		return fmt.Errorf("failed to marshal input data: %w", err)
	}
	variablesJSON, err := json.Marshal(workflowRun.Variables)
	if err != nil {
		return fmt.Errorf("failed to marshal variables: %w", err)
	}
	retryPolicyJSON, err := json.Marshal(workflowRun.RetryPolicy)
	if err != nil {
		return fmt.Errorf("failed to marshal retry policy: %w", err)
	}

	// Begin transaction for atomic trigger update, workflow run creation and queueing
	ctx := context.Background()
	tx, err := db.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Will be no-op if tx.Commit() succeeds

	// Update last_checked timestamp within transaction
	if _, err := tx.ExecContext(ctx, `UPDATE triggers SET last_checked = now() WHERE id = $1`, triggerID); err != nil {
		return fmt.Errorf("failed to update last_checked timestamp: %w", err)
	}

	// Insert workflow run
	query := `
		INSERT INTO workflow_runs (
			id, agent_id, version_id, trigger_id, status, input_data, 
			variables, timeout_seconds, retry_policy
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)`

	if _, err := tx.ExecContext(ctx, query,
		runID, agentUUID, versionUUID, triggerUUID, workflowRun.Status,
		inputDataJSON, variablesJSON, workflowRun.TimeoutSeconds, retryPolicyJSON); err != nil {
		return fmt.Errorf("failed to create workflow run: %w", err)
	}

	// Queue the run for execution
	queueItemID := uuid.New()
	queueQuery := `
		INSERT INTO workflow_queue (
			id, run_id, queue_type, priority, available_at, max_attempts, payload
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)`

	payloadJSON, err := json.Marshal(map[string]interface{}{})
	if err != nil {
		return fmt.Errorf("failed to marshal queue payload: %w", err)
	}

	if _, err := tx.ExecContext(ctx, queueQuery,
		queueItemID, runID, "start_run", 5, time.Now(), 3, payloadJSON); err != nil {
		return fmt.Errorf("failed to queue workflow run: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("trigger engine fired %s, created workflow run %s", triggerID, runID)
	return nil
}
