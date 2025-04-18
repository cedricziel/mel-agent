package api

import (
    "database/sql"
    "encoding/json"
    "net/http"
    "time"

    "github.com/go-chi/chi/v5"
    "github.com/google/uuid"
    "github.com/your-org/agentsaas/internal/db"
)

// Handler returns a router with API routes mounted.
func Handler() http.Handler {
    r := chi.NewRouter()

    // Agent endpoints.
    r.Get("/agents", listAgents)
    r.Post("/agents", createAgent)
    r.Post("/agents/{agentID}/versions", createAgentVersion)

    // Connection endpoints.
    r.Get("/connections", listConnections)
    r.Post("/connections", createConnection)

    // Integrations (read‑only)
    r.Get("/integrations", listIntegrations)
    // Node type definitions for builder
    r.Get("/node-types", listNodeTypes)
    // WebSocket for collaborative updates
    r.Get("/ws/agents/{agentID}", wsHandler)
    // Test-run an agent workflow
    r.Post("/agents/{agentID}/runs/test", testRunHandler)
    // Fetch latest version (graph) for an agent
    r.Get("/agents/{agentID}/versions/latest", getLatestAgentVersionHandler)
    // Execute a single node with provided input data (stub implementation)
    r.Post("/agents/{agentID}/nodes/{nodeID}/execute", executeNodeHandler)

    return r
}

// Helper to write JSON responses.
func writeJSON(w http.ResponseWriter, status int, v any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    _ = json.NewEncoder(w).Encode(v)
}

// --- Agent handlers ---

type Agent struct {
    ID          string  `db:"id" json:"id"`
    UserID      string  `db:"user_id" json:"user_id"`
    Name        string  `db:"name" json:"name"`
    Description *string `json:"description,omitempty"`
    IsActive    bool    `db:"is_active" json:"is_active"`
}

func listAgents(w http.ResponseWriter, r *http.Request) {
    rows, err := db.DB.Query(`SELECT id, user_id, name, description, is_active FROM agents ORDER BY created_at DESC`)
    if err != nil {
        writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
        return
    }
    defer rows.Close()

    // initialize slice to ensure JSON encodes as [] instead of null when empty
    agents := []Agent{}
    for rows.Next() {
        var a Agent
        var desc sql.NullString
        if err := rows.Scan(&a.ID, &a.UserID, &a.Name, &desc, &a.IsActive); err != nil {
            writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
            return
        }
        if desc.Valid {
            a.Description = &desc.String
        }
        agents = append(agents, a)
    }
    writeJSON(w, http.StatusOK, agents)
}

func createAgent(w http.ResponseWriter, r *http.Request) {
    var in struct {
        UserID      string `json:"user_id"`
        Name        string `json:"name"`
        Description string `json:"description"`
    }
    if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
        writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
        return
    }
    if in.UserID == "" {
        // until auth is added default to a stub user
        in.UserID = "00000000-0000-0000-0000-000000000001"
    }

    var id string
    query := `INSERT INTO agents (user_id, name, description) VALUES ($1,$2,$3) RETURNING id`
    if err := db.DB.QueryRow(query, in.UserID, in.Name, sql.NullString{String: in.Description, Valid: in.Description != ""}).Scan(&id); err != nil {
        writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
        return
    }
    writeJSON(w, http.StatusCreated, map[string]string{"id": id})
}

// --- Agent version handlers ---

func createAgentVersion(w http.ResponseWriter, r *http.Request) {
    agentID := chi.URLParam(r, "agentID")
    // Body: {"semantic_version":"1.0.0","graph":{...},"default_params":{}}
    var in struct {
        SemanticVersion string          `json:"semantic_version"`
        Graph           json.RawMessage `json:"graph"`
        DefaultParams   json.RawMessage `json:"default_params"`
    }
    if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
        writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
        return
    }
    if in.SemanticVersion == "" {
        in.SemanticVersion = "0.1.0"
    }
    var versionID string
    err := db.Tx(func(tx *sql.Tx) error {
        if err := tx.QueryRow(`INSERT INTO agent_versions (agent_id, semantic_version, graph, default_params) VALUES ($1,$2,$3::jsonb,$4::jsonb) RETURNING id`, agentID, in.SemanticVersion, string(in.Graph), string(in.DefaultParams)).Scan(&versionID); err != nil {
            return err
        }
        if _, err := tx.Exec(`UPDATE agents SET latest_version_id=$1 WHERE id=$2`, versionID, agentID); err != nil {
            return err
        }
        return nil
    })
    if err != nil {
        writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
        return
    }

    writeJSON(w, http.StatusCreated, map[string]string{"id": versionID})
}

