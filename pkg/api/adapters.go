package api

import (
	"fmt"
)

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

	// Check for node definition (envelope-based)
	if node, ok := def.(NodeDefinition); ok {
		return node
	}

	return nil
}

