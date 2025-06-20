package api

import (
	"bytes"
	"database/sql"
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

// getTestIntegrationID gets the first available integration ID for testing
func getTestIntegrationID(t *testing.T, db any) uuid.UUID {
	t.Helper()

	var integrationID string
	err := db.(*sql.DB).QueryRow("SELECT id FROM integrations LIMIT 1").Scan(&integrationID)
	require.NoError(t, err, "Failed to get test integration ID")

	parsedID, err := uuid.Parse(integrationID)
	require.NoError(t, err, "Failed to parse integration ID")

	return parsedID
}

// TestOpenAPICreateConnection tests creating a new connection
func TestOpenAPICreateConnection(t *testing.T) {
	db, cleanup := testutil.SetupOpenAPITestDB(t)
	mockEngine := execution.NewMockExecutionEngine()
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	// Get a test integration ID
	integrationID := getTestIntegrationID(t, db)

	createReq := CreateConnectionRequest{
		Name:          "Test Connection",
		IntegrationId: integrationID,
		Secret: &map[string]interface{}{
			"api_key": "test-secret-key",
		},
		Config: &map[string]interface{}{
			"base_url": "https://api.example.com",
		},
		UsageLimitMonth: testutil.IntPtr(1000),
		IsDefault:       testutil.BoolPtr(false),
	}

	reqBody, err := json.Marshal(createReq)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/connections", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response Connection
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "Test Connection", *response.Name)
	assert.Equal(t, integrationID, *response.IntegrationId)
	assert.NotNil(t, response.Secret)
	assert.Equal(t, "test-secret-key", (*response.Secret)["api_key"])
	assert.NotNil(t, response.Config)
	assert.Equal(t, "https://api.example.com", (*response.Config)["base_url"])
	assert.NotNil(t, response.UsageLimitMonth)
	assert.Equal(t, 1000, *response.UsageLimitMonth)
	assert.NotNil(t, response.IsDefault)
	assert.Equal(t, false, *response.IsDefault)
	assert.Equal(t, ConnectionStatusValid, *response.Status)
	assert.NotEqual(t, uuid.Nil, *response.Id)
	assert.False(t, response.CreatedAt.IsZero())

	// Verify the connection was actually created in the database
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM connections WHERE id = $1", response.Id.String()).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

// TestOpenAPICreateConnectionMinimal tests creating a connection with minimal data
func TestOpenAPICreateConnectionMinimal(t *testing.T) {
	db, cleanup := testutil.SetupOpenAPITestDB(t)
	mockEngine := execution.NewMockExecutionEngine()
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	// Get a test integration ID
	integrationID := getTestIntegrationID(t, db)

	createReq := CreateConnectionRequest{
		Name:          "Minimal Connection",
		IntegrationId: integrationID,
	}

	reqBody, err := json.Marshal(createReq)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/connections", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Logf("Response body: %s", w.Body.String())
	}
	assert.Equal(t, http.StatusCreated, w.Code)

	var response Connection
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "Minimal Connection", *response.Name)
	assert.Equal(t, integrationID, *response.IntegrationId)
	assert.Nil(t, response.Secret)
	assert.Nil(t, response.Config)
	assert.Nil(t, response.UsageLimitMonth)
	assert.NotNil(t, response.IsDefault)
	assert.Equal(t, false, *response.IsDefault)
	assert.Equal(t, ConnectionStatusValid, *response.Status)
}

// TestOpenAPIListConnections tests listing connections
func TestOpenAPIListConnections(t *testing.T) {
	db, cleanup := testutil.SetupOpenAPITestDB(t)
	mockEngine := execution.NewMockExecutionEngine()
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	// Get a test integration ID
	integrationID := getTestIntegrationID(t, db)

	// Create test connections
	testConnections := []string{"Connection 1", "Connection 2", "Connection 3"}
	for _, name := range testConnections {
		createReq := CreateConnectionRequest{
			Name:          name,
			IntegrationId: integrationID,
		}
		reqBody, _ := json.Marshal(createReq)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/connections", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)
	}

	// List connections
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/connections", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []Connection
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Len(t, response, 3)

	// Check that we have all the expected connections
	connectionNames := make(map[string]bool)
	for _, conn := range response {
		connectionNames[*conn.Name] = true
	}

	for _, expectedName := range testConnections {
		assert.True(t, connectionNames[expectedName], "Should find connection %s", expectedName)
	}
}

