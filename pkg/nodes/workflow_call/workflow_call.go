package workflow_call

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cedricziel/mel-agent/pkg/api"
)

type workflowCallDefinition struct{}

func (workflowCallDefinition) Meta() api.NodeType {
	return api.NodeType{
		Type:     "workflow_call",
		Label:    "Workflow Call",
		Icon:     "📞",
		Category: "Workflow",
		Parameters: []api.ParameterDefinition{
			api.NewStringParameter("targetWorkflowId", "Target Workflow ID", true).
				WithGroup("Target").
				WithDescription("ID of the workflow to call"),
			api.NewStringParameter("targetWorkflowName", "Target Workflow Name", false).
				WithGroup("Target").
				WithDescription("Optional: Name of the workflow for reference"),
			api.NewObjectParameter("callData", "Call Data", false).
				WithDefault("{}").
				WithGroup("Data").
				WithDescription("Data to pass to the target workflow").
				WithValidators(api.ValidatorSpec{Type: "json"}),
			api.NewEnumParameter("callMode", "Call Mode", []string{"async", "sync"}, false).
				WithDefault("async").
				WithGroup("Execution").
				WithDescription("Whether to wait for the workflow to complete"),
			api.NewNumberParameter("timeoutSeconds", "Timeout (seconds)", false).
				WithDefault(30).
				WithGroup("Execution").
				WithDescription("Timeout for synchronous calls").
				WithVisibilityCondition("callMode=='sync'"),
			api.NewBooleanParameter("passCurrentData", "Pass Current Data", false).
				WithDefault(true).
				WithGroup("Data").
				WithDescription("Include current workflow data in the call"),
		},
	}
}

// ExecuteEnvelope calls another workflow and optionally waits for response.
func (d workflowCallDefinition) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[interface{}]) (*api.Envelope[interface{}], error) {
	targetWorkflowId, ok := node.Data["targetWorkflowId"].(string)
	if !ok || targetWorkflowId == "" {
		return nil, fmt.Errorf("targetWorkflowId is required")
	}

	callMode, _ := node.Data["callMode"].(string)
	if callMode == "" {
		callMode = "async"
	}

	passCurrentData, _ := node.Data["passCurrentData"].(bool)
	timeoutSeconds, _ := node.Data["timeoutSeconds"].(float64)
	if timeoutSeconds <= 0 {
		timeoutSeconds = 30
	}

	// Prepare call data
	callData := make(map[string]interface{})

	// Add custom call data if provided
	if customData, ok := node.Data["callData"].(string); ok && customData != "" {
		var parsed interface{}
		if err := json.Unmarshal([]byte(customData), &parsed); err == nil {
			if parsedMap, ok := parsed.(map[string]interface{}); ok {
				for k, v := range parsedMap {
					callData[k] = v
				}
			}
		}
	}

	// Include current workflow data if requested
	if passCurrentData {
		callData["sourceWorkflowId"] = ctx.AgentID
		callData["sourceRunId"] = ctx.RunID
		callData["sourceNodeId"] = node.ID
		callData["sourceData"] = envelope.Data
		callData["sourceTrace"] = envelope.Trace
	}

	// Use the platform's workflow calling capability
	workflowReq := api.WorkflowCallRequest{
		TargetWorkflowID: targetWorkflowId,
		CallData:         callData,
		CallMode:         callMode,
		TimeoutSeconds:   int(timeoutSeconds),
		SourceContext:    ctx,
	}

	response, err := ctx.Mel.CallWorkflow(context.Background(), workflowReq)
	if err != nil {
		return nil, fmt.Errorf("workflow call failed: %w", err)
	}

	result := envelope.Clone()
	result.Trace = envelope.Trace.Next(node.ID)
	result.DataType = "object"
	result.Data = map[string]interface{}{
		"callInfo": map[string]interface{}{
			"callId":           response.CallID,
			"targetWorkflowId": targetWorkflowId,
			"callMode":         callMode,
			"calledAt":         time.Now().Format(time.RFC3339),
			"sourceWorkflow":   ctx.AgentID,
			"sourceRun":        ctx.RunID,
			"sourceNode":       node.ID,
			"status":           response.Status,
			"completedAt":      response.CompletedAt.Format(time.RFC3339),
		},
		"response": response.Data,
		"message":  response.Message,
		"success":  response.Status == "success" || response.Status == "sent",
	}

	return result, nil
}

func (workflowCallDefinition) Initialize(mel api.Mel) error {
	return nil
}

func init() {
	api.RegisterNodeDefinition(workflowCallDefinition{})
}

// assert that workflowCallDefinition implements the interface
var _ api.NodeDefinition = (*workflowCallDefinition)(nil)
