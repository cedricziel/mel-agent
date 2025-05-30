package delay

import (
	"testing"
	"time"

	"github.com/cedricziel/mel-agent/pkg/api"
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

func TestDelayDefinition_Execute(t *testing.T) {
	def := delayDefinition{}
	ctx := api.ExecutionContext{
		AgentID:   "test-agent",
		Variables: make(map[string]interface{}),
	}

	tests := []struct {
		name          string
		node          api.Node
		input         interface{}
		expectedOutput interface{}
		minDuration   time.Duration
		maxDuration   time.Duration
	}{
		{
			name: "valid duration",
			node: api.Node{
				Data: map[string]interface{}{
					"duration": float64(100),
				},
			},
			input:         "test input",
			expectedOutput: "test input",
			minDuration:   90 * time.Millisecond,
			maxDuration:   150 * time.Millisecond,
		},
		{
			name: "zero duration",
			node: api.Node{
				Data: map[string]interface{}{
					"duration": float64(0),
				},
			},
			input:         "test input",
			expectedOutput: "test input",
			minDuration:   0,
			maxDuration:   10 * time.Millisecond,
		},
		{
			name: "missing duration",
			node: api.Node{
				Data: map[string]interface{}{},
			},
			input:         "test input",
			expectedOutput: "test input",
			minDuration:   0,
			maxDuration:   10 * time.Millisecond,
		},
		{
			name: "invalid duration type",
			node: api.Node{
				Data: map[string]interface{}{
					"duration": "invalid",
				},
			},
			input:         "test input",
			expectedOutput: "test input",
			minDuration:   0,
			maxDuration:   10 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			output, err := def.Execute(ctx, tt.node, tt.input)
			duration := time.Since(start)

			if err != nil {
				t.Errorf("Execute() error = %v", err)
				return
			}
			if output != tt.expectedOutput {
				t.Errorf("Execute() output = %v, expected %v", output, tt.expectedOutput)
			}
			if duration < tt.minDuration {
				t.Errorf("Execute() duration %v is less than expected minimum %v", duration, tt.minDuration)
			}
			if duration > tt.maxDuration {
				t.Errorf("Execute() duration %v is greater than expected maximum %v", duration, tt.maxDuration)
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