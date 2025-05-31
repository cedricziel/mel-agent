package core

import (
	"fmt"
	"sync"

	"github.com/cedricziel/mel-agent/pkg/api"
)

// SplitterNode implements the split pattern for envelope arrays
type SplitterNode struct{}

// Meta returns metadata for the Splitter node
func (s *SplitterNode) Meta() api.NodeType {
	return api.NodeType{
		Type:     "envelope_splitter",
		Label:    "Split Array",
		Icon:     "🔀",
		Category: "Control Flow",
		Parameters: []api.ParameterDefinition{
			api.NewStringParameter("keyPath", "Array Key Path", false).
				WithDescription("JSONPath to array field (leave empty for root array)").
				WithGroup("Settings"),
			api.NewBooleanParameter("preserveEmpty", "Preserve Empty Arrays", false).
				WithDefault(false).
				WithDescription("Emit empty result when input array is empty").
				WithGroup("Settings"),
		},
	}
}

// Initialize implements api.EnvelopeNodeDefinition
func (s *SplitterNode) Initialize(mel api.Mel) error {
	return nil
}

// ExecuteEnvelope splits an envelope containing an array into multiple envelopes
func (s *SplitterNode) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[interface{}]) (*api.Envelope[interface{}], error) {
	keyPath, _ := node.Data["keyPath"].(string)
	preserveEmpty, _ := node.Data["preserveEmpty"].(bool)
	
	var arrayData []interface{}
	
	// Extract array from envelope data
	if keyPath == "" {
		// Root data should be an array
		if arr, ok := envelope.Data.([]interface{}); ok {
			arrayData = arr
		} else {
			return envelope, api.NewNodeError(node.ID, node.Type, "input data is not an array")
		}
	} else {
		// Extract array from specified path
		arrayData = extractArrayFromPath(envelope.Data, keyPath)
		if arrayData == nil {
			return envelope, api.NewNodeError(node.ID, node.Type, fmt.Sprintf("no array found at path: %s", keyPath))
		}
	}
	
	// Handle empty arrays
	if len(arrayData) == 0 && !preserveEmpty {
		return envelope, api.NewNodeError(node.ID, node.Type, "input array is empty")
	}
	
	// Split the array into individual envelopes
	splitEnvelopes := SplitEnvelope(&api.Envelope[[]interface{}]{
		ID:        envelope.ID,
		IssuedAt:  envelope.IssuedAt,
		Version:   envelope.Version,
		DataType:  "array",
		Data:      arrayData,
		Binary:    envelope.Binary,
		Meta:      envelope.Meta,
		Variables: envelope.Variables,
		Trace:     envelope.Trace,
		Errors:    envelope.Errors,
	})
	
	// Return the first envelope and store the rest for subsequent calls
	// Note: In a real implementation, this would need a proper queue system
	if len(splitEnvelopes) > 0 {
		// Convert back to generic envelope
		result := &api.Envelope[interface{}]{
			ID:        splitEnvelopes[0].ID,
			IssuedAt:  splitEnvelopes[0].IssuedAt,
			Version:   splitEnvelopes[0].Version,
			DataType:  splitEnvelopes[0].DataType,
			Data:      splitEnvelopes[0].Data,
			Binary:    splitEnvelopes[0].Binary,
			Meta:      splitEnvelopes[0].Meta,
			Variables: splitEnvelopes[0].Variables,
			Trace:     splitEnvelopes[0].Trace,
			Errors:    splitEnvelopes[0].Errors,
		}
		
		// Add metadata about the split operation
		if result.Meta == nil {
			result.Meta = make(map[string]string)
		}
		result.Meta["split_operation"] = "true"
		result.Meta["split_total"] = fmt.Sprintf("%d", len(splitEnvelopes))
		
		return result, nil
	}
	
	return envelope, nil
}

// AggregatorNode implements the aggregate pattern for collecting envelopes
type AggregatorNode struct {
	storage map[string][]*api.Envelope[interface{}] // Keyed by workflow+step
	mu      sync.RWMutex
}

// NewAggregatorNode creates a new aggregator node
func NewAggregatorNode() *AggregatorNode {
	return &AggregatorNode{
		storage: make(map[string][]*api.Envelope[interface{}]),
	}
}

// Meta returns metadata for the Aggregator node
func (a *AggregatorNode) Meta() api.NodeType {
	return api.NodeType{
		Type:     "envelope_aggregator",
		Label:    "Aggregate Items",
		Icon:     "🔗",
		Category: "Control Flow",
		Parameters: []api.ParameterDefinition{
			api.NewIntegerParameter("expectedCount", "Expected Count", false).
				WithDescription("Number of items expected (0 = auto-detect from split metadata)").
				WithGroup("Settings"),
			api.NewStringParameter("timeoutMs", "Timeout (ms)", false).
				WithDescription("Maximum time to wait for all items").
				WithGroup("Settings"),
			api.NewBooleanParameter("partialResults", "Allow Partial Results", false).
				WithDefault(false).
				WithDescription("Emit results even if not all items received").
				WithGroup("Settings"),
		},
	}
}

// Initialize implements api.EnvelopeNodeDefinition
func (a *AggregatorNode) Initialize(mel api.Mel) error {
	return nil
}

