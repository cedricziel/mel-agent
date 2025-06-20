package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cedricziel/mel-agent/internal/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOpenAPIHandleWebhook tests handling a valid webhook request
func TestOpenAPIHandleWebhook(t *testing.T) {
	db, mockEngine, cleanup := testutil.SetupOpenAPITestDB(t)
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	// Create a test workflow first
	createWorkflowReq := CreateWorkflowRequest{
		Name:        "Webhook Test Workflow",
		Description: testutil.StringPtr("A workflow for webhook testing"),
	}
	reqBody, _ := json.Marshal(createWorkflowReq)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/workflows", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var createdWorkflow Workflow
	json.NewDecoder(w.Body).Decode(&createdWorkflow)

	// Create a webhook trigger
	webhookToken := "test-webhook-token-123"
	config := map[string]interface{}{
		"token":  webhookToken,
		"method": "POST",
		"secret": "webhook-secret",
	}
	configJson, _ := json.Marshal(config)

	// Use default user_id for testing
	defaultUserID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	_, err := db.Exec(`
		INSERT INTO triggers (id, user_id, provider, name, type, workflow_id, config, enabled) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		uuid.New(), defaultUserID, "webhook", "Test Webhook Trigger", "webhook", createdWorkflow.Id, configJson, true)
	require.NoError(t, err)

	// Test webhook payload
	webhookPayload := map[string]interface{}{
		"event":     "user.created",
		"user_id":   "user_123",
		"email":     "test@example.com",
		"timestamp": "2023-06-20T12:00:00Z",
		"metadata": map[string]interface{}{
			"source": "registration_form",
			"ip":     "192.168.1.100",
		},
	}
	payloadJson, _ := json.Marshal(webhookPayload)

	// Send webhook request
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", fmt.Sprintf("/webhooks/%s", webhookToken), bytes.NewBuffer(payloadJson))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "GitHub-Hookshot/1.0")
	req.Header.Set("X-GitHub-Event", "push")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify webhook event was stored in database
	var eventCount int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM webhook_events 
		WHERE payload @> $1`,
		payloadJson).Scan(&eventCount)
	require.NoError(t, err)
	assert.Equal(t, 1, eventCount)

	// Verify the stored payload contains correct data
	var storedPayload []byte
	err = db.QueryRow(`
		SELECT payload FROM webhook_events 
		WHERE payload @> $1`,
		payloadJson).Scan(&storedPayload)
	require.NoError(t, err)

	var stored map[string]interface{}
	json.Unmarshal(storedPayload, &stored)
	assert.Equal(t, "user.created", stored["event"])
	assert.Equal(t, "test@example.com", stored["email"])
}

// TestOpenAPIHandleWebhookNotFound tests webhook with invalid token
func TestOpenAPIHandleWebhookNotFound(t *testing.T) {
	db, mockEngine, cleanup := testutil.SetupOpenAPITestDB(t)
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	webhookPayload := map[string]interface{}{
		"event": "test",
		"data":  "some data",
	}
	payloadJson, _ := json.Marshal(webhookPayload)

	// Send webhook request with non-existent token
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/webhooks/non-existent-token", bytes.NewBuffer(payloadJson))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response Error
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "not found", *response.Error)
	assert.Equal(t, "Webhook not found or disabled", *response.Message)

	// Verify no webhook event was stored
	var eventCount int
	err = db.QueryRow("SELECT COUNT(*) FROM webhook_events").Scan(&eventCount)
	require.NoError(t, err)
	assert.Equal(t, 0, eventCount)
}

