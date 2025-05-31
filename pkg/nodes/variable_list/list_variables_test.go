package variable_list

import (
	"testing"

	"github.com/cedricziel/mel-agent/pkg/api"
)

func TestListVariablesMeta(t *testing.T) {
	def := listVariablesDefinition{}
	meta := def.Meta()

	if meta.Type != "variable_list" {
		t.Errorf("Expected type 'variable_list', got '%s'", meta.Type)
	}

	if meta.Label != "List Variables" {
		t.Errorf("Expected label 'List Variables', got '%s'", meta.Label)
	}

	if meta.Category != "Variables" {
		t.Errorf("Expected category 'Variables', got '%s'", meta.Category)
	}

	if len(meta.Parameters) == 0 {
		t.Error("Expected parameters to be defined")
	}
}

func TestListVariables_AllScopes(t *testing.T) {
	// Reset variable store and set test variables
	api.SetVariableStore(api.NewMemoryVariableStore())
	
	ctx := api.ExecutionContext{
		AgentID: "test-agent",
		RunID:   "test-run",
	}
	varCtx := api.CreateVariableContext("test-agent", "test-run", "test-node")
	
	// Set variables in different scopes
	api.SetVariable(varCtx, api.RunScope, "runVar", "run value")
	api.SetVariable(varCtx, api.WorkflowScope, "workflowVar", "workflow value")
	api.SetVariable(varCtx, api.GlobalScope, "globalVar", "global value")

	def := listVariablesDefinition{}
	node := api.Node{
		ID:   "test-node",
		Type: "variable_list",
		Data: map[string]interface{}{
			"scope": "all",
		},
	}

	input := map[string]interface{}{"original": "data"}
	result, err := def.Execute(ctx, node, input)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}

	// Should have input preserved
	inputResult, ok := resultMap["input"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected input to be preserved as map")
	}
	if inputResult["original"] != "data" {
		t.Error("Expected input to be preserved")
	}

	// Should have variables from all scopes
	runVars, ok := resultMap["run_variables"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected run_variables to be a map")
	}
	if runVars["runVar"] != "run value" {
		t.Error("Expected runVar in run scope")
	}

	workflowVars, ok := resultMap["workflow_variables"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected workflow_variables to be a map")
	}
	if workflowVars["workflowVar"] != "workflow value" {
		t.Error("Expected workflowVar in workflow scope")
	}

	globalVars, ok := resultMap["global_variables"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected global_variables to be a map")
	}
	if globalVars["globalVar"] != "global value" {
		t.Error("Expected globalVar in global scope")
	}
}

func TestListVariables_SpecificScope(t *testing.T) {
	// Reset variable store and set test variables
	api.SetVariableStore(api.NewMemoryVariableStore())
	
	ctx := api.ExecutionContext{
		AgentID: "test-agent",
		RunID:   "test-run",
	}
	varCtx := api.CreateVariableContext("test-agent", "test-run", "test-node")
	
	// Set variables in run scope
	api.SetVariable(varCtx, api.RunScope, "var1", "value1")
	api.SetVariable(varCtx, api.RunScope, "var2", 42)

	def := listVariablesDefinition{}
	node := api.Node{
		ID:   "test-node",
		Type: "variable_list",
		Data: map[string]interface{}{
			"scope": "run",
		},
	}

	result, err := def.Execute(ctx, node, "input data")
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}

	// Should have input preserved
	if resultMap["input"] != "input data" {
		t.Error("Expected input to be preserved")
	}

	// Should have variables from run scope only
	variables, ok := resultMap["variables"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected variables to be a map")
	}

	if variables["var1"] != "value1" {
		t.Error("Expected var1 in variables")
	}

	if variables["var2"] != 42 {
		t.Error("Expected var2 in variables")
	}

	// Should have scope info
	if resultMap["scope"] != "run" {
		t.Error("Expected scope to be 'run'")
	}
}

func TestListVariables_EmptyScope(t *testing.T) {
	// Reset variable store for clean test
	api.SetVariableStore(api.NewMemoryVariableStore())

	def := listVariablesDefinition{}
	ctx := api.ExecutionContext{
		AgentID: "test-agent",
		RunID:   "test-run",
	}
	node := api.Node{
		ID:   "test-node",
		Type: "variable_list",
		Data: map[string]interface{}{
			"scope":        "run",
			"includeEmpty": false,
		},
	}

	result, err := def.Execute(ctx, node, nil)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}

	// Should have variables as empty map
	variables, ok := resultMap["variables"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected variables to be a map")
	}

	if len(variables) != 0 {
		t.Error("Expected empty variables map")
	}
}

