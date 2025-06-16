package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestMelHTTPRequest(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request details
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Expected Authorization header, got %s", r.Header.Get("Authorization"))
		}

		// Read and verify body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("Failed to read request body: %v", err)
			return
		}
		expectedBody := `{"test": "data"}`
		if string(body) != expectedBody {
			t.Errorf("Expected body %s, got %s", expectedBody, string(body))
		}

		// Send response
		w.Header().Set("X-Test-Header", "test-value")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"result": "success"})
	}))
	defer server.Close()

	mel := NewMel()

	req := HTTPRequest{
		Method: "POST",
		URL:    server.URL,
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer test-token",
		},
		Body:    strings.NewReader(`{"test": "data"}`),
		Timeout: 5 * time.Second,
	}

	ctx := context.Background()
	resp, err := mel.HTTPRequest(ctx, req)

	if err != nil {
		t.Fatalf("HTTP request failed: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	if resp.Headers["X-Test-Header"] != "test-value" {
		t.Errorf("Expected X-Test-Header test-value, got %s", resp.Headers["X-Test-Header"])
	}

	var responseData map[string]interface{}
	if err := json.Unmarshal(resp.Body, &responseData); err != nil {
		t.Fatalf("Failed to parse response body: %v", err)
	}

	if responseData["result"] != "success" {
		t.Errorf("Expected result success, got %v", responseData["result"])
	}

	if resp.Duration <= 0 {
		t.Error("Expected positive duration")
	}
}

func TestMelHTTPRequestTimeout(t *testing.T) {
	// Create a slow server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // Longer than our timeout
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	mel := NewMel()

	req := HTTPRequest{
		Method:  "GET",
		URL:     server.URL,
		Timeout: 100 * time.Millisecond, // Very short timeout
	}

	ctx := context.Background()
	_, err := mel.HTTPRequest(ctx, req)

	if err == nil {
		t.Error("Expected timeout error")
	}

	if !strings.Contains(err.Error(), "timeout") && !strings.Contains(err.Error(), "deadline") {
		t.Errorf("Expected timeout error, got: %v", err)
	}
}

func TestMelDataStorage(t *testing.T) {
	mel := NewMel()
	ctx := context.Background()

	// Test storing and retrieving data
	testData := map[string]interface{}{
		"name":  "test",
		"value": 42,
	}

	err := mel.StoreData(ctx, "test-key", testData, 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to store data: %v", err)
	}

	// Retrieve the data
	retrieved, err := mel.RetrieveData(ctx, "test-key")
	if err != nil {
		t.Fatalf("Failed to retrieve data: %v", err)
	}

	retrievedMap, ok := retrieved.(map[string]interface{})
	if !ok {
		t.Fatalf("Retrieved data is not a map: %T", retrieved)
	}

	if retrievedMap["name"] != "test" {
		t.Errorf("Expected name 'test', got %v", retrievedMap["name"])
	}

	if retrievedMap["value"] != 42 {
		t.Errorf("Expected value 42, got %v", retrievedMap["value"])
	}
}

func TestMelDataStorageExpiration(t *testing.T) {
	mel := NewMel()
	ctx := context.Background()

	// Store data with very short TTL
	err := mel.StoreData(ctx, "expire-key", "test-data", 50*time.Millisecond)
	if err != nil {
		t.Fatalf("Failed to store data: %v", err)
	}

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Try to retrieve expired data
	_, err = mel.RetrieveData(ctx, "expire-key")
	if err == nil {
		t.Error("Expected error for expired data")
	}

	if !strings.Contains(err.Error(), "expired") {
		t.Errorf("Expected expiration error, got: %v", err)
	}
}

func TestMelDataStorageNotFound(t *testing.T) {
	mel := NewMel()
	ctx := context.Background()

	// Try to retrieve non-existent data
	_, err := mel.RetrieveData(ctx, "non-existent-key")
	if err == nil {
		t.Error("Expected error for non-existent data")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected not found error, got: %v", err)
	}
}

func TestMelDataStorageDelete(t *testing.T) {
	mel := NewMel()
	ctx := context.Background()

	// Store data
	err := mel.StoreData(ctx, "delete-key", "test-data", 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to store data: %v", err)
	}

	// Verify it exists
	_, err = mel.RetrieveData(ctx, "delete-key")
	if err != nil {
		t.Fatalf("Data should exist: %v", err)
	}

	// Delete it
	err = mel.DeleteData(ctx, "delete-key")
	if err != nil {
		t.Fatalf("Failed to delete data: %v", err)
	}

	// Verify it's gone
	_, err = mel.RetrieveData(ctx, "delete-key")
	if err == nil {
		t.Error("Expected error for deleted data")
	}
}

