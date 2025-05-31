package core

import (
	"fmt"
	"testing"
	"time"

	"github.com/cedricziel/mel-agent/pkg/api"
)

func TestNewEnvelope(t *testing.T) {
	trace := api.Trace{
		AgentID: "test-agent",
		RunID:   "test-run",
		NodeID:  "test-node",
		Step:    "test-step",
		Attempt: 1,
	}

	data := map[string]interface{}{
		"message": "hello world",
		"count":   42,
	}

	envelope := NewEnvelope(data, trace)

	if envelope == nil {
		t.Fatal("Envelope should not be nil")
	}

	if envelope.ID == "" {
		t.Error("Envelope ID should not be empty")
	}

	if envelope.IssuedAt.IsZero() {
		t.Error("Envelope IssuedAt should not be zero")
	}

	if envelope.Version != 1 {
		t.Errorf("Expected version 1, got %d", envelope.Version)
	}

	if envelope.DataType != "object" {
		t.Errorf("Expected DataType 'object', got '%s'", envelope.DataType)
	}

	if envelope.Trace.AgentID != "test-agent" {
		t.Errorf("Expected AgentID 'test-agent', got '%s'", envelope.Trace.AgentID)
	}

	dataMap := envelope.Data
	if dataMap["message"] != "hello world" {
		t.Error("Expected message 'hello world'")
	}

	if dataMap["count"] != 42 {
		t.Error("Expected count 42")
	}
}

func TestEnvelopeVariables(t *testing.T) {
	// Create a mock variable store
	api.SetVariableStore(api.NewMemoryVariableStore())

	trace := api.Trace{
		AgentID: "test-agent",
		RunID:   "test-run",
		NodeID:  "test-node",
		Step:    "test-step",
		Attempt: 1,
	}

	envelope := NewEnvelope("test data", trace)

	// Test setting run-level variable in envelope
	err := envelope.SetVariable(api.RunScope, "testVar", "test value")
	if err != nil {
		t.Fatalf("Failed to set variable: %v", err)
	}

	// Test getting variable from envelope
	value, exists := envelope.GetVariable(api.RunScope, "testVar")
	if !exists {
		t.Error("Variable should exist")
	}

	if value != "test value" {
		t.Errorf("Expected 'test value', got %v", value)
	}

	// Test setting global scope variable
	err = envelope.SetVariable(api.GlobalScope, "globalVar", "global value")
	if err != nil {
		t.Fatalf("Failed to set global variable: %v", err)
	}

	// Test getting global variable
	globalValue, exists := envelope.GetVariable(api.GlobalScope, "globalVar")
	if !exists {
		t.Error("Global variable should exist")
	}

	if globalValue != "global value" {
		t.Errorf("Expected 'global value', got %v", globalValue)
	}
}

func TestEnvelopeClone(t *testing.T) {
	trace := api.Trace{
		AgentID: "test-agent",
		RunID:   "test-run",
		NodeID:  "test-node",
		Step:    "test-step",
		Attempt: 1,
	}

	original := NewEnvelope("test data", trace)
	original.SetMeta("key1", "value1")
	original.Variables = map[string]interface{}{"var1": "val1"}
	original.Binary = map[string][]byte{"file1": []byte("content")}

	clone := original.Clone()

	// Check that IDs are different
	if clone.ID == original.ID {
		t.Error("Clone should have different ID")
	}

	// Check that data is copied
	if clone.Data != original.Data {
		t.Error("Clone should have same data")
	}

	// Check that metadata is copied but independent
	if clone.Meta["key1"] != "value1" {
		t.Error("Metadata should be copied")
	}

	clone.SetMeta("key2", "value2")
	if _, exists := original.GetMeta("key2"); exists {
		t.Error("Original should not have new metadata from clone")
	}

	// Check that variables are copied but independent
	if clone.Variables["var1"] != "val1" {
		t.Error("Variables should be copied")
	}

	clone.Variables["var2"] = "val2"
	if _, exists := original.Variables["var2"]; exists {
		t.Error("Original should not have new variables from clone")
	}

	// Check that binary data is copied
	if string(clone.Binary["file1"]) != "content" {
		t.Error("Binary data should be copied")
	}
}

func TestEnvelopeErrors(t *testing.T) {
	trace := api.Trace{
		AgentID: "test-agent",
		RunID:   "test-run",
		NodeID:  "test-node",
		Step:    "test-step",
		Attempt: 1,
	}

	envelope := NewEnvelope("test data", trace)

	if envelope.HasErrors() {
		t.Error("New envelope should not have errors")
	}

	if envelope.LastError() != nil {
		t.Error("New envelope should not have last error")
	}

	// Add an error
	nodeErr := api.NewNodeError("test-node", "test-type", "test error")
	envelope.AddError("test-node", "Test error occurred", nodeErr)

	if !envelope.HasErrors() {
		t.Error("Envelope should have errors after adding one")
	}

	lastError := envelope.LastError()
	if lastError == nil {
		t.Fatal("Should have last error")
	}

	if lastError.NodeID != "test-node" {
		t.Errorf("Expected NodeID 'test-node', got '%s'", lastError.NodeID)
	}

	if lastError.Message != "Test error occurred" {
		t.Errorf("Expected message 'Test error occurred', got '%s'", lastError.Message)
	}
}