// ExecuteEnvelope collects envelopes and emits when quorum is met
func (a *AggregatorNode) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[interface{}]) (*api.Envelope[interface{}], error) {
	expectedCount, _ := node.Data["expectedCount"].(int)
	partialResults, _ := node.Data["partialResults"].(bool)
	
	// Generate storage key from run and node context
	key := fmt.Sprintf("%s:%s:%s", envelope.Trace.AgentID, envelope.Trace.RunID, node.ID)
	
	a.mu.Lock()
	defer a.mu.Unlock()
	
	// Add envelope to collection
	a.storage[key] = append(a.storage[key], envelope)
	
	// Determine expected count if not specified
	if expectedCount == 0 {
		if totalStr, exists := envelope.GetMeta("split_total"); exists {
			fmt.Sscanf(totalStr, "%d", &expectedCount)
		} else {
			expectedCount = 1 // Default to single item
		}
	}
	
	currentCount := len(a.storage[key])
	
	// Check if we have enough items or should emit partial results
	shouldEmit := currentCount >= expectedCount || 
		(partialResults && currentCount > 0)
	
	if shouldEmit {
		// Collect all items
		items := a.storage[key]
		var data []interface{}
		for _, item := range items {
			data = append(data, item.Data)
		}
		
		// Create aggregated envelope
		result := &api.Envelope[interface{}]{
			ID:       GenerateEnvelopeID(),
			IssuedAt: envelope.IssuedAt,
			Version:  envelope.Version,
			DataType: "array",
			Data:     data,
			Trace:    envelope.Trace.Next(node.ID),
		}
		
		// Merge metadata from all items
		mergedMeta := make(map[string]string)
		mergedVars := make(map[string]interface{})
		var allErrors []api.ExecutionError
		
		for _, item := range items {
			// Merge metadata
			for k, v := range item.Meta {
				mergedMeta[k] = v
			}
			
			// Merge variables
			for k, v := range item.Variables {
				mergedVars[k] = v
			}
			
			// Collect errors
			allErrors = append(allErrors, item.Errors...)
		}
		
		if len(mergedMeta) > 0 {
			result.Meta = mergedMeta
			// Add aggregation metadata
			result.Meta["aggregated_count"] = fmt.Sprintf("%d", len(items))
			result.Meta["aggregation_complete"] = fmt.Sprintf("%t", currentCount >= expectedCount)
		}
		
		if len(mergedVars) > 0 {
			result.Variables = mergedVars
		}
		
		if len(allErrors) > 0 {
			result.Errors = allErrors
		}
		
		// Clean up storage
		delete(a.storage, key)
		
		return result, nil
	}
	
	// Not ready to emit yet - return nil to indicate waiting
	return nil, nil
}

// BatchProcessorNode processes items in batches
type BatchProcessorNode struct{}

// Meta returns metadata for the Batch Processor node
func (b *BatchProcessorNode) Meta() api.NodeType {
	return api.NodeType{
		Type:     "envelope_batch_processor",
		Label:    "Batch Processor",
		Icon:     "📦",
		Category: "Control Flow",
		Parameters: []api.ParameterDefinition{
			api.NewIntegerParameter("batchSize", "Batch Size", true).
				WithDescription("Number of items per batch").
				WithGroup("Settings"),
			api.NewBooleanParameter("processPartial", "Process Partial Batches", false).
				WithDefault(true).
				WithDescription("Process remaining items even if batch is not full").
				WithGroup("Settings"),
		},
	}
}

// Initialize implements api.EnvelopeNodeDefinition
func (b *BatchProcessorNode) Initialize(mel api.Mel) error {
	return nil
}

// ExecuteEnvelope processes items in configurable batches
func (b *BatchProcessorNode) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[interface{}]) (*api.Envelope[interface{}], error) {
	batchSize, _ := node.Data["batchSize"].(int)
	if batchSize <= 0 {
		return envelope, api.NewNodeError(node.ID, node.Type, "batchSize must be greater than 0")
	}
	
	// Extract array from envelope
	var items []interface{}
	if arr, ok := envelope.Data.([]interface{}); ok {
		items = arr
	} else {
		return envelope, api.NewNodeError(node.ID, node.Type, "input data must be an array")
	}
	
	// Split into batches
	var batches [][]interface{}
	for i := 0; i < len(items); i += batchSize {
		end := i + batchSize
		if end > len(items) {
			end = len(items)
		}
		batches = append(batches, items[i:end])
	}
	
	// Return first batch (in real implementation, would queue the rest)
	if len(batches) > 0 {
		result := envelope.Clone()
		result.Data = batches[0]
		result.Trace = envelope.Trace.Next(node.ID)
		
		// Add batch metadata
		result.SetMeta("batch_size", fmt.Sprintf("%d", len(batches[0])))
		result.SetMeta("total_batches", fmt.Sprintf("%d", len(batches)))
		result.SetMeta("batch_index", "0")
		
		return result, nil
	}
	
	return envelope, nil
}

// Helper function to extract array from JSONPath-like path
func extractArrayFromPath(data interface{}, path string) []interface{} {
	if path == "" {
		if arr, ok := data.([]interface{}); ok {
			return arr
		}
		return nil
	}
	
	// Simple path traversal for demonstration
	// In production, would use a proper JSONPath library
	current := data
	parts := splitPath(path)
	
	for _, part := range parts {
		if part == "" {
			continue
		}
		
		switch v := current.(type) {
		case map[string]interface{}:
			current = v[part]
		default:
			return nil
		}
		
		if current == nil {
			return nil
		}
	}
	
	if arr, ok := current.([]interface{}); ok {
		return arr
	}
	
	return nil
}

// Helper function to split path by dots
func splitPath(path string) []string {
	if path == "" {
		return nil
	}
	
	var parts []string
	var current string
	
	for _, char := range path {
		if char == '.' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	
	if current != "" {
		parts = append(parts, current)
	}
	
	return parts
}