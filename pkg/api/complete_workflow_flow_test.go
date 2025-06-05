package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestCompleteWorkflowCallFlow tests the complete flow of workflow_call -> workflow_trigger -> workflow_return
func TestCompleteWorkflowCallFlow(t *testing.T) {
	// Create a mock server that simulates the target workflow execution environment
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Just acknowledge the trigger request - the actual workflow execution
		// happens in the background and ends with workflow_return calling mel.ReturnToWorkflow
		response := map[string]interface{}{
			"runId":   "complete-flow-test",
			"status":  "triggered",
			"message": "Workflow started successfully",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	mel := NewMelWithConfig(30*time.Second, server.URL)
	ctx := context.Background()

	t.Run("CompleteWorkflowFlow", func(t *testing.T) {
		// Simulate a workflow_call node calling another workflow
		req := WorkflowCallRequest{
			TargetWorkflowID: "target-workflow-complete",
			CallData: map[string]interface{}{
				"userInput":      "Hello from calling workflow",
				"priority":       "high",
				"requestId":      "req-12345",
				"batchSize":      50,
				"processingMode": "async",
			},
			CallMode:       "sync",
			TimeoutSeconds: 15,
			SourceContext: ExecutionContext{
				AgentID: "source-workflow-complete",
				RunID:   "run-complete-123",
				Mel:     mel,
			},
		}

		// Start the workflow call in background
		responseChan := make(chan *WorkflowCallResponse, 1)
		errorChan := make(chan error, 1)

		go func() {
			response, err := mel.CallWorkflow(ctx, req)
			if err != nil {
				errorChan <- err
				return
			}
			responseChan <- response
		}()

		// Wait for the workflow to be triggered and pending call registered
		time.Sleep(150 * time.Millisecond)

		// Simulate the target workflow executing and reaching a workflow_return node
		// The workflow_return node would extract the callId from the trigger payload
		// and call mel.ReturnToWorkflow with the results

		// Find the pending call ID
		mel.(*melImpl).pendingCallsMu.RLock()
		var callID string
		for id := range mel.(*melImpl).pendingCalls {
			callID = id
			break
		}
		mel.(*melImpl).pendingCallsMu.RUnlock()

		if callID == "" {
			t.Fatal("No pending call found - workflow call may not have been registered properly")
		}

		// Simulate complex workflow processing and return rich data
		workflowResults := map[string]interface{}{
			"success": true,
			"processedData": map[string]interface{}{
				"originalInput": req.CallData,
				"processedAt":   time.Now().Format(time.RFC3339),
				"processedBy":   "workflow-engine-v2",
				"outputData": map[string]interface{}{
					"transformedInput": strings.ToUpper(req.CallData["userInput"].(string)),
					"itemsProcessed":   req.CallData["batchSize"],
					"status":           "completed",
					"metadata": map[string]interface{}{
						"executionTime": "2.5s",
						"memoryUsed":    "45MB",
						"cpuTime":       "1.2s",
					},
				},
			},
			"warnings": []string{
				"Large batch size detected",
				"High priority processing used additional resources",
			},
			"nextSteps": []string{
				"Data has been processed and stored",
				"Notification sent to user",
				"Cleanup scheduled for tomorrow",
			},
		}

		// Simulate workflow_return node calling ReturnToWorkflow
		err := mel.ReturnToWorkflow(context.Background(), callID, workflowResults, "success")
		if err != nil {
			t.Fatalf("ReturnToWorkflow failed: %v", err)
		}

		// Wait for the original workflow call to complete
		var response *WorkflowCallResponse
		select {
		case response = <-responseChan:
			// Success
		case err := <-errorChan:
			t.Fatalf("Workflow call failed: %v", err)
		case <-time.After(10 * time.Second):
			t.Fatal("Timeout waiting for workflow call to complete")
		}

		// Verify the complete response
		if response.Status != "success" {
			t.Errorf("Expected status 'success', got %s", response.Status)
		}

		if response.CallID == "" {
			t.Error("Expected non-empty call ID")
		}

		// Verify rich workflow results
		if success, exists := response.Data["success"]; !exists || success != true {
			t.Error("Expected success=true in response data")
		}

		if processedData, exists := response.Data["processedData"]; exists {
			if dataMap, ok := processedData.(map[string]interface{}); ok {
				if originalInput, exists := dataMap["originalInput"]; exists {
					if inputMap, ok := originalInput.(map[string]interface{}); ok {
						if userInput := inputMap["userInput"]; userInput != "Hello from calling workflow" {
							t.Errorf("Expected original userInput to be preserved, got %v", userInput)
						}
						if batchSize := inputMap["batchSize"]; batchSize != 50 {
							t.Errorf("Expected original batchSize to be preserved, got %v", batchSize)
						}
					}
				}

				if outputData, exists := dataMap["outputData"]; exists {
					if outMap, ok := outputData.(map[string]interface{}); ok {
						if transformed := outMap["transformedInput"]; transformed != "HELLO FROM CALLING WORKFLOW" {
							t.Errorf("Expected transformed input 'HELLO FROM CALLING WORKFLOW', got %v", transformed)
						}
						if processed := outMap["itemsProcessed"]; processed != 50 {
							t.Errorf("Expected itemsProcessed 50, got %v", processed)
						}
					}
				}
			}
		} else {
			t.Error("Expected processedData in response")
		}

		if warnings, exists := response.Data["warnings"]; exists {
			if warningsList, ok := warnings.([]interface{}); ok {
				if len(warningsList) != 2 {
					t.Errorf("Expected 2 warnings, got %d", len(warningsList))
				}
			}
		}

		t.Logf("Complete workflow flow test completed successfully with call ID: %s", response.CallID)
	})
}