package manual_trigger

import (
	"encoding/json"
	"time"

	"github.com/cedricziel/mel-agent/pkg/api"
)

type manualTriggerDefinition struct{}

func (manualTriggerDefinition) Meta() api.NodeType {
	return api.NodeType{
		Type:       "manual_trigger",
		Label:      "Manual Trigger",
		Icon:       "▶️",
		Category:   "Triggers",
		EntryPoint: true,
		Parameters: []api.ParameterDefinition{
			api.NewStringParameter("label", "Trigger Label", false).WithDefault("Manual Trigger").WithGroup("Configuration").WithDescription("Custom label for this trigger"),
			api.NewStringParameter("data", "Payload Data", false).WithDefault("{}").WithGroup("Data").WithDescription("JSON data to inject into the workflow"),
			api.NewEnumParameter("format", "Data Format", []string{"json", "text", "number"}, false).WithDefault("json").WithGroup("Data"),
			api.NewBooleanParameter("timestamp", "Include Timestamp", false).WithDefault(true).WithGroup("Data").WithDescription("Add timestamp to the output"),
		},
	}
}

// ExecuteEnvelope returns manual trigger data with optional custom payload.
func (d manualTriggerDefinition) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[interface{}]) (*api.Envelope[interface{}], error) {
	now := time.Now()
	
	// Start with the base trigger data
	triggerData := map[string]interface{}{
		"triggerType": "manual",
		"triggeredBy": "user",
	}

	// Add timestamp if enabled
	if timestamp, ok := node.Data["timestamp"].(bool); ok && timestamp {
		triggerData["timestamp"] = now.Format(time.RFC3339)
		triggerData["unix"] = now.Unix()
		triggerData["triggeredAt"] = now.Format(time.RFC3339)
	}

	// Parse custom data based on format
	if dataStr, ok := node.Data["data"].(string); ok && dataStr != "" {
		format, _ := node.Data["format"].(string)
		if format == "" {
			format = "json"
		}

		switch format {
		case "json":
			// Try to parse as JSON, fallback to string if invalid
			var jsonData interface{}
			if err := json.Unmarshal([]byte(dataStr), &jsonData); err == nil {
				triggerData["payload"] = jsonData
			} else {
				triggerData["payload"] = dataStr
				triggerData["parseError"] = "Invalid JSON, treated as string"
			}
		case "number":
			// Try to parse as number, fallback to string if invalid
			if num, err := parseNumber(dataStr); err == nil {
				triggerData["payload"] = num
			} else {
				triggerData["payload"] = dataStr
				triggerData["parseError"] = "Invalid number, treated as string"
			}
		case "text":
			triggerData["payload"] = dataStr
		default:
			triggerData["payload"] = dataStr
		}
	} else {
		// No custom data provided, use empty object
		triggerData["payload"] = map[string]interface{}{}
	}

	// Add the trigger label
	if label, ok := node.Data["label"].(string); ok && label != "" {
		triggerData["label"] = label
	}

	result := envelope.Clone()
	result.Trace = envelope.Trace.Next(node.ID)
	result.Data = triggerData
	result.DataType = "object"
	
	return result, nil
}

func (manualTriggerDefinition) Initialize(mel api.Mel) error {
	return nil
}

// parseNumber attempts to parse a string as either int or float
func parseNumber(s string) (interface{}, error) {
	// Try int first
	if i, err := parseInt(s); err == nil {
		return i, nil
	}
	// Try float
	if f, err := parseFloat(s); err == nil {
		return f, nil
	}
	return nil, &parseError{s: s}
}

type parseError struct {
	s string
}

func (e *parseError) Error() string {
	return "cannot parse " + e.s + " as number"
}

// Simple int parser to avoid importing strconv
func parseInt(s string) (int, error) {
	var result int
	var sign = 1
	var i = 0
	
	if len(s) == 0 {
		return 0, &parseError{s: s}
	}
	
	if s[0] == '-' {
		sign = -1
		i = 1
	} else if s[0] == '+' {
		i = 1
	}
	
	for ; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return 0, &parseError{s: s}
		}
		result = result*10 + int(s[i]-'0')
	}
	
	return result * sign, nil
}

// Simple float parser to avoid importing strconv
func parseFloat(s string) (float64, error) {
	var result float64
	var sign float64 = 1
	var i = 0
	var decimal = false
	var decimalPlace float64 = 1
	
	if len(s) == 0 {
		return 0, &parseError{s: s}
	}
	
	if s[0] == '-' {
		sign = -1
		i = 1
	} else if s[0] == '+' {
		i = 1
	}
	
	for ; i < len(s); i++ {
		if s[i] == '.' {
			if decimal {
				return 0, &parseError{s: s} // Multiple decimal points
			}
			decimal = true
			continue
		}
		
		if s[i] < '0' || s[i] > '9' {
			return 0, &parseError{s: s}
		}
		
		digit := float64(s[i] - '0')
		
		if decimal {
			decimalPlace *= 10
			result += digit / decimalPlace
		} else {
			result = result*10 + digit
		}
	}
	
	return result * sign, nil
}

func init() {
	api.RegisterNodeDefinition(manualTriggerDefinition{})
}

// assert that manualTriggerDefinition implements both interfaces
var _ api.NodeDefinition = (*manualTriggerDefinition)(nil)