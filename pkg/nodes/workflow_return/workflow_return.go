package workflow_return

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cedricziel/mel-agent/pkg/api"
)

type workflowReturnDefinition struct{}

func (workflowReturnDefinition) Meta() api.NodeType {
	return api.NodeType{
		Type:     "workflow_return",
		Label:    "Workflow Return",
		Icon:     "↩️",
		Category: "Workflow",
		Parameters: []api.ParameterDefinition{
			api.NewObjectParameter("returnData", "Return Data", false).
				WithDefault("{}").
				WithGroup("Response").
				WithDescription("Data to return to the calling workflow").
				WithValidators(api.ValidatorSpec{Type: "json"}),
			api.NewBooleanParameter("returnCurrentData", "Return Current Data", false).
				WithDefault(true).
				WithGroup("Response").
				WithDescription("Include current workflow data in the response"),
			api.NewEnumParameter("returnStatus", "Return Status", []string{"success", "error", "warning"}, false).
				WithDefault("success").
				WithGroup("Response").
				WithDescription("Status to return to the calling workflow"),
			api.NewStringParameter("returnMessage", "Return Message", false).
				WithDefault("").
				WithGroup("Response").
				WithDescription("Optional message to include with the response"),
			api.NewBooleanParameter("terminateWorkflow", "Terminate Workflow", false).
				WithDefault(true).
				WithGroup("Execution").
				WithDescription("Whether to stop the current workflow after returning"),
		},
	}
}

// ExecuteEnvelope returns data to a calling workflow and optionally terminates current execution.
func (d workflowReturnDefinition) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[interface{}]) (*api.Envelope[interface{}], error) {
	returnCurrentData, _ := node.Data["returnCurrentData"].(bool)
	returnStatus, _ := node.Data["returnStatus"].(string)
	returnMessage, _ := node.Data["returnMessage"].(string)
	terminateWorkflow, _ := node.Data["terminateWorkflow"].(bool)

	if returnStatus == "" {
		returnStatus = "success"
	}

	// Prepare return data
	returnData := make(map[string]interface{})
	
	// Add custom return data if provided
	if customData, ok := node.Data["returnData"].(string); ok && customData != "" {
		var parsed interface{}
		if err := json.Unmarshal([]byte(customData), &parsed); err == nil {
			if parsedMap, ok := parsed.(map[string]interface{}); ok {
				for k, v := range parsedMap {
					returnData[k] = v
				}
			}
		}
	}

	// Include current workflow data if requested
	if returnCurrentData {
		returnData["workflowData"] = envelope.Data
		returnData["workflowTrace"] = envelope.Trace
	}

	// Extract calling workflow information from current data
	var callId string
	if currentData, ok := envelope.Data.(map[string]interface{}); ok {
		if id, exists := currentData["callId"]; exists {
			if idStr, ok := id.(string); ok {
				callId = idStr
			}
		}
	}

	// Use the platform's workflow return capability if this workflow was called by another workflow
	if callId != "" {
		err := ctx.Mel.ReturnToWorkflow(context.Background(), callId, returnData, returnStatus)
		if err != nil {
			return nil, fmt.Errorf("workflow return failed: %w", err)
		}

		result := envelope.Clone()
		result.Trace = envelope.Trace.Next(node.ID)
		result.DataType = "object"
		result.Data = map[string]interface{}{
			"returnResponse": map[string]interface{}{
				"returnId":         fmt.Sprintf("return-%d", time.Now().UnixNano()),
				"returnedAt":       time.Now().Format(time.RFC3339),
				"sourceWorkflow":   ctx.AgentID,
				"sourceRun":        ctx.RunID,
				"sourceNode":       node.ID,
				"status":           returnStatus,
				"callId":           callId,
				"terminateWorkflow": terminateWorkflow,
			},
			"data":       returnData,
			"message":    fmt.Sprintf("Workflow return sent with status: %s", returnStatus),
			"success":    true,
			"terminated": terminateWorkflow,
		}

		if terminateWorkflow {
			result.Data.(map[string]interface{})["__terminate_workflow"] = true
		}

		return result, nil
	}

	// Fallback implementation
	returnResponse := map[string]interface{}{
		"returnId":         fmt.Sprintf("return-%d", time.Now().UnixNano()),
		"returnedAt":       time.Now().Format(time.RFC3339),
		"sourceWorkflow":   ctx.AgentID,
		"sourceRun":        ctx.RunID,
		"sourceNode":       node.ID,
		"status":           returnStatus,
		"data":             returnData,
		"terminateWorkflow": terminateWorkflow,
		"mode":             "fallback",
	}

	if returnMessage != "" {
		returnResponse["message"] = returnMessage
	}

	result := envelope.Clone()
	result.Trace = envelope.Trace.Next(node.ID)
	result.DataType = "object"
	result.Data = map[string]interface{}{
		"returnResponse": returnResponse,
		"message":        fmt.Sprintf("Workflow return executed with status: %s (fallback mode)", returnStatus),
		"success":        returnStatus == "success",
		"terminated":     terminateWorkflow,
	}

	if terminateWorkflow {
		result.Data.(map[string]interface{})["__terminate_workflow"] = true
	}

	return result, nil
}

func (workflowReturnDefinition) Initialize(mel api.Mel) error {
	return nil
}

func init() {
	api.RegisterNodeDefinition(workflowReturnDefinition{})
}

// assert that workflowReturnDefinition implements the interface
var _ api.NodeDefinition = (*workflowReturnDefinition)(nil)