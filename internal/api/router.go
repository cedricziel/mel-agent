package api

import (
    "database/sql"
    "encoding/json"
    "net/http"

    "github.com/go-chi/chi/v5"
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
