package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/cedricziel/mel-agent/internal/db"
	"github.com/cedricziel/mel-agent/internal/plugin"
	"github.com/cedricziel/mel-agent/pkg/api"
	_ "github.com/cedricziel/mel-agent/pkg/credentials" // Register credential definitions
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Handler returns a router with API routes mounted.
func Handler() http.Handler {
	r := chi.NewRouter()

	// Agent endpoints.
	r.Get("/agents", listAgents)
	r.Post("/agents", createAgent)
	r.Post("/agents/{agentID}/versions", createAgentVersion)

	// Draft workflow endpoints for auto-persistence
	r.Route("/agents/{agentID}/draft", func(r chi.Router) {
		r.Get("/", getDraftHandler)
		r.Put("/", updateDraftHandler)
		r.Post("/nodes/{nodeID}/test", testDraftNodeHandler)
	})

	// Version deployment
	r.Post("/agents/{agentID}/deploy", deployVersionHandler)

	// Workflow management (agents are workflows)
	r.Route("/workflows", func(r chi.Router) {
		r.Get("/", listAgents)   // reuse existing handler
		r.Post("/", createAgent) // reuse existing handler
		r.Route("/{workflowID}", func(r chi.Router) {
			r.Get("/", getWorkflow)
			r.Put("/", updateWorkflow)
			r.Delete("/", deleteWorkflow)

			// Node management
			r.Get("/nodes", listWorkflowNodes)
			r.Post("/nodes", createWorkflowNode)
			r.Route("/nodes/{nodeID}", func(r chi.Router) {
				r.Get("/", getWorkflowNode)
				r.Put("/", updateWorkflowNode)
				r.Delete("/", deleteWorkflowNode)
			})

			// Edge management
			r.Get("/edges", listWorkflowEdges)
			r.Post("/edges", createWorkflowEdge)
			r.Delete("/edges/{edgeID}", deleteWorkflowEdge)

			// Layout management
			r.Post("/layout", autoLayoutWorkflow)
		})
	})

	// Connection endpoints.
	r.Get("/connections", listConnections)
	r.Post("/connections", createConnection)
	r.Get("/connections/{connectionID}", getConnection)
	r.Put("/connections/{connectionID}", updateConnection)
	r.Delete("/connections/{connectionID}", deleteConnection)
	// Credentials for selection in nodes
	r.Get("/credentials", listCredentialsForSelection)

	// Trigger endpoints.
	r.Get("/triggers", listTriggers)
	r.Post("/triggers", createTrigger)
	r.Get("/triggers/{triggerID}", getTrigger)
	r.Patch("/triggers/{triggerID}", updateTrigger)
	r.Delete("/triggers/{triggerID}", deleteTrigger)

	// Webhook entrypoint for external events: POST /webhooks/{provider}/{triggerID}
	r.Post("/webhooks/{provider}/{triggerID}", WebhookHandler)

	// Integrations (read‑only)
	r.Get("/integrations", listIntegrations)
	// Credential type definitions
	r.Get("/credential-types", listCredentialTypes)
	r.Get("/credential-types/schema/{type}", getCredentialTypeSchemaHandler)
	r.Post("/credential-types/{type}/test", testCredentialsHandler)
	// Node type definitions for builder
	r.Get("/node-types", listNodeTypes)
	// JSON Schema for node types
	r.Get("/node-types/schema/{type}", getNodeTypeSchemaHandler)
	// Dynamic options for node parameters
	r.Get("/node-types/{type}/parameters/{parameter}/options", getDynamicOptionsHandler)
	// WebSocket for collaborative updates
	r.Get("/ws/agents/{agentID}", wsHandler)
	// Create a run via trigger or API
	r.Post("/agents/{agentID}/runs", createRunHandler)
	// Test-run an agent workflow
	r.Post("/agents/{agentID}/runs/test", testRunHandler)
	// List past runs
	r.Get("/agents/{agentID}/runs", listRunsHandler)
	// Get specific run details
	r.Get("/agents/{agentID}/runs/{runID}", getRunHandler)
	// Fetch latest version (graph) for an agent
	r.Get("/agents/{agentID}/versions/latest", getLatestAgentVersionHandler)
	// Execute a single node with provided input data (stub implementation)
	r.Post("/agents/{agentID}/nodes/{nodeID}/execute", executeNodeHandler)
	// Chat assistant endpoint for builder
	r.Post("/agents/{agentID}/assistant/chat", assistantChatHandler)

	// Plugin extensions catalog
	r.Get("/extensions", listExtensionsHandler)

	return r
}

