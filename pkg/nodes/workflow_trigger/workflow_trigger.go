package workflow_trigger

import (
	"time"

	"github.com/cedricziel/mel-agent/pkg/api"
)

type workflowTriggerDefinition struct{}

func (workflowTriggerDefinition) Meta() api.NodeType {
	return api.NodeType{
		Type:       "workflow_trigger",
		Label:      "Workflow Trigger",
		Icon:       "🔗",
		Category:   "Triggers",
		EntryPoint: true,
		Parameters: []api.ParameterDefinition{
			api.NewStringParameter("triggerName", "Trigger Name", false).
				WithDefault("Workflow Call Trigger").
				WithGroup("Configuration").
				WithDescription("Name for this workflow trigger"),
			api.NewStringParameter("allowedCallers", "Allowed Callers", false).
				WithDefault("*").
				WithGroup("Security").
				WithDescription("Comma-separated list of workflow IDs allowed to call this workflow, or * for any"),
			api.NewBooleanParameter("requireAuthentication", "Require Authentication", false).
				WithDefault(false).
				WithGroup("Security").
				WithDescription("Whether calling workflows must be authenticated"),
			api.NewBooleanParameter("logCalls", "Log Calls", false).
				WithDefault(true).
				WithGroup("Monitoring").
				WithDescription("Whether to log workflow call events"),
			api.NewObjectParameter("defaultData", "Default Data", false).
				WithDefault("{}").
				WithGroup("Data").
				WithDescription("Default data to merge with incoming call data").
				WithValidators(api.ValidatorSpec{Type: "json"}),
		},
	}
}

// ExecuteEnvelope processes incoming workflow calls and prepares data for downstream nodes.
func (d workflowTriggerDefinition) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[interface{}]) (*api.Envelope[interface{}], error) {
	now := time.Now()
	triggerName, _ := node.Data["triggerName"].(string)
	if triggerName == "" {
		triggerName = "Workflow Call Trigger"
	}

	allowedCallers, _ := node.Data["allowedCallers"].(string)
	if allowedCallers == "" {
		allowedCallers = "*"
	}

	requireAuth, _ := node.Data["requireAuthentication"].(bool)
	logCalls, _ := node.Data["logCalls"].(bool)

	// Create trigger data with workflow call information
	triggerData := map[string]interface{}{
		"triggerType":    "workflow_call",
		"triggerName":    triggerName,
		"triggeredAt":    now.Format(time.RFC3339),
		"triggeredBy":    "workflow_call",
		"timestamp":      now.Format(time.RFC3339),
		"unix":           now.Unix(),
		"allowedCallers": allowedCallers,
		"requireAuth":    requireAuth,
		"logCalls":       logCalls,
	}

	// Extract call information from the envelope data
	// This would typically be set by the workflow execution engine when a workflow_call triggers this workflow
	if callData, ok := envelope.Data.(map[string]interface{}); ok {
		// Extract calling workflow information
		if sourceWorkflowId, exists := callData["sourceWorkflowId"]; exists {
			triggerData["callingWorkflow"] = sourceWorkflowId
		}
		if sourceRunId, exists := callData["sourceRunId"]; exists {
			triggerData["callingRun"] = sourceRunId
		}
		if sourceNodeId, exists := callData["sourceNodeId"]; exists {
			triggerData["callingNode"] = sourceNodeId
		}
		if callId, exists := callData["callId"]; exists {
			triggerData["callId"] = callId
		}

		// Include the call data payload
		if sourceData, exists := callData["sourceData"]; exists {
			triggerData["callData"] = sourceData
		}

		// Security check: verify caller is allowed
		if allowedCallers != "*" {
			if sourceWorkflowId, ok := callData["sourceWorkflowId"].(string); ok {
				allowed := false
				// Simple comma-separated check (could be enhanced with regex or more complex rules)
				if sourceWorkflowId != "" {
					// For now, just check if the source workflow ID is in the allowed list
					// TODO: Implement proper caller validation
					allowed = true // Simplified for initial implementation
				}

				if !allowed {
					triggerData["error"] = "Caller not authorized"
					triggerData["authorized"] = false
				} else {
					triggerData["authorized"] = true
				}
			}
		} else {
			triggerData["authorized"] = true
		}
	} else {
		// No call data means this might be a test run or manual trigger
		triggerData["callData"] = map[string]interface{}{}
		triggerData["authorized"] = true
		triggerData["testRun"] = true
	}

	// Merge with default data if provided
	if defaultDataStr, ok := node.Data["defaultData"].(string); ok && defaultDataStr != "" {
		// TODO: Parse and merge default data
		triggerData["hasDefaultData"] = true
	}

	// Log the call if enabled
	if logCalls {
		triggerData["logged"] = true
		triggerData["loggedAt"] = now.Format(time.RFC3339)
		// TODO: Implement actual logging to database or external system
	}

	result := envelope.Clone()
	result.Trace = envelope.Trace.Next(node.ID)
	result.Data = triggerData
	result.DataType = "object"

	return result, nil
}

func (workflowTriggerDefinition) Initialize(mel api.Mel) error {
	return nil
}

func init() {
	api.RegisterNodeDefinition(workflowTriggerDefinition{})
}

// assert that workflowTriggerDefinition implements both interfaces
var _ api.NodeDefinition = (*workflowTriggerDefinition)(nil)
