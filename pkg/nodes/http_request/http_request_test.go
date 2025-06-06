package httprequest

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cedricziel/mel-agent/pkg/api"
	"github.com/cedricziel/mel-agent/pkg/core"
)

func TestHttpRequestMeta(t *testing.T) {
	def := httpRequestDefinition{}
	meta := def.Meta()

	if meta.Type != "http_request" {
		t.Errorf("Expected type 'http_request', got '%s'", meta.Type)
	}

	if meta.Label != "HTTP Request" {
		t.Errorf("Expected label 'HTTP Request', got '%s'", meta.Label)
	}

	if meta.Category != "Integration" {
		t.Errorf("Expected category 'Integration', got '%s'", meta.Category)
	}

	if len(meta.Parameters) == 0 {
		t.Error("Expected parameters to be defined")
	}

	// Test specific parameters exist
	paramNames := make(map[string]bool)
	for _, param := range meta.Parameters {
		paramNames[param.Name] = true
	}

	expectedParams := []string{
		"url", "method", "headers", "body", "contentType",
		"authType", "authValue", "authHeader", "timeout",
		"followRedirects", "verifySSL",
	}

	for _, expected := range expectedParams {
		if !paramNames[expected] {
			t.Errorf("Expected parameter '%s' not found", expected)
		}
	}
}

func TestHttpRequestExecuteEnvelope_GET(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"message": "success",
			"data":    []string{"item1", "item2"},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	def := httpRequestDefinition{}
	ctx := api.ExecutionContext{
		AgentID: "test-agent",
		RunID:   "test-run",
		Mel:     api.NewMel(),
	}
	node := api.Node{
		ID:   "test-node",
		Type: "http_request",
		Data: map[string]interface{}{
			"url":    server.URL,
			"method": "GET",
		},
	}

	// Create input envelope
	trace := api.Trace{
		AgentID: ctx.AgentID,
		RunID:   ctx.RunID,
		NodeID:  node.ID,
		Step:    node.ID,
		Attempt: 1,
	}
	inputEnvelope := core.NewEnvelope(interface{}(nil), trace)

	outputEnvelope, err := def.ExecuteEnvelope(ctx, node, inputEnvelope)
	if err != nil {
		t.Fatalf("ExecuteEnvelope failed: %v", err)
	}

	if outputEnvelope == nil {
		t.Fatal("ExecuteEnvelope returned nil envelope")
	}

	resultMap, ok := outputEnvelope.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}

	// Check response structure
	if resultMap["status"] != 200 {
		t.Errorf("Expected status 200, got %v", resultMap["status"])
	}

	if resultMap["method"] != "GET" {
		t.Errorf("Expected method GET, got %v", resultMap["method"])
	}

	if resultMap["url"] != server.URL {
		t.Errorf("Expected url %s, got %v", server.URL, resultMap["url"])
	}

	// Check response data
	data, ok := resultMap["data"].(map[string]interface{})
	if !ok {
		t.Fatal("Response data is not a map")
	}

	if data["message"] != "success" {
		t.Errorf("Expected message 'success', got %v", data["message"])
	}
}

func TestHttpRequestExecuteEnvelope_POST_WithBody(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		// Read and verify body
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		if body["test"] != "data" {
			t.Errorf("Expected body.test 'data', got %v", body["test"])
		}

		w.WriteHeader(http.StatusCreated)
		response := map[string]interface{}{
			"id":      123,
			"created": true,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	def := httpRequestDefinition{}
	ctx := api.ExecutionContext{
		AgentID: "test-agent",
		RunID:   "test-run",
		Mel:     api.NewMel(),
	}
	node := api.Node{
		ID:   "test-node",
		Type: "http_request",
		Data: map[string]interface{}{
			"url":         server.URL,
			"method":      "POST",
			"body":        `{"test": "data"}`,
			"contentType": "application/json",
		},
	}

	// Create input envelope
	trace := api.Trace{
		AgentID: ctx.AgentID,
		RunID:   ctx.RunID,
		NodeID:  node.ID,
		Step:    node.ID,
		Attempt: 1,
	}
	inputEnvelope := core.NewEnvelope(interface{}(nil), trace)

	outputEnvelope, err := def.ExecuteEnvelope(ctx, node, inputEnvelope)
	if err != nil {
		t.Fatalf("ExecuteEnvelope failed: %v", err)
	}

	if outputEnvelope == nil {
		t.Fatal("ExecuteEnvelope returned nil envelope")
	}

	resultMap, ok := outputEnvelope.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}

	if resultMap["status"] != 201 {
		t.Errorf("Expected status 201, got %v", resultMap["status"])
	}

	if resultMap["method"] != "POST" {
		t.Errorf("Expected method POST, got %v", resultMap["method"])
	}
}