// Helper to write JSON responses.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// listExtensionsHandler returns the catalog of registered plugins (GoPlugins and MCP servers).
func listExtensionsHandler(w http.ResponseWriter, r *http.Request) {
	metas := plugin.GetAllPlugins()
	writeJSON(w, http.StatusOK, metas)
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
		// Insert new agent version and update latest pointer
		if err := tx.QueryRow(
			`INSERT INTO agent_versions (agent_id, semantic_version, graph, default_params) VALUES ($1,$2,$3::jsonb,$4::jsonb) RETURNING id`,
			agentID, in.SemanticVersion, string(in.Graph), string(in.DefaultParams),
		).Scan(&versionID); err != nil {
			return err
		}
		if _, err := tx.Exec(`UPDATE agents SET latest_version_id=$1 WHERE id=$2`, versionID, agentID); err != nil {
			return err
		}
		// Sync triggers based on graph's entry-point nodes
		type graphNode struct {
			ID   string                 `json:"id"`
			Type string                 `json:"type"`
			Data map[string]interface{} `json:"data"`
		}
		var graph struct {
			Nodes []graphNode `json:"nodes"`
		}
		if err := json.Unmarshal(in.Graph, &graph); err != nil {
			return err
		}
		// Load existing triggers for this agent
		existing := map[string]string{} // node_id -> trigger_id
		rows, err := tx.Query(`SELECT id, node_id FROM triggers WHERE agent_id = $1`, agentID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var tid, nid string
			if err := rows.Scan(&tid, &nid); err != nil {
				return err
			}
			existing[nid] = tid
		}
		// Get agent's user_id for new triggers
		var userID string
		if err := tx.QueryRow(`SELECT user_id FROM agents WHERE id = $1`, agentID).Scan(&userID); err != nil {
			return err
		}
		// Track processed node IDs
		processed := map[string]bool{}
		for _, nd := range graph.Nodes {
			// Only sync entry-point (trigger) node types
			def := api.FindDefinition(nd.Type)
			if def == nil || !def.Meta().EntryPoint {
				continue
			}
			processed[nd.ID] = true
			// Encode node Data as JSON
			cfgBytes, err := json.Marshal(nd.Data)
			if err != nil {
				return err
			}
			if tid, ok := existing[nd.ID]; ok {
				// Update existing trigger row
				if _, err := tx.Exec(
					`UPDATE triggers SET config=$1::jsonb, provider=$2, enabled=true, updated_at=now() WHERE id=$3`,
					string(cfgBytes), nd.Type, tid,
				); err != nil {
					return err
				}
			} else {
				// Insert new trigger row
				newID := uuid.New().String()
				if _, err := tx.Exec(
					`INSERT INTO triggers (id, user_id, agent_id, node_id, provider, config, enabled) VALUES ($1,$2,$3,$4,$5,$6::jsonb,$7)`,
					newID, userID, agentID, nd.ID, nd.Type, string(cfgBytes), true,
				); err != nil {
					return err
				}
			}
		}
		// Disable triggers removed from graph
		for nid, tid := range existing {
			if !processed[nid] {
				if _, err := tx.Exec(
					`UPDATE triggers SET enabled=false, updated_at=now() WHERE id=$1`,
					tid,
				); err != nil {
					return err
				}
			}
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
	ID             string `db:"id" json:"id"`
	Name           string `db:"name" json:"name"`
	AuthType       string `db:"auth_type" json:"auth_type"`
	CredentialType string `db:"credential_type" json:"credential_type"`
}

func listIntegrations(w http.ResponseWriter, r *http.Request) {
	// initialize slice to ensure JSON encodes as [] instead of null when empty
	integrations := []Integration{}

	// Check if credential_type column exists
	var hasCredentialType bool
	err := db.DB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'integrations' AND column_name = 'credential_type'
		)
	`).Scan(&hasCredentialType)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	var query string
	if hasCredentialType {
		query = `SELECT id, name, auth_type, COALESCE(credential_type, '') FROM integrations ORDER BY name`
	} else {
		query = `SELECT id, name, auth_type FROM integrations ORDER BY name`
	}

	rows, err := db.DB.Query(query)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var i Integration
		if hasCredentialType {
			if err := rows.Scan(&i.ID, &i.Name, &i.AuthType, &i.CredentialType); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}
		} else {
			if err := rows.Scan(&i.ID, &i.Name, &i.AuthType); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}
			// Set default credential type based on auth_type for backward compatibility
			switch i.AuthType {
			case "api_key", "token":
				i.CredentialType = "api_key"
			case "custom":
				if i.Name == "baserow" {
					i.CredentialType = "baserow_jwt"
				}
			}
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
		IntegrationID  string          `json:"integration_id"`
		UserID         string          `json:"user_id"`
		Name           string          `json:"name"`
		Secret         json.RawMessage `json:"secret"` // kept opaque
		Config         json.RawMessage `json:"config"`
		IsDefault      bool            `json:"is_default"`
		CredentialType string          `json:"credential_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if in.UserID == "" {
		in.UserID = "00000000-0000-0000-0000-000000000001"
	}

	// If credential_type is provided, validate and transform the credentials
	var finalSecret json.RawMessage = in.Secret
	if in.CredentialType != "" {
		// Parse the secret data for validation
		var secretData map[string]interface{}
		if err := json.Unmarshal(in.Secret, &secretData); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid secret format"})
			return
		}

		// Validate credentials using the credential definition
		if err := api.ValidateCredentials(in.CredentialType, secretData); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "credential validation failed: " + err.Error()})
			return
		}

		// Transform credentials (e.g., exchange username/password for token)
		transformedData, err := api.TransformCredentials(in.CredentialType, secretData)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "credential transformation failed: " + err.Error()})
			return
		}

		// Re-encode the transformed data
		finalSecret, err = json.Marshal(transformedData)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to encode transformed credentials"})
			return
		}
	}

	var id string
	// Check if credential_type column exists in connections table
	var hasCredentialType bool
	err := db.DB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'connections' AND column_name = 'credential_type'
		)
	`).Scan(&hasCredentialType)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	var query string
	if hasCredentialType {
		query = `INSERT INTO connections (user_id,integration_id,name,secret,config,is_default,credential_type) VALUES ($1,$2,$3,$4::jsonb,$5::jsonb,$6,$7) RETURNING id`
		if err := db.DB.QueryRow(query, in.UserID, in.IntegrationID, in.Name, string(finalSecret), string(in.Config), in.IsDefault, in.CredentialType).Scan(&id); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
	} else {
		query = `INSERT INTO connections (user_id,integration_id,name,secret,config,is_default) VALUES ($1,$2,$3,$4::jsonb,$5::jsonb,$6) RETURNING id`
		if err := db.DB.QueryRow(query, in.UserID, in.IntegrationID, in.Name, string(finalSecret), string(in.Config), in.IsDefault).Scan(&id); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
	}
	writeJSON(w, http.StatusCreated, map[string]string{"id": id})
}

// getConnection retrieves a single connection by ID
func getConnection(w http.ResponseWriter, r *http.Request) {
	connectionID := chi.URLParam(r, "connectionID")

	// Check if credential_type column exists in connections table
	var hasCredentialType bool
	err := db.DB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'connections' AND column_name = 'credential_type'
		)
	`).Scan(&hasCredentialType)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	var query string
	if hasCredentialType {
		query = `SELECT id, integration_id, user_id, name, secret, config, is_default, COALESCE(credential_type, '') as credential_type FROM connections WHERE id = $1`
	} else {
		query = `SELECT id, integration_id, user_id, name, secret, config, is_default FROM connections WHERE id = $1`
	}

	var c struct {
		ID             string          `json:"id"`
		IntegrationID  string          `json:"integration_id"`
		UserID         string          `json:"user_id"`
		Name           string          `json:"name"`
		Secret         json.RawMessage `json:"secret"`
		Config         json.RawMessage `json:"config"`
		IsDefault      bool            `json:"is_default"`
		CredentialType string          `json:"credential_type,omitempty"`
	}

	var secretBytes, configBytes []byte
	if hasCredentialType {
		err = db.DB.QueryRow(query, connectionID).Scan(
			&c.ID, &c.IntegrationID, &c.UserID, &c.Name,
			&secretBytes, &configBytes, &c.IsDefault, &c.CredentialType,
		)
	} else {
		err = db.DB.QueryRow(query, connectionID).Scan(
			&c.ID, &c.IntegrationID, &c.UserID, &c.Name,
			&secretBytes, &configBytes, &c.IsDefault,
		)
	}

	if err != nil {
		if err == sql.ErrNoRows {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "connection not found"})
		} else {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return
	}

	c.Secret = secretBytes
	c.Config = configBytes

	writeJSON(w, http.StatusOK, c)
}

