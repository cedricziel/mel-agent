package triggers

import (
   "database/sql"
   "encoding/json"
   "log"
   "sync"
   "time"

   "github.com/google/uuid"
   "github.com/robfig/cron/v3"

   "github.com/cedricziel/mel-agent/internal/db"
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
func (e *Engine) Start() {
   e.scheduler.Start()
   go e.watch()
}

// watch polls the triggers table periodically to sync jobs.
func (e *Engine) watch() {
   ticker := time.NewTicker(time.Minute)
   defer ticker.Stop()
   e.sync()
   for range ticker.C {
       e.sync()
   }
}

// sync loads schedule triggers and updates scheduler jobs.
func (e *Engine) sync() {
   rows, err := db.DB.Query(
       `SELECT id, agent_id, node_id, provider, config, enabled FROM triggers WHERE provider = 'schedule'`,
   )
   if err != nil {
       log.Printf("trigger engine sync error: %v", err)
       return
   }
   defer rows.Close()
   current := map[string]struct{}{}
   for rows.Next() {
       var id, agentID, nodeID, provider string
       var configRaw []byte
       var enabled bool
       if err := rows.Scan(&id, &agentID, &nodeID, &provider, &configRaw, &enabled); err != nil {
           log.Printf("trigger engine scan error: %v", err)
           continue
       }
       if !enabled {
           e.removeJob(id)
           continue
       }
       var cfg map[string]interface{}
       if err := json.Unmarshal(configRaw, &cfg); err != nil {
           log.Printf("trigger engine unmarshal config error for %s: %v", id, err)
           continue
       }
       cronSpec, ok := cfg["cron"].(string)
       if !ok || cronSpec == "" {
           log.Printf("trigger engine missing cron for %s", id)
           continue
       }
       current[id] = struct{}{}
       e.addJob(id, agentID, nodeID, cronSpec)
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
   // update last_checked timestamp
   if _, err := db.DB.Exec(`UPDATE triggers SET last_checked = now() WHERE id = $1`, triggerID); err != nil {
       log.Printf("trigger engine update last_checked error: %v", err)
   }
   // get latest version for agent
   var versionID sql.NullString
   if err := db.DB.QueryRow(`SELECT latest_version_id FROM agents WHERE id = $1`, agentID).Scan(&versionID); err != nil {
       log.Printf("trigger engine query latest_version_id error: %v", err)
       return
   }
   if !versionID.Valid {
       log.Printf("trigger engine no version for agent %s", agentID)
       return
   }
   // construct payload for run
   payload := map[string]interface{}{
       "versionId":   versionID.String,
       "startNodeId": nodeID,
       "input": map[string]interface{}{
           "triggerId": triggerID,
           "timestamp": time.Now().UTC().Format(time.RFC3339),
       },
   }
   raw, err := json.Marshal(payload)
   if err != nil {
       log.Printf("trigger engine marshal payload error: %v", err)
       return
   }
   runID := uuid.New().String()
   if _, err := db.DB.Exec(`INSERT INTO agent_runs (id, agent_id, payload) VALUES ($1, $2, $3::jsonb)`, runID, agentID, string(raw)); err != nil {
       log.Printf("trigger engine insert run error: %v", err)
       return
   }
   log.Printf("trigger engine fired %s, created run %s", triggerID, runID)
}