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

    // Connection endpoints.
    r.Get("/connections", listConnections)
    r.Post("/connections", createConnection)

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
    ID          string         `db:"id"          json:"id"`
    UserID      string         `db:"user_id"      json:"user_id"`
    Name        string         `db:"name"         json:"name"`
    Description sql.NullString `db:"description"  json:"description,omitempty"`
    IsActive    bool           `db:"is_active"    json:"is_active"`
}

func listAgents(w http.ResponseWriter, r *http.Request) {
    var agents []Agent
    if err := db.DB.Select(&agents, `SELECT id, user_id, name, description, is_active FROM agents ORDER BY created_at DESC`); err != nil {
        writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
        return
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

// --- Connection handlers ---

type Connection struct {
    ID            string `db:"id"             json:"id"`
    IntegrationID string `db:"integration_id" json:"integration_id"`
    UserID        string `db:"user_id"        json:"user_id"`
    Name          string `db:"name"          json:"name"`
    IsDefault     bool   `db:"is_default"    json:"is_default"`
}

func listConnections(w http.ResponseWriter, r *http.Request) {
    var conns []Connection
    if err := db.DB.Select(&conns, `SELECT id, integration_id, user_id, name, is_default FROM connections ORDER BY created_at DESC`); err != nil {
        writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
        return
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