// --- Trigger handlers ---
// Trigger represents a user-configured event trigger for starting workflows.
type Trigger struct {
	ID          string          `db:"id" json:"id"`
	UserID      string          `db:"user_id" json:"user_id"`
	Provider    string          `db:"provider" json:"provider"`
	Config      json.RawMessage `db:"config" json:"config"`
	Enabled     bool            `db:"enabled" json:"enabled"`
	NodeID      string          `db:"node_id" json:"node_id"`
	AgentID     string          `db:"agent_id" json:"agent_id"`
	LastChecked *time.Time      `db:"last_checked" json:"last_checked,omitempty"`
	CreatedAt   time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time       `db:"updated_at" json:"updated_at"`
}

// listTriggers returns all triggers.
func listTriggers(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query(
		`SELECT id, user_id, provider, config, enabled, node_id, agent_id, last_checked, created_at, updated_at
        FROM triggers
        ORDER BY created_at DESC`,
	)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()

	triggers := []Trigger{}
	for rows.Next() {
		var t Trigger
		var configBytes []byte
		var nt sql.NullTime
		if err := rows.Scan(&t.ID, &t.UserID, &t.Provider, &configBytes, &t.Enabled, &t.NodeID, &t.AgentID, &nt, &t.CreatedAt, &t.UpdatedAt); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		t.Config = configBytes
		if nt.Valid {
			t.LastChecked = &nt.Time
		}
		triggers = append(triggers, t)
	}
	writeJSON(w, http.StatusOK, triggers)
}

// getTrigger fetches a single trigger by ID.
func getTrigger(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "triggerID")
	var t Trigger
	var configBytes []byte
	var nt sql.NullTime
	err := db.DB.QueryRow(
		`SELECT id, user_id, provider, config, enabled, node_id, agent_id, last_checked, created_at, updated_at
        FROM triggers WHERE id = $1`, id,
	).Scan(&t.ID, &t.UserID, &t.Provider, &configBytes, &t.Enabled, &t.NodeID, &t.AgentID, &nt, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "trigger not found"})
		} else {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return
	}
	t.Config = configBytes
	if nt.Valid {
		t.LastChecked = &nt.Time
	}
	writeJSON(w, http.StatusOK, t)
}

// createTrigger registers a new trigger.
func createTrigger(w http.ResponseWriter, r *http.Request) {
	var in struct {
		UserID   string          `json:"user_id"`
		Provider string          `json:"provider"`
		Config   json.RawMessage `json:"config"`
		Enabled  *bool           `json:"enabled,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if in.UserID == "" {
		in.UserID = "00000000-0000-0000-0000-000000000001"
	}
	enabled := true
	if in.Enabled != nil {
		enabled = *in.Enabled
	}
	id := uuid.New().String()
	query := `INSERT INTO triggers (id, user_id, provider, config, enabled) VALUES ($1,$2,$3,$4::jsonb,$5) RETURNING id`
	if err := db.DB.QueryRow(query, id, in.UserID, in.Provider, string(in.Config), enabled).Scan(&id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"id": id})
}

// updateTrigger modifies an existing trigger.
func updateTrigger(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "triggerID")
	var in struct {
		Provider *string          `json:"provider,omitempty"`
		Config   *json.RawMessage `json:"config,omitempty"`
		Enabled  *bool            `json:"enabled,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if in.Provider != nil {
		if _, err := db.DB.Exec(`UPDATE triggers SET provider=$1, updated_at=now() WHERE id=$2`, *in.Provider, id); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
	}
	if in.Config != nil {
		if _, err := db.DB.Exec(`UPDATE triggers SET config=$1::jsonb, updated_at=now() WHERE id=$2`, string(*in.Config), id); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
	}
	if in.Enabled != nil {
		if _, err := db.DB.Exec(`UPDATE triggers SET enabled=$1, updated_at=now() WHERE id=$2`, *in.Enabled, id); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]string{"id": id})
}

// deleteTrigger removes a trigger.
func deleteTrigger(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "triggerID")
	if _, err := db.DB.Exec(`DELETE FROM triggers WHERE id=$1`, id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// WebhookHandler handles external webhook events and enqueues a run.
func WebhookHandler(w http.ResponseWriter, r *http.Request) {
	// Indicate this webhook request is handled by our agent engine
	w.Header().Set("X-Agent-Processed", "true")
	provider := chi.URLParam(r, "provider")
	triggerID := chi.URLParam(r, "triggerID")
	// Lookup trigger instance
	var enabled bool
	var agentID, nodeID string
	var configRaw []byte
	err := db.DB.QueryRow(
		`SELECT enabled, agent_id, node_id, config FROM triggers WHERE id=$1 AND provider=$2`,
		triggerID, provider,
	).Scan(&enabled, &agentID, &nodeID, &configRaw)
	if err != nil {
		if err == sql.ErrNoRows {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "trigger not found"})
		} else {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return
	}
	if !enabled {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "trigger disabled"})
		return
	}
	// Enforce allowed HTTP method from node config
	var methodCfg map[string]interface{}
	if err := json.Unmarshal(configRaw, &methodCfg); err != nil {
		methodCfg = map[string]interface{}{}
	}
	methodAllowed, _ := methodCfg["method"].(string)
	if methodAllowed == "" {
		methodAllowed = "POST"
	}
	if methodAllowed != "ANY" && !strings.EqualFold(r.Method, methodAllowed) {
		w.Header().Set("Allow", methodAllowed)
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	// Read raw body
	rawBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	// Build plugin payload
	triggerPayload := map[string]interface{}{
		"trigger_id":  triggerID,
		"agent_id":    agentID,
		"node_id":     nodeID,
		"http_method": r.Method,
		"headers":     r.Header,
		"body_raw":    rawBody,
	}
	// Invoke the webhook trigger plugin
	p, ok := plugin.GetTriggerPlugin(provider)
	if !ok {
		writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "no webhook plugin"})
		return
	}
	runIDRaw, err := p.OnTrigger(r.Context(), triggerPayload)
	if err != nil {
		// Treat method not allowed specially
		if err.Error() == "method not allowed" {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": err.Error()})
		} else {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return
	}
	runID, _ := runIDRaw.(string)
	// Respond with run ID and default accepted status
	writeJSON(w, http.StatusAccepted, map[string]string{"runId": runID})
	return
}

