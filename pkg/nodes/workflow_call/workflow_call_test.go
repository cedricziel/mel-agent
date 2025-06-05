package workflow_call

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cedricziel/mel-agent/pkg/api"
)

func TestWorkflowCallDefinition_Meta(t *testing.T) {
	def := workflowCallDefinition{}
	meta := def.Meta()

	if meta.Type != "workflow_call" {
		t.Errorf("Expected type 'workflow_call', got '%s'", meta.Type)
	}

	if meta.Label != "Workflow Call" {
		t.Errorf("Expected label 'Workflow Call', got '%s'", meta.Label)
	}

	if meta.Category != "Workflow" {
		t.Errorf("Expected category 'Workflow', got '%s'", meta.Category)
	}

	if len(meta.Parameters) == 0 {
		t.Error("Expected parameters to be defined")
	}
}

func TestWorkflowCallDefinition_ExecuteEnvelope(t *testing.T) {
	// Create a mock server for workflow calls
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"runId":   "test-run-123",
			"status":  "triggered",
			"message": "Workflow started successfully",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	def := workflowCallDefinition{}
	
	// Test with missing targetWorkflowId
	ctx := api.ExecutionContext{
		AgentID: "test-agent",
		RunID:   "test-run",
		Mel:     api.NewMelWithConfig(30*time.Second, server.URL), // Use mock server
	}
	
	node := api.Node{
		ID:   "test-node",
		Type: "workflow_call",
		Data: map[string]interface{}{
			// Missing targetWorkflowId
		},
	}
	
	envelope := &api.Envelope[interface{}]{
		ID:       "test-envelope",
		IssuedAt: time.Now(),
		Version:  1,
		DataType: "object",
		Data:     map[string]interface{}{"test": "data"},
		Trace: api.Trace{
			AgentID: ctx.AgentID,
			RunID:   ctx.RunID,
			NodeID:  node.ID,
		},
	}
	
	_, err := def.ExecuteEnvelope(ctx, node, envelope)
	if err == nil {
		t.Error("Expected error for missing targetWorkflowId")
	}
	
	// Test with valid targetWorkflowId
	node.Data["targetWorkflowId"] = "target-workflow-123"
	node.Data["callMode"] = "async"
	
	result, err := def.ExecuteEnvelope(ctx, node, envelope)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	if result == nil {
		t.Error("Expected result to be non-nil")
	}
	
	if result.DataType != "object" {
		t.Errorf("Expected DataType 'object', got '%s'", result.DataType)
	}
	
	// Check result data structure
	if data, ok := result.Data.(map[string]interface{}); ok {
		if callInfo, exists := data["callInfo"]; !exists {
			t.Error("Expected callInfo in result data")
		} else if callInfoMap, ok := callInfo.(map[string]interface{}); ok {
			if targetId, exists := callInfoMap["targetWorkflowId"]; !exists {
				t.Error("Expected targetWorkflowId in callInfo")
			} else if targetId != "target-workflow-123" {
				t.Errorf("Expected targetWorkflowId 'target-workflow-123', got '%v'", targetId)
			}
		}
	} else {
		t.Error("Expected result data to be a map")
	}
}