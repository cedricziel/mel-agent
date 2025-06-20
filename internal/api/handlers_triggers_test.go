package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cedricziel/mel-agent/internal/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getTestAgentID gets the first available agent ID for testing (using agent as workflow)
func getTestAgentID(t *testing.T, db any) uuid.UUID {
	t.Helper()

	var agentID string
	err := db.(*sql.DB).QueryRow("SELECT id FROM agents LIMIT 1").Scan(&agentID)
	if err != nil {
		// If no agents exist, create one for testing
		testAgentID := uuid.New()
		userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

		_, err = db.(*sql.DB).Exec("INSERT INTO agents (id, user_id, name, created_at) VALUES ($1, $2, $3, $4)",
			testAgentID, userID, "Test Agent for Triggers", time.Now())
		require.NoError(t, err, "Failed to create test agent")

		return testAgentID
	}

	parsedID, err := uuid.Parse(agentID)
	require.NoError(t, err, "Failed to parse agent ID")

	return parsedID
}

// TestOpenAPICreateTrigger tests creating a new trigger
func TestOpenAPICreateTrigger(t *testing.T) {
	db, mockEngine, cleanup := testutil.SetupOpenAPITestDB(t)
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	// Get a test agent ID (workflow ID)
	agentID := getTestAgentID(t, db)

	createReq := CreateTriggerRequest{
		Name:       "Test Schedule Trigger",
		Type:       CreateTriggerRequestTypeSchedule,
		WorkflowId: agentID,
		Config: &map[string]interface{}{
			"schedule": "0 */6 * * *",
			"timezone": "UTC",
		},
		Enabled: testutil.BoolPtr(true),
	}

	reqBody, err := json.Marshal(createReq)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/triggers", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Logf("Response body: %s", w.Body.String())
	}
	assert.Equal(t, http.StatusCreated, w.Code)

	var response Trigger
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "Test Schedule Trigger", *response.Name)
	assert.Equal(t, TriggerTypeSchedule, *response.Type)
	assert.Equal(t, agentID, *response.WorkflowId)
	assert.NotNil(t, response.Config)
	assert.Equal(t, "0 */6 * * *", (*response.Config)["schedule"])
	assert.Equal(t, "UTC", (*response.Config)["timezone"])
	assert.NotNil(t, response.Enabled)
	assert.Equal(t, true, *response.Enabled)
	assert.NotEqual(t, uuid.Nil, *response.Id)
	assert.False(t, response.CreatedAt.IsZero())
	assert.False(t, response.UpdatedAt.IsZero())

	// Verify the trigger was actually created in the database
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM triggers WHERE id = $1", response.Id.String()).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

// TestOpenAPICreateTriggerWebhook tests creating a webhook trigger
func TestOpenAPICreateTriggerWebhook(t *testing.T) {
	db, mockEngine, cleanup := testutil.SetupOpenAPITestDB(t)
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	// Get a test agent ID (workflow ID)
	agentID := getTestAgentID(t, db)

	createReq := CreateTriggerRequest{
		Name:       "Test Webhook Trigger",
		Type:       CreateTriggerRequestTypeWebhook,
		WorkflowId: agentID,
		Config: &map[string]interface{}{
			"path":   "/webhook/test",
			"method": "POST",
		},
		Enabled: testutil.BoolPtr(false),
	}

	reqBody, err := json.Marshal(createReq)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/triggers", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response Trigger
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "Test Webhook Trigger", *response.Name)
	assert.Equal(t, TriggerTypeWebhook, *response.Type)
	assert.Equal(t, agentID, *response.WorkflowId)
	assert.NotNil(t, response.Config)
	assert.Equal(t, "/webhook/test", (*response.Config)["path"])
	assert.Equal(t, "POST", (*response.Config)["method"])
	assert.NotNil(t, response.Enabled)
	assert.Equal(t, false, *response.Enabled)
}