// executeNodeHandler handles running a single node with provided input
func executeNodeHandler(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	nodeID := chi.URLParam(r, "nodeID")
	
	var input interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	// Get the latest workflow version to find node configuration
	var versionID sql.NullString
	if err := db.DB.QueryRow(`SELECT latest_version_id FROM agents WHERE id = $1`, agentID).Scan(&versionID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get agent version"})
		return
	}
	
	if !versionID.Valid {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "no workflow version available"})
		return
	}

	// Load the workflow graph to find the specific node configuration
	var graphRaw []byte
	if err := db.DB.QueryRow(`SELECT graph FROM agent_versions WHERE id = $1`, versionID.String).Scan(&graphRaw); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load workflow graph"})
		return
	}

	// Parse the graph to find the target node
	var graphStruct struct {
		Nodes []api.Node `json:"nodes"`
	}
	if err := json.Unmarshal(graphRaw, &graphStruct); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to parse workflow graph"})
		return
	}

	// Find the specific node
	var targetNode *api.Node
	for _, node := range graphStruct.Nodes {
		if node.ID == nodeID {
			targetNode = &node
			break
		}
	}

	if targetNode == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "node not found in workflow"})
		return
	}

	// Execute the node
	result, err := executeNodeLocal(agentID, *targetNode, input)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	// Return the execution result
	response := map[string]interface{}{
		"agent_id": agentID,
		"node_id":  nodeID,
		"output":   result,
		"success":  true,
	}
	writeJSON(w, http.StatusOK, response)
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
		"id":               versionID.String,
		"semantic_version": semanticVersion,
		"graph":            graph,
		"default_params":   defaultParams,
	})
}

// createRunHandler starts a new run for an agent triggered by external event or scheduler.
func createRunHandler(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	var in struct {
		VersionID   string      `json:"versionId"`
		StartNodeID string      `json:"startNodeId"`
		Input       interface{} `json:"input"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	// Verify version belongs to agent
	var exists bool
	if err := db.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM agent_versions WHERE id = $1 AND agent_id = $2)`, in.VersionID, agentID).Scan(&exists); err != nil || !exists {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid versionId"})
		return
	}
	// Prepare payload for run
	payload := map[string]interface{}{
		"versionId":   in.VersionID,
		"startNodeId": in.StartNodeID,
		"input":       in.Input,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	runID := uuid.New().String()
	if _, err := db.DB.Exec(`INSERT INTO agent_runs (id, agent_id, payload) VALUES ($1, $2, $3::jsonb)`, runID, agentID, string(raw)); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"runId": runID})
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
	// Unmarshal full graph (nodes + edges) for rendering
	var graphData interface{}
	if err := json.Unmarshal(graphRaw, &graphData); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	// Unmarshal nodes sequence for execution order
	var graphStruct struct {
		Nodes []api.Node `json:"nodes"`
	}
	if err := json.Unmarshal(graphRaw, &graphStruct); err != nil {
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
	// Define execution payload with full graph and trace
	type Item struct {
		ID   string      `json:"id"`
		Data interface{} `json:"data"`
	}
	type Step struct {
		NodeID string `json:"nodeId"`
		Input  []Item `json:"input"`
		Output []Item `json:"output"`
	}
	type Payload struct {
		RunID   string                 `json:"runId"`
		Graph   interface{}            `json:"graph"`
		Context map[string]interface{} `json:"context"`
		Meta    map[string]interface{} `json:"meta"`
		Trace   []Step                 `json:"trace"`
	}
	runID := uuid.NewString()
	// Prepare initial payload
	initialItem := Item{ID: uuid.NewString(), Data: defaultParams}
	payload := Payload{
		RunID:   runID,
		Graph:   graphData,
		Context: map[string]interface{}{},
		Meta: map[string]interface{}{
			"startTime": time.Now().UTC().Format(time.RFC3339),
		},
		Trace: []Step{},
	}
	currentItems := []Item{initialItem}
	// Execute nodes in order
	// Execute nodes in order, recording trace
	for _, node := range graphStruct.Nodes {
		// Broadcast node execution start
		hub := GetHub(agentID)
		if startMsg, err := json.Marshal(map[string]string{"type": "nodeExecution", "nodeId": node.ID, "phase": "start"}); err == nil {
			hub.broadcast(startMsg)
		}
		inputItems := currentItems
		var nextItems []Item
		for _, item := range inputItems {
			output, err := executeNodeLocal(agentID, node, item.Data)
			if err != nil {
				nextItems = append(nextItems, Item{ID: item.ID, Data: map[string]interface{}{"error": err.Error()}})
			} else if output != nil {
				nextItems = append(nextItems, Item{ID: item.ID, Data: output})
			}
		}
		// Broadcast node execution end
		if endMsg, err := json.Marshal(map[string]string{"type": "nodeExecution", "nodeId": node.ID, "phase": "end"}); err == nil {
			hub.broadcast(endMsg)
		}
		// Append step to trace
		payload.Trace = append(payload.Trace, Step{
			NodeID: node.ID,
			Input:  inputItems,
			Output: nextItems,
		})
		// Prepare for next iteration
		currentItems = nextItems
		payload.Meta["lastNode"] = node.ID
	}
	// Persist run record
	// Embed final items under meta
	payload.Meta["finalItems"] = currentItems
	raw, err := json.Marshal(payload)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if _, err := db.DB.Exec(`INSERT INTO agent_runs (id, agent_id, payload) VALUES ($1, $2, $3::jsonb)`, payload.RunID, agentID, string(raw)); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, payload)
}

