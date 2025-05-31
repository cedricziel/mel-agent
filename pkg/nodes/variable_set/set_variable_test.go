package variable_set

import (
	"testing"

	"github.com/cedricziel/mel-agent/pkg/api"
)

func TestSetVariableMeta(t *testing.T) {
	def := setVariableDefinition{}
	meta := def.Meta()

	if meta.Type != "variable_set" {
		t.Errorf("Expected type 'variable_set', got '%s'", meta.Type)
	}

	if meta.Label != "Set Variable" {
		t.Errorf("Expected label 'Set Variable', got '%s'", meta.Label)
	}

	if meta.Category != "Variables" {
		t.Errorf("Expected category 'Variables', got '%s'", meta.Category)
	}

	if len(meta.Parameters) == 0 {
		t.Error("Expected parameters to be defined")
	}
}

func TestSetVariable_RunScope(t *testing.T) {
	// Reset variable store for clean test
	api.SetVariableStore(api.NewMemoryVariableStore())

	def := setVariableDefinition{}
	ctx := api.ExecutionContext{
		AgentID: "test-agent",
		RunID:   "test-run",
	}
	node := api.Node{
		ID:   "test-node",
		Type: "variable_set",
		Data: map[string]interface{}{
			"key":         "testVar",
			"scope":       "run",
			"value":       "test value",
			"passthrough": true,
		},
	}

	input := map[string]interface{}{"original": "data"}
	result, err := def.Execute(ctx, node, input)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Should return input when passthrough is true
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}

	if resultMap["original"] != "data" {
		t.Errorf("Expected original data to be preserved")
	}

	// Verify variable was set
	varCtx := api.CreateVariableContext("test-agent", "test-run", "test-node")
	value, exists, err := api.GetVariable(varCtx, api.RunScope, "testVar")
	if err != nil {
		t.Fatalf("Failed to get variable: %v", err)
	}

	if !exists {
		t.Error("Variable was not set")
	}

	if value != "test value" {
		t.Errorf("Expected variable value 'test value', got %v", value)
	}
}

func TestSetVariable_WorkflowScope(t *testing.T) {
	// Reset variable store for clean test
	api.SetVariableStore(api.NewMemoryVariableStore())

	def := setVariableDefinition{}
	ctx := api.ExecutionContext{
		AgentID: "test-agent",
		RunID:   "test-run",
	}
	node := api.Node{
		ID:   "test-node",
		Type: "variable_set",
		Data: map[string]interface{}{
			"key":         "workflowVar",
			"scope":       "workflow",
			"value":       "workflow value",
			"passthrough": false,
		},
	}

	result, err := def.Execute(ctx, node, "input data")
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Should return variable info when passthrough is false
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}

	if resultMap["variable_set"] != true {
		t.Error("Expected variable_set to be true")
	}

	if resultMap["key"] != "workflowVar" {
		t.Error("Expected key to be workflowVar")
	}

	if resultMap["scope"] != "workflow" {
		t.Error("Expected scope to be workflow")
	}

	// Verify variable was set in workflow scope
	varCtx := api.CreateVariableContext("test-agent", "", "")
	value, exists, err := api.GetVariable(varCtx, api.WorkflowScope, "workflowVar")
	if err != nil {
		t.Fatalf("Failed to get variable: %v", err)
	}

	if !exists {
		t.Error("Variable was not set")
	}

	if value != "workflow value" {
		t.Errorf("Expected variable value 'workflow value', got %v", value)
	}
}

func TestSetVariable_GlobalScope(t *testing.T) {
	// Reset variable store for clean test
	api.SetVariableStore(api.NewMemoryVariableStore())

	def := setVariableDefinition{}
	ctx := api.ExecutionContext{
		AgentID: "test-agent",
		RunID:   "test-run",
	}
	node := api.Node{
		ID:   "test-node",
		Type: "variable_set",
		Data: map[string]interface{}{
			"key":   "globalVar",
			"scope": "global",
			"value": "global value",
		},
	}

	_, err := def.Execute(ctx, node, nil)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify variable was set in global scope
	varCtx := api.CreateVariableContext("test-agent", "", "test-node")
	value, exists, err := api.GetVariable(varCtx, api.GlobalScope, "globalVar")
	if err != nil {
		t.Fatalf("Failed to get variable: %v", err)
	}

	if !exists {
		t.Error("Variable was not set")
	}

	if value != "global value" {
		t.Errorf("Expected variable value 'global value', got %v", value)
	}
}

func TestSetVariable_ValueFromInput(t *testing.T) {
	// Reset variable store for clean test
	api.SetVariableStore(api.NewMemoryVariableStore())

	def := setVariableDefinition{}
	ctx := api.ExecutionContext{
		AgentID: "test-agent",
		RunID:   "test-run",
	}
	node := api.Node{
		ID:   "test-node",
		Type: "variable_set",
		Data: map[string]interface{}{
			"key":       "extractedVar",
			"scope":     "run",
			"valueFrom": "data.result",
		},
	}

	input := map[string]interface{}{
		"data": map[string]interface{}{
			"result": "extracted value",
		},
	}

	_, err := def.Execute(ctx, node, input)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify extracted value was set
	varCtx := api.CreateVariableContext("test-agent", "test-run", "test-node")
	value, exists, err := api.GetVariable(varCtx, api.RunScope, "extractedVar")
	if err != nil {
		t.Fatalf("Failed to get variable: %v", err)
	}

	if !exists {
		t.Error("Variable was not set")
	}

	if value != "extracted value" {
		t.Errorf("Expected variable value 'extracted value', got %v", value)
	}
}