// TestOpenAPIGetConnection tests retrieving a single connection
func TestOpenAPIGetConnection(t *testing.T) {
	db, cleanup := testutil.SetupOpenAPITestDB(t)
	mockEngine := execution.NewMockExecutionEngine()
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	// Get a test integration ID
	integrationID := getTestIntegrationID(t, db)

	// Create test connection
	createReq := CreateConnectionRequest{
		Name:          "Get Test Connection",
		IntegrationId: integrationID,
		Secret: &map[string]interface{}{
			"token": "secret-token",
		},
	}
	reqBody, _ := json.Marshal(createReq)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/connections", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var createdConnection Connection
	json.NewDecoder(w.Body).Decode(&createdConnection)

	// Get the connection
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/connections/%s", createdConnection.Id), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response Connection
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, *createdConnection.Id, *response.Id)
	assert.Equal(t, "Get Test Connection", *response.Name)
	assert.Equal(t, "secret-token", (*response.Secret)["token"])
}

// TestOpenAPIGetConnectionNotFound tests retrieving a non-existent connection
func TestOpenAPIGetConnectionNotFound(t *testing.T) {
	db, cleanup := testutil.SetupOpenAPITestDB(t)
	mockEngine := execution.NewMockExecutionEngine()
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	nonExistentID := uuid.New()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/connections/%s", nonExistentID), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response Error
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "not found", *response.Error)
}

// TestOpenAPIUpdateConnection tests updating a connection
func TestOpenAPIUpdateConnection(t *testing.T) {
	db, cleanup := testutil.SetupOpenAPITestDB(t)
	mockEngine := execution.NewMockExecutionEngine()
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	// Get a test integration ID
	integrationID := getTestIntegrationID(t, db)

	// Create test connection
	createReq := CreateConnectionRequest{
		Name:          "Original Connection",
		IntegrationId: integrationID,
		Secret: &map[string]interface{}{
			"api_key": "original-key",
		},
		UsageLimitMonth: testutil.IntPtr(500),
	}
	reqBody, _ := json.Marshal(createReq)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/connections", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var createdConnection Connection
	json.NewDecoder(w.Body).Decode(&createdConnection)

	// Update the connection
	updateReq := UpdateConnectionRequest{
		Name: testutil.StringPtr("Updated Connection"),
		Secret: &map[string]interface{}{
			"api_key": "updated-key",
		},
		UsageLimitMonth: testutil.IntPtr(1000),
		Status:          func() *UpdateConnectionRequestStatus { s := Invalid; return &s }(),
	}
	reqBody, _ = json.Marshal(updateReq)

	w = httptest.NewRecorder()
	req = httptest.NewRequest("PUT", fmt.Sprintf("/api/connections/%s", createdConnection.Id), bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response Connection
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, *createdConnection.Id, *response.Id)
	assert.Equal(t, "Updated Connection", *response.Name)
	assert.Equal(t, "updated-key", (*response.Secret)["api_key"])
	assert.Equal(t, 1000, *response.UsageLimitMonth)
	assert.Equal(t, ConnectionStatusInvalid, *response.Status)
}

// TestOpenAPIUpdateConnectionNotFound tests updating a non-existent connection
func TestOpenAPIUpdateConnectionNotFound(t *testing.T) {
	db, cleanup := testutil.SetupOpenAPITestDB(t)
	mockEngine := execution.NewMockExecutionEngine()
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	nonExistentID := uuid.New()
	updateReq := UpdateConnectionRequest{
		Name: testutil.StringPtr("Updated Connection"),
	}
	reqBody, _ := json.Marshal(updateReq)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", fmt.Sprintf("/api/connections/%s", nonExistentID), bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response Error
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "not found", *response.Error)
}

// TestOpenAPIDeleteConnection tests deleting a connection
func TestOpenAPIDeleteConnection(t *testing.T) {
	db, cleanup := testutil.SetupOpenAPITestDB(t)
	mockEngine := execution.NewMockExecutionEngine()
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	// Get a test integration ID
	integrationID := getTestIntegrationID(t, db)

	// Create test connection
	createReq := CreateConnectionRequest{
		Name:          "Delete Test Connection",
		IntegrationId: integrationID,
	}
	reqBody, _ := json.Marshal(createReq)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/connections", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var createdConnection Connection
	json.NewDecoder(w.Body).Decode(&createdConnection)

	// Delete the connection
	w = httptest.NewRecorder()
	req = httptest.NewRequest("DELETE", fmt.Sprintf("/api/connections/%s", createdConnection.Id), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	// Verify the connection was deleted
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM connections WHERE id = $1", createdConnection.Id.String()).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

// TestOpenAPIDeleteConnectionNotFound tests deleting a non-existent connection
func TestOpenAPIDeleteConnectionNotFound(t *testing.T) {
	db, cleanup := testutil.SetupOpenAPITestDB(t)
	mockEngine := execution.NewMockExecutionEngine()
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	nonExistentID := uuid.New()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/connections/%s", nonExistentID), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response Error
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "not found", *response.Error)
}