// executeNodeLocal invokes the node logic via the node definitions.
func executeNodeLocal(agentID string, node api.Node, input interface{}) (interface{}, error) {
	if def := api.FindDefinition(node.Type); def != nil {
		ctx := api.ExecutionContext{
			AgentID: agentID,
			Mel:     api.NewMel(), // Provide platform utilities
		}

		// Create envelope from input
		envelope := &api.Envelope[interface{}]{
			ID:       "local-" + node.ID,
			IssuedAt: time.Now(),
			Version:  1,
			DataType: "unknown",
			Data:     input,
			Trace: api.Trace{
				AgentID: ctx.AgentID,
				RunID:   ctx.RunID,
				NodeID:  node.ID,
				Step:    node.ID,
				Attempt: 1,
			},
			Variables: ctx.Variables,
		}

		result, err := def.ExecuteEnvelope(ctx, node, envelope)
		if err != nil {
			return input, err
		}

		return result.Data, nil
	}
	// Fallback: return input unchanged
	return input, nil
}

// listRunsHandler returns a list of past runs for an agent.
func listRunsHandler(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	rows, err := db.DB.Query(
		`SELECT id, created_at FROM agent_runs WHERE agent_id = $1 ORDER BY created_at DESC`, agentID,
	)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()
	type runMeta struct {
		ID        string `json:"id"`
		CreatedAt string `json:"created_at"`
	}
	var runs []runMeta
	for rows.Next() {
		var rm runMeta
		var t sql.NullTime
		if err := rows.Scan(&rm.ID, &t); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		if t.Valid {
			rm.CreatedAt = t.Time.UTC().Format(time.RFC3339)
		}
		runs = append(runs, rm)
	}
	writeJSON(w, http.StatusOK, runs)
}

// getRunHandler returns the payload of a specific run.
func getRunHandler(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	runID := chi.URLParam(r, "runID")
	var payloadRaw []byte
	err := db.DB.QueryRow(
		`SELECT payload FROM agent_runs WHERE agent_id = $1 AND id = $2`, agentID, runID,
	).Scan(&payloadRaw)
	if err != nil {
		if err == sql.ErrNoRows {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "run not found"})
		} else {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return
	}
	// Decode payload into a map to allow enriching with graph when absent
	var payloadMap map[string]interface{}
	if err := json.Unmarshal(payloadRaw, &payloadMap); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	// If graph structure is not embedded, attach it from the agent_versions table
	if _, hasGraph := payloadMap["graph"]; !hasGraph {
		if ver, ok := payloadMap["versionId"].(string); ok && ver != "" {
			var graphRaw []byte
			// fetch the stored graph JSON
			if err := db.DB.QueryRow(`SELECT graph FROM agent_versions WHERE id = $1`, ver).Scan(&graphRaw); err == nil {
				var graph interface{}
				if err2 := json.Unmarshal(graphRaw, &graph); err2 == nil {
					payloadMap["graph"] = graph
				}
			}
		}
	}
	writeJSON(w, http.StatusOK, payloadMap)
}

// --- Workflow API handlers ---

// WorkflowNode represents a node in a workflow
type WorkflowNode struct {
	ID        string                 `db:"id" json:"id"`
	AgentID   string                 `db:"agent_id" json:"agent_id"`
	NodeID    string                 `db:"node_id" json:"node_id"`
	NodeType  string                 `db:"node_type" json:"node_type"`
	PositionX float64                `db:"position_x" json:"position_x"`
	PositionY float64                `db:"position_y" json:"position_y"`
	Config    map[string]interface{} `json:"config"`
	CreatedAt time.Time              `db:"created_at" json:"created_at"`
	UpdatedAt time.Time              `db:"updated_at" json:"updated_at"`
}

// WorkflowEdge represents an edge between nodes
type WorkflowEdge struct {
	ID           string    `db:"id" json:"id"`
	AgentID      string    `db:"agent_id" json:"agent_id"`
	EdgeID       string    `db:"edge_id" json:"edge_id"`
	SourceNodeID string    `db:"source_node_id" json:"source_node_id"`
	TargetNodeID string    `db:"target_node_id" json:"target_node_id"`
	SourceHandle *string   `db:"source_handle" json:"source_handle,omitempty"`
	TargetHandle *string   `db:"target_handle" json:"target_handle,omitempty"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}

func getWorkflow(w http.ResponseWriter, r *http.Request) {
	workflowID := chi.URLParam(r, "workflowID")
	var agent Agent
	var desc sql.NullString
	err := db.DB.QueryRow(`SELECT id, user_id, name, description, is_active FROM agents WHERE id = $1`, workflowID).
		Scan(&agent.ID, &agent.UserID, &agent.Name, &desc, &agent.IsActive)
	if err != nil {
		if err == sql.ErrNoRows {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "workflow not found"})
		} else {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return
	}
	if desc.Valid {
		agent.Description = &desc.String
	}
	writeJSON(w, http.StatusOK, agent)
}

func updateWorkflow(w http.ResponseWriter, r *http.Request) {
	workflowID := chi.URLParam(r, "workflowID")
	var in struct {
		Name        *string `json:"name,omitempty"`
		Description *string `json:"description,omitempty"`
		IsActive    *bool   `json:"is_active,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if in.Name != nil {
		if _, err := db.DB.Exec(`UPDATE agents SET name = $1 WHERE id = $2`, *in.Name, workflowID); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
	}
	if in.Description != nil {
		if _, err := db.DB.Exec(`UPDATE agents SET description = $1 WHERE id = $2`, sql.NullString{String: *in.Description, Valid: *in.Description != ""}, workflowID); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
	}
	if in.IsActive != nil {
		if _, err := db.DB.Exec(`UPDATE agents SET is_active = $1 WHERE id = $2`, *in.IsActive, workflowID); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
	}

	writeJSON(w, http.StatusOK, map[string]string{"id": workflowID})
}

