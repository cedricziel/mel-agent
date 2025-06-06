package core

import (
	"fmt"
	"testing"

	"github.com/cedricziel/mel-agent/pkg/api"
)

// ExampleHTTPNode demonstrates a new envelope-based node
type ExampleHTTPNode struct{}

func (n *ExampleHTTPNode) Meta() api.NodeType {
	return api.NodeType{
		Type:     "example_http",
		Label:    "Example HTTP Request",
		Icon:     "🌐",
		Category: "HTTP",
		Parameters: []api.ParameterDefinition{
			api.NewStringParameter("url", "URL", true).
				WithDescription("The URL to request"),
			api.NewEnumParameter("method", "Method", []string{"GET", "POST", "PUT", "DELETE"}, true).
				WithDefault("GET"),
		},
	}
}

func (n *ExampleHTTPNode) Initialize(mel api.Mel) error {
	return nil
}

// ExecuteEnvelope implements the new envelope-based interface
func (n *ExampleHTTPNode) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[interface{}]) (*api.Envelope[interface{}], error) {
	url, _ := node.Data["url"].(string)
	method, _ := node.Data["method"].(string)

	// Create HTTP payload
	httpPayload := HTTPPayload{
		Method:  method,
		URL:     url,
		Headers: map[string]string{"Content-Type": "application/json"},
		Body:    envelope.Data,
		Status:  200, // Simulated response
	}

	// Create response envelope
	result := &api.Envelope[interface{}]{
		ID:        GenerateEnvelopeID(),
		IssuedAt:  envelope.IssuedAt,
		Version:   envelope.Version,
		DataType:  "http_response",
		Data:      httpPayload,
		Trace:     envelope.Trace.Next(node.ID),
		Meta:      envelope.Meta,
		Variables: envelope.Variables,
	}

	// Add some metadata about the operation
	result.SetMeta("http_method", method)
	result.SetMeta("http_url", url)
	result.SetMeta("http_status", "200")

	return result, nil
}

// ExampleTypedNode demonstrates a strongly-typed envelope node
type ExampleTypedNode struct{}

func (n *ExampleTypedNode) Meta() api.NodeType {
	return api.NodeType{
		Type:     "example_typed",
		Label:    "Example Typed Node",
		Icon:     "🔤",
		Category: "Data",
	}
}

func (n *ExampleTypedNode) Initialize(mel api.Mel) error {
	return nil
}

// ExecuteTyped implements the typed interface
func (n *ExampleTypedNode) ExecuteTyped(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[HTTPPayload]) (*api.Envelope[string], error) {
	// Process the HTTP payload and return a string result
	result := fmt.Sprintf("Processed %s request to %s with status %d",
		envelope.Data.Method,
		envelope.Data.URL,
		envelope.Data.Status)

	return &api.Envelope[string]{
		ID:        GenerateEnvelopeID(),
		IssuedAt:  envelope.IssuedAt,
		Version:   envelope.Version,
		DataType:  "string",
		Data:      result,
		Trace:     envelope.Trace.Next(node.ID),
		Meta:      envelope.Meta,
		Variables: envelope.Variables,
	}, nil
}