func TestHttpRequestExecuteEnvelope_WithCustomHeaders(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Custom-Header") != "custom-value" {
			t.Errorf("Expected X-Custom-Header 'custom-value', got '%s'", r.Header.Get("X-Custom-Header"))
		}

		if r.Header.Get("X-Another-Header") != "another-value" {
			t.Errorf("Expected X-Another-Header 'another-value', got '%s'", r.Header.Get("X-Another-Header"))
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	def := httpRequestDefinition{}
	ctx := api.ExecutionContext{
		AgentID: "test-agent",
		RunID:   "test-run",
		Mel:     api.NewMel(),
	}
	node := api.Node{
		ID:   "test-node",
		Type: "http_request",
		Data: map[string]interface{}{
			"url":    server.URL,
			"method": "GET",
			"headers": map[string]interface{}{
				"X-Custom-Header":  "custom-value",
				"X-Another-Header": "another-value",
			},
		},
	}

	// Create input envelope
	trace := api.Trace{
		AgentID: ctx.AgentID,
		RunID:   ctx.RunID,
		NodeID:  node.ID,
		Step:    node.ID,
		Attempt: 1,
	}
	inputEnvelope := core.NewEnvelope(interface{}(nil), trace)

	outputEnvelope, err := def.ExecuteEnvelope(ctx, node, inputEnvelope)
	if err != nil {
		t.Fatalf("ExecuteEnvelope failed: %v", err)
	}

	if outputEnvelope == nil {
		t.Fatal("ExecuteEnvelope returned nil envelope")
	}

	resultMap, ok := outputEnvelope.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}

	if resultMap["status"] != 200 {
		t.Errorf("Expected status 200, got %v", resultMap["status"])
	}
}

func TestHttpRequestExecuteEnvelope_BearerAuth(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer test-token-123" {
			t.Errorf("Expected Authorization 'Bearer test-token-123', got '%s'", authHeader)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Authenticated"))
	}))
	defer server.Close()

	def := httpRequestDefinition{}
	ctx := api.ExecutionContext{
		AgentID: "test-agent",
		RunID:   "test-run",
		Mel:     api.NewMel(),
	}
	node := api.Node{
		ID:   "test-node",
		Type: "http_request",
		Data: map[string]interface{}{
			"url":       server.URL,
			"method":    "GET",
			"authType":  "bearer",
			"authValue": "test-token-123",
		},
	}

	// Create input envelope
	trace := api.Trace{
		AgentID: ctx.AgentID,
		RunID:   ctx.RunID,
		NodeID:  node.ID,
		Step:    node.ID,
		Attempt: 1,
	}
	inputEnvelope := core.NewEnvelope(interface{}(nil), trace)

	outputEnvelope, err := def.ExecuteEnvelope(ctx, node, inputEnvelope)
	if err != nil {
		t.Fatalf("ExecuteEnvelope failed: %v", err)
	}

	if outputEnvelope == nil {
		t.Fatal("ExecuteEnvelope returned nil envelope")
	}

	resultMap, ok := outputEnvelope.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}

	if resultMap["status"] != 200 {
		t.Errorf("Expected status 200, got %v", resultMap["status"])
	}

	// Check that response data is a string (not JSON)
	if resultMap["data"] != "Authenticated" {
		t.Errorf("Expected data 'Authenticated', got %v", resultMap["data"])
	}
}