func deleteWorkflow(w http.ResponseWriter, r *http.Request) {
	workflowID := chi.URLParam(r, "workflowID")
	if _, err := db.DB.Exec(`DELETE FROM agents WHERE id = $1`, workflowID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Node management handlers

func listWorkflowNodes(w http.ResponseWriter, r *http.Request) {
	workflowID := chi.URLParam(r, "workflowID")
	rows, err := db.DB.Query(`SELECT id, agent_id, node_id, node_type, position_x, position_y, config, created_at, updated_at FROM workflow_nodes WHERE agent_id = $1 ORDER BY created_at`, workflowID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()

	nodes := []WorkflowNode{}
	for rows.Next() {
		var node WorkflowNode
		var configBytes []byte
		if err := rows.Scan(&node.ID, &node.AgentID, &node.NodeID, &node.NodeType, &node.PositionX, &node.PositionY, &configBytes, &node.CreatedAt, &node.UpdatedAt); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		if err := json.Unmarshal(configBytes, &node.Config); err != nil {
			node.Config = map[string]interface{}{}
		}
		nodes = append(nodes, node)
	}
	writeJSON(w, http.StatusOK, nodes)
}

func createWorkflowNode(w http.ResponseWriter, r *http.Request) {
	workflowID := chi.URLParam(r, "workflowID")
	var in struct {
		NodeID    string                 `json:"node_id"`
		NodeType  string                 `json:"node_type"`
		PositionX float64                `json:"position_x"`
		PositionY float64                `json:"position_y"`
		Config    map[string]interface{} `json:"config"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if in.Config == nil {
		in.Config = map[string]interface{}{}
	}
	configBytes, err := json.Marshal(in.Config)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid config"})
		return
	}

	var id string
	query := `INSERT INTO workflow_nodes (agent_id, node_id, node_type, position_x, position_y, config) VALUES ($1, $2, $3, $4, $5, $6::jsonb) RETURNING id`
	if err := db.DB.QueryRow(query, workflowID, in.NodeID, in.NodeType, in.PositionX, in.PositionY, string(configBytes)).Scan(&id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func getWorkflowNode(w http.ResponseWriter, r *http.Request) {
	workflowID := chi.URLParam(r, "workflowID")
	nodeID := chi.URLParam(r, "nodeID")

	var node WorkflowNode
	var configBytes []byte
	err := db.DB.QueryRow(`SELECT id, agent_id, node_id, node_type, position_x, position_y, config, created_at, updated_at FROM workflow_nodes WHERE agent_id = $1 AND node_id = $2`, workflowID, nodeID).
		Scan(&node.ID, &node.AgentID, &node.NodeID, &node.NodeType, &node.PositionX, &node.PositionY, &configBytes, &node.CreatedAt, &node.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "node not found"})
		} else {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return
	}

	if err := json.Unmarshal(configBytes, &node.Config); err != nil {
		node.Config = map[string]interface{}{}
	}
	writeJSON(w, http.StatusOK, node)
}

func updateWorkflowNode(w http.ResponseWriter, r *http.Request) {
	workflowID := chi.URLParam(r, "workflowID")
	nodeID := chi.URLParam(r, "nodeID")

	var in struct {
		NodeType  *string                 `json:"node_type,omitempty"`
		PositionX *float64                `json:"position_x,omitempty"`
		PositionY *float64                `json:"position_y,omitempty"`
		Config    *map[string]interface{} `json:"config,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if in.NodeType != nil {
		if _, err := db.DB.Exec(`UPDATE workflow_nodes SET node_type = $1 WHERE agent_id = $2 AND node_id = $3`, *in.NodeType, workflowID, nodeID); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
	}
	if in.PositionX != nil {
		if _, err := db.DB.Exec(`UPDATE workflow_nodes SET position_x = $1 WHERE agent_id = $2 AND node_id = $3`, *in.PositionX, workflowID, nodeID); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
	}
	if in.PositionY != nil {
		if _, err := db.DB.Exec(`UPDATE workflow_nodes SET position_y = $1 WHERE agent_id = $2 AND node_id = $3`, *in.PositionY, workflowID, nodeID); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
	}
	if in.Config != nil {
		configBytes, err := json.Marshal(*in.Config)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid config"})
			return
		}
		if _, err := db.DB.Exec(`UPDATE workflow_nodes SET config = $1::jsonb WHERE agent_id = $2 AND node_id = $3`, string(configBytes), workflowID, nodeID); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
	}

	writeJSON(w, http.StatusOK, map[string]string{"node_id": nodeID})
}

func deleteWorkflowNode(w http.ResponseWriter, r *http.Request) {
	workflowID := chi.URLParam(r, "workflowID")
	nodeID := chi.URLParam(r, "nodeID")

	if _, err := db.DB.Exec(`DELETE FROM workflow_nodes WHERE agent_id = $1 AND node_id = $2`, workflowID, nodeID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Edge management handlers

func listWorkflowEdges(w http.ResponseWriter, r *http.Request) {
	workflowID := chi.URLParam(r, "workflowID")
	rows, err := db.DB.Query(`SELECT id, agent_id, edge_id, source_node_id, target_node_id, source_handle, target_handle, created_at FROM workflow_edges WHERE agent_id = $1 ORDER BY created_at`, workflowID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()

	edges := []WorkflowEdge{}
	for rows.Next() {
		var edge WorkflowEdge
		var sourceHandle, targetHandle sql.NullString
		if err := rows.Scan(&edge.ID, &edge.AgentID, &edge.EdgeID, &edge.SourceNodeID, &edge.TargetNodeID, &sourceHandle, &targetHandle, &edge.CreatedAt); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		if sourceHandle.Valid {
			edge.SourceHandle = &sourceHandle.String
		}
		if targetHandle.Valid {
			edge.TargetHandle = &targetHandle.String
		}
		edges = append(edges, edge)
	}
	writeJSON(w, http.StatusOK, edges)
}

func createWorkflowEdge(w http.ResponseWriter, r *http.Request) {
	workflowID := chi.URLParam(r, "workflowID")
	var in struct {
		EdgeID       string  `json:"edge_id"`
		SourceNodeID string  `json:"source_node_id"`
		TargetNodeID string  `json:"target_node_id"`
		SourceHandle *string `json:"source_handle,omitempty"`
		TargetHandle *string `json:"target_handle,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	var id string
	query := `INSERT INTO workflow_edges (agent_id, edge_id, source_node_id, target_node_id, source_handle, target_handle) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	if err := db.DB.QueryRow(query, workflowID, in.EdgeID, in.SourceNodeID, in.TargetNodeID,
		sql.NullString{String: stringDeref(in.SourceHandle), Valid: in.SourceHandle != nil},
		sql.NullString{String: stringDeref(in.TargetHandle), Valid: in.TargetHandle != nil}).Scan(&id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func deleteWorkflowEdge(w http.ResponseWriter, r *http.Request) {
	workflowID := chi.URLParam(r, "workflowID")
	edgeID := chi.URLParam(r, "edgeID")

	if _, err := db.DB.Exec(`DELETE FROM workflow_edges WHERE agent_id = $1 AND edge_id = $2`, workflowID, edgeID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Layout management handler

func autoLayoutWorkflow(w http.ResponseWriter, r *http.Request) {
	workflowID := chi.URLParam(r, "workflowID")

	// Get all nodes
	nodeRows, err := db.DB.Query(`SELECT node_id FROM workflow_nodes WHERE agent_id = $1`, workflowID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	defer nodeRows.Close()

	var allNodes []string
	nodeSet := make(map[string]bool)
	for nodeRows.Next() {
		var nodeID string
		if err := nodeRows.Scan(&nodeID); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		allNodes = append(allNodes, nodeID)
		nodeSet[nodeID] = true
	}

	// Get all edges to build the graph
	edgeRows, err := db.DB.Query(`SELECT source_node_id, target_node_id FROM workflow_edges WHERE agent_id = $1`, workflowID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	defer edgeRows.Close()

	// Build adjacency list and in-degree count
	adjacencyList := make(map[string][]string)
	inDegree := make(map[string]int)

	// Initialize in-degree for all nodes
	for _, nodeID := range allNodes {
		inDegree[nodeID] = 0
	}

	for edgeRows.Next() {
		var sourceID, targetID string
		if err := edgeRows.Scan(&sourceID, &targetID); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}

		// Only process edges between existing nodes
		if nodeSet[sourceID] && nodeSet[targetID] {
			adjacencyList[sourceID] = append(adjacencyList[sourceID], targetID)
			inDegree[targetID]++
		}
	}

	// Create a layered layout using BFS to group nodes by their distance from start nodes
	layers := make(map[int][]string) // layer index -> list of node IDs
	nodeLayer := make(map[string]int) // node ID -> layer index
	visited := make(map[string]bool)
	
	// Find all nodes with no incoming edges (start nodes) - these go in layer 0
	var startNodes []string
	for _, nodeID := range allNodes {
		if inDegree[nodeID] == 0 {
			startNodes = append(startNodes, nodeID)
			layers[0] = append(layers[0], nodeID)
			nodeLayer[nodeID] = 0
			visited[nodeID] = true
		}
	}

	// Use BFS to assign layers to remaining nodes
	currentLayer := 0
	for {
		nextLayerNodes := []string{}
		
		// Process all nodes in current layer
		for _, nodeID := range layers[currentLayer] {
			// For each neighbor of current node
			for _, neighbor := range adjacencyList[nodeID] {
				if !visited[neighbor] {
					// Check if all predecessors of neighbor have been visited
					allPredecessorsVisited := true
					for _, potentialPred := range allNodes {
						for _, target := range adjacencyList[potentialPred] {
							if target == neighbor && !visited[potentialPred] {
								allPredecessorsVisited = false
								break
							}
						}
						if !allPredecessorsVisited {
							break
						}
					}
					
					if allPredecessorsVisited {
						nextLayerNodes = append(nextLayerNodes, neighbor)
						visited[neighbor] = true
						nodeLayer[neighbor] = currentLayer + 1
					}
				}
			}
		}
		
		if len(nextLayerNodes) == 0 {
			break
		}
		
		currentLayer++
		layers[currentLayer] = nextLayerNodes
	}
	
	// Add any remaining unvisited nodes (disconnected components) to the last layer
	maxLayer := currentLayer
	for _, nodeID := range allNodes {
		if !visited[nodeID] {
			maxLayer++
			layers[maxLayer] = append(layers[maxLayer], nodeID)
			nodeLayer[nodeID] = maxLayer
		}
	}

	// Layout constants
	const spacingX = 300.0 // Horizontal spacing between layers
	const spacingY = 120.0 // Vertical spacing between nodes in same layer
	const startX = 50.0    // Starting X position
	const startY = 50.0    // Starting Y position

	// Position nodes layer by layer
	for layerIndex := 0; layerIndex <= maxLayer; layerIndex++ {
		layerNodes := layers[layerIndex]
		if len(layerNodes) == 0 {
			continue
		}
		
		// Calculate X position for this layer
		x := startX + float64(layerIndex)*spacingX
		
		// Calculate starting Y position to center the nodes vertically
		totalHeight := float64(len(layerNodes)-1) * spacingY
		startYForLayer := startY - totalHeight/2
		
		// Position each node in this layer
		for i, nodeID := range layerNodes {
			y := startYForLayer + float64(i)*spacingY
			
			if _, err := db.DB.Exec(`UPDATE workflow_nodes SET position_x = $1, position_y = $2 WHERE agent_id = $3 AND node_id = $4`, x, y, workflowID, nodeID); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"updated_nodes": len(allNodes),
		"layers":        layers,
		"max_layer":     maxLayer,
	})
}

// Helper function to dereference string pointer
func stringDeref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// getDynamicOptionsHandler provides a unified endpoint for loading dynamic options for any node parameter
func getDynamicOptionsHandler(w http.ResponseWriter, r *http.Request) {
	nodeType := chi.URLParam(r, "type")
	parameterName := chi.URLParam(r, "parameter")

	// Find the node definition
	def := api.FindDefinition(nodeType)
	if def == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "node type not found"})
		return
	}

	// Check if the node implements DynamicOptionsProvider
	optionsProvider, ok := def.(api.DynamicOptionsProvider)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "node does not support dynamic options"})
		return
	}

	// Parse dependencies from query parameters
	dependencies := make(map[string]interface{})
	for key, values := range r.URL.Query() {
		if len(values) > 0 {
			dependencies[key] = values[0]
		}
	}

	// Create execution context (we might need a user context in the future)
	ctx := api.ExecutionContext{}

	// Get dynamic options
	options, err := optionsProvider.GetDynamicOptions(ctx, parameterName, dependencies)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"options": options,
	})
}

// listCredentialTypes returns all available credential type definitions
func listCredentialTypes(w http.ResponseWriter, r *http.Request) {
	types := api.ListCredentialDefinitions()
	writeJSON(w, http.StatusOK, types)
}

// getCredentialTypeSchemaHandler returns the JSON schema for a specific credential type
func getCredentialTypeSchemaHandler(w http.ResponseWriter, r *http.Request) {
	credentialType := chi.URLParam(r, "type")

	def := api.FindCredentialDefinition(credentialType)
	if def == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "credential type not found"})
		return
	}

	// Build JSON schema from parameter definitions
	schema := map[string]interface{}{
		"type":       "object",
		"properties": make(map[string]interface{}),
		"required":   make([]string, 0),
	}

	properties := schema["properties"].(map[string]interface{})
	required := schema["required"].([]string)

	for _, param := range def.Parameters() {
		properties[param.Name] = param.ToJSONSchema()
		if param.Required {
			required = append(required, param.Name)
		}
	}

	schema["required"] = required
	writeJSON(w, http.StatusOK, schema)
}

