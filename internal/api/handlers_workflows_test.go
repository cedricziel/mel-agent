package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cedricziel/mel-agent/internal/testutil"
	"github.com/cedricziel/mel-agent/pkg/execution"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOpenAPICreateWorkflow tests creating a new workflow
func TestOpenAPICreateWorkflow(t *testing.T) {
	db, cleanup := testutil.SetupOpenAPITestDB(t)
	mockEngine := execution.NewMockExecutionEngine()
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	// Create a test workflow definition
	definition := WorkflowDefinition{
		Nodes: &[]WorkflowNode{
			{
				Id:   "node-1",
				Type: "start",
				Name: "Start Node",
				Config: map[string]interface{}{
					"trigger": "manual",
				},
			},
			{
				Id:   "node-2",
				Type: "action",
				Name: "Action Node",
				Config: map[string]interface{}{
					"action": "send_email",
				},
			},
		},
		Edges: &[]WorkflowEdge{
			{
				Id:     "edge-1",
				Source: "node-1",
				Target: "node-2",
			},
		},
	}

	createReq := CreateWorkflowRequest{
		Name:        "Test Workflow",
		Description: testutil.StringPtr("A test workflow for testing"),
		Definition:  &definition,
	}

	reqBody, err := json.Marshal(createReq)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/workflows", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Logf("Response body: %s", w.Body.String())
	}
	assert.Equal(t, http.StatusCreated, w.Code)

	var response Workflow
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "Test Workflow", response.Name)
	assert.Equal(t, "A test workflow for testing", *response.Description)
	assert.NotNil(t, response.Definition)
	assert.NotNil(t, response.Definition.Nodes)
	assert.Len(t, *response.Definition.Nodes, 2)
	assert.NotEqual(t, uuid.Nil, response.Id)
	assert.False(t, response.CreatedAt.IsZero())
	assert.False(t, response.UpdatedAt.IsZero())

	// Verify the workflow was actually created in the database
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM workflows WHERE id = $1", response.Id.String()).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

// TestOpenAPICreateWorkflowMinimal tests creating a workflow with minimal data
func TestOpenAPICreateWorkflowMinimal(t *testing.T) {
	db, cleanup := testutil.SetupOpenAPITestDB(t)
	mockEngine := execution.NewMockExecutionEngine()
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	createReq := CreateWorkflowRequest{
		Name: "Minimal Workflow",
	}

	reqBody, err := json.Marshal(createReq)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/workflows", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Logf("Response body: %s", w.Body.String())
	}
	assert.Equal(t, http.StatusCreated, w.Code)

	var response Workflow
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "Minimal Workflow", response.Name)
	assert.Nil(t, response.Description)
	assert.Nil(t, response.Definition)
}

// TestOpenAPIListWorkflows tests listing workflows with pagination
func TestOpenAPIListWorkflows(t *testing.T) {
	db, cleanup := testutil.SetupOpenAPITestDB(t)
	mockEngine := execution.NewMockExecutionEngine()
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	// Create test workflows
	testWorkflows := []string{"Workflow 1", "Workflow 2", "Workflow 3", "Workflow 4", "Workflow 5"}
	for _, name := range testWorkflows {
		createReq := CreateWorkflowRequest{
			Name:        name,
			Description: testutil.StringPtr(fmt.Sprintf("Description for %s", name)),
		}
		reqBody, _ := json.Marshal(createReq)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/workflows", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)
	}

	// List workflows with default pagination
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/workflows", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response WorkflowList
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.NotNil(t, response.Workflows)
	assert.Len(t, *response.Workflows, 5)
	assert.Equal(t, 5, *response.Total)
	assert.Equal(t, 1, *response.Page)
	assert.Equal(t, 20, *response.Limit)

	// Check that we have all the expected workflows
	workflowNames := make(map[string]bool)
	for _, workflow := range *response.Workflows {
		workflowNames[workflow.Name] = true
	}

	for _, expectedName := range testWorkflows {
		assert.True(t, workflowNames[expectedName], "Should find workflow %s", expectedName)
	}
}

