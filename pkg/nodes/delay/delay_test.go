package delay

import (
	"testing"
	"time"

	"github.com/cedricziel/mel-agent/pkg/api"
	"github.com/cedricziel/mel-agent/pkg/core"
)

func TestDelayDefinition_Meta(t *testing.T) {
	def := delayDefinition{}
	meta := def.Meta()

	if meta.Type != "delay" {
		t.Errorf("expected type 'delay', got %s", meta.Type)
	}
	if meta.Label != "Delay" {
		t.Errorf("expected label 'Delay', got %s", meta.Label)
	}
	if meta.Category != "Control" {
		t.Errorf("expected category 'Control', got %s", meta.Category)
	}
	if len(meta.Parameters) != 1 {
		t.Errorf("expected 1 parameter, got %d", len(meta.Parameters))
	}

	param := meta.Parameters[0]
	if param.Name != "duration" {
		t.Errorf("expected parameter name 'duration', got %s", param.Name)
	}
	if param.Type != "number" {
		t.Errorf("expected parameter type 'number', got %s", param.Type)
	}
	if !param.Required {
		t.Error("expected parameter to be required")
	}
	if param.Default != 1000 {
		t.Errorf("expected default value 1000, got %v", param.Default)
	}
}

func TestDelayDefinition_ExecuteEnvelope(t *testing.T) {
	def := delayDefinition{}
	ctx := api.ExecutionContext{
		AgentID:   "test-agent",
		RunID:     "test-run",
		Variables: make(map[string]interface{}),
	}

	tests := []struct {
		name           string
		node           api.Node
		input          interface{}
		expectedOutput interface{}
		minDuration    time.Duration
		maxDuration    time.Duration
	}{
		{
			name: "valid duration",
			node: api.Node{
				ID:   "delay-node",
				Type: "delay",
				Data: map[string]interface{}{
					"duration": float64(100),
				},
			},
			input:          "test input",
			expectedOutput: "test input",
			minDuration:    90 * time.Millisecond,
			maxDuration:    150 * time.Millisecond,
		},
		{
			name: "zero duration",
			node: api.Node{
				ID:   "delay-node",
				Type: "delay",
				Data: map[string]interface{}{
					"duration": float64(0),
				},
			},
			input:          "test input",
			expectedOutput: "test input",
			minDuration:    0,
			maxDuration:    10 * time.Millisecond,
		},
		{
			name: "missing duration",
			node: api.Node{
				ID:   "delay-node",
				Type: "delay",
				Data: map[string]interface{}{},
			},
			input:          "test input",
			expectedOutput: "test input",
			minDuration:    0,
			maxDuration:    10 * time.Millisecond,
		},
		{
			name: "invalid duration type",
			node: api.Node{
				ID:   "delay-node",
				Type: "delay",
				Data: map[string]interface{}{
					"duration": "invalid",
				},
			},
			input:          "test input",
			expectedOutput: "test input",
			minDuration:    0,
			maxDuration:    10 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create input envelope
			trace := api.Trace{
				AgentID: ctx.AgentID,
				RunID:   ctx.RunID,
				NodeID:  tt.node.ID,
				Step:    tt.node.ID,
				Attempt: 1,
			}
			inputEnvelope := core.NewEnvelope(tt.input, trace)

			start := time.Now()
			outputEnvelope, err := def.ExecuteEnvelope(ctx, tt.node, inputEnvelope)
			duration := time.Since(start)

			if err != nil {
				t.Errorf("ExecuteEnvelope() error = %v", err)
				return
			}

			if outputEnvelope == nil {
				t.Error("ExecuteEnvelope() returned nil envelope")
				return
			}

			if outputEnvelope.Data != tt.expectedOutput {
				t.Errorf("ExecuteEnvelope() output = %v, expected %v", outputEnvelope.Data, tt.expectedOutput)
			}

			if duration < tt.minDuration {
				t.Errorf("ExecuteEnvelope() duration %v is less than expected minimum %v", duration, tt.minDuration)
			}
			if duration > tt.maxDuration {
				t.Errorf("ExecuteEnvelope() duration %v is greater than expected maximum %v", duration, tt.maxDuration)
			}

			// Verify trace is properly updated
			if outputEnvelope.Trace.NodeID != tt.node.ID {
				t.Errorf("Expected trace NodeID %s, got %s", tt.node.ID, outputEnvelope.Trace.NodeID)
			}
		})
	}
}

func TestDelayDefinition_Initialize(t *testing.T) {
	def := delayDefinition{}
	err := def.Initialize(nil)
	if err != nil {
		t.Errorf("Initialize() error = %v, expected nil", err)
	}
}

func TestDelayDefinition_ImplementsInterface(t *testing.T) {
	var _ api.NodeDefinition = (*delayDefinition)(nil)
}