// TestOpenAPIHandleWebhookDisabled tests webhook with disabled trigger
func TestOpenAPIHandleWebhookDisabled(t *testing.T) {
	db, mockEngine, cleanup := testutil.SetupOpenAPITestDB(t)
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	// Create a test workflow
	createWorkflowReq := CreateWorkflowRequest{
		Name: "Disabled Webhook Test Workflow",
	}
	reqBody, _ := json.Marshal(createWorkflowReq)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/workflows", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var createdWorkflow Workflow
	json.NewDecoder(w.Body).Decode(&createdWorkflow)

	// Create a disabled webhook trigger
	webhookToken := "disabled-webhook-token"
	config := map[string]interface{}{
		"token": webhookToken,
	}
	configJson, _ := json.Marshal(config)

	// Use default user_id for testing
	defaultUserID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	_, err := db.Exec(`
		INSERT INTO triggers (id, user_id, provider, name, type, workflow_id, config, enabled) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		uuid.New(), defaultUserID, "webhook", "Disabled Webhook Trigger", "webhook", createdWorkflow.Id, configJson, false)
	require.NoError(t, err)

	webhookPayload := map[string]interface{}{
		"event": "test",
	}
	payloadJson, _ := json.Marshal(webhookPayload)

	// Send webhook request to disabled trigger
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", fmt.Sprintf("/webhooks/%s", webhookToken), bytes.NewBuffer(payloadJson))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response Error
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "not found", *response.Error)
	assert.Equal(t, "Webhook not found or disabled", *response.Message)
}

// TestOpenAPIHandleWebhookMultipleFormats tests webhook with different payload formats
func TestOpenAPIHandleWebhookMultipleFormats(t *testing.T) {
	db, mockEngine, cleanup := testutil.SetupOpenAPITestDB(t)
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	// Create a test workflow
	createWorkflowReq := CreateWorkflowRequest{
		Name: "Multi-format Webhook Test Workflow",
	}
	reqBody, _ := json.Marshal(createWorkflowReq)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/workflows", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var createdWorkflow Workflow
	json.NewDecoder(w.Body).Decode(&createdWorkflow)

	// Create webhook trigger
	webhookToken := "multi-format-token"
	config := map[string]interface{}{
		"token": webhookToken,
	}
	configJson, _ := json.Marshal(config)

	// Use default user_id for testing
	defaultUserID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	_, err := db.Exec(`
		INSERT INTO triggers (id, user_id, provider, name, type, workflow_id, config, enabled) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		uuid.New(), defaultUserID, "webhook", "Multi-format Webhook Trigger", "webhook", createdWorkflow.Id, configJson, true)
	require.NoError(t, err)

	// Test different payload formats
	testCases := []struct {
		name    string
		payload interface{}
	}{
		{
			name: "Simple object",
			payload: map[string]interface{}{
				"message": "Hello World",
			},
		},
		{
			name: "Complex nested object",
			payload: map[string]interface{}{
				"user": map[string]interface{}{
					"id":    123,
					"name":  "John Doe",
					"email": "john@example.com",
					"roles": []string{"admin", "user"},
				},
				"action": "login",
				"metadata": map[string]interface{}{
					"ip":        "192.168.1.1",
					"timestamp": "2023-06-20T15:30:00Z",
					"source":    "web",
				},
			},
		},
		{
			name: "Array payload",
			payload: []map[string]interface{}{
				{"id": 1, "name": "Item 1"},
				{"id": 2, "name": "Item 2"},
			},
		},
		{
			name:    "String payload",
			payload: "Simple string payload",
		},
		{
			name:    "Number payload",
			payload: 42,
		},
		{
			name:    "Boolean payload",
			payload: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			payloadJson, _ := json.Marshal(tc.payload)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", fmt.Sprintf("/webhooks/%s", webhookToken), bytes.NewBuffer(payloadJson))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
		})
	}

	// Verify all webhook events were stored
	var eventCount int
	err = db.QueryRow("SELECT COUNT(*) FROM webhook_events").Scan(&eventCount)
	require.NoError(t, err)
	assert.Equal(t, len(testCases), eventCount)
}