// testCredentialsHandler tests credentials for a specific credential type
func testCredentialsHandler(w http.ResponseWriter, r *http.Request) {
	credentialType := chi.URLParam(r, "type")

	var requestData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}

	// Test the credentials
	if err := api.TestCredentials(credentialType, requestData); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "success", "message": "credentials are valid"})
}

// listCredentialsForSelection returns credentials suitable for selection in nodes
func listCredentialsForSelection(w http.ResponseWriter, r *http.Request) {
	// Get optional credential_type filter
	credentialType := r.URL.Query().Get("credential_type")

	var query string
	var args []interface{}

	// Build query based on whether we have credential_type column
	var hasCredentialTypeColumn bool
	err := db.DB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'connections' AND column_name = 'credential_type'
		)
	`).Scan(&hasCredentialTypeColumn)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	if hasCredentialTypeColumn && credentialType != "" {
		query = `SELECT c.id, c.name, c.integration_id, i.name as integration_name, c.credential_type
		         FROM connections c
		         JOIN integrations i ON c.integration_id = i.id
		         WHERE c.credential_type = $1
		         ORDER BY c.name`
		args = append(args, credentialType)
	} else if hasCredentialTypeColumn {
		query = `SELECT c.id, c.name, c.integration_id, i.name as integration_name, COALESCE(c.credential_type, '') as credential_type
		         FROM connections c
		         JOIN integrations i ON c.integration_id = i.id
		         ORDER BY c.name`
	} else {
		query = `SELECT c.id, c.name, c.integration_id, i.name as integration_name
		         FROM connections c
		         JOIN integrations i ON c.integration_id = i.id
		         ORDER BY c.name`
	}

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()

	type CredentialOption struct {
		ID              string `json:"id"`
		Name            string `json:"name"`
		IntegrationID   string `json:"integration_id"`
		IntegrationName string `json:"integration_name"`
		CredentialType  string `json:"credential_type,omitempty"`
	}

	credentials := []CredentialOption{}
	for rows.Next() {
		var cred CredentialOption
		if hasCredentialTypeColumn {
			if err := rows.Scan(&cred.ID, &cred.Name, &cred.IntegrationID, &cred.IntegrationName, &cred.CredentialType); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}
		} else {
			if err := rows.Scan(&cred.ID, &cred.Name, &cred.IntegrationID, &cred.IntegrationName); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}
		}
		credentials = append(credentials, cred)
	}

	writeJSON(w, http.StatusOK, credentials)
}

// updateConnection updates an existing connection/credential
func updateConnection(w http.ResponseWriter, r *http.Request) {
	connectionID := chi.URLParam(r, "connectionID")

	var in struct {
		Name           *string         `json:"name,omitempty"`
		Secret         json.RawMessage `json:"secret,omitempty"`
		Config         json.RawMessage `json:"config,omitempty"`
		CredentialType *string         `json:"credential_type,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	// Check if connection exists
	var exists bool
	err := db.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM connections WHERE id = $1)`, connectionID).Scan(&exists)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if !exists {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "connection not found"})
		return
	}

	// Build dynamic update query
	updates := []string{}
	args := []interface{}{}
	argIndex := 1

	if in.Name != nil {
		updates = append(updates, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, *in.Name)
		argIndex++
	}

	if in.Secret != nil {
		// If credential_type is provided, validate and transform the credentials
		var finalSecret json.RawMessage = in.Secret
		if in.CredentialType != nil && *in.CredentialType != "" {
			// Parse the secret data for validation
			var secretData map[string]interface{}
			if err := json.Unmarshal(in.Secret, &secretData); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid secret format"})
				return
			}

			// Validate credentials using the credential definition
			if err := api.ValidateCredentials(*in.CredentialType, secretData); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "credential validation failed: " + err.Error()})
				return
			}

			// Transform credentials (e.g., exchange username/password for token)
			transformedData, err := api.TransformCredentials(*in.CredentialType, secretData)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "credential transformation failed: " + err.Error()})
				return
			}

			// Re-encode the transformed data
			finalSecret, err = json.Marshal(transformedData)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to encode transformed credentials"})
				return
			}
		}

		updates = append(updates, fmt.Sprintf("secret = $%d::jsonb", argIndex))
		args = append(args, string(finalSecret))
		argIndex++
	}

	if in.Config != nil {
		updates = append(updates, fmt.Sprintf("config = $%d::jsonb", argIndex))
		args = append(args, string(in.Config))
		argIndex++
	}

	if in.CredentialType != nil {
		// Check if credential_type column exists
		var hasCredentialType bool
		err := db.DB.QueryRow(`
			SELECT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name = 'connections' AND column_name = 'credential_type'
			)
		`).Scan(&hasCredentialType)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}

		if hasCredentialType {
			updates = append(updates, fmt.Sprintf("credential_type = $%d", argIndex))
			args = append(args, *in.CredentialType)
			argIndex++
		}
	}

	if len(updates) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "no fields to update"})
		return
	}

	// Add WHERE clause
	query := fmt.Sprintf("UPDATE connections SET %s WHERE id = $%d",
		strings.Join(updates, ", "), argIndex)
	args = append(args, connectionID)

	_, err = db.DB.Exec(query, args...)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"id": connectionID})
}

// deleteConnection deletes a connection/credential
func deleteConnection(w http.ResponseWriter, r *http.Request) {
	connectionID := chi.URLParam(r, "connectionID")

	// Check if connection exists and delete it
	result, err := db.DB.Exec(`DELETE FROM connections WHERE id = $1`, connectionID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	if rowsAffected == 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "connection not found"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