// TestOpenAPICreateTriggerMinimal tests creating a trigger with minimal data
func TestOpenAPICreateTriggerMinimal(t *testing.T) {
	db, mockEngine, cleanup := testutil.SetupOpenAPITestDB(t)
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	// Get a test agent ID (workflow ID)
	agentID := getTestAgentID(t, db)

	createReq := CreateTriggerRequest{
		Name:       "Minimal Trigger",
		Type:       CreateTriggerRequestTypeSchedule,
		WorkflowId: agentID,
	}

	reqBody, err := json.Marshal(createReq)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/triggers", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Logf("Response body: %s", w.Body.String())
	}
	assert.Equal(t, http.StatusCreated, w.Code)

	var response Trigger
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "Minimal Trigger", *response.Name)
	assert.Equal(t, TriggerTypeSchedule, *response.Type)
	assert.Equal(t, agentID, *response.WorkflowId)
	assert.Nil(t, response.Config)
	assert.NotNil(t, response.Enabled)
	assert.Equal(t, true, *response.Enabled) // Default is true
}

// TestOpenAPIListTriggers tests listing triggers
func TestOpenAPIListTriggers(t *testing.T) {
	db, mockEngine, cleanup := testutil.SetupOpenAPITestDB(t)
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	// Get a test agent ID (workflow ID)
	agentID := getTestAgentID(t, db)

	// Create test triggers
	testTriggers := []struct {
		name        string
		triggerType CreateTriggerRequestType
	}{
		{"Schedule Trigger 1", CreateTriggerRequestTypeSchedule},
		{"Webhook Trigger 1", CreateTriggerRequestTypeWebhook},
		{"Schedule Trigger 2", CreateTriggerRequestTypeSchedule},
	}

	for _, trigger := range testTriggers {
		createReq := CreateTriggerRequest{
			Name:       trigger.name,
			Type:       trigger.triggerType,
			WorkflowId: agentID,
		}
		reqBody, _ := json.Marshal(createReq)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/triggers", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)
	}

	// List triggers
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/triggers", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []Trigger
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Len(t, response, 3)

	// Check that we have all the expected triggers
	triggerNames := make(map[string]bool)
	for _, trigger := range response {
		triggerNames[*trigger.Name] = true
	}

	for _, expectedTrigger := range testTriggers {
		assert.True(t, triggerNames[expectedTrigger.name], "Should find trigger %s", expectedTrigger.name)
	}
}

// TestOpenAPIGetTrigger tests retrieving a single trigger
func TestOpenAPIGetTrigger(t *testing.T) {
	db, mockEngine, cleanup := testutil.SetupOpenAPITestDB(t)
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	// Get a test agent ID (workflow ID)
	agentID := getTestAgentID(t, db)

	// Create test trigger
	createReq := CreateTriggerRequest{
		Name:       "Get Test Trigger",
		Type:       CreateTriggerRequestTypeWebhook,
		WorkflowId: agentID,
		Config: &map[string]interface{}{
			"path": "/webhook/get-test",
		},
		Enabled: testutil.BoolPtr(false),
	}
	reqBody, _ := json.Marshal(createReq)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/triggers", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var createdTrigger Trigger
	json.NewDecoder(w.Body).Decode(&createdTrigger)

	// Get the trigger
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/triggers/%s", createdTrigger.Id), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response Trigger
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, *createdTrigger.Id, *response.Id)
	assert.Equal(t, "Get Test Trigger", *response.Name)
	assert.Equal(t, TriggerTypeWebhook, *response.Type)
	assert.Equal(t, agentID, *response.WorkflowId)
	assert.Equal(t, "/webhook/get-test", (*response.Config)["path"])
	assert.Equal(t, false, *response.Enabled)
}

// TestOpenAPIGetTriggerNotFound tests retrieving a non-existent trigger
func TestOpenAPIGetTriggerNotFound(t *testing.T) {
	db, mockEngine, cleanup := testutil.SetupOpenAPITestDB(t)
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	nonExistentID := uuid.New()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/triggers/%s", nonExistentID), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response Error
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "not found", *response.Error)
}

