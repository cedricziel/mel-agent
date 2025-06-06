package variable_get

import (
	"testing"

	"github.com/cedricziel/mel-agent/pkg/api"
	"github.com/cedricziel/mel-agent/pkg/core"
)

func TestGetVariableMeta(t *testing.T) {
	def := getVariableDefinition{}
	meta := def.Meta()

	if meta.Type != "variable_get" {
		t.Errorf("Expected type 'variable_get', got '%s'", meta.Type)
	}

	if meta.Label != "Get Variable" {
		t.Errorf("Expected label 'Get Variable', got '%s'", meta.Label)
	}

	if meta.Category != "Variables" {
		t.Errorf("Expected category 'Variables', got '%s'", meta.Category)
	}

	if len(meta.Parameters) == 0 {
		t.Error("Expected parameters to be defined")
	}
}

func TestGetVariableExecuteEnvelope_ValueOnly(t *testing.T) {
	// Reset variable store and set a test variable
	api.SetVariableStore(api.NewMemoryVariableStore())

	ctx := api.ExecutionContext{
		AgentID: "test-agent",
		RunID:   "test-run",
	}
	varCtx := api.CreateVariableContext("test-agent", "test-run", "test-node")
	api.SetVariable(varCtx, api.RunScope, "testVar", "test value")

	def := getVariableDefinition{}
	node := api.Node{
		ID:   "test-node",
		Type: "variable_get",
		Data: map[string]interface{}{
			"key":        "testVar",
			"scope":      "run",
			"outputMode": "value_only",
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
	inputEnvelope := core.NewEnvelope(interface{}("input data"), trace)

	outputEnvelope, err := def.ExecuteEnvelope(ctx, node, inputEnvelope)
	if err != nil {
		t.Fatalf("ExecuteEnvelope failed: %v", err)
	}

	if outputEnvelope == nil {
		t.Fatal("ExecuteEnvelope returned nil envelope")
	}

	result := outputEnvelope.Data

	if result != "test value" {
		t.Errorf("Expected 'test value', got %v", result)
	}
}

func TestGetVariableExecuteEnvelope_WithMetadata(t *testing.T) {
	// Reset variable store and set a test variable
	api.SetVariableStore(api.NewMemoryVariableStore())

	ctx := api.ExecutionContext{
		AgentID: "test-agent",
		RunID:   "test-run",
	}
	varCtx := api.CreateVariableContext("test-agent", "test-run", "test-node")
	api.SetVariable(varCtx, api.WorkflowScope, "workflowVar", "workflow value")

	def := getVariableDefinition{}
	node := api.Node{
		ID:   "test-node",
		Type: "variable_get",
		Data: map[string]interface{}{
			"key":        "workflowVar",
			"scope":      "workflow",
			"outputMode": "with_metadata",
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

	result := outputEnvelope.Data

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}

	if resultMap["value"] != "workflow value" {
		t.Errorf("Expected value 'workflow value', got %v", resultMap["value"])
	}

	if resultMap["exists"] != true {
		t.Error("Expected exists to be true")
	}

	if resultMap["key"] != "workflowVar" {
		t.Error("Expected key to be workflowVar")
	}

	if resultMap["scope"] != "workflow" {
		t.Error("Expected scope to be workflow")
	}
}

func TestGetVariableExecuteEnvelope_MergeInput(t *testing.T) {
	// Reset variable store and set a test variable
	api.SetVariableStore(api.NewMemoryVariableStore())

	ctx := api.ExecutionContext{
		AgentID: "test-agent",
		RunID:   "test-run",
	}
	varCtx := api.CreateVariableContext("test-agent", "test-run", "test-node")
	api.SetVariable(varCtx, api.RunScope, "myVar", 42)

	def := getVariableDefinition{}
	node := api.Node{
		ID:   "test-node",
		Type: "variable_get",
		Data: map[string]interface{}{
			"key":        "myVar",
			"scope":      "run",
			"outputMode": "merge_input",
		},
	}

	input := map[string]interface{}{
		"existing": "data",
		"count":    1,
	}

	// Create input envelope
	trace := api.Trace{
		AgentID: ctx.AgentID,
		RunID:   ctx.RunID,
		NodeID:  node.ID,
		Step:    node.ID,
		Attempt: 1,
	}
	inputEnvelope := core.NewEnvelope(interface{}(input), trace)

	outputEnvelope, err := def.ExecuteEnvelope(ctx, node, inputEnvelope)
	if err != nil {
		t.Fatalf("ExecuteEnvelope failed: %v", err)
	}

	if outputEnvelope == nil {
		t.Fatal("ExecuteEnvelope returned nil envelope")
	}

	result := outputEnvelope.Data

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}

	// Should have original input data
	if resultMap["existing"] != "data" {
		t.Error("Expected existing input to be preserved")
	}

	if resultMap["count"] != 1 {
		t.Error("Expected existing input to be preserved")
	}

	// Should have the variable value
	if resultMap["myVar"] != 42 {
		t.Errorf("Expected myVar to be 42, got %v", resultMap["myVar"])
	}
}

func TestGetVariableExecuteEnvelope_MergeInputNonMap(t *testing.T) {
	// Reset variable store and set a test variable
	api.SetVariableStore(api.NewMemoryVariableStore())

	ctx := api.ExecutionContext{
		AgentID: "test-agent",
		RunID:   "test-run",
	}
	varCtx := api.CreateVariableContext("test-agent", "test-run", "test-node")
	api.SetVariable(varCtx, api.RunScope, "myVar", "variable value")

	def := getVariableDefinition{}
	node := api.Node{
		ID:   "test-node",
		Type: "variable_get",
		Data: map[string]interface{}{
			"key":        "myVar",
			"scope":      "run",
			"outputMode": "merge_input",
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
	inputEnvelope := core.NewEnvelope(interface{}("string input"), trace)

	outputEnvelope, err := def.ExecuteEnvelope(ctx, node, inputEnvelope)
	if err != nil {
		t.Fatalf("ExecuteEnvelope failed: %v", err)
	}

	if outputEnvelope == nil {
		t.Fatal("ExecuteEnvelope returned nil envelope")
	}

	result := outputEnvelope.Data

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}

	// Should have input under "input" key
	if resultMap["input"] != "string input" {
		t.Error("Expected input to be preserved under 'input' key")
	}

	// Should have the variable value
	if resultMap["myVar"] != "variable value" {
		t.Errorf("Expected myVar to be 'variable value', got %v", resultMap["myVar"])
	}
}

func TestGetVariableExecuteEnvelope_MissingVariable(t *testing.T) {
	// Reset variable store for clean test
	api.SetVariableStore(api.NewMemoryVariableStore())

	def := getVariableDefinition{}
	ctx := api.ExecutionContext{
		AgentID: "test-agent",
		RunID:   "test-run",
	}
	node := api.Node{
		ID:   "test-node",
		Type: "variable_get",
		Data: map[string]interface{}{
			"key":   "missingVar",
			"scope": "run",
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
	inputEnvelope := core.NewEnvelope(interface{}("input data"), trace)

	outputEnvelope, err := def.ExecuteEnvelope(ctx, node, inputEnvelope)
	if err != nil {
		t.Fatalf("ExecuteEnvelope failed: %v", err)
	}

	if outputEnvelope == nil {
		t.Fatal("ExecuteEnvelope returned nil envelope")
	}

	result := outputEnvelope.Data

	// Should merge with input and set variable to nil
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}

	if resultMap["input"] != "input data" {
		t.Error("Expected input to be preserved")
	}

	if resultMap["missingVar"] != nil {
		t.Errorf("Expected missingVar to be nil, got %v", resultMap["missingVar"])
	}
}

func TestGetVariableExecuteEnvelope_MissingVariableWithDefault(t *testing.T) {
	// Reset variable store for clean test
	api.SetVariableStore(api.NewMemoryVariableStore())

	def := getVariableDefinition{}
	ctx := api.ExecutionContext{
		AgentID: "test-agent",
		RunID:   "test-run",
	}
	node := api.Node{
		ID:   "test-node",
		Type: "variable_get",
		Data: map[string]interface{}{
			"key":          "missingVar",
			"scope":        "run",
			"defaultValue": "default value",
			"outputMode":   "value_only",
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

	result := outputEnvelope.Data

	if result != "default value" {
		t.Errorf("Expected 'default value', got %v", result)
	}
}

func TestGetVariableExecuteEnvelope_MissingVariableFailIfMissing(t *testing.T) {
	// Reset variable store for clean test
	api.SetVariableStore(api.NewMemoryVariableStore())

	def := getVariableDefinition{}
	ctx := api.ExecutionContext{
		AgentID: "test-agent",
		RunID:   "test-run",
	}
	node := api.Node{
		ID:   "test-node",
		Type: "variable_get",
		Data: map[string]interface{}{
			"key":           "missingVar",
			"scope":         "run",
			"failIfMissing": true,
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
		t.Fatal("Expected error for missing variable")
	}

	nodeErr, ok := err.(*api.NodeError)
	if !ok {
		t.Fatal("Expected NodeError")
	}

	expectedMessage := "variable 'missingVar' not found in run scope"
	if nodeErr.Message != expectedMessage {
		t.Errorf("Expected error message '%s', got: %s", expectedMessage, nodeErr.Message)
	}
}

func TestGetVariableExecuteEnvelope_MissingKey(t *testing.T) {
	def := getVariableDefinition{}
	ctx := api.ExecutionContext{
		AgentID: "test-agent",
		RunID:   "test-run",
	}
	node := api.Node{
		ID:   "test-node",
		Type: "variable_get",
		Data: map[string]interface{}{
			"scope": "run",
			// Missing key
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

func TestGetVariableExecuteEnvelope_InvalidScope(t *testing.T) {
	def := getVariableDefinition{}
	ctx := api.ExecutionContext{
		AgentID: "test-agent",
		RunID:   "test-run",
	}
	node := api.Node{
		ID:   "test-node",
		Type: "variable_get",
		Data: map[string]interface{}{
			"key":   "testVar",
			"scope": "invalid",
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

func TestGetVariableExecuteEnvelope_InvalidOutputMode(t *testing.T) {
	// Reset variable store and set a test variable
	api.SetVariableStore(api.NewMemoryVariableStore())

	ctx := api.ExecutionContext{
		AgentID: "test-agent",
		RunID:   "test-run",
	}
	varCtx := api.CreateVariableContext("test-agent", "test-run", "test-node")
	api.SetVariable(varCtx, api.RunScope, "testVar", "test value")

	def := getVariableDefinition{}
	node := api.Node{
		ID:   "test-node",
		Type: "variable_get",
		Data: map[string]interface{}{
			"key":        "testVar",
			"scope":      "run",
			"outputMode": "invalid",
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
		t.Fatal("Expected error for invalid output mode")
	}

	nodeErr, ok := err.(*api.NodeError)
	if !ok {
		t.Fatal("Expected NodeError")
	}

	if nodeErr.Message != "invalid output mode: invalid" {
		t.Errorf("Expected error message about invalid output mode, got: %s", nodeErr.Message)
	}
}