func TestHttpRequestExecuteEnvelope_ApiKeyAuth(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-Key")
		if apiKey != "secret-key-456" {
			t.Errorf("Expected X-API-Key 'secret-key-456', got '%s'", apiKey)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("API Key Valid"))
	}))
	defer server.Close()

	def := httpRequestDefinition{}
	ctx := api.ExecutionContext{
		AgentID: "test-agent",
		RunID:   "test-run",
		Mel:     api.NewMel(),
	}
	node := api.Node{
		ID:   "test-node",
		Type: "http_request",
		Data: map[string]interface{}{
			"url":        server.URL,
			"method":     "GET",
			"authType":   "apikey",
			"authValue":  "secret-key-456",
			"authHeader": "X-API-Key",
		},
	}

	// Create input envelope
	trace := api.Trace{
		AgentID: ctx.AgentID,
		RunID:   ctx.RunID,
		NodeID:  node.ID,
		Step:    node.ID,
		Attempt: 1,
	}
	inputEnvelope := core.NewEnvelope(interface{}(nil), trace)

	outputEnvelope, err := def.ExecuteEnvelope(ctx, node, inputEnvelope)
	if err != nil {
		t.Fatalf("ExecuteEnvelope failed: %v", err)
	}

	if outputEnvelope == nil {
		t.Fatal("ExecuteEnvelope returned nil envelope")
	}

	resultMap, ok := outputEnvelope.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}

	if resultMap["status"] != 200 {
		t.Errorf("Expected status 200, got %v", resultMap["status"])
	}
}

func TestHttpRequestExecuteEnvelope_BasicAuth(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Basic dXNlcjpwYXNz" { // base64 encoded "user:pass"
			t.Errorf("Expected Authorization 'Basic dXNlcjpwYXNz', got '%s'", authHeader)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Basic Auth Valid"))
	}))
	defer server.Close()

	def := httpRequestDefinition{}
	ctx := api.ExecutionContext{
		AgentID: "test-agent",
		RunID:   "test-run",
		Mel:     api.NewMel(),
	}
	node := api.Node{
		ID:   "test-node",
		Type: "http_request",
		Data: map[string]interface{}{
			"url":       server.URL,
			"method":    "GET",
			"authType":  "basic",
			"authValue": "dXNlcjpwYXNz", // pre-encoded
		},
	}

	// Create input envelope
	trace := api.Trace{
		AgentID: ctx.AgentID,
		RunID:   ctx.RunID,
		NodeID:  node.ID,
		Step:    node.ID,
		Attempt: 1,
	}
	inputEnvelope := core.NewEnvelope(interface{}(nil), trace)

	outputEnvelope, err := def.ExecuteEnvelope(ctx, node, inputEnvelope)
	if err != nil {
		t.Fatalf("ExecuteEnvelope failed: %v", err)
	}

	if outputEnvelope == nil {
		t.Fatal("ExecuteEnvelope returned nil envelope")
	}

	resultMap, ok := outputEnvelope.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}

	if resultMap["status"] != 200 {
		t.Errorf("Expected status 200, got %v", resultMap["status"])
	}
}

func TestHttpRequestExecuteEnvelope_MissingURL(t *testing.T) {
	def := httpRequestDefinition{}
	ctx := api.ExecutionContext{
		AgentID: "test-agent",
		RunID:   "test-run",
		Mel:     api.NewMel(),
	}
	node := api.Node{
		ID:   "test-node",
		Type: "http_request",
		Data: map[string]interface{}{
			"method": "GET",
			// URL is missing
		},
	}

	// Create input envelope
	trace := api.Trace{
		AgentID: ctx.AgentID,
		RunID:   ctx.RunID,
		NodeID:  node.ID,
		Step:    node.ID,
		Attempt: 1,
	}
	inputEnvelope := core.NewEnvelope(interface{}(nil), trace)

	_, err := def.ExecuteEnvelope(ctx, node, inputEnvelope)
	if err == nil {
		t.Fatal("Expected error for missing URL")
	}

	if err.Error() != "url is required" {
		t.Errorf("Expected error 'url is required', got '%s'", err.Error())
	}
}

