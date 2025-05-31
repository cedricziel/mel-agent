package core

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/cedricziel/mel-agent/pkg/api"
)

// GenerateEnvelopeID generates a UUID-like ID for envelopes
func GenerateEnvelopeID() string {
	// Generate 16 random bytes
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID if random fails
		return fmt.Sprintf("env-%d", time.Now().UnixNano())
	}

	// Format as UUID-like string
	return fmt.Sprintf("env-%x-%x-%x-%x-%x",
		bytes[0:4],
		bytes[4:6],
		bytes[6:8],
		bytes[8:10],
		bytes[10:16])
}

// NewEnvelope creates a new envelope with the given data and trace information
func NewEnvelope[T any](data T, trace api.Trace) *api.Envelope[T] {
	return &api.Envelope[T]{
		ID:       GenerateEnvelopeID(),
		IssuedAt: time.Now(),
		Version:  1,
		DataType: inferDataType(data),
		Data:     data,
		Trace:    trace,
	}
}

// NewEnvelopeFromContext creates a new envelope from execution context
func NewEnvelopeFromContext[T any](data T, ctx api.ExecutionContext, nodeID string) *api.Envelope[T] {
	trace := api.Trace{
		AgentID: ctx.AgentID,
		RunID:   ctx.RunID,
		NodeID:  nodeID,
		Step:    nodeID,
		Attempt: 1,
	}

	envelope := NewEnvelope(data, trace)

	// Copy variables from context
	if len(ctx.Variables) > 0 {
		envelope.Variables = make(map[string]interface{})
		for k, v := range ctx.Variables {
			envelope.Variables[k] = v
		}
	}

	return envelope
}

// NewGenericEnvelope creates a new envelope with interface{} data type
func NewGenericEnvelope(data interface{}, trace api.Trace) *api.Envelope[interface{}] {
	return NewEnvelope(data, trace)
}

// NewGenericEnvelopeFromContext creates a new generic envelope from execution context
func NewGenericEnvelopeFromContext(data interface{}, ctx api.ExecutionContext, nodeID string) *api.Envelope[interface{}] {
	return NewEnvelopeFromContext(data, ctx, nodeID)
}

// TransformEnvelope converts an envelope from one type to another
func TransformEnvelope[TIn, TOut any](envelope *api.Envelope[TIn], transformer func(TIn) TOut) *api.Envelope[TOut] {
	transformed := transformer(envelope.Data)

	result := &api.Envelope[TOut]{
		ID:       GenerateEnvelopeID(),
		IssuedAt: time.Now(),
		Version:  envelope.Version,
		DataType: inferDataType(transformed),
		Data:     transformed,
		Trace:    envelope.Trace,
	}

	// Copy metadata
	if len(envelope.Binary) > 0 {
		result.Binary = make(map[string][]byte)
		for k, v := range envelope.Binary {
			result.Binary[k] = v
		}
	}

	if len(envelope.Meta) > 0 {
		result.Meta = make(map[string]string)
		for k, v := range envelope.Meta {
			result.Meta[k] = v
		}
	}

	if len(envelope.Variables) > 0 {
		result.Variables = make(map[string]interface{})
		for k, v := range envelope.Variables {
			result.Variables[k] = v
		}
	}

	if len(envelope.Errors) > 0 {
		result.Errors = make([]api.ExecutionError, len(envelope.Errors))
		copy(result.Errors, envelope.Errors)
	}

	return result
}

// MergeEnvelopes combines multiple envelopes into a single envelope with array data
func MergeEnvelopes[T any](envelopes []*api.Envelope[T]) *api.Envelope[[]T] {
	if len(envelopes) == 0 {
		return nil
	}

	var data []T
	var combinedTrace api.Trace
	var combinedErrors []api.ExecutionError
	combinedMeta := make(map[string]string)
	combinedVars := make(map[string]interface{})

	for i, env := range envelopes {
		data = append(data, env.Data)

		// Use first envelope's trace as base
		if i == 0 {
			combinedTrace = env.Trace
		}

		// Combine errors
		combinedErrors = append(combinedErrors, env.Errors...)

		// Merge metadata (later values override earlier ones)
		for k, v := range env.Meta {
			combinedMeta[k] = v
		}

		// Merge variables (later values override earlier ones)
		for k, v := range env.Variables {
			combinedVars[k] = v
		}
	}

	result := &api.Envelope[[]T]{
		ID:       GenerateEnvelopeID(),
		IssuedAt: time.Now(),
		Version:  envelopes[0].Version,
		DataType: "array",
		Data:     data,
		Trace:    combinedTrace,
		Errors:   combinedErrors,
	}

	if len(combinedMeta) > 0 {
		result.Meta = combinedMeta
	}

	if len(combinedVars) > 0 {
		result.Variables = combinedVars
	}

	return result
}

// SplitEnvelope splits an envelope containing an array into multiple envelopes
func SplitEnvelope[T any](envelope *api.Envelope[[]T]) []*api.Envelope[T] {
	var results []*api.Envelope[T]

	for i, item := range envelope.Data {
		childTrace := envelope.Trace.Child(envelope.Trace.NodeID, i)

		result := &api.Envelope[T]{
			ID:       GenerateEnvelopeID(),
			IssuedAt: time.Now(),
			Version:  envelope.Version,
			DataType: inferDataType(item),
			Data:     item,
			Trace:    childTrace,
		}

		// Copy metadata and variables to each split envelope
		result.Meta = make(map[string]string)
		if len(envelope.Meta) > 0 {
			for k, v := range envelope.Meta {
				result.Meta[k] = v
			}
		}
		// Add split-specific metadata
		result.Meta["split_index"] = fmt.Sprintf("%d", i)
		result.Meta["split_total"] = fmt.Sprintf("%d", len(envelope.Data))

		if len(envelope.Variables) > 0 {
			result.Variables = make(map[string]interface{})
			for k, v := range envelope.Variables {
				result.Variables[k] = v
			}
		}

		if len(envelope.Binary) > 0 {
			result.Binary = make(map[string][]byte)
			for k, v := range envelope.Binary {
				result.Binary[k] = v
			}
		}

		// Don't copy errors to split envelopes (start fresh)

		results = append(results, result)
	}

	return results
}

// ValidateEnvelope performs basic validation on an envelope
func ValidateEnvelope[T any](envelope *api.Envelope[T]) error {
	if envelope == nil {
		return fmt.Errorf("envelope is nil")
	}

	if envelope.ID == "" {
		return fmt.Errorf("envelope ID is required")
	}

	if envelope.IssuedAt.IsZero() {
		return fmt.Errorf("envelope issuedAt timestamp is required")
	}

	if envelope.Trace.AgentID == "" {
		return fmt.Errorf("envelope trace AgentID is required")
	}

	if envelope.Trace.RunID == "" {
		return fmt.Errorf("envelope trace RunID is required")
	}

	if envelope.Trace.NodeID == "" {
		return fmt.Errorf("envelope trace NodeID is required")
	}

	return nil
}

// inferDataType attempts to determine a data type string from the given data
func inferDataType(data interface{}) string {
	if data == nil {
		return "null"
	}

	switch data.(type) {
	case string:
		return "string"
	case int, int8, int16, int32, int64:
		return "integer"
	case uint, uint8, uint16, uint32, uint64:
		return "integer"
	case float32, float64:
		return "number"
	case bool:
		return "boolean"
	case []interface{}:
		return "array"
	case map[string]interface{}:
		return "object"
	default:
		return "unknown"
	}
}
