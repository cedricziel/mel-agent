package api

import (
	"time"
)

// Trace holds IDs for logging, metrics, and correlation
type Trace struct {
	AgentID  string `json:"agentId"`            // Workflow/Agent identifier
	RunID    string `json:"runId"`              // Execution run identifier
	ParentID string `json:"parentId,omitempty"` // Parent workflow for sub-flows
	NodeID   string `json:"nodeId"`             // Current processing node
	Step     string `json:"step"`               // Step name/identifier
	Attempt  int    `json:"attempt"`            // Retry attempt number
}

// Next creates a new trace for the next step in the workflow
func (t Trace) Next(nodeID string) Trace {
	return Trace{
		AgentID:  t.AgentID,
		RunID:    t.RunID,
		ParentID: t.ParentID,
		NodeID:   nodeID,
		Step:     nodeID, // Use nodeID as step by default
		Attempt:  1,      // Reset attempt for new step
	}
}

// Child creates a new trace for a child workflow or split operation
func (t Trace) Child(nodeID string, index int) Trace {
	return Trace{
		AgentID:  t.AgentID,
		RunID:    t.RunID,
		ParentID: t.NodeID, // Current node becomes parent
		NodeID:   nodeID,
		Step:     nodeID,
		Attempt:  1,
	}
}

// Retry creates a new trace for a retry attempt
func (t Trace) Retry() Trace {
	retry := t
	retry.Attempt++
	return retry
}

// ExecutionError captures step-level failures for DLQ & observability
type ExecutionError struct {
	Time    time.Time `json:"time"`
	NodeID  string    `json:"nodeId"`
	Message string    `json:"message"`
	Stack   string    `json:"stack,omitempty"`
	Code    string    `json:"code,omitempty"`
}

// Envelope is the generic unit that flows through the MEL engine
type Envelope[T any] struct {
	ID        string                 `json:"id"` // UUID v4
	IssuedAt  time.Time              `json:"issuedAt"`
	Version   int                    `json:"version"`             // Schema version
	DataType  string                 `json:"dataType"`            // e.g. "HttpRequest", "UserCreated"
	Data      T                      `json:"data"`                // Strongly-typed business payload
	Binary    map[string][]byte      `json:"binary,omitempty"`    // Optional binary attachments
	Meta      map[string]string      `json:"meta,omitempty"`      // Arbitrary metadata
	Variables map[string]interface{} `json:"variables,omitempty"` // Context variables
	Trace     Trace                  `json:"trace"`
	Errors    []ExecutionError       `json:"errors,omitempty"`
}

// AddError appends an execution error to the envelope
func (e *Envelope[T]) AddError(nodeID, message string, err error) {
	execError := ExecutionError{
		Time:    time.Now(),
		NodeID:  nodeID,
		Message: message,
	}

	if nodeErr, ok := err.(*NodeError); ok {
		execError.Code = nodeErr.Code
		// Use the provided message, not the NodeError message
	}

	e.Errors = append(e.Errors, execError)
}

// HasErrors returns true if the envelope contains any errors
func (e *Envelope[T]) HasErrors() bool {
	return len(e.Errors) > 0
}

// LastError returns the most recent error, or nil if no errors exist
func (e *Envelope[T]) LastError() *ExecutionError {
	if len(e.Errors) == 0 {
		return nil
	}
	return &e.Errors[len(e.Errors)-1]
}

// GetVariable retrieves a variable from the envelope or global store
func (e *Envelope[T]) GetVariable(scope VariableScope, key string) (interface{}, bool) {
	// Try envelope-local variables first for run scope
	if scope == RunScope {
		if val, exists := e.Variables[key]; exists {
			return val, true
		}
	}

	// Fall back to global variable store
	ctx := CreateVariableContext(e.Trace.AgentID, e.Trace.RunID, e.Trace.NodeID)
	val, exists, _ := GetVariable(ctx, scope, key)
	return val, exists
}

// SetVariable stores a variable in the envelope or global store
func (e *Envelope[T]) SetVariable(scope VariableScope, key string, value interface{}) error {
	if scope == RunScope {
		if e.Variables == nil {
			e.Variables = make(map[string]interface{})
		}
		e.Variables[key] = value
		return nil
	}

	// Use global variable store for other scopes
	ctx := CreateVariableContext(e.Trace.AgentID, e.Trace.RunID, e.Trace.NodeID)
	return SetVariable(ctx, scope, key, value)
}

// Clone creates a deep copy of the envelope with a new ID and timestamp
func (e *Envelope[T]) Clone() *Envelope[T] {
	clone := &Envelope[T]{
		ID:       generateEnvelopeID(),
		IssuedAt: time.Now(),
		Version:  e.Version,
		DataType: e.DataType,
		Data:     e.Data, // Note: T might need deep copying depending on type
		Trace:    e.Trace,
	}

	// Copy maps to avoid shared references
	if len(e.Binary) > 0 {
		clone.Binary = make(map[string][]byte, len(e.Binary))
		for k, v := range e.Binary {
			clone.Binary[k] = make([]byte, len(v))
			copy(clone.Binary[k], v)
		}
	}

	if len(e.Meta) > 0 {
		clone.Meta = make(map[string]string, len(e.Meta))
		for k, v := range e.Meta {
			clone.Meta[k] = v
		}
	}

	if len(e.Variables) > 0 {
		clone.Variables = make(map[string]interface{}, len(e.Variables))
		for k, v := range e.Variables {
			clone.Variables[k] = v
		}
	}

	if len(e.Errors) > 0 {
		clone.Errors = make([]ExecutionError, len(e.Errors))
		copy(clone.Errors, e.Errors)
	}

	return clone
}

// SetMeta sets a metadata value
func (e *Envelope[T]) SetMeta(key, value string) {
	if e.Meta == nil {
		e.Meta = make(map[string]string)
	}
	e.Meta[key] = value
}

// GetMeta retrieves a metadata value
func (e *Envelope[T]) GetMeta(key string) (string, bool) {
	if e.Meta == nil {
		return "", false
	}
	val, exists := e.Meta[key]
	return val, exists
}

// generateEnvelopeID generates a new UUID for envelopes
// This is a simple implementation - the core package has a more robust version
func generateEnvelopeID() string {
	// Simple timestamp-based ID for now
	return "env-" + time.Now().Format("20060102150405")
}