// TestOpenAPIListWorkflowsWithPagination tests listing workflows with custom pagination
func TestOpenAPIListWorkflowsWithPagination(t *testing.T) {
	db, cleanup := testutil.SetupOpenAPITestDB(t)
	mockEngine := execution.NewMockExecutionEngine()
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	// Create 7 test workflows
	for i := 1; i <= 7; i++ {
		createReq := CreateWorkflowRequest{
			Name: fmt.Sprintf("Paginated Workflow %d", i),
		}
		reqBody, _ := json.Marshal(createReq)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/workflows", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)
	}

	// List workflows with pagination: page 1, limit 3
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/workflows?page=1&limit=3", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response WorkflowList
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.NotNil(t, response.Workflows)
	assert.Len(t, *response.Workflows, 3)
	assert.Equal(t, 7, *response.Total)
	assert.Equal(t, 1, *response.Page)
	assert.Equal(t, 3, *response.Limit)

	// List workflows with pagination: page 2, limit 3
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/workflows?page=2&limit=3", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.NotNil(t, response.Workflows)
	assert.Len(t, *response.Workflows, 3)
	assert.Equal(t, 7, *response.Total)
	assert.Equal(t, 2, *response.Page)
	assert.Equal(t, 3, *response.Limit)
}

// TestOpenAPIGetWorkflow tests retrieving a single workflow
func TestOpenAPIGetWorkflow(t *testing.T) {
	db, cleanup := testutil.SetupOpenAPITestDB(t)
	mockEngine := execution.NewMockExecutionEngine()
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	// Create a test workflow with definition
	definition := WorkflowDefinition{
		Nodes: &[]WorkflowNode{
			{
				Id:   "get-node-1",
				Type: "trigger",
				Name: "Webhook Trigger",
				Config: map[string]interface{}{
					"trigger_type": "webhook",
				},
			},
		},
	}

	createReq := CreateWorkflowRequest{
		Name:        "Get Test Workflow",
		Description: testutil.StringPtr("A workflow for get testing"),
		Definition:  &definition,
	}
	reqBody, _ := json.Marshal(createReq)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/workflows", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var createdWorkflow Workflow
	json.NewDecoder(w.Body).Decode(&createdWorkflow)

	// Get the workflow
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/workflows/%s", createdWorkflow.Id), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response Workflow
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, createdWorkflow.Id, response.Id)
	assert.Equal(t, "Get Test Workflow", response.Name)
	assert.Equal(t, "A workflow for get testing", *response.Description)
	assert.NotNil(t, response.Definition)
	assert.NotNil(t, response.Definition.Nodes)
	assert.Len(t, *response.Definition.Nodes, 1)
	assert.Equal(t, "get-node-1", (*response.Definition.Nodes)[0].Id)
}

// TestOpenAPIGetWorkflowNotFound tests retrieving a non-existent workflow
func TestOpenAPIGetWorkflowNotFound(t *testing.T) {
	db, cleanup := testutil.SetupOpenAPITestDB(t)
	mockEngine := execution.NewMockExecutionEngine()
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	nonExistentID := uuid.New()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/workflows/%s", nonExistentID), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response Error
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "not found", *response.Error)
}

// TestOpenAPIUpdateWorkflow tests updating a workflow
func TestOpenAPIUpdateWorkflow(t *testing.T) {
	db, cleanup := testutil.SetupOpenAPITestDB(t)
	mockEngine := execution.NewMockExecutionEngine()
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	// Create a test workflow
	createReq := CreateWorkflowRequest{
		Name:        "Original Workflow",
		Description: testutil.StringPtr("Original description"),
	}
	reqBody, _ := json.Marshal(createReq)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/workflows", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var createdWorkflow Workflow
	json.NewDecoder(w.Body).Decode(&createdWorkflow)

	// Update the workflow
	newDefinition := WorkflowDefinition{
		Nodes: &[]WorkflowNode{
			{
				Id:   "updated-node",
				Type: "action",
				Name: "Updated Action",
				Config: map[string]interface{}{
					"action": "updated_action",
				},
			},
		},
	}

	updateReq := UpdateWorkflowRequest{
		Name:        testutil.StringPtr("Updated Workflow"),
		Description: testutil.StringPtr("Updated description"),
		Definition:  &newDefinition,
	}
	reqBody, _ = json.Marshal(updateReq)

	w = httptest.NewRecorder()
	req = httptest.NewRequest("PUT", fmt.Sprintf("/api/workflows/%s", createdWorkflow.Id), bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response Workflow
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, createdWorkflow.Id, response.Id)
	assert.Equal(t, "Updated Workflow", response.Name)
	assert.Equal(t, "Updated description", *response.Description)
	assert.NotNil(t, response.Definition)
	assert.NotNil(t, response.Definition.Nodes)
	assert.Len(t, *response.Definition.Nodes, 1)
	assert.True(t, response.UpdatedAt.After(createdWorkflow.UpdatedAt))
}