// --- Connection handlers ---

type Connection struct {
    ID            string `db:"id"             json:"id"`
    IntegrationID string `db:"integration_id" json:"integration_id"`
    UserID        string `db:"user_id"        json:"user_id"`
    Name          string `db:"name"          json:"name"`
    IsDefault     bool   `db:"is_default"    json:"is_default"`
}

type Integration struct {
    ID   string `db:"id" json:"id"`
    Name string `db:"name" json:"name"`
}

func listIntegrations(w http.ResponseWriter, r *http.Request) {
    // initialize slice to ensure JSON encodes as [] instead of null when empty
    integrations := []Integration{}
    rows, err := db.DB.Query(`SELECT id, name FROM integrations ORDER BY name`)
    if err != nil {
        writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
        return
    }
    defer rows.Close()
    for rows.Next() {
        var i Integration
        if err := rows.Scan(&i.ID, &i.Name); err != nil {
            writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
            return
        }
        integrations = append(integrations, i)
    }
    writeJSON(w, http.StatusOK, integrations)
}

func listConnections(w http.ResponseWriter, r *http.Request) {
    // Query connections and scan manually since using database/sql
    rows, err := db.DB.Query(
        `SELECT id, integration_id, user_id, name, is_default
         FROM connections
         ORDER BY created_at DESC`,
    )
    if err != nil {
        writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
        return
    }
    defer rows.Close()

    // initialize slice to ensure JSON encodes as [] instead of null when empty
    conns := []Connection{}
    for rows.Next() {
        var c Connection
        if err := rows.Scan(&c.ID, &c.IntegrationID, &c.UserID, &c.Name, &c.IsDefault); err != nil {
            writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
            return
        }
        conns = append(conns, c)
    }
    writeJSON(w, http.StatusOK, conns)
}