func TestSetVariable_InputAsValue(t *testing.T) {
	// Reset variable store for clean test
	api.SetVariableStore(api.NewMemoryVariableStore())

	def := setVariableDefinition{}
	ctx := api.ExecutionContext{
		AgentID: "test-agent",
		RunID:   "test-run",
	}
	node := api.Node{
		ID:   "test-node",
		Type: "variable_set",
		Data: map[string]interface{}{
			"key":   "inputVar",
			"scope": "run",
			// No value or valueFrom - should use entire input
		},
	}

	input := map[string]interface{}{
		"message": "hello world",
		"count":   42,
	}

	_, err := def.Execute(ctx, node, input)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify entire input was set as value
	varCtx := api.CreateVariableContext("test-agent", "test-run", "test-node")
	value, exists, err := api.GetVariable(varCtx, api.RunScope, "inputVar")
	if err != nil {
		t.Fatalf("Failed to get variable: %v", err)
	}

	if !exists {
		t.Error("Variable was not set")
	}

	valueMap, ok := value.(map[string]interface{})
	if !ok {
		t.Fatal("Expected value to be a map")
	}

	if valueMap["message"] != "hello world" {
		t.Error("Expected input to be preserved")
	}

	if valueMap["count"] != 42 {
		t.Error("Expected input to be preserved")
	}
}

func TestSetVariable_MissingKey(t *testing.T) {
	def := setVariableDefinition{}
	ctx := api.ExecutionContext{
		AgentID: "test-agent",
		RunID:   "test-run",
	}
	node := api.Node{
		ID:   "test-node",
		Type: "variable_set",
		Data: map[string]interface{}{
			"scope": "run",
			"value": "test value",
			// Missing key
		},
	}

	_, err := def.Execute(ctx, node, nil)
	if err == nil {
		t.Fatal("Expected error for missing key")
	}

	nodeErr, ok := err.(*api.NodeError)
	if !ok {
		t.Fatal("Expected NodeError")
	}

	if nodeErr.Message != "variable name is required" {
		t.Errorf("Expected error message about missing key, got: %s", nodeErr.Message)
	}
}

func TestSetVariable_InvalidScope(t *testing.T) {
	def := setVariableDefinition{}
	ctx := api.ExecutionContext{
		AgentID: "test-agent",
		RunID:   "test-run",
	}
	node := api.Node{
		ID:   "test-node",
		Type: "variable_set",
		Data: map[string]interface{}{
			"key":   "testVar",
			"scope": "invalid",
			"value": "test value",
		},
	}

	_, err := def.Execute(ctx, node, nil)
	if err == nil {
		t.Fatal("Expected error for invalid scope")
	}

	nodeErr, ok := err.(*api.NodeError)
	if !ok {
		t.Fatal("Expected NodeError")
	}

	if nodeErr.Message != "invalid scope: invalid" {
		t.Errorf("Expected error message about invalid scope, got: %s", nodeErr.Message)
	}
}

func TestExtractValueFromInput(t *testing.T) {
	input := map[string]interface{}{
		"data": map[string]interface{}{
			"result": "success",
			"count":  42,
		},
		"status": "ok",
	}

	// Test simple path
	value := extractValueFromInput(input, "status")
	if value != "ok" {
		t.Errorf("Expected 'ok', got %v", value)
	}

	// Test nested path
	value = extractValueFromInput(input, "data.result")
	if value != "success" {
		t.Errorf("Expected 'success', got %v", value)
	}

	// Test missing path
	value = extractValueFromInput(input, "missing.path")
	if value != nil {
		t.Errorf("Expected nil for missing path, got %v", value)
	}

	// Test empty path
	value = extractValueFromInput(input, "")
	inputMap, ok := value.(map[string]interface{})
	if !ok {
		t.Error("Expected entire input for empty path")
	} else if inputMap["status"] != "ok" {
		t.Error("Expected entire input for empty path")
	}

	// Test dot path
	value = extractValueFromInput(input, ".")
	inputMap, ok = value.(map[string]interface{})
	if !ok {
		t.Error("Expected entire input for dot path")
	} else if inputMap["status"] != "ok" {
		t.Error("Expected entire input for dot path")
	}
}

func TestSplitPath(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"", nil},
		{"simple", []string{"simple"}},
		{"data.result", []string{"data", "result"}},
		{"a.b.c.d", []string{"a", "b", "c", "d"}},
		{"..empty..", []string{"empty"}},
	}

	for _, test := range tests {
		result := splitPath(test.input)
		if len(result) != len(test.expected) {
			t.Errorf("For input '%s', expected %d parts, got %d", test.input, len(test.expected), len(result))
			continue
		}

		for i, expected := range test.expected {
			if result[i] != expected {
				t.Errorf("For input '%s', expected part %d to be '%s', got '%s'", test.input, i, expected, result[i])
			}
		}
	}
}