func TestMelWorkflowCall(t *testing.T) {
	// Create a mock server to simulate the workflow API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mock successful workflow execution response
		response := map[string]interface{}{
			"runId": "test-run-123",
			"trace": []map[string]interface{}{
				{
					"nodeId": "final-node",
					"output": []map[string]interface{}{
						{
							"data": map[string]interface{}{
								"result": "workflow-completed",
								"value":  42,
							},
						},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create Mel instance with mock server endpoint
	mel := NewMelWithConfig(30*time.Second, server.URL)
	ctx := context.Background()

	req := WorkflowCallRequest{
		TargetWorkflowID: "test-workflow-123",
		CallData: map[string]interface{}{
			"input": "test-data",
		},
		CallMode:       "async",
		TimeoutSeconds: 30,
		SourceContext: ExecutionContext{
			AgentID: "source-agent",
			RunID:   "source-run",
		},
	}

	// Test async call
	resp, err := mel.CallWorkflow(ctx, req)
	if err != nil {
		t.Fatalf("Async workflow call failed: %v", err)
	}

	if resp.Status != "sent" {
		t.Errorf("Expected status 'sent', got %s", resp.Status)
	}

	if resp.CallID == "" {
		t.Error("Expected non-empty call ID")
	}

	// Test sync call - this test doesn't simulate workflow_return, so we test with a short timeout
	// and expect it to timeout since no workflow_return node calls ReturnToWorkflow
	req.CallMode = "sync"
	req.TimeoutSeconds = 1 // Short timeout for testing
	resp, err = mel.CallWorkflow(ctx, req)

	// For this simple test without workflow_return simulation, we expect a timeout
	if err == nil {
		t.Error("Expected timeout error for sync call without workflow_return, but call succeeded")
	} else if !strings.Contains(err.Error(), "timeout") {
		t.Errorf("Expected timeout error, got: %v", err)
	}
}

func TestMelWorkflowReturn(t *testing.T) {
	mel := NewMel()
	ctx := context.Background()

	returnData := map[string]interface{}{
		"result": "completed",
		"value":  123,
	}

	// Test workflow return (current placeholder implementation)
	err := mel.ReturnToWorkflow(ctx, "test-call-id", returnData, "success")
	if err != nil {
		t.Fatalf("Workflow return failed: %v", err)
	}
}

func TestMelWithConfig(t *testing.T) {
	timeout := 10 * time.Second
	endpoint := "http://custom-endpoint:8080/api"

	mel := NewMelWithConfig(timeout, endpoint)

	// Test that the custom config is applied
	melImpl, ok := mel.(*melImpl)
	if !ok {
		t.Fatal("Expected melImpl type")
	}

	if melImpl.httpClient.Timeout != timeout {
		t.Errorf("Expected timeout %v, got %v", timeout, melImpl.httpClient.Timeout)
	}

	if melImpl.workflowEndpoint != endpoint {
		t.Errorf("Expected endpoint %s, got %s", endpoint, melImpl.workflowEndpoint)
	}
}

func TestMelNodeDefinitionRegistry(t *testing.T) {
	mel := NewMel()

	// Test node definition registration
	testDef := &testNodeDefinition{nodeType: "test-node"}
	mel.RegisterNodeDefinition(testDef)

	// Test listing definitions
	definitions := mel.ListNodeDefinitions()
	if len(definitions) != 1 {
		t.Errorf("Expected 1 definition, got %d", len(definitions))
	}

	// Test finding definition
	found := mel.FindDefinition("test-node")
	if found == nil {
		t.Error("Expected to find test-node definition")
	}

	if found != testDef {
		t.Error("Found definition doesn't match registered definition")
	}

	// Test finding non-existent definition
	notFound := mel.FindDefinition("non-existent")
	if notFound != nil {
		t.Error("Expected nil for non-existent definition")
	}

	// Test listing node types
	nodeTypes := mel.ListNodeTypes()
	if len(nodeTypes) != 1 {
		t.Errorf("Expected 1 node type, got %d", len(nodeTypes))
	}

	if nodeTypes[0].Type != "test-node" {
		t.Errorf("Expected type 'test-node', got %s", nodeTypes[0].Type)
	}
}

// Test helper node definition
type testNodeDefinition struct {
	nodeType string
}

func (d *testNodeDefinition) Meta() NodeType {
	return NodeType{
		Type:     d.nodeType,
		Label:    "Test Node",
		Category: "Test",
	}
}

func (d *testNodeDefinition) ExecuteEnvelope(ctx ExecutionContext, node Node, envelope *Envelope[interface{}]) (*Envelope[interface{}], error) {
	result := envelope.Clone()
	result.Data = map[string]interface{}{"test": "executed"}
	return result, nil
}

func (d *testNodeDefinition) Initialize(mel Mel) error {
	return nil
}