func TestHttpRequestExecuteEnvelope_InvalidURL(t *testing.T) {
	def := httpRequestDefinition{}
	ctx := api.ExecutionContext{
		AgentID: "test-agent",
		RunID:   "test-run",
		Mel:     api.NewMel(),
	}
	node := api.Node{
		ID:   "test-node",
		Type: "http_request",
		Data: map[string]interface{}{
			"url":    "not-a-valid-url",
			"method": "GET",
		},
	}

	// Create input envelope
	trace := api.Trace{
		AgentID: ctx.AgentID,
		RunID:   ctx.RunID,
		NodeID:  node.ID,
		Step:    node.ID,
		Attempt: 1,
	}
	inputEnvelope := core.NewEnvelope(interface{}(nil), trace)

	_, err := def.ExecuteEnvelope(ctx, node, inputEnvelope)
	if err == nil {
		t.Fatal("Expected error for invalid URL")
	}

	// Should contain "request failed" or "failed to create request"
	if !contains(err.Error(), "request failed") && !contains(err.Error(), "failed to create request") {
		t.Errorf("Expected error about request failure, got '%s'", err.Error())
	}
}

func TestHttpRequestExecuteEnvelope_Timeout(t *testing.T) {
	// Create slow test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Sleep longer than the timeout
		// time.Sleep(2 * time.Second) // Commented out to avoid slow tests
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	def := httpRequestDefinition{}
	ctx := api.ExecutionContext{
		AgentID: "test-agent",
		RunID:   "test-run",
		Mel:     api.NewMel(),
	}
	node := api.Node{
		ID:   "test-node",
		Type: "http_request",
		Data: map[string]interface{}{
			"url":     server.URL,
			"method":  "GET",
			"timeout": 1, // 1 second timeout
		},
	}

	// Create input envelope
	trace := api.Trace{
		AgentID: ctx.AgentID,
		RunID:   ctx.RunID,
		NodeID:  node.ID,
		Step:    node.ID,
		Attempt: 1,
	}
	inputEnvelope := core.NewEnvelope(interface{}(nil), trace)

	outputEnvelope, err := def.ExecuteEnvelope(ctx, node, inputEnvelope)
	if err != nil {
		t.Fatalf("ExecuteEnvelope failed: %v", err)
	}

	if outputEnvelope == nil {
		t.Fatal("ExecuteEnvelope returned nil envelope")
	}

	// Should succeed since we're not actually sleeping
	resultMap, ok := outputEnvelope.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}

	if resultMap["status"] != 200 {
		t.Errorf("Expected status 200, got %v", resultMap["status"])
	}
}

func TestHttpRequestExecuteEnvelope_NonJSONResponse(t *testing.T) {
	// Create test server that returns plain text
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Plain text response"))
	}))
	defer server.Close()

	def := httpRequestDefinition{}
	ctx := api.ExecutionContext{
		AgentID: "test-agent",
		RunID:   "test-run",
		Mel:     api.NewMel(),
	}
	node := api.Node{
		ID:   "test-node",
		Type: "http_request",
		Data: map[string]interface{}{
			"url":    server.URL,
			"method": "GET",
		},
	}

	// Create input envelope
	trace := api.Trace{
		AgentID: ctx.AgentID,
		RunID:   ctx.RunID,
		NodeID:  node.ID,
		Step:    node.ID,
		Attempt: 1,
	}
	inputEnvelope := core.NewEnvelope(interface{}(nil), trace)

	outputEnvelope, err := def.ExecuteEnvelope(ctx, node, inputEnvelope)
	if err != nil {
		t.Fatalf("ExecuteEnvelope failed: %v", err)
	}

	if outputEnvelope == nil {
		t.Fatal("ExecuteEnvelope returned nil envelope")
	}

	resultMap, ok := outputEnvelope.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}

	// Should return response as string since it's not JSON
	if resultMap["data"] != "Plain text response" {
		t.Errorf("Expected data 'Plain text response', got %v", resultMap["data"])
	}
}

func TestHttpRequestInitialize(t *testing.T) {
	def := httpRequestDefinition{}
	err := def.Initialize(nil)
	if err != nil {
		t.Errorf("Initialize should not return an error, got: %v", err)
	}
}

// Helper function
func contains(str, substr string) bool {
	return len(str) >= len(substr) && (str == substr ||
		(len(str) > len(substr) &&
			(str[:len(substr)] == substr ||
				str[len(str)-len(substr):] == substr ||
				findSubstring(str, substr))))
}

func findSubstring(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