func createConnection(w http.ResponseWriter, r *http.Request) {
    var in struct {
        IntegrationID string          `json:"integration_id"`
        UserID        string          `json:"user_id"`
        Name          string          `json:"name"`
        Secret        json.RawMessage `json:"secret"` // kept opaque
        Config        json.RawMessage `json:"config"`
        IsDefault     bool            `json:"is_default"`
    }
    if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
        writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
        return
    }
    if in.UserID == "" {
        in.UserID = "00000000-0000-0000-0000-000000000001"
    }
    var id string
    query := `INSERT INTO connections (user_id,integration_id,name,secret,config,is_default) VALUES ($1,$2,$3,$4::jsonb,$5::jsonb,$6) RETURNING id`
    if err := db.DB.QueryRow(query, in.UserID, in.IntegrationID, in.Name, string(in.Secret), string(in.Config), in.IsDefault).Scan(&id); err != nil {
        writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
        return
    }
    writeJSON(w, http.StatusCreated, map[string]string{"id": id})
}
// executeNodeHandler handles running a single node with provided input (stub implementation)
func executeNodeHandler(w http.ResponseWriter, r *http.Request) {
    agentID := chi.URLParam(r, "agentID")
    nodeID := chi.URLParam(r, "nodeID")
    var input interface{}
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
        return
    }
    // TODO: integrate real execution engine. For now, echo input as output.
    result := map[string]interface{}{"agent_id": agentID, "node_id": nodeID, "output": input}
    writeJSON(w, http.StatusOK, result)
}
// getLatestAgentVersionHandler returns the latest saved graph for an agent.
func getLatestAgentVersionHandler(w http.ResponseWriter, r *http.Request) {
    agentID := chi.URLParam(r, "agentID")
    var versionID sql.NullString
    err := db.DB.QueryRow(`SELECT latest_version_id FROM agents WHERE id = $1`, agentID).Scan(&versionID)
    if err != nil {
        if err == sql.ErrNoRows {
            writeJSON(w, http.StatusNotFound, map[string]string{"error": "agent not found"})
        } else {
            writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
        }
        return
    }
    if !versionID.Valid {
        // no version saved yet: return empty graph
        writeJSON(w, http.StatusOK, map[string]interface{}{"graph": map[string]interface{}{"nodes": []interface{}{}, "edges": []interface{}{}}, "default_params": map[string]interface{}{}})
        return
    }
    var semanticVersion string
    var graphRaw, defaultRaw []byte
    err = db.DB.QueryRow(
        `SELECT semantic_version, graph, default_params FROM agent_versions WHERE id = $1`, versionID.String,
    ).Scan(&semanticVersion, &graphRaw, &defaultRaw)
    if err != nil {
        writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
        return
    }
    var graph interface{}
    if err := json.Unmarshal(graphRaw, &graph); err != nil {
        writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
        return
    }
    var defaultParams interface{}
    if len(defaultRaw) > 0 {
        if err := json.Unmarshal(defaultRaw, &defaultParams); err != nil {
            writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
            return
        }
    } else {
        defaultParams = map[string]interface{}{}
    }
    writeJSON(w, http.StatusOK, map[string]interface{}{ 
        "id": versionID.String,
        "semantic_version": semanticVersion,
        "graph": graph,
        "default_params": defaultParams,
    })
}
// testRunHandler executes the entire agent workflow sequentially.
func testRunHandler(w http.ResponseWriter, r *http.Request) {
    agentID := chi.URLParam(r, "agentID")
    // Fetch latest version ID
    var versionID sql.NullString
    if err := db.DB.QueryRow(`SELECT latest_version_id FROM agents WHERE id = $1`, agentID).Scan(&versionID); err != nil {
        writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
        return
    }
    if !versionID.Valid {
        writeJSON(w, http.StatusBadRequest, map[string]string{"error": "no version available for agent"})
        return
    }
    // Load graph and default params
    var graphRaw, defaultRaw []byte
    if err := db.DB.QueryRow(
        `SELECT graph, default_params FROM agent_versions WHERE id = $1`, versionID.String,
    ).Scan(&graphRaw, &defaultRaw); err != nil {
        writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
        return
    }
    // Unmarshal graph nodes
    var graph struct {
        Nodes []Node `json:"nodes"`
        // Edges ignored for linear execution
    }
    if err := json.Unmarshal(graphRaw, &graph); err != nil {
        writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
        return
    }
    // Unmarshal default params into initial data
    var defaultParams map[string]interface{}
    if len(defaultRaw) > 0 {
        if err := json.Unmarshal(defaultRaw, &defaultParams); err != nil {
            writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
            return
        }
    } else {
        defaultParams = map[string]interface{}{}
    }
    // Initialize execution payload
    type Item struct {
        ID   string      `json:"id"`
        Data interface{} `json:"data"`
    }
    type Payload struct {
        RunID   string                 `json:"runId"`
        Items   []Item                 `json:"items"`
        Context map[string]interface{} `json:"context"`
        Meta    map[string]interface{} `json:"meta"`
    }
    runID := uuid.NewString()
    payload := Payload{
        RunID:   runID,
        Items:   []Item{{ID: uuid.NewString(), Data: defaultParams}},
        Context: map[string]interface{}{},
        Meta: map[string]interface{}{
            "startTime": time.Now().UTC().Format(time.RFC3339),
        },
    }
    // Execute nodes in order
    for _, node := range graph.Nodes {
        // Process each item through the node
        var nextItems []Item
        for _, item := range payload.Items {
            output, err := executeNodeLocal(agentID, node, item.Data)
            if err != nil {
                // propagate error in data field
                nextItems = append(nextItems, Item{ID: item.ID, Data: map[string]interface{}{"error": err.Error()}})
            } else if output != nil {
                nextItems = append(nextItems, Item{ID: item.ID, Data: output})
            }
            // if output is nil, item is filtered out
        }
        payload.Items = nextItems
        payload.Meta["lastNode"] = node.ID
    }
    writeJSON(w, http.StatusOK, payload)
}

// executeNodeLocal invokes the node logic via registered executors.
func executeNodeLocal(agentID string, node Node, input interface{}) (interface{}, error) {
    executor := getExecutor(node.Type)
    return executor.Execute(agentID, node, input)
}
