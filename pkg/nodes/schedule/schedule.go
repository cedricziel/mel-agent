package schedule

import (
	"time"

	"github.com/cedricziel/mel-agent/pkg/api"
)

type scheduleDefinition struct{}

func (scheduleDefinition) Meta() api.NodeType {
	return api.NodeType{
		Type:       "schedule",
		Label:      "Schedule",
		Icon:       "🗓️",
		Category:   "Triggers",
		EntryPoint: true,
		Parameters: []api.ParameterDefinition{
			api.NewStringParameter("cron", "Cron Expression", true).WithDefault("").WithGroup("Schedule").WithDescription("Cron schedule to run"),
			api.NewEnumParameter("mode", "Mode", []string{"async", "sync"}, true).WithDefault("async").WithGroup("Execution"),
			api.NewNumberParameter("statusCode", "Response Status", false).WithDefault(202).WithGroup("Response"),
			api.NewStringParameter("responseBody", "Response Body", false).WithDefault("").WithGroup("Response"),
		},
	}
}

// ExecuteEnvelope returns schedule data including current timestamp.
func (d scheduleDefinition) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[interface{}]) (*api.Envelope[interface{}], error) {
	// Create schedule execution data with current timestamp
	now := time.Now()
	scheduleData := map[string]interface{}{
		"timestamp":     now.Format(time.RFC3339),
		"unix":          now.Unix(),
		"unixNano":      now.UnixNano(),
		"year":          now.Year(),
		"month":         int(now.Month()),
		"monthName":     now.Month().String(),
		"day":           now.Day(),
		"weekday":       int(now.Weekday()),
		"weekdayName":   now.Weekday().String(),
		"hour":          now.Hour(),
		"minute":        now.Minute(),
		"second":        now.Second(),
		"timezone":      now.Location().String(),
		"scheduledAt":   now.Format(time.RFC3339),
		"cron":          node.Data["cron"],
		"triggerType":   "schedule",
	}

	result := envelope.Clone()
	result.Trace = envelope.Trace.Next(node.ID)
	result.Data = scheduleData
	result.DataType = "object"
	
	return result, nil
}

func (scheduleDefinition) Initialize(mel api.Mel) error {
	return nil
}

func init() {
	api.RegisterNodeDefinition(scheduleDefinition{})
}

// assert that scheduleDefinition implements both interfaces
var _ api.NodeDefinition = (*scheduleDefinition)(nil)
