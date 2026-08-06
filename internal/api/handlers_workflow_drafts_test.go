package api

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cedricziel/mel-agent/internal/testutil"
)

func TestWorkflowDrafts(t *testing.T) {
	ctx := context.Background()
	_, db, cleanup := testutil.SetupPostgresWithMigrations(ctx, t)
	defer cleanup()

	handlers := &OpenAPIHandlers{db: db}

	// Helper function to create test user and workflow
	createTestWorkflow := func(t *testing.T, definition []byte) (uuid.UUID, uuid.UUID) {
		t.Helper()
		// Create test user with unique email
		userID := uuid.New()
		email := userID.String() + "@test.com"
		_, err := db.Exec(`
			INSERT INTO users (id, name, email, created_at)
			VALUES ($1, $2, $3, $4)`,
			userID.String(), "Test User", email, time.Now())
		require.NoError(t, err)

		// Create workflow
		workflowID := uuid.New()
		var definitionParam interface{}
		if definition == nil {
			definitionParam = nil
		} else {
			definitionParam = definition
		}
		_, err = db.Exec(`
			INSERT INTO workflows (id, user_id, name, description, definition, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			workflowID.String(), userID.String(), "Test Workflow", "Test Description", definitionParam, time.Now(), time.Now())
		require.NoError(t, err)

		return userID, workflowID
	}

	t.Run("GetWorkflowDraft - creates empty draft for workflow with no definition", func(t *testing.T) {
		_, workflowID := createTestWorkflow(t, nil)

		req := GetWorkflowDraftRequestObject{
			WorkflowId: workflowID,
		}

		resp, err := handlers.GetWorkflowDraft(context.Background(), req)
		require.NoError(t, err)

		// Should return 200 with empty definition
		draft, ok := resp.(GetWorkflowDraft200JSONResponse)
		require.True(t, ok)

		assert.Equal(t, workflowID, *draft.WorkflowId)
		assert.NotNil(t, draft.Definition)
		assert.NotNil(t, draft.Definition.Nodes)
		assert.NotNil(t, draft.Definition.Edges)
		assert.Empty(t, draft.Definition.Nodes)
		assert.Empty(t, draft.Definition.Edges)
		assert.NotNil(t, draft.UpdatedAt)
	})

	t.Run("GetWorkflowDraft - returns 404 for non-existent workflow", func(t *testing.T) {
		workflowID := uuid.New()

		req := GetWorkflowDraftRequestObject{
			WorkflowId: workflowID,
		}

		resp, err := handlers.GetWorkflowDraft(context.Background(), req)
		require.NoError(t, err)

		// Should return 404
		_, ok := resp.(GetWorkflowDraft404JSONResponse)
		assert.True(t, ok)
	})

	t.Run("UpdateWorkflowDraft - creates and retrieves draft", func(t *testing.T) {
		_, workflowID := createTestWorkflow(t, nil)

		// Create draft with test content
		nodeID := uuid.New()
		definition := WorkflowDefinition{
			Nodes: []WorkflowNode{
				{
					Id:     nodeID.String(),
					Name:   "Test Node",
					Type:   "test_node",
					Config: NodeConfig{},
				},
			},
			Edges: []WorkflowEdge{},
		}

		updateReq := UpdateWorkflowDraftRequestObject{
			WorkflowId: workflowID,
			Body: &UpdateWorkflowDraftRequest{
				Definition: &definition,
			},
		}

		updateResp, err := handlers.UpdateWorkflowDraft(context.Background(), updateReq)
		require.NoError(t, err)

		// Should return 200
		updateDraft, ok := updateResp.(UpdateWorkflowDraft200JSONResponse)
		require.True(t, ok)

		assert.Equal(t, workflowID, *updateDraft.WorkflowId)
		assert.NotNil(t, updateDraft.Definition)
		assert.Len(t, updateDraft.Definition.Nodes, 1)
		assert.Equal(t, "test_node", (updateDraft.Definition.Nodes)[0].Type)

		// Verify we can retrieve the draft
		getReq := GetWorkflowDraftRequestObject{
			WorkflowId: workflowID,
		}

		getResp, err := handlers.GetWorkflowDraft(context.Background(), getReq)
		require.NoError(t, err)

		getDraft, ok := getResp.(GetWorkflowDraft200JSONResponse)
		require.True(t, ok)

		assert.Equal(t, workflowID, *getDraft.WorkflowId)
		assert.Len(t, getDraft.Definition.Nodes, 1)
		assert.Equal(t, "test_node", (getDraft.Definition.Nodes)[0].Type)

		// Verify draft was saved to database
		var savedDefinition []byte
		err = db.QueryRow(`
			SELECT definition FROM workflow_drafts WHERE workflow_id = $1`,
			workflowID.String()).Scan(&savedDefinition)
		require.NoError(t, err)

		var savedDef WorkflowDefinition
		err = json.Unmarshal(savedDefinition, &savedDef)
		require.NoError(t, err)
		assert.Len(t, savedDef.Nodes, 1)
		assert.Equal(t, "test_node", (savedDef.Nodes)[0].Type)
	})

	t.Run("UpdateWorkflowDraft - returns 404 for non-existent workflow", func(t *testing.T) {
		workflowID := uuid.New()

		definition := WorkflowDefinition{
			Nodes: []WorkflowNode{},
			Edges: []WorkflowEdge{},
		}

		req := UpdateWorkflowDraftRequestObject{
			WorkflowId: workflowID,
			Body: &UpdateWorkflowDraftRequest{
				Definition: &definition,
			},
		}

		resp, err := handlers.UpdateWorkflowDraft(context.Background(), req)
		require.NoError(t, err)

		// Should return 404
		_, ok := resp.(UpdateWorkflowDraft404JSONResponse)
		assert.True(t, ok)
	})
}
