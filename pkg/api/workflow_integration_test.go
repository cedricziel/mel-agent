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

// TestWorkflowCallIntegration tests the full workflow calling flow
func TestWorkflowCallIntegration(t *testing.T) {
	// Create a mock server that simulates a target workflow
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse the incoming call data
		var payload map[string]interface{}
		json.NewDecoder(r.Body).Decode(&payload)

		// Simulate a workflow that processes the call data and returns a response
		response := map[string]interface{}{
			"runId": "integration-test-run",
			"trace": []map[string]interface{}{
				{
					"nodeId": "workflow_trigger",
					"input": []map[string]interface{}{
						{"data": payload},
					},
					"output": []map[string]interface{}{
						{
							"data": map[string]interface{}{
								"triggerType":     "workflow_call",
								"callId":          payload["callId"],
								"callingWorkflow": payload["sourceWorkflowId"],
								"callData":        payload["callData"],
							},
						},
					},
				},
				{
					"nodeId": "processing_node",
					"input": []map[string]interface{}{
						{
							"data": map[string]interface{}{
								"callData": payload["callData"],
							},
						},
					},
					"output": []map[string]interface{}{
						{
							"data": map[string]interface{}{
								"processed": true,
								"inputData": payload["callData"],
								"result":    "processing completed",
							},
						},
					},
				},
				{
					"nodeId": "workflow_return",
					"input": []map[string]interface{}{
						{
							"data": map[string]interface{}{
								"processed": true,
								"result":    "processing completed",
							},
						},
					},
					"output": []map[string]interface{}{
						{
							"data": map[string]interface{}{
								"finalResult":    "workflow execution completed",
								"processedInput": payload["callData"],
								"timestamp":      time.Now().Format(time.RFC3339),
								"returnResponse": map[string]interface{}{
									"callId": payload["callId"],
									"status": "success",
								},
								"success": true,
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

	// Create Mel instance with mock server
	mel := NewMelWithConfig(30*time.Second, server.URL)

	// Test the full workflow calling flow
	t.Run("SynchronousWorkflowCall", func(t *testing.T) {
		ctx := context.Background()

		// Prepare workflow call request
		req := WorkflowCallRequest{
			TargetWorkflowID: "target-workflow-123",
			CallData: map[string]interface{}{
				"operation":  "process_data",
				"inputValue": 42,
				"metadata": map[string]interface{}{
					"priority": "high",
					"source":   "integration-test",
				},
			},
			CallMode:       "sync",
			TimeoutSeconds: 10,
			SourceContext: ExecutionContext{
				AgentID: "source-workflow-456",
				RunID:   "source-run-789",
				Mel:     mel,
			},
		}

		// Execute the workflow call in a goroutine since it will block waiting for response
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

		// Give the workflow call time to register the pending call and trigger the workflow
		time.Sleep(100 * time.Millisecond)

		// Simulate workflow_return node calling ReturnToWorkflow
		// We need to extract the callID from the payload that was sent to the mock server
		// For this test, we'll simulate the return by directly calling ReturnToWorkflow

		// First, we need to find the pending call to get the callID
		mel.(*melImpl).pendingCallsMu.RLock()
		var callID string
		for id := range mel.(*melImpl).pendingCalls {
			callID = id
			break
		}
		mel.(*melImpl).pendingCallsMu.RUnlock()

		if callID == "" {
			t.Fatal("No pending call found")
		}

		// Simulate workflow_return node calling ReturnToWorkflow
		returnData := map[string]interface{}{
			"finalResult":    "workflow execution completed",
			"processedInput": req.CallData,
			"timestamp":      time.Now().Format(time.RFC3339),
		}

		err := mel.ReturnToWorkflow(context.Background(), callID, returnData, "success")
		if err != nil {
			t.Fatalf("ReturnToWorkflow failed: %v", err)
		}

		// Wait for the workflow call to complete
		var response *WorkflowCallResponse
		select {
		case response = <-responseChan:
			// Success
		case err := <-errorChan:
			t.Fatalf("Workflow call failed: %v", err)
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for workflow call to complete")
		}

		// Verify the response
		if response.Status != "success" {
			t.Errorf("Expected status 'success', got %s", response.Status)
		}

		if response.CallID == "" {
			t.Error("Expected non-empty call ID")
		}

		// Verify the response data contains the expected final result
		// Since this comes from ReturnToWorkflow, the data is directly in response.Data
		if finalResult, exists := response.Data["finalResult"]; exists {
			if finalResult != "workflow execution completed" {
				t.Errorf("Expected final result 'workflow execution completed', got %v", finalResult)
			}
		} else {
			t.Error("Expected finalResult in response data")
		}

		// Verify that the input data was processed correctly
		if processedInput, exists := response.Data["processedInput"]; exists {
			if inputMap, ok := processedInput.(map[string]interface{}); ok {
				if operation, exists := inputMap["operation"]; exists && operation != "process_data" {
					t.Errorf("Expected operation 'process_data', got %v", operation)
				}
				if inputValue, exists := inputMap["inputValue"]; exists && inputValue != 42 {
					t.Errorf("Expected inputValue 42, got %v", inputValue)
				}
			} else {
				t.Error("Expected processedInput to be a map")
			}
		} else {
			t.Error("Expected processedInput in response data")
		}
	})

	t.Run("SynchronousWorkflowCallTimeout", func(t *testing.T) {
		ctx := context.Background()

		// Prepare workflow call request with very short timeout
		req := WorkflowCallRequest{
			TargetWorkflowID: "timeout-workflow-123",
			CallData: map[string]interface{}{
				"operation": "long_running_task",
			},
			CallMode:       "sync",
			TimeoutSeconds: 1, // Very short timeout
			SourceContext: ExecutionContext{
				AgentID: "timeout-source-workflow",
				RunID:   "timeout-run-123",
				Mel:     mel,
			},
		}

		// Execute the workflow call - this should timeout
		_, err := mel.CallWorkflow(ctx, req)
		if err == nil {
			t.Error("Expected timeout error, but call succeeded")
		}

		if !strings.Contains(err.Error(), "timeout") {
			t.Errorf("Expected timeout error, got: %v", err)
		}
	})

	t.Run("AsynchronousWorkflowCall", func(t *testing.T) {
		ctx := context.Background()

		// Prepare async workflow call
		req := WorkflowCallRequest{
			TargetWorkflowID: "async-target-workflow",
			CallData: map[string]interface{}{
				"asyncOperation": "background_process",
				"batchSize":      100,
			},
			CallMode:       "async",
			TimeoutSeconds: 5,
			SourceContext: ExecutionContext{
				AgentID: "async-source-workflow",
				RunID:   "async-run-123",
				Mel:     mel,
			},
		}

		// Execute the async workflow call
		response, err := mel.CallWorkflow(ctx, req)
		if err != nil {
			t.Fatalf("Async workflow call failed: %v", err)
		}

		// Verify async response
		if response.Status != "sent" {
			t.Errorf("Expected status 'sent' for async call, got %s", response.Status)
		}

		if response.CallID == "" {
			t.Error("Expected non-empty call ID for async call")
		}

		// For async calls, we should get a confirmation, not the final result
		if message, exists := response.Data["message"]; exists {
			if message != "Async workflow triggered" {
				t.Errorf("Expected async trigger message, got %v", message)
			}
		}
	})
}

// TestWorkflowReturnIntegration tests the workflow return mechanism
func TestWorkflowReturnIntegration(t *testing.T) {
	mel := NewMel()
	ctx := context.Background()

	// Test storing and retrieving workflow return data
	callID := "test-return-call-123"
	returnData := map[string]interface{}{
		"result":    "operation completed",
		"count":     42,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	// Test ReturnToWorkflow when no pending call exists (should store data)
	err := mel.ReturnToWorkflow(ctx, callID, returnData, "success")
	if err != nil {
		t.Fatalf("ReturnToWorkflow failed: %v", err)
	}

	// Verify that the return data was stored
	storedData, err := mel.RetrieveData(ctx, "workflow_return:"+callID)
	if err != nil {
		t.Fatalf("Failed to retrieve stored return data: %v", err)
	}

	if storedMap, ok := storedData.(map[string]interface{}); ok {
		if returnedData, exists := storedMap["data"]; exists {
			if dataMap, ok := returnedData.(map[string]interface{}); ok {
				if result, exists := dataMap["result"]; exists && result != "operation completed" {
					t.Errorf("Expected result 'operation completed', got %v", result)
				}
			}
		}

		if status, exists := storedMap["status"]; exists && status != "success" {
			t.Errorf("Expected status 'success', got %v", status)
		}
	} else {
		t.Error("Expected stored data to be a map")
	}
}

// TestWorkflowDataFlow tests the complete data flow between workflow nodes
func TestWorkflowDataFlow(t *testing.T) {
	ctx := context.Background()

	// Simulate workflow_call node execution
	t.Run("WorkflowCallNode", func(t *testing.T) {
		// Create a mock server for the target workflow
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := map[string]interface{}{
				"runId": "dataflow-test",
				"trace": []map[string]interface{}{
					{
						"nodeId": "final",
						"output": []map[string]interface{}{
							{
								"data": map[string]interface{}{
									"flowResult": "data flow successful",
									"nodeCount":  3,
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

		melWithMock := NewMelWithConfig(10*time.Second, server.URL)

		// Test that workflow call properly passes data through the platform
		execCtx := ExecutionContext{
			AgentID: "dataflow-source",
			RunID:   "dataflow-run",
			Mel:     melWithMock,
		}

		// Create a simple envelope to test data passing
		envelope := &Envelope[interface{}]{
			Data: map[string]interface{}{
				"inputToWorkflow": "test data for workflow",
				"processingMode":  "batch",
			},
		}

		// Simulate workflow call data
		workflowReq := WorkflowCallRequest{
			TargetWorkflowID: "dataflow-target",
			CallData: map[string]interface{}{
				"envelope": envelope.Data,
				"context":  "dataflow test",
			},
			CallMode:       "sync",
			TimeoutSeconds: 5,
			SourceContext:  execCtx,
		}

		// Execute the workflow call in a goroutine since it will block waiting for response
		responseChan := make(chan *WorkflowCallResponse, 1)
		errorChan := make(chan error, 1)

		go func() {
			response, err := melWithMock.CallWorkflow(ctx, workflowReq)
			if err != nil {
				errorChan <- err
				return
			}
			responseChan <- response
		}()

		// Give the workflow call time to register the pending call and trigger the workflow
		time.Sleep(100 * time.Millisecond)

		// Find the pending call to get the callID
		melWithMock.(*melImpl).pendingCallsMu.RLock()
		var callID string
		for id := range melWithMock.(*melImpl).pendingCalls {
			callID = id
			break
		}
		melWithMock.(*melImpl).pendingCallsMu.RUnlock()

		if callID == "" {
			t.Fatal("No pending call found")
		}

		// Simulate workflow_return node calling ReturnToWorkflow
		returnData := map[string]interface{}{
			"flowResult": "data flow successful",
			"nodeCount":  3,
		}

		err := melWithMock.ReturnToWorkflow(context.Background(), callID, returnData, "success")
		if err != nil {
			t.Fatalf("ReturnToWorkflow failed: %v", err)
		}

		// Wait for the workflow call to complete
		var response *WorkflowCallResponse
		select {
		case response = <-responseChan:
			// Success
		case err := <-errorChan:
			t.Fatalf("Dataflow workflow call failed: %v", err)
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for workflow call to complete")
		}

		// Verify that data flowed through correctly
		// Since this comes from ReturnToWorkflow, the data is directly in response.Data
		if flowResult, exists := response.Data["flowResult"]; exists {
			if flowResult != "data flow successful" {
				t.Errorf("Expected flow result 'data flow successful', got %v", flowResult)
			}
		} else {
			t.Error("Expected flowResult in response data")
		}

		if nodeCount, exists := response.Data["nodeCount"]; exists {
			if nodeCount != 3 {
				t.Errorf("Expected nodeCount 3, got %v", nodeCount)
			}
		}
	})
}