// TestOpenAPIUpdateWorkflowNotFound tests updating a non-existent workflow
func TestOpenAPIUpdateWorkflowNotFound(t *testing.T) {
	db, cleanup := testutil.SetupOpenAPITestDB(t)
	mockEngine := execution.NewMockExecutionEngine()
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	nonExistentID := uuid.New()
	updateReq := UpdateWorkflowRequest{
		Name: testutil.StringPtr("Updated Workflow"),
	}
	reqBody, _ := json.Marshal(updateReq)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", fmt.Sprintf("/api/workflows/%s", nonExistentID), bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response Error
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "not found", *response.Error)
}

// TestOpenAPIDeleteWorkflow tests deleting a workflow
func TestOpenAPIDeleteWorkflow(t *testing.T) {
	db, cleanup := testutil.SetupOpenAPITestDB(t)
	mockEngine := execution.NewMockExecutionEngine()
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	// Create a test workflow
	createReq := CreateWorkflowRequest{
		Name: "Delete Test Workflow",
	}
	reqBody, _ := json.Marshal(createReq)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/workflows", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var createdWorkflow Workflow
	json.NewDecoder(w.Body).Decode(&createdWorkflow)

	// Delete the workflow
	w = httptest.NewRecorder()
	req = httptest.NewRequest("DELETE", fmt.Sprintf("/api/workflows/%s", createdWorkflow.Id), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	// Verify the workflow was deleted
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM workflows WHERE id = $1", createdWorkflow.Id.String()).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

// TestOpenAPIDeleteWorkflowNotFound tests deleting a non-existent workflow
func TestOpenAPIDeleteWorkflowNotFound(t *testing.T) {
	db, cleanup := testutil.SetupOpenAPITestDB(t)
	mockEngine := execution.NewMockExecutionEngine()
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	nonExistentID := uuid.New()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/workflows/%s", nonExistentID), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response Error
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "not found", *response.Error)
}

// TestOpenAPIExecuteWorkflow tests executing a workflow
func TestOpenAPIExecuteWorkflow(t *testing.T) {
	db, cleanup := testutil.SetupOpenAPITestDB(t)
	mockEngine := execution.NewMockExecutionEngine()
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	// Create a test workflow
	createReq := CreateWorkflowRequest{
		Name:        "Executable Workflow",
		Description: testutil.StringPtr("A workflow for execution testing"),
	}
	reqBody, _ := json.Marshal(createReq)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/workflows", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var createdWorkflow Workflow
	json.NewDecoder(w.Body).Decode(&createdWorkflow)

	// Execute the workflow
	executeReq := ExecuteWorkflowJSONBody{
		Input: &map[string]interface{}{
			"message": "Hello, World!",
			"count":   42,
		},
	}
	reqBody, _ = json.Marshal(executeReq)

	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", fmt.Sprintf("/api/workflows/%s/execute", createdWorkflow.Id), bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response WorkflowExecution
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.NotNil(t, response.Id)
	assert.Equal(t, createdWorkflow.Id, *response.WorkflowId)
	assert.Equal(t, WorkflowExecutionStatusPending, *response.Status)
	assert.NotNil(t, response.StartedAt)
	assert.NotNil(t, response.Result)
	assert.Equal(t, "Hello, World!", (*response.Result)["message"])
	assert.Equal(t, float64(42), (*response.Result)["count"])

	// Verify the execution was created in the database
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM workflow_executions WHERE id = $1", response.Id.String()).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

// TestOpenAPIExecuteWorkflowNotFound tests executing a non-existent workflow
func TestOpenAPIExecuteWorkflowNotFound(t *testing.T) {
	db, cleanup := testutil.SetupOpenAPITestDB(t)
	mockEngine := execution.NewMockExecutionEngine()
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	nonExistentID := uuid.New()
	executeReq := ExecuteWorkflowJSONBody{
		Input: &map[string]interface{}{
			"test": "data",
		},
	}
	reqBody, _ := json.Marshal(executeReq)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", fmt.Sprintf("/api/workflows/%s/execute", nonExistentID), bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response Error
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "not found", *response.Error)
}
