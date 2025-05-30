package if_node

import (
	"strings"
	"testing"

	"github.com/cedricziel/mel-agent/pkg/api"
)

func TestIfNode_BasicConditions(t *testing.T) {
	def := ifDefinition{}

	tests := []struct {
		name           string
		conditions     []interface{}
		hasElse        bool
		input          interface{}
		expectedBranch string
		expectedMatch  bool
	}{
		{
			name: "simple true condition",
			conditions: []interface{}{
				map[string]interface{}{"expression": "true", "branch": "always"},
			},
			hasElse:        false,
			input:          map[string]interface{}{"value": 10},
			expectedBranch: "always",
			expectedMatch:  true,
		},
		{
			name: "simple false condition with else",
			conditions: []interface{}{
				map[string]interface{}{"expression": "false", "branch": "never"},
			},
			hasElse:        true,
			input:          map[string]interface{}{"value": 10},
			expectedBranch: "else",
			expectedMatch:  false,
		},
		{
			name: "number comparison - greater than",
			conditions: []interface{}{
				map[string]interface{}{"expression": "input.value > 5", "branch": "high"},
			},
			hasElse:        false,
			input:          map[string]interface{}{"value": 10},
			expectedBranch: "high",
			expectedMatch:  true,
		},
		{
			name: "number comparison - less than",
			conditions: []interface{}{
				map[string]interface{}{"expression": "input.value < 5", "branch": "low"},
			},
			hasElse:        false,
			input:          map[string]interface{}{"value": 3},
			expectedBranch: "low",
			expectedMatch:  true,
		},
		{
			name: "string equality",
			conditions: []interface{}{
				map[string]interface{}{"expression": "input.name == \"test\"", "branch": "match"},
			},
			hasElse:        false,
			input:          map[string]interface{}{"name": "test"},
			expectedBranch: "match",
			expectedMatch:  true,
		},
		{
			name: "multiple conditions - first matches",
			conditions: []interface{}{
				map[string]interface{}{"expression": "input.value > 15", "branch": "very_high"},
				map[string]interface{}{"expression": "input.value > 10", "branch": "high"},
				map[string]interface{}{"expression": "input.value > 5", "branch": "medium"},
			},
			hasElse:        false,
			input:          map[string]interface{}{"value": 20},
			expectedBranch: "very_high",
			expectedMatch:  true,
		},
		{
			name: "multiple conditions - middle matches",
			conditions: []interface{}{
				map[string]interface{}{"expression": "input.value > 15", "branch": "very_high"},
				map[string]interface{}{"expression": "input.value > 10", "branch": "high"},
				map[string]interface{}{"expression": "input.value > 5", "branch": "medium"},
			},
			hasElse:        false,
			input:          map[string]interface{}{"value": 12},
			expectedBranch: "high",
			expectedMatch:  true,
		},
		{
			name: "multiple conditions - none match, with else",
			conditions: []interface{}{
				map[string]interface{}{"expression": "input.value > 15", "branch": "very_high"},
				map[string]interface{}{"expression": "input.value > 10", "branch": "high"},
			},
			hasElse:        true,
			input:          map[string]interface{}{"value": 5},
			expectedBranch: "else",
			expectedMatch:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := api.Node{
				Data: map[string]interface{}{
					"conditions": tt.conditions,
					"hasElse":    tt.hasElse,
				},
			}

			ctx := api.ExecutionContext{
				AgentID:   "test-agent",
				Variables: map[string]interface{}{},
			}

			result, err := def.Execute(ctx, node, tt.input)
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			ifResult, ok := result.(*IfResult)
			if !ok {
				t.Fatalf("Expected *IfResult, got %T", result)
			}

			if ifResult.Branch != tt.expectedBranch {
				t.Errorf("Expected branch %q, got %q", tt.expectedBranch, ifResult.Branch)
			}

			if ifResult.Matched != tt.expectedMatch {
				t.Errorf("Expected matched %v, got %v", tt.expectedMatch, ifResult.Matched)
			}
		})
	}
}

func TestIfNode_ExpressionEvaluation(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		input      interface{}
		expected   bool
		shouldErr  bool
	}{
		{
			name:       "literal true",
			expression: "true",
			input:      nil,
			expected:   true,
		},
		{
			name:       "literal false",
			expression: "false",
			input:      nil,
			expected:   false,
		},
		{
			name:       "property existence",
			expression: "input.name",
			input:      map[string]interface{}{"name": "test"},
			expected:   true,
		},
		{
			name:       "property non-existence",
			expression: "input.missing",
			input:      map[string]interface{}{"name": "test"},
			expected:   false,
		},
		{
			name:       "greater than - true",
			expression: "input.value > 10",
			input:      map[string]interface{}{"value": 15},
			expected:   true,
		},
		{
			name:       "greater than - false",
			expression: "input.value > 10",
			input:      map[string]interface{}{"value": 5},
			expected:   false,
		},
		{
			name:       "equality - true",
			expression: "input.status == \"active\"",
			input:      map[string]interface{}{"status": "active"},
			expected:   true,
		},
		{
			name:       "equality - false",
			expression: "input.status == \"active\"",
			input:      map[string]interface{}{"status": "inactive"},
			expected:   false,
		},
		{
			name:       "inequality - true",
			expression: "input.status != \"active\"",
			input:      map[string]interface{}{"status": "inactive"},
			expected:   true,
		},
		{
			name:       "less than or equal - true",
			expression: "input.value <= 10",
			input:      map[string]interface{}{"value": 10},
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := api.ExecutionContext{}
			result, err := evaluateExpression(tt.expression, tt.input, ctx)

			if tt.shouldErr {
				if err == nil {
					t.Errorf("Expected error, got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestIfNode_Meta(t *testing.T) {
	def := ifDefinition{}
	meta := def.Meta()

	if meta.Type != "if" {
		t.Errorf("Expected type 'if', got %s", meta.Type)
	}

	if meta.Label != "If" {
		t.Errorf("Expected label 'If', got %s", meta.Label)
	}

	if !meta.Branching {
		t.Error("Expected branching to be true")
	}

	if len(meta.Parameters) != 2 {
		t.Errorf("Expected 2 parameters, got %d", len(meta.Parameters))
	}
}

func TestIfNode_InvalidConditions(t *testing.T) {
	def := ifDefinition{}
	
	tests := []struct {
		name       string
		conditions interface{}
		shouldErr  bool
		errMsg     string
	}{
		{
			name:       "non-array conditions",
			conditions: "invalid type",
			shouldErr:  true,
			errMsg:     "conditions must be an array",
		},
		{
			name:       "array with non-object condition",
			conditions: []interface{}{"invalid"},
			shouldErr:  true,
			errMsg:     "condition 0 must be an object",
		},
		{
			name:       "condition missing expression",
			conditions: []interface{}{map[string]interface{}{"branch": "test"}},
			shouldErr:  true,
			errMsg:     "condition 0 missing expression string",
		},
		{
			name:       "condition missing branch",
			conditions: []interface{}{map[string]interface{}{"expression": "true"}},
			shouldErr:  true,
			errMsg:     "condition 0 missing branch string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := api.Node{
				Data: map[string]interface{}{
					"conditions": tt.conditions,
					"hasElse":    false,
				},
			}

			ctx := api.ExecutionContext{}
			_, err := def.Execute(ctx, node, map[string]interface{}{})

			if tt.shouldErr {
				if err == nil {
					t.Errorf("Expected error containing '%s', got none", tt.errMsg)
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}