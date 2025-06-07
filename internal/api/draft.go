package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/cedricziel/mel-agent/internal/db"
	"github.com/cedricziel/mel-agent/pkg/api"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// WorkflowDraft represents the auto-persisting draft state of a workflow
type WorkflowDraft struct {
	AgentID   uuid.UUID              `json:"agent_id"`
	Nodes     []api.Node             `json:"nodes"`
	Edges     []WorkflowEdge         `json:"edges"`
	Layout    map[string]interface{} `json:"layout,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// DraftUpdateRequest represents a request to update the draft
type DraftUpdateRequest struct {
	Nodes    []api.Node             `json:"nodes,omitempty"`
	Edges    []WorkflowEdge         `json:"edges,omitempty"`
	Layout   map[string]interface{} `json:"layout,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// DraftNodeTestRequest represents a request to test a single node in draft
type DraftNodeTestRequest struct {
	NodeID   string                 `json:"node_id"`
	TestData map[string]interface{} `json:"test_data,omitempty"`
}

// getDraftHandler returns the current draft for an agent
func getDraftHandler(w http.ResponseWriter, r *http.Request) {
	agentID, err := uuid.Parse(chi.URLParam(r, "agentID"))
	if err != nil {
		http.Error(w, "Invalid agent ID", http.StatusBadRequest)
		return
	}

	draft, err := getDraft(agentID)
	if err != nil {
		if err == sql.ErrNoRows {
			// No draft exists, create empty one
			draft = &WorkflowDraft{
				AgentID:   agentID,
				Nodes:     []api.Node{},
				Edges:     []WorkflowEdge{},
				Layout:    make(map[string]interface{}),
				Metadata:  make(map[string]interface{}),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
		} else {
			http.Error(w, fmt.Sprintf("Failed to get draft: %v", err), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(draft)
}

// updateDraftHandler updates the draft with auto-persistence
func updateDraftHandler(w http.ResponseWriter, r *http.Request) {
	agentID, err := uuid.Parse(chi.URLParam(r, "agentID"))
	if err != nil {
		http.Error(w, "Invalid agent ID", http.StatusBadRequest)
		return
	}

	var req DraftUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get existing draft or create new one
	draft, err := getDraft(agentID)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, fmt.Sprintf("Failed to get draft: %v", err), http.StatusInternalServerError)
		return
	}

	if draft == nil {
		draft = &WorkflowDraft{
			AgentID:   agentID,
			Nodes:     []api.Node{},
			Edges:     []WorkflowEdge{},
			Layout:    make(map[string]interface{}),
			Metadata:  make(map[string]interface{}),
			CreatedAt: time.Now(),
		}
	}

	// Update fields if provided
	if req.Nodes != nil {
		draft.Nodes = req.Nodes
	}
	if req.Edges != nil {
		draft.Edges = req.Edges
	}
	if req.Layout != nil {
		draft.Layout = req.Layout
	}
	if req.Metadata != nil {
		draft.Metadata = req.Metadata
	}

	draft.UpdatedAt = time.Now()

	// Save draft
	if err := saveDraft(draft); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save draft: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(draft)
}

// testDraftNodeHandler tests a single node in the draft context
func testDraftNodeHandler(w http.ResponseWriter, r *http.Request) {
	agentID, err := uuid.Parse(chi.URLParam(r, "agentID"))
	if err != nil {
		http.Error(w, "Invalid agent ID", http.StatusBadRequest)
		return
	}

	nodeID := chi.URLParam(r, "nodeID")
	if nodeID == "" {
		http.Error(w, "Node ID is required", http.StatusBadRequest)
		return
	}

	var req DraftNodeTestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get draft
	draft, err := getDraft(agentID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "No draft found. Please save the workflow first.", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to get draft: %v", err), http.StatusInternalServerError)
		return
	}

	// Find the node in draft
	var targetNode *api.Node
	for i := range draft.Nodes {
		if draft.Nodes[i].ID == nodeID {
			targetNode = &draft.Nodes[i]
			break
		}
	}

	if targetNode == nil {
		http.Error(w, "Node not found in draft", http.StatusNotFound)
		return
	}

	// Get node definition
	nodeDef := api.FindDefinition(targetNode.Type)
	if nodeDef == nil {
		http.Error(w, fmt.Sprintf("Unknown node type: %s", targetNode.Type), http.StatusBadRequest)
		return
	}

	// Initialize the node definition if needed
	if err := nodeDef.Initialize(nil); err != nil {
		http.Error(w, fmt.Sprintf("Failed to initialize node: %v", err), http.StatusInternalServerError)
		return
	}

	// Create test envelope with provided data or empty data
	testData := req.TestData
	if testData == nil {
		testData = make(map[string]interface{})
	}

	envelope := &api.Envelope[interface{}]{
		ID:       uuid.New().String(),
		IssuedAt: time.Now(),
		Version:  1,
		Data:     testData,
		Variables: make(map[string]interface{}),
		Trace: api.Trace{
			AgentID: agentID.String(),
			RunID:   "draft-test-" + uuid.New().String(),
			NodeID:  nodeID,
			Step:    "draft-test",
			Attempt: 1,
		},
	}

	// Create execution context
	ctx := api.ExecutionContext{
		AgentID: agentID.String(),
		RunID:   envelope.Trace.RunID,
		// Note: Mel interface would need to be injected here in real implementation
	}

	// Execute the node
	result, err := nodeDef.ExecuteEnvelope(ctx, *targetNode, envelope)
	if err != nil {
		// Return the error as a structured response rather than HTTP error
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK) // Still 200, but with error in response
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
			"node_id": nodeID,
		})
		return
	}

	// Return successful result
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"result":  result,
		"node_id": nodeID,
		"trace":   result.Trace,
	})
}