// TestOpenAPIHandleWebhookGitHubFormat tests webhook with GitHub-style payload
func TestOpenAPIHandleWebhookGitHubFormat(t *testing.T) {
	db, mockEngine, cleanup := testutil.SetupOpenAPITestDB(t)
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	// Create a test workflow
	createWorkflowReq := CreateWorkflowRequest{
		Name: "GitHub Webhook Test Workflow",
	}
	reqBody, _ := json.Marshal(createWorkflowReq)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/workflows", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var createdWorkflow Workflow
	json.NewDecoder(w.Body).Decode(&createdWorkflow)

	// Create webhook trigger for GitHub
	webhookToken := "github-webhook-token"
	config := map[string]interface{}{
		"token":  webhookToken,
		"source": "github",
		"events": []string{"push", "pull_request"},
	}
	configJson, _ := json.Marshal(config)

	// Use default user_id for testing
	defaultUserID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	_, err := db.Exec(`
		INSERT INTO triggers (id, user_id, provider, name, type, workflow_id, config, enabled) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		uuid.New(), defaultUserID, "webhook", "GitHub Webhook Trigger", "webhook", createdWorkflow.Id, configJson, true)
	require.NoError(t, err)

	// GitHub push event payload (simplified)
	githubPayload := map[string]interface{}{
		"ref":    "refs/heads/main",
		"before": "0000000000000000000000000000000000000000",
		"after":  "1234567890123456789012345678901234567890",
		"repository": map[string]interface{}{
			"id":        123456789,
			"name":      "test-repo",
			"full_name": "user/test-repo",
			"owner": map[string]interface{}{
				"login": "user",
				"id":    987654321,
			},
		},
		"pusher": map[string]interface{}{
			"name":  "user",
			"email": "user@example.com",
		},
		"commits": []map[string]interface{}{
			{
				"id":      "1234567890123456789012345678901234567890",
				"message": "Initial commit",
				"author": map[string]interface{}{
					"name":  "Developer",
					"email": "dev@example.com",
				},
				"added":    []string{"README.md"},
				"removed":  []string{},
				"modified": []string{},
			},
		},
	}
	payloadJson, _ := json.Marshal(githubPayload)

	// Send GitHub webhook request
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", fmt.Sprintf("/webhooks/%s", webhookToken), bytes.NewBuffer(payloadJson))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "GitHub-Hookshot/abc123")
	req.Header.Set("X-GitHub-Event", "push")
	req.Header.Set("X-GitHub-Delivery", "12345678-1234-1234-1234-123456789012")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify the GitHub payload was stored correctly
	var storedPayload []byte
	err = db.QueryRow(`
		SELECT payload FROM webhook_events 
		WHERE payload->'repository'->>'name' = 'test-repo'`).Scan(&storedPayload)
	require.NoError(t, err)

	var stored map[string]interface{}
	json.Unmarshal(storedPayload, &stored)
	assert.Equal(t, "refs/heads/main", stored["ref"])
	assert.Equal(t, "test-repo", stored["repository"].(map[string]interface{})["name"])
	assert.Equal(t, "user@example.com", stored["pusher"].(map[string]interface{})["email"])
}

// TestOpenAPIHandleWebhookInvalidJSON tests webhook with malformed JSON
func TestOpenAPIHandleWebhookInvalidJSON(t *testing.T) {
	db, mockEngine, cleanup := testutil.SetupOpenAPITestDB(t)
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	// Create a test workflow and webhook trigger
	createWorkflowReq := CreateWorkflowRequest{
		Name: "Invalid JSON Test Workflow",
	}
	reqBody, _ := json.Marshal(createWorkflowReq)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/workflows", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var createdWorkflow Workflow
	json.NewDecoder(w.Body).Decode(&createdWorkflow)

	// Create webhook trigger
	webhookToken := "invalid-json-token"
	config := map[string]interface{}{
		"token": webhookToken,
	}
	configJson, _ := json.Marshal(config)

	// Use default user_id for testing
	defaultUserID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	_, err := db.Exec(`
		INSERT INTO triggers (id, user_id, provider, name, type, workflow_id, config, enabled) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		uuid.New(), defaultUserID, "webhook", "Invalid JSON Test Trigger", "webhook", createdWorkflow.Id, configJson, true)
	require.NoError(t, err)

	// Send webhook request with invalid JSON
	invalidJSON := `{"invalid": json, "missing": "quotes"}`

	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", fmt.Sprintf("/webhooks/%s", webhookToken), bytes.NewBufferString(invalidJSON))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// OpenAPI framework validates JSON and returns 400 for invalid JSON
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestOpenAPIHandleWebhookConcurrency tests multiple concurrent webhook requests
func TestOpenAPIHandleWebhookConcurrency(t *testing.T) {
	db, mockEngine, cleanup := testutil.SetupOpenAPITestDB(t)
	defer cleanup()

	router := NewOpenAPIRouter(db, mockEngine)

	// Create a test workflow
	createWorkflowReq := CreateWorkflowRequest{
		Name: "Concurrency Test Workflow",
	}
	reqBody, _ := json.Marshal(createWorkflowReq)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/workflows", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var createdWorkflow Workflow
	json.NewDecoder(w.Body).Decode(&createdWorkflow)

	// Create webhook trigger
	webhookToken := "concurrency-test-token"
	config := map[string]interface{}{
		"token": webhookToken,
	}
	configJson, _ := json.Marshal(config)

	// Use default user_id for testing
	defaultUserID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	_, err := db.Exec(`
		INSERT INTO triggers (id, user_id, provider, name, type, workflow_id, config, enabled) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		uuid.New(), defaultUserID, "webhook", "Concurrency Test Trigger", "webhook", createdWorkflow.Id, configJson, true)
	require.NoError(t, err)

	// Send multiple concurrent webhook requests
	numRequests := 10
	results := make(chan int, numRequests)

	for i := 0; i < numRequests; i++ {
		go func(index int) {
			payload := map[string]interface{}{
				"request_id": index,
				"message":    fmt.Sprintf("Concurrent request %d", index),
			}
			payloadJson, _ := json.Marshal(payload)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", fmt.Sprintf("/webhooks/%s", webhookToken), bytes.NewBuffer(payloadJson))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			results <- w.Code
		}(i)
	}

	// Collect results
	successCount := 0
	for i := 0; i < numRequests; i++ {
		statusCode := <-results
		if statusCode == http.StatusOK {
			successCount++
		}
	}

	assert.Equal(t, numRequests, successCount, "All concurrent requests should succeed")

	// Verify all webhook events were stored
	var eventCount int
	err = db.QueryRow("SELECT COUNT(*) FROM webhook_events").Scan(&eventCount)
	require.NoError(t, err)
	assert.Equal(t, numRequests, eventCount)
}
