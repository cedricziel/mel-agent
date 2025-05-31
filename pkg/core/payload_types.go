package core

// HTTPPayload for HTTP request/response data
type HTTPPayload struct {
	Method   string            `json:"method"`
	URL      string            `json:"url"`
	Headers  map[string]string `json:"headers"`
	Body     interface{}       `json:"body"`
	Status   int              `json:"status,omitempty"`
}

// DatabasePayload for database operations
type DatabasePayload struct {
	Query     string                 `json:"query"`
	Params    []interface{}         `json:"params,omitempty"`
	Results   []map[string]interface{} `json:"results,omitempty"`
	RowCount  int                   `json:"rowCount,omitempty"`
}

// FilePayload for file operations
type FilePayload struct {
	Path      string `json:"path"`
	Name      string `json:"name"`
	MimeType  string `json:"mimeType"`
	Size      int64  `json:"size"`
	BinaryKey string `json:"binaryKey,omitempty"` // Reference to Binary map
}

// GenericPayload for flexible data
type GenericPayload map[string]interface{}

// EmailPayload for email operations
type EmailPayload struct {
	To      []string `json:"to"`
	CC      []string `json:"cc,omitempty"`
	BCC     []string `json:"bcc,omitempty"`
	Subject string   `json:"subject"`
	Body    string   `json:"body"`
	HTML    bool     `json:"html,omitempty"`
}

// SlackPayload for Slack messaging
type SlackPayload struct {
	Channel string `json:"channel"`
	Text    string `json:"text"`
	Blocks  interface{} `json:"blocks,omitempty"`
}

// VariablePayload for variable operations
type VariablePayload struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
	Scope string      `json:"scope"`
}

// TimerPayload for timer/delay operations
type TimerPayload struct {
	Duration  int64  `json:"duration"`  // milliseconds
	Unit      string `json:"unit"`      // "ms", "s", "m", "h"
	Timestamp int64  `json:"timestamp"` // Unix timestamp when timer expires
}

// ErrorPayload for error information
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// WebhookPayload for webhook trigger data
type WebhookPayload struct {
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Query   map[string]string `json:"query"`
	Body    interface{}       `json:"body"`
	Path    string           `json:"path"`
}

// SchedulePayload for scheduled trigger data
type SchedulePayload struct {
	CronExpression string `json:"cronExpression"`
	Timezone       string `json:"timezone,omitempty"`
	LastRun        string `json:"lastRun,omitempty"`
	NextRun        string `json:"nextRun,omitempty"`
}

// LogPayload for logging operations
type LogPayload struct {
	Level   string      `json:"level"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ScriptPayload for script execution
type ScriptPayload struct {
	Language string      `json:"language"`
	Code     string      `json:"code"`
	Result   interface{} `json:"result,omitempty"`
	Output   string      `json:"output,omitempty"`
}

// ConditionPayload for conditional logic
type ConditionPayload struct {
	Expression string      `json:"expression"`
	Result     bool        `json:"result"`
	Context    interface{} `json:"context,omitempty"`
}

// LoopPayload for iteration operations
type LoopPayload struct {
	Items   []interface{} `json:"items"`
	Index   int          `json:"index"`
	Current interface{}   `json:"current"`
	Total   int          `json:"total"`
}

// AggregatePayload for collecting multiple inputs
type AggregatePayload struct {
	Items []interface{} `json:"items"`
	Count int          `json:"count"`
	Keys  []string     `json:"keys,omitempty"`
}

// TransformPayload for data transformation
type TransformPayload struct {
	Input   interface{} `json:"input"`
	Output  interface{} `json:"output"`
	Mapping interface{} `json:"mapping,omitempty"`
}