// deployVersionHandler deploys a specific version
func deployVersionHandler(w http.ResponseWriter, r *http.Request) {
	agentID, err := uuid.Parse(chi.URLParam(r, "agentID"))
	if err != nil {
		http.Error(w, "Invalid agent ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Version int    `json:"version"`
		Notes   string `json:"notes,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Deploy the version
	success, err := deployVersion(agentID, req.Version, req.Notes)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to deploy version: %v", err), http.StatusInternalServerError)
		return
	}

	if !success {
		http.Error(w, "Version not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Version %d deployed successfully", req.Version),
	})
}

// Database operations

func getDraft(agentID uuid.UUID) (*WorkflowDraft, error) {
	query := `
		SELECT agent_id, nodes, edges, layout, metadata, created_at, updated_at
		FROM workflow_drafts 
		WHERE agent_id = $1
	`

	var draft WorkflowDraft
	var nodesJSON, edgesJSON, layoutJSON, metadataJSON []byte

	err := db.DB.QueryRow(query, agentID).Scan(
		&draft.AgentID,
		&nodesJSON,
		&edgesJSON,
		&layoutJSON,
		&metadataJSON,
		&draft.CreatedAt,
		&draft.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Parse JSON fields
	if err := json.Unmarshal(nodesJSON, &draft.Nodes); err != nil {
		return nil, fmt.Errorf("failed to parse nodes: %w", err)
	}
	if err := json.Unmarshal(edgesJSON, &draft.Edges); err != nil {
		return nil, fmt.Errorf("failed to parse edges: %w", err)
	}
	if err := json.Unmarshal(layoutJSON, &draft.Layout); err != nil {
		return nil, fmt.Errorf("failed to parse layout: %w", err)
	}
	if err := json.Unmarshal(metadataJSON, &draft.Metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}

	return &draft, nil
}

func saveDraft(draft *WorkflowDraft) error {
	nodesJSON, _ := json.Marshal(draft.Nodes)
	edgesJSON, _ := json.Marshal(draft.Edges)
	layoutJSON, _ := json.Marshal(draft.Layout)
	metadataJSON, _ := json.Marshal(draft.Metadata)

	query := `
		INSERT INTO workflow_drafts (agent_id, nodes, edges, layout, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (agent_id) 
		DO UPDATE SET 
			nodes = EXCLUDED.nodes,
			edges = EXCLUDED.edges,
			layout = EXCLUDED.layout,
			metadata = EXCLUDED.metadata,
			updated_at = EXCLUDED.updated_at
	`

	_, err := db.DB.Exec(query,
		draft.AgentID,
		nodesJSON,
		edgesJSON,
		layoutJSON,
		metadataJSON,
		draft.CreatedAt,
		draft.UpdatedAt,
	)

	return err
}

func deployVersion(agentID uuid.UUID, version int, notes string) (bool, error) {
	query := `SELECT deploy_workflow_version($1, $2, $3)`
	var success bool
	err := db.DB.QueryRow(query, agentID, version, notes).Scan(&success)
	return success, err
}