func TestListVariables_AllScopesWithEmpty(t *testing.T) {
	// Reset variable store and set only workflow variable
	api.SetVariableStore(api.NewMemoryVariableStore())
	
	ctx := api.ExecutionContext{
		AgentID: "test-agent",
		RunID:   "test-run",
	}
	varCtx := api.CreateVariableContext("test-agent", "test-run", "test-node")
	
	// Set only workflow variable
	api.SetVariable(varCtx, api.WorkflowScope, "workflowVar", "workflow value")

	def := listVariablesDefinition{}
	node := api.Node{
		ID:   "test-node",
		Type: "variable_list",
		Data: map[string]interface{}{
			"scope":        "all",
			"includeEmpty": true,
		},
	}

	result, err := def.Execute(ctx, node, nil)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}

	// Should have workflow variables
	workflowVars, ok := resultMap["workflow_variables"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected workflow_variables to be a map")
	}
	if workflowVars["workflowVar"] != "workflow value" {
		t.Error("Expected workflowVar in workflow scope")
	}

	// Should have empty run and global variables since includeEmpty is true
	runVars, ok := resultMap["run_variables"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected run_variables to be a map")
	}
	if len(runVars) != 0 {
		t.Error("Expected empty run_variables")
	}

	globalVars, ok := resultMap["global_variables"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected global_variables to be a map")
	}
	if len(globalVars) != 0 {
		t.Error("Expected empty global_variables")
	}
}

func TestListVariables_AllScopesWithoutEmpty(t *testing.T) {
	// Reset variable store and set only workflow variable
	api.SetVariableStore(api.NewMemoryVariableStore())
	
	ctx := api.ExecutionContext{
		AgentID: "test-agent",
		RunID:   "test-run",
	}
	varCtx := api.CreateVariableContext("test-agent", "test-run", "test-node")
	
	// Set only workflow variable
	api.SetVariable(varCtx, api.WorkflowScope, "workflowVar", "workflow value")

	def := listVariablesDefinition{}
	node := api.Node{
		ID:   "test-node",
		Type: "variable_list",
		Data: map[string]interface{}{
			"scope":        "all",
			"includeEmpty": false,
		},
	}

	result, err := def.Execute(ctx, node, nil)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}

	// Should have workflow variables
	workflowVars, ok := resultMap["workflow_variables"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected workflow_variables to be a map")
	}
	if workflowVars["workflowVar"] != "workflow value" {
		t.Error("Expected workflowVar in workflow scope")
	}

	// Should NOT have empty run and global variables since includeEmpty is false
	if _, exists := resultMap["run_variables"]; exists {
		t.Error("Expected run_variables to be omitted when empty")
	}

	if _, exists := resultMap["global_variables"]; exists {
		t.Error("Expected global_variables to be omitted when empty")
	}
}

func TestListVariables_DefaultScope(t *testing.T) {
	// Reset variable store and set test variables
	api.SetVariableStore(api.NewMemoryVariableStore())
	
	ctx := api.ExecutionContext{
		AgentID: "test-agent",
		RunID:   "test-run",
	}
	varCtx := api.CreateVariableContext("test-agent", "test-run", "test-node")
	
	// Set variables in different scopes
	api.SetVariable(varCtx, api.RunScope, "runVar", "run value")
	api.SetVariable(varCtx, api.WorkflowScope, "workflowVar", "workflow value")

	def := listVariablesDefinition{}
	node := api.Node{
		ID:   "test-node",
		Type: "variable_list",
		Data: map[string]interface{}{
			// No scope specified - should default to "all"
		},
	}

	result, err := def.Execute(ctx, node, nil)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}

	// Should have variables from all scopes (default behavior)
	if _, exists := resultMap["run_variables"]; !exists {
		t.Error("Expected run_variables to exist")
	}

	if _, exists := resultMap["workflow_variables"]; !exists {
		t.Error("Expected workflow_variables to exist")
	}
}

func TestListVariables_InvalidScope(t *testing.T) {
	def := listVariablesDefinition{}
	ctx := api.ExecutionContext{
		AgentID: "test-agent",
		RunID:   "test-run",
	}
	node := api.Node{
		ID:   "test-node",
		Type: "variable_list",
		Data: map[string]interface{}{
			"scope": "invalid",
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