func TestEnvelopeWorkflow(t *testing.T) {
	// Initialize variable store
	api.SetVariableStore(api.NewMemoryVariableStore())

	// Create execution context
	ctx := api.ExecutionContext{
		AgentID: "test-workflow",
		RunID:   "run-001",
		Variables: map[string]interface{}{
			"api_key": "secret-key-123",
		},
	}

	// Create initial data envelope
	initialData := map[string]interface{}{
		"user_id": 12345,
		"action":  "create_user",
	}

	envelope := NewGenericEnvelopeFromContext(initialData, ctx, "start")
	envelope.SetVariable(api.RunScope, "start_time", "2024-01-01T00:00:00Z")

	// Node 1: HTTP Request (envelope-based)
	httpNode := &ExampleHTTPNode{}
	httpNodeConfig := api.Node{
		ID:   "http-1",
		Type: "example_http",
		Data: map[string]interface{}{
			"url":    "https://api.example.com/users",
			"method": "POST",
		},
	}

	envelope, err := httpNode.ExecuteEnvelope(ctx, httpNodeConfig, envelope)
	if err != nil {
		t.Fatalf("HTTP node failed: %v", err)
	}

	// Verify HTTP response envelope
	httpPayload, ok := envelope.Data.(HTTPPayload)
	if !ok {
		t.Fatal("Expected HTTPPayload")
	}

	if httpPayload.Method != "POST" {
		t.Errorf("Expected method POST, got %s", httpPayload.Method)
	}

	if httpPayload.URL != "https://api.example.com/users" {
		t.Errorf("Expected URL https://api.example.com/users, got %s", httpPayload.URL)
	}

	// Check metadata
	if method, exists := envelope.GetMeta("http_method"); !exists || method != "POST" {
		t.Errorf("Expected http_method metadata 'POST', got '%s'", method)
	}

	// Node 2: Transform to typed node (manually convert for this example)
	typedNodeConfig := api.Node{
		ID:   "typed-1",
		Type: "example_typed",
		Data: map[string]interface{}{},
	}

	// Convert to typed envelope manually for demonstration
	typedEnvelope := &api.Envelope[HTTPPayload]{
		ID:        envelope.ID,
		IssuedAt:  envelope.IssuedAt,
		Version:   envelope.Version,
		DataType:  envelope.DataType,
		Data:      httpPayload,
		Trace:     envelope.Trace,
		Meta:      envelope.Meta,
		Variables: envelope.Variables,
	}

	typedNode := &ExampleTypedNode{}
	stringResult, err := typedNode.ExecuteTyped(ctx, typedNodeConfig, typedEnvelope)
	if err != nil {
		t.Fatalf("Typed node failed: %v", err)
	}

	// Convert back to generic envelope
	envelope = &api.Envelope[interface{}]{
		ID:        stringResult.ID,
		IssuedAt:  stringResult.IssuedAt,
		Version:   stringResult.Version,
		DataType:  stringResult.DataType,
		Data:      stringResult.Data,
		Trace:     stringResult.Trace,
		Meta:      stringResult.Meta,
		Variables: stringResult.Variables,
	}

	// Verify string result
	result, ok := envelope.Data.(string)
	if !ok {
		t.Fatal("Expected string result")
	}

	expected := "Processed POST request to https://api.example.com/users with status 200"
	if result != expected {
		t.Errorf("Expected result '%s', got '%s'", expected, result)
	}

	// Test variable access
	startTime, exists := envelope.GetVariable(api.RunScope, "start_time")
	if !exists {
		t.Error("Expected start_time variable to exist")
	}

	if startTime != "2024-01-01T00:00:00Z" {
		t.Errorf("Expected start_time '2024-01-01T00:00:00Z', got %v", startTime)
	}

	// Test serialization
	converter := NewJSONConverter()
	data, err := converter.Marshal(envelope)
	if err != nil {
		t.Fatalf("Failed to marshal envelope: %v", err)
	}

	restored, err := converter.Unmarshal(data)
	if err != nil {
		t.Fatalf("Failed to unmarshal envelope: %v", err)
	}

	if restored.ID != envelope.ID {
		t.Errorf("Expected restored ID '%s', got '%s'", envelope.ID, restored.ID)
	}

	if restored.DataType != envelope.DataType {
		t.Errorf("Expected restored DataType '%s', got '%s'", envelope.DataType, restored.DataType)
	}

	// Test direct envelope execution (without legacy adapter)
	directResult, err := httpNode.ExecuteEnvelope(ctx, httpNodeConfig, envelope)
	if err != nil {
		t.Fatalf("Direct envelope execution failed: %v", err)
	}

	if directPayload, ok := directResult.Data.(HTTPPayload); !ok {
		t.Error("Direct execution should return HTTPPayload")
	} else if directPayload.Method != "POST" {
		t.Errorf("Direct result method should be POST, got %s", directPayload.Method)
	}
}

