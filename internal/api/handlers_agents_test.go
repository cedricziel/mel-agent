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

// TestOpenAPICreateAgent tests creating a new agent
func TestOpenAPICreateAgent(t *testing.T) {
	db, cleanup := testutil.SetupOpenAPITestDB(t)
	mockEngine := execution.NewMockExecutionEngine()
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	createReq := CreateAgentRequest{
		Name:        "Test Agent",
		Description: testutil.StringPtr("A test agent for testing"),
	}

	reqBody, err := json.Marshal(createReq)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/agents", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response Agent
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "Test Agent", response.Name)
	assert.NotNil(t, response.Description)
	assert.Equal(t, "A test agent for testing", *response.Description)
	assert.NotEqual(t, uuid.Nil, response.Id)
	assert.False(t, response.CreatedAt.IsZero())

	// Verify the agent was actually created in the database
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM agents WHERE id = $1", response.Id.String()).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

// TestOpenAPICreateAgentMinimal tests creating an agent with minimal data
func TestOpenAPICreateAgentMinimal(t *testing.T) {
	db, cleanup := testutil.SetupOpenAPITestDB(t)
	mockEngine := execution.NewMockExecutionEngine()
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	createReq := CreateAgentRequest{
		Name: "Minimal Agent",
	}

	reqBody, err := json.Marshal(createReq)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/agents", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response Agent
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "Minimal Agent", response.Name)
	assert.Nil(t, response.Description)
	assert.NotEqual(t, uuid.Nil, response.Id)
	assert.False(t, response.CreatedAt.IsZero())
}

// TestOpenAPIListAgents tests listing agents
func TestOpenAPIListAgents(t *testing.T) {
	db, cleanup := testutil.SetupOpenAPITestDB(t)
	mockEngine := execution.NewMockExecutionEngine()
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	// Create test agents
	testAgents := []string{"Agent 1", "Agent 2", "Agent 3"}
	for _, name := range testAgents {
		createReq := CreateAgentRequest{Name: name}
		reqBody, _ := json.Marshal(createReq)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/agents", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)
	}

	// List agents
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/agents", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response AgentList
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.NotNil(t, response.Agents)
	assert.Len(t, *response.Agents, 3)
	assert.Equal(t, 3, *response.Total)
}

// TestOpenAPIListAgentsWithPagination tests listing agents with pagination
func TestOpenAPIListAgentsWithPagination(t *testing.T) {
	db, cleanup := testutil.SetupOpenAPITestDB(t)
	mockEngine := execution.NewMockExecutionEngine()
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	// Create test agents
	for i := 0; i < 5; i++ {
		createReq := CreateAgentRequest{Name: fmt.Sprintf("Agent %d", i+1)}
		reqBody, _ := json.Marshal(createReq)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/agents", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)
	}

	// List agents with pagination
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/agents?page=1&limit=2", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response AgentList
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.NotNil(t, response.Agents)
	assert.Len(t, *response.Agents, 2)
	assert.Equal(t, 5, *response.Total)
	assert.Equal(t, 1, *response.Page)
	assert.Equal(t, 2, *response.Limit)
}

// TestOpenAPIGetAgent tests retrieving a single agent
func TestOpenAPIGetAgent(t *testing.T) {
	db, cleanup := testutil.SetupOpenAPITestDB(t)
	mockEngine := execution.NewMockExecutionEngine()
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	// Create test agent
	createReq := CreateAgentRequest{
		Name:        "Get Test Agent",
		Description: testutil.StringPtr("Agent for get testing"),
	}
	reqBody, _ := json.Marshal(createReq)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/agents", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var createdAgent Agent
	json.NewDecoder(w.Body).Decode(&createdAgent)

	// Get the agent
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/agents/%s", createdAgent.Id), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response Agent
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, createdAgent.Id, response.Id)
	assert.Equal(t, "Get Test Agent", response.Name)
	assert.Equal(t, "Agent for get testing", *response.Description)
}

// TestOpenAPIGetAgentNotFound tests retrieving a non-existent agent
func TestOpenAPIGetAgentNotFound(t *testing.T) {
	db, cleanup := testutil.SetupOpenAPITestDB(t)
	mockEngine := execution.NewMockExecutionEngine()
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	nonExistentID := uuid.New()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/agents/%s", nonExistentID), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response Error
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "not found", *response.Error)
}

// TestOpenAPIUpdateAgent tests updating an agent
func TestOpenAPIUpdateAgent(t *testing.T) {
	db, cleanup := testutil.SetupOpenAPITestDB(t)
	mockEngine := execution.NewMockExecutionEngine()
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	// Create test agent
	createReq := CreateAgentRequest{
		Name:        "Original Agent",
		Description: testutil.StringPtr("Original description"),
	}
	reqBody, _ := json.Marshal(createReq)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/agents", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var createdAgent Agent
	json.NewDecoder(w.Body).Decode(&createdAgent)

	// Update the agent
	updateReq := UpdateAgentRequest{
		Name:        testutil.StringPtr("Updated Agent"),
		Description: testutil.StringPtr("Updated description"),
	}
	reqBody, _ = json.Marshal(updateReq)

	w = httptest.NewRecorder()
	req = httptest.NewRequest("PUT", fmt.Sprintf("/api/agents/%s", createdAgent.Id), bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response Agent
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, createdAgent.Id, response.Id)
	assert.Equal(t, "Updated Agent", response.Name)
	assert.NotNil(t, response.Description)
	assert.Equal(t, "Updated description", *response.Description)
}

// TestOpenAPIUpdateAgentNotFound tests updating a non-existent agent
func TestOpenAPIUpdateAgentNotFound(t *testing.T) {
	db, cleanup := testutil.SetupOpenAPITestDB(t)
	mockEngine := execution.NewMockExecutionEngine()
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	nonExistentID := uuid.New()
	updateReq := UpdateAgentRequest{
		Name: testutil.StringPtr("Updated Agent"),
	}
	reqBody, _ := json.Marshal(updateReq)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", fmt.Sprintf("/api/agents/%s", nonExistentID), bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response Error
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "not found", *response.Error)
}

// TestOpenAPIDeleteAgent tests deleting an agent
func TestOpenAPIDeleteAgent(t *testing.T) {
	db, cleanup := testutil.SetupOpenAPITestDB(t)
	mockEngine := execution.NewMockExecutionEngine()
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	// Create test agent
	createReq := CreateAgentRequest{Name: "Delete Test Agent"}
	reqBody, _ := json.Marshal(createReq)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/agents", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var createdAgent Agent
	json.NewDecoder(w.Body).Decode(&createdAgent)

	// Delete the agent
	w = httptest.NewRecorder()
	req = httptest.NewRequest("DELETE", fmt.Sprintf("/api/agents/%s", createdAgent.Id), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	// Verify the agent was deleted
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM agents WHERE id = $1", createdAgent.Id.String()).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

// TestOpenAPIDeleteAgentNotFound tests deleting a non-existent agent
func TestOpenAPIDeleteAgentNotFound(t *testing.T) {
	db, cleanup := testutil.SetupOpenAPITestDB(t)
	mockEngine := execution.NewMockExecutionEngine()
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	nonExistentID := uuid.New()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/agents/%s", nonExistentID), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response Error
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "not found", *response.Error)
}