// TestOpenAPIUpdateTrigger tests updating a trigger
func TestOpenAPIUpdateTrigger(t *testing.T) {
	db, mockEngine, cleanup := testutil.SetupOpenAPITestDB(t)
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	// Get a test agent ID (workflow ID)
	agentID := getTestAgentID(t, db)

	// Create test trigger
	createReq := CreateTriggerRequest{
		Name:       "Original Trigger",
		Type:       CreateTriggerRequestTypeSchedule,
		WorkflowId: agentID,
		Config: &map[string]interface{}{
			"schedule": "0 0 * * *",
		},
		Enabled: testutil.BoolPtr(true),
	}
	reqBody, _ := json.Marshal(createReq)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/triggers", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var createdTrigger Trigger
	json.NewDecoder(w.Body).Decode(&createdTrigger)

	// Update the trigger
	updateReq := UpdateTriggerRequest{
		Name: testutil.StringPtr("Updated Trigger"),
		Config: &map[string]interface{}{
			"schedule": "0 12 * * *",
			"timezone": "America/New_York",
		},
		Enabled: testutil.BoolPtr(false),
	}
	reqBody, _ = json.Marshal(updateReq)

	w = httptest.NewRecorder()
	req = httptest.NewRequest("PUT", fmt.Sprintf("/api/triggers/%s", createdTrigger.Id), bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response Trigger
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, *createdTrigger.Id, *response.Id)
	assert.Equal(t, "Updated Trigger", *response.Name)
	assert.Equal(t, TriggerTypeSchedule, *response.Type)
	assert.Equal(t, agentID, *response.WorkflowId)
	assert.Equal(t, "0 12 * * *", (*response.Config)["schedule"])
	assert.Equal(t, "America/New_York", (*response.Config)["timezone"])
	assert.Equal(t, false, *response.Enabled)
	assert.True(t, response.UpdatedAt.After(*createdTrigger.UpdatedAt))
}

// TestOpenAPIUpdateTriggerNotFound tests updating a non-existent trigger
func TestOpenAPIUpdateTriggerNotFound(t *testing.T) {
	db, mockEngine, cleanup := testutil.SetupOpenAPITestDB(t)
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	nonExistentID := uuid.New()
	updateReq := UpdateTriggerRequest{
		Name: testutil.StringPtr("Updated Trigger"),
	}
	reqBody, _ := json.Marshal(updateReq)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", fmt.Sprintf("/api/triggers/%s", nonExistentID), bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response Error
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "not found", *response.Error)
}

// TestOpenAPIDeleteTrigger tests deleting a trigger
func TestOpenAPIDeleteTrigger(t *testing.T) {
	db, mockEngine, cleanup := testutil.SetupOpenAPITestDB(t)
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	// Get a test agent ID (workflow ID)
	agentID := getTestAgentID(t, db)

	// Create test trigger
	createReq := CreateTriggerRequest{
		Name:       "Delete Test Trigger",
		Type:       CreateTriggerRequestTypeSchedule,
		WorkflowId: agentID,
	}
	reqBody, _ := json.Marshal(createReq)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/triggers", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var createdTrigger Trigger
	json.NewDecoder(w.Body).Decode(&createdTrigger)

	// Delete the trigger
	w = httptest.NewRecorder()
	req = httptest.NewRequest("DELETE", fmt.Sprintf("/api/triggers/%s", createdTrigger.Id), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	// Verify the trigger was deleted
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM triggers WHERE id = $1", createdTrigger.Id.String()).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

// TestOpenAPIDeleteTriggerNotFound tests deleting a non-existent trigger
func TestOpenAPIDeleteTriggerNotFound(t *testing.T) {
	db, mockEngine, cleanup := testutil.SetupOpenAPITestDB(t)
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	nonExistentID := uuid.New()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/triggers/%s", nonExistentID), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response Error
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "not found", *response.Error)
}