// Demonstrate split/aggregate pattern
func TestSplitAggregatePattern(t *testing.T) {
	// Create array data
	arrayData := []interface{}{"item1", "item2", "item3"}
	trace := api.Trace{
		AgentID: "test-workflow",
		RunID:   "run-split",
		NodeID:  "start",
		Step:    "start",
		Attempt: 1,
	}

	envelope := NewGenericEnvelope(arrayData, trace)

	// Test splitter
	splitter := &SplitterNode{}
	splitterConfig := api.Node{
		ID:   "splitter-1",
		Type: "envelope_splitter",
		Data: map[string]interface{}{},
	}

	ctx := api.ExecutionContext{
		AgentID: "test-workflow",
		RunID:   "run-split",
	}

	splitResult, err := splitter.ExecuteEnvelope(ctx, splitterConfig, envelope)
	if err != nil {
		t.Fatalf("Splitter failed: %v", err)
	}

	// Verify split metadata
	if splitOp, exists := splitResult.GetMeta("split_operation"); !exists || splitOp != "true" {
		t.Error("Expected split_operation metadata")
	}

	if splitTotal, exists := splitResult.GetMeta("split_total"); !exists || splitTotal != "3" {
		t.Errorf("Expected split_total '3', got '%s'", splitTotal)
	}

	// Test aggregator
	aggregator := NewAggregatorNode()
	aggregatorConfig := api.Node{
		ID:   "aggregator-1",
		Type: "envelope_aggregator",
		Data: map[string]interface{}{
			"expectedCount": 3,
		},
	}

	// Simulate receiving split items
	splitEnvelopes := []*api.Envelope[interface{}]{
		NewGenericEnvelope("item1", trace.Next("splitter-1")),
		NewGenericEnvelope("item2", trace.Next("splitter-1")),
		NewGenericEnvelope("item3", trace.Next("splitter-1")),
	}

	// Add split metadata to each
	for i, env := range splitEnvelopes {
		env.SetMeta("split_total", "3")
		env.SetMeta("split_index", fmt.Sprintf("%d", i))
	}

	var aggregatedResult *api.Envelope[interface{}]

	// Send each split envelope to aggregator
	for i, splitEnv := range splitEnvelopes {
		result, err := aggregator.ExecuteEnvelope(ctx, aggregatorConfig, splitEnv)
		if err != nil {
			t.Fatalf("Aggregator failed on item %d: %v", i, err)
		}

		if i < 2 {
			// Should return nil for first two items (not ready)
			if result != nil {
				t.Errorf("Aggregator should return nil for item %d", i)
			}
		} else {
			// Should return aggregated result for last item
			if result == nil {
				t.Fatal("Aggregator should return result for last item")
			}
			aggregatedResult = result
		}
	}

	// Verify aggregated result
	if aggregatedResult.DataType != "array" {
		t.Errorf("Expected aggregated DataType 'array', got '%s'", aggregatedResult.DataType)
	}

	aggregatedData := aggregatedResult.Data.([]interface{})
	if len(aggregatedData) != 3 {
		t.Errorf("Expected 3 aggregated items, got %d", len(aggregatedData))
	}

	expectedItems := []interface{}{"item1", "item2", "item3"}
	for i, item := range aggregatedData {
		if item != expectedItems[i] {
			t.Errorf("Expected aggregated item '%v', got '%v'", expectedItems[i], item)
		}
	}

	// Check aggregation metadata
	if count, exists := aggregatedResult.GetMeta("aggregated_count"); !exists || count != "3" {
		t.Errorf("Expected aggregated_count '3', got '%s'", count)
	}

	if complete, exists := aggregatedResult.GetMeta("aggregation_complete"); !exists || complete != "true" {
		t.Errorf("Expected aggregation_complete 'true', got '%s'", complete)
	}
}
