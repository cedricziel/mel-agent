package api

import (
	"fmt"
	"time"
)

// LegacyToEnvelopeAdapter wraps a legacy NodeDefinition to work with envelopes
type LegacyToEnvelopeAdapter struct {
	legacy NodeDefinition
}

// NewLegacyToEnvelopeAdapter creates an adapter for legacy nodes
func NewLegacyToEnvelopeAdapter(legacy NodeDefinition) *LegacyToEnvelopeAdapter {
	return &LegacyToEnvelopeAdapter{legacy: legacy}
}

// Initialize delegates to the legacy node
func (a *LegacyToEnvelopeAdapter) Initialize(mel Mel) error {
	return a.legacy.Initialize(mel)
}

// Meta delegates to the legacy node
func (a *LegacyToEnvelopeAdapter) Meta() NodeType {
	return a.legacy.Meta()
}

// ExecuteEnvelope adapts the legacy Execute method to work with envelopes
func (a *LegacyToEnvelopeAdapter) ExecuteEnvelope(ctx ExecutionContext, node Node, envelope *Envelope[interface{}]) (*Envelope[interface{}], error) {
	// Extract data for legacy node
	output, err := a.legacy.Execute(ctx, node, envelope.Data)
	if err != nil {
		// Add error to envelope
		envelope.AddError(node.ID, "Legacy node execution failed", err)
		return envelope, err
	}
	
	// Create new envelope with output
	result := &Envelope[interface{}]{
		ID:        generateEnvelopeID(),
		IssuedAt:  time.Now(),
		Version:   envelope.Version,
		DataType:  inferDataType(output),
		Data:      output,
		Trace:     envelope.Trace.Next(node.ID),
		Variables: envelope.Variables, // Carry forward variables
	}
	
	// Copy metadata and binary data
	if len(envelope.Meta) > 0 {
		result.Meta = make(map[string]string)
		for k, v := range envelope.Meta {
			result.Meta[k] = v
		}
	}
	
	if len(envelope.Binary) > 0 {
		result.Binary = make(map[string][]byte)
		for k, v := range envelope.Binary {
			result.Binary[k] = v
		}
	}
	
	// Copy errors but mark this step as successful
	if len(envelope.Errors) > 0 {
		result.Errors = make([]ExecutionError, len(envelope.Errors))
		copy(result.Errors, envelope.Errors)
	}
	
	return result, nil
}

// EnvelopeToLegacyAdapter wraps an EnvelopeNodeDefinition to work with legacy interface
type EnvelopeToLegacyAdapter struct {
	envelope EnvelopeNodeDefinition
}

// NewEnvelopeToLegacyAdapter creates an adapter for envelope nodes to legacy interface
func NewEnvelopeToLegacyAdapter(envelope EnvelopeNodeDefinition) *EnvelopeToLegacyAdapter {
	return &EnvelopeToLegacyAdapter{envelope: envelope}
}

// Initialize delegates to the envelope node
func (a *EnvelopeToLegacyAdapter) Initialize(mel Mel) error {
	return a.envelope.Initialize(mel)
}

// Meta delegates to the envelope node
func (a *EnvelopeToLegacyAdapter) Meta() NodeType {
	return a.envelope.Meta()
}

// Execute adapts the envelope ExecuteEnvelope method to work with legacy interface
func (a *EnvelopeToLegacyAdapter) Execute(ctx ExecutionContext, node Node, input interface{}) (interface{}, error) {
	// Create envelope from input
	envelope := &Envelope[interface{}]{
		ID:       generateEnvelopeID(),
		IssuedAt: time.Now(),
		Version:  1,
		DataType: inferDataType(input),
		Data:     input,
		Trace: Trace{
			AgentID: ctx.AgentID,
			RunID:   ctx.RunID,
			NodeID:  node.ID,
			Step:    node.ID,
			Attempt: 1,
		},
		Variables: ctx.Variables,
	}
	
	// Execute with envelope
	result, err := a.envelope.ExecuteEnvelope(ctx, node, envelope)
	if err != nil {
		return input, err
	}
	
	// Return just the data for legacy compatibility
	return result.Data, nil
}

// TypedToEnvelopeAdapter wraps a TypedNodeDefinition to work with generic envelopes
type TypedToEnvelopeAdapter[TIn, TOut any] struct {
	typed TypedNodeDefinition[TIn, TOut]
}

// NewTypedToEnvelopeAdapter creates an adapter for typed nodes
func NewTypedToEnvelopeAdapter[TIn, TOut any](typed TypedNodeDefinition[TIn, TOut]) *TypedToEnvelopeAdapter[TIn, TOut] {
	return &TypedToEnvelopeAdapter[TIn, TOut]{typed: typed}
}

// Initialize delegates to the typed node
func (a *TypedToEnvelopeAdapter[TIn, TOut]) Initialize(mel Mel) error {
	return a.typed.Initialize(mel)
}

// Meta delegates to the typed node
func (a *TypedToEnvelopeAdapter[TIn, TOut]) Meta() NodeType {
	return a.typed.Meta()
}

// ExecuteEnvelope adapts the typed ExecuteTyped method to work with generic envelopes
func (a *TypedToEnvelopeAdapter[TIn, TOut]) ExecuteEnvelope(ctx ExecutionContext, node Node, envelope *Envelope[interface{}]) (*Envelope[interface{}], error) {
	// Attempt to cast input data to expected type
	typedData, ok := envelope.Data.(TIn)
	if !ok {
		err := fmt.Errorf("input data type mismatch: expected %T, got %T", *new(TIn), envelope.Data)
		envelope.AddError(node.ID, "Type conversion failed", err)
		return envelope, err
	}
	
	// Create typed envelope
	typedEnvelope := &Envelope[TIn]{
		ID:        envelope.ID,
		IssuedAt:  envelope.IssuedAt,
		Version:   envelope.Version,
		DataType:  envelope.DataType,
		Data:      typedData,
		Binary:    envelope.Binary,
		Meta:      envelope.Meta,
		Variables: envelope.Variables,
		Trace:     envelope.Trace,
		Errors:    envelope.Errors,
	}
	
	// Execute with typed envelope
	typedResult, err := a.typed.ExecuteTyped(ctx, node, typedEnvelope)
	if err != nil {
		envelope.AddError(node.ID, "Typed node execution failed", err)
		return envelope, err
	}
	
	// Convert back to generic envelope
	result := &Envelope[interface{}]{
		ID:        typedResult.ID,
		IssuedAt:  typedResult.IssuedAt,
		Version:   typedResult.Version,
		DataType:  typedResult.DataType,
		Data:      any(typedResult.Data),
		Binary:    typedResult.Binary,
		Meta:      typedResult.Meta,
		Variables: typedResult.Variables,
		Trace:     typedResult.Trace,
		Errors:    typedResult.Errors,
	}
	
	return result, nil
}

// AutoDetectNodeType attempts to determine the best interface for a given node
func AutoDetectNodeType(def interface{}) interface{} {
	// Check for typed node definition first (most specific)
	if typed, ok := def.(interface {
		Initialize(mel Mel) error
		Meta() NodeType
		ExecuteTyped(ctx ExecutionContext, node Node, envelope *Envelope[interface{}]) (*Envelope[interface{}], error)
	}); ok {
		return typed
	}
	
	// Check for envelope node definition
	if env, ok := def.(EnvelopeNodeDefinition); ok {
		return env
	}
	
	// Fall back to legacy node definition
	if legacy, ok := def.(NodeDefinition); ok {
		return NewLegacyToEnvelopeAdapter(legacy)
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