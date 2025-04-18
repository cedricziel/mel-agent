package runs

import (
   "database/sql"
   "encoding/json"
   "fmt"
   "log"
   "time"

   "github.com/cedricziel/mel-agent/internal/api"
   "github.com/cedricziel/mel-agent/internal/db"
)

// Runner polls agent_runs and executes pending runs.
type Runner struct {
   interval time.Duration
}

// NewRunner creates a new Runner with default polling interval.
func NewRunner() *Runner {
   return &Runner{interval: time.Second}
}

// Start begins the runner loop in a background goroutine.
func (r *Runner) Start() {
   go r.loop()
}

// loop continuously processes pending runs.
func (r *Runner) loop() {
   for {
       for {
           runID, agentID, payload := r.fetchPending()
           if runID == "" {
               break
           }
           r.runOne(runID, agentID, payload)
       }
       time.Sleep(r.interval)
   }
}

// fetchPending selects one pending run and marks it as running.
func (r *Runner) fetchPending() (string, string, []byte) {
   tx, err := db.DB.Begin()
   if err != nil {
       log.Printf("runner begin tx error: %v", err)
       return "", "", nil
   }
   defer tx.Rollback()
   var runID, agentID string
   var payload []byte
   err = tx.QueryRow(
       `SELECT id, agent_id, payload
        FROM agent_runs
        WHERE status = 'pending'
        ORDER BY created_at
        FOR UPDATE SKIP LOCKED
        LIMIT 1`,
   ).Scan(&runID, &agentID, &payload)
   if err != nil {
       if err != sql.ErrNoRows {
           log.Printf("runner fetch pending error: %v", err)
       }
       return "", "", nil
   }
   if _, err := tx.Exec(`UPDATE agent_runs SET status = 'running' WHERE id = $1`, runID); err != nil {
       log.Printf("runner mark running error for %s: %v", runID, err)
       return "", "", nil
   }
   if err := tx.Commit(); err != nil {
       log.Printf("runner commit error: %v", err)
       return "", "", nil
   }
   return runID, agentID, payload
}

// runOne executes the workflow for a single run.
func (r *Runner) runOne(runID, agentID string, payloadRaw []byte) {
   // Parse run payload
   type RunPayload struct {
       VersionID   string      `json:"versionId"`
       StartNodeID string      `json:"startNodeId"`
       Input       interface{} `json:"input"`
   }
   // Parse run payload
   var in RunPayload
   if err := json.Unmarshal(payloadRaw, &in); err != nil {
       log.Printf("runner unmarshal payload error for run %s: %v", runID, err)
       r.markFailed(runID, err)
       return
   }
   // Skip runs with missing versionId
   if in.VersionID == "" {
       err := fmt.Errorf("missing versionId in run payload")
       log.Printf("runner missing versionId for run %s", runID)
       r.markFailed(runID, err)
       return
   }
   // Fetch graph for version
   var graphRaw []byte
   err := db.DB.QueryRow(`SELECT graph FROM agent_versions WHERE id = $1`, in.VersionID).Scan(&graphRaw)
   if err != nil {
       log.Printf("runner fetch graph error for run %s: %v", runID, err)
       r.markFailed(runID, err)
       return
   }
   // Unmarshal graph nodes and edges
   type Graph struct {
       Nodes []api.Node `json:"nodes"`
       Edges []struct {
           Source       string `json:"source"`
           Target       string `json:"target"`
           SourceHandle string `json:"sourceHandle,omitempty"`
           TargetHandle string `json:"targetHandle,omitempty"`
       } `json:"edges"`
   }
   var graph Graph
   if err := json.Unmarshal(graphRaw, &graph); err != nil {
       log.Printf("runner unmarshal graph error for run %s: %v", runID, err)
       r.markFailed(runID, err)
       return
   }
   // Build adjacency list
   children := make(map[string][]string)
   for _, e := range graph.Edges {
       children[e.Source] = append(children[e.Source], e.Target)
   }
   // Define execution step
   type Step struct {
       NodeID string      `json:"nodeId"`
       Input  interface{} `json:"input"`
       Output interface{} `json:"output"`
   }
   var trace []Step
   // BFS/DFS queue
   type QueueItem struct {
       NodeID string
       Data   interface{}
   }
   queue := []QueueItem{{NodeID: in.StartNodeID, Data: in.Input}}
   // Execute nodes
   for len(queue) > 0 {
       item := queue[0]
       queue = queue[1:]
       // Find node in graph
       var node api.Node
       for _, n := range graph.Nodes {
           if n.ID == item.NodeID {
               node = n
               break
           }
       }
       // Notify start
       hub := api.GetHub(agentID)
       if msg, err := json.Marshal(map[string]string{"type": "nodeExecution", "nodeId": node.ID, "phase": "start"}); err == nil {
           hub.Broadcast(msg)
       }
       // Execute node
       def := api.FindDefinition(node.Type)
       var output interface{}
       if def != nil {
           out, err := def.Execute(agentID, node, item.Data)
           if err != nil {
               log.Printf("runner execute node error for run %s, node %s: %v", runID, node.ID, err)
           }
           output = out
       }
       // Notify end
       if msg, err := json.Marshal(map[string]string{"type": "nodeExecution", "nodeId": node.ID, "phase": "end"}); err == nil {
           hub.Broadcast(msg)
       }
       // Record step
       trace = append(trace, Step{NodeID: node.ID, Input: item.Data, Output: output})
       // Enqueue children
       if output != nil {
           for _, child := range children[node.ID] {
               queue = append(queue, QueueItem{NodeID: child, Data: output})
           }
       }
   }
   // Prepare result payload
   result := struct {
       RunID string `json:"runId"`
       Trace []Step `json:"trace"`
   }{
       RunID: runID,
       Trace: trace,
   }
   rawResult, err := json.Marshal(result)
   if err != nil {
       log.Printf("runner marshal result error for run %s: %v", runID, err)
       r.markFailed(runID, err)
       return
   }
   // Persist completed run
   if _, err := db.DB.Exec(`UPDATE agent_runs SET payload = $2::jsonb, status = 'completed' WHERE id = $1`, runID, string(rawResult)); err != nil {
       log.Printf("runner update run record error for run %s: %v", runID, err)
   }
}

// markFailed updates the run record with failure status and error info.
func (r *Runner) markFailed(runID string, execErr error) {
   errInfo := map[string]interface{}{"error": execErr.Error()}
   payload := struct {
       RunID string                 `json:"runId"`
       Error map[string]interface{} `json:"error"`
   }{
       RunID: runID,
       Error: errInfo,
   }
   raw, _ := json.Marshal(payload)
   if _, err := db.DB.Exec(`UPDATE agent_runs SET payload = $2::jsonb, status = 'failed' WHERE id = $1`, runID, string(raw)); err != nil {
       log.Printf("runner markFailed update error for run %s: %v", runID, err)
   }
}