func TestSplitEnvelope(t *testing.T) {
	trace := api.Trace{
		AgentID: "test-agent",
		RunID:   "test-run",
		NodeID:  "test-node",
		Step:    "test-step",
		Attempt: 1,
	}

	data := []interface{}{"item1", "item2", "item3"}
	envelope := &api.Envelope[[]interface{}]{
		ID:       GenerateEnvelopeID(),
		IssuedAt: time.Now(),
		Version:  1,
		DataType: "array",
		Data:     data,
		Trace:    trace,
	}

	split := SplitEnvelope(envelope)

	if len(split) != 3 {
		t.Errorf("Expected 3 split envelopes, got %d", len(split))
	}

	for i, item := range split {
		expectedValue := fmt.Sprintf("item%d", i+1)
		if item.Data != expectedValue {
			t.Errorf("Expected data '%s', got '%v'", expectedValue, item.Data)
		}

		if item.Trace.ParentID != "test-node" {
			t.Errorf("Expected ParentID 'test-node', got '%s'", item.Trace.ParentID)
		}

		if splitIndex, exists := item.GetMeta("split_index"); !exists || splitIndex != fmt.Sprintf("%d", i) {
			t.Errorf("Expected split_index '%d', got '%s'", i, splitIndex)
		}

		if splitTotal, exists := item.GetMeta("split_total"); !exists || splitTotal != "3" {
			t.Errorf("Expected split_total '3', got '%s'", splitTotal)
		}
	}
}

func TestMergeEnvelopes(t *testing.T) {
	trace := api.Trace{
		AgentID: "test-agent",
		RunID:   "test-run",
		NodeID:  "test-node",
		Step:    "test-step",
		Attempt: 1,
	}

	// Create multiple envelopes
	env1 := NewGenericEnvelope("item1", trace)
	env1.SetMeta("source", "first")

	env2 := NewGenericEnvelope("item2", trace)
	env2.SetMeta("source", "second")

	env3 := NewGenericEnvelope("item3", trace)
	env3.SetMeta("source", "third")

	envelopes := []*api.Envelope[interface{}]{env1, env2, env3}
	merged := MergeEnvelopes(envelopes)

	if merged == nil {
		t.Fatal("Merged envelope should not be nil")
	}

	if merged.DataType != "array" {
		t.Errorf("Expected DataType 'array', got '%s'", merged.DataType)
	}

	dataArray := merged.Data
	if len(dataArray) != 3 {
		t.Errorf("Expected 3 items in merged array, got %d", len(dataArray))
	}

	expectedItems := []string{"item1", "item2", "item3"}
	for i, item := range dataArray {
		if item != expectedItems[i] {
			t.Errorf("Expected item '%s', got '%v'", expectedItems[i], item)
		}
	}

	// Check that metadata from last envelope is preserved
	if source, exists := merged.GetMeta("source"); !exists || source != "third" {
		t.Errorf("Expected source metadata 'third', got '%s'", source)
	}
}

func TestJSONConverter(t *testing.T) {
	trace := api.Trace{
		AgentID: "test-agent",
		RunID:   "test-run",
		NodeID:  "test-node",
		Step:    "test-step",
		Attempt: 1,
	}

	data := map[string]interface{}{
		"message": "hello",
		"count":   42,
	}

	envelope := NewGenericEnvelope(data, trace)

	converter := NewJSONConverter()

	// Test marshaling
	bytes, err := converter.Marshal(envelope)
	if err != nil {
		t.Fatalf("Failed to marshal envelope: %v", err)
	}

	if len(bytes) == 0 {
		t.Error("Marshaled bytes should not be empty")
	}

	// Test unmarshaling
	restored, err := converter.Unmarshal(bytes)
	if err != nil {
		t.Fatalf("Failed to unmarshal envelope: %v", err)
	}

	if restored.ID != envelope.ID {
		t.Errorf("Expected ID '%s', got '%s'", envelope.ID, restored.ID)
	}

	if restored.DataType != envelope.DataType {
		t.Errorf("Expected DataType '%s', got '%s'", envelope.DataType, restored.DataType)
	}

	if restored.Trace.AgentID != envelope.Trace.AgentID {
		t.Errorf("Expected AgentID '%s', got '%s'", envelope.Trace.AgentID, restored.Trace.AgentID)
	}
}
