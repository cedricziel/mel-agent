package code

import (
	"testing"
	"time"

	"github.com/cedricziel/mel-agent/pkg/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCodeNode_Integration(t *testing.T) {
	// Test that the node is properly registered
	def := api.FindDefinition("code")
	require.NotNil(t, def, "Code node should be registered")

	meta := def.Meta()
	assert.Equal(t, "code", meta.Type)
	assert.Equal(t, "Code", meta.Label)
	assert.Equal(t, "Code", meta.Category)
	assert.Len(t, meta.Parameters, 4)

	// Initialize the definition
	err := def.Initialize(nil)
	require.NoError(t, err)
}

func TestCodeNode_RealWorldScenarios(t *testing.T) {
	def := NewCodeDefinition().(*codeDefinition)
	require.NoError(t, def.Initialize(nil))

	// Debug: Check if JavaScript runtime is available
	jsRuntime, exists := def.runtimes["javascript"]
	require.True(t, exists, "JavaScript runtime should be available")
	require.NotNil(t, jsRuntime, "JavaScript runtime should not be nil")
	t.Logf("Available runtimes: %v", len(def.runtimes))

	ctx := api.ExecutionContext{
		AgentID: "integration-test-agent",
		RunID:   "integration-test-run",
	}

	tests := []struct {
		name     string
		code     string
		data     interface{}
		vars     map[string]interface{}
		expected interface{}
	}{
		{
			name: "Data validation and cleaning",
			code: `
				const user = input.data;
				const errors = [];
				
				// Validate email
				if (!user.email || !user.email.includes('@')) {
					errors.push('Invalid email');
				}
				
				// Validate age
				if (!user.age || user.age < 0 || user.age > 150) {
					errors.push('Invalid age');
				}
				
				// Clean phone
				const cleanPhone = user.phone ? user.phone.replace(/[^\d]/g, '') : '';
				
				return {
					isValid: errors.length === 0,
					errors: errors,
					cleanedUser: {
						...user,
						email: user.email ? user.email.toLowerCase() : '',
						phone: cleanPhone
					}
				};
			`,
			data: map[string]interface{}{
				"email": "User@Example.COM",
				"age":   25,
				"phone": "+1 (555) 123-4567",
				"name":  "John Doe",
			},
			expected: map[string]interface{}{
				"isValid": true,
				"errors":  []interface{}{},
				"cleanedUser": map[string]interface{}{
					"email": "user@example.com",
					"age":   int64(25),
					"phone": "15551234567",
					"name":  "John Doe",
				},
			},
		},
		{
			name: "API response transformation",
			code: `
				const apiResponse = input.data;
				const users = apiResponse.data || [];
				
				const transformed = users
					.filter(user => user.active)
					.map(user => ({
						id: user.id,
						displayName: user.first_name + ' ' + user.last_name,
						email: user.email_address,
						role: user.user_role || 'user',
						lastSeen: new Date(user.last_login_timestamp).toISOString()
					}));
				
				return {
					users: transformed,
					totalCount: users.length,
					activeCount: transformed.length,
					transformedAt: new Date().toISOString()
				};
			`,
			data: map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{
						"id":                   1,
						"first_name":           "Alice",
						"last_name":            "Smith",
						"email_address":        "alice@example.com",
						"user_role":            "admin",
						"active":               true,
						"last_login_timestamp": "2024-01-15T10:30:00Z",
					},
					map[string]interface{}{
						"id":                   2,
						"first_name":           "Bob",
						"last_name":            "Jones",
						"email_address":        "bob@example.com",
						"active":               false,
						"last_login_timestamp": "2024-01-10T15:45:00Z",
					},
					map[string]interface{}{
						"id":                   3,
						"first_name":           "Carol",
						"last_name":            "Wilson",
						"email_address":        "carol@example.com",
						"user_role":            "editor",
						"active":               true,
						"last_login_timestamp": "2024-01-20T08:15:00Z",
					},
				},
			},
			expected: map[string]interface{}{
				"totalCount":  int64(3),
				"activeCount": int64(2),
			},
		},
		{
			name: "Working with variables",
			code: `
				const settings = input.variables.settings || {};
				const data = input.data;
				
				// Apply settings-based transformation
				let result = data.value;
				
				if (settings.multiply) {
					result *= settings.multiplier || 1;
				}
				
				if (settings.addPrefix) {
					result = (settings.prefix || 'RESULT') + ': ' + result;
				}
				
				return {
					original: data.value,
					processed: result,
					settingsApplied: Object.keys(settings)
				};
			`,
			data: map[string]interface{}{
				"value": 42,
			},
			vars: map[string]interface{}{
				"settings": map[string]interface{}{
					"multiply":   true,
					"multiplier": 2,
					"addPrefix":  true,
					"prefix":     "CALC",
				},
			},
			expected: map[string]interface{}{
				"original":  int64(42),
				"processed": "CALC: 84",
			},
		},
		{
			name: "Utility functions usage",
			code: `
				const data = input.data;
				
				// Parse JSON from string field
				const config = utils.parseJSON(data.configJson);
				
				// Generate tracking ID
				const trackingId = utils.generateUUID();
				
				// Hash sensitive data
				const hashedEmail = utils.md5(data.email);
				
				return {
					config: config,
					trackingId: trackingId,
					hashedEmail: hashedEmail,
					trackingIdLength: trackingId.length,
					processedAt: new Date().toISOString()
				};
			`,
			data: map[string]interface{}{
				"email":      "user@example.com",
				"configJson": `{"theme": "dark", "notifications": true, "timeout": 30}`,
			},
			expected: map[string]interface{}{
				"config": map[string]interface{}{
					"theme":         "dark",
					"notifications": true,
					"timeout":       30.0, // JSON parsing returns float64 for numbers
				},
				"hashedEmail":      "b58996c504c5638798eb6b511e6f49af", // MD5 of "user@example.com"
				"trackingIdLength": int64(36),                          // UUID length
			},
		},
		{
			name: "Error handling in code",
			code: `
				try {
					const data = input.data;
					
					if (!data.required_field) {
						throw new Error('Required field is missing');
					}
					
					// Simulate some processing
					const result = data.value * 2;
					
					return {
						success: true,
						result: result
					};
				} catch (error) {
					return {
						success: false,
						error: error.message,
						timestamp: new Date().toISOString()
					};
				}
			`,
			data: map[string]interface{}{
				"value": 10,
				// missing required_field intentionally
			},
			expected: map[string]interface{}{
				"success": false,
				"error":   "Required field is missing",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envelope := &api.Envelope[interface{}]{
				ID:        "integration-test-envelope",
				IssuedAt:  time.Now(),
				Version:   1,
				Data:      tt.data,
				Variables: tt.vars,
				Trace:     newTestTrace(),
			}

			node := api.Node{
				ID:   "integration-test-node",
				Type: "code",
				Data: map[string]interface{}{
					"language": "javascript",
					"code":     tt.code,
					"timeout":  30.0,
				},
			}

			result, err := def.ExecuteEnvelope(ctx, node, envelope)
			require.NoError(t, err, "Code execution should not fail")
			require.NotNil(t, result, "Result should not be nil")

			resultData, ok := result.Data.(map[string]interface{})
			require.True(t, ok, "Result data should be a map")

			// Check expected fields exist
			for key, expectedValue := range tt.expected.(map[string]interface{}) {
				actualValue, exists := resultData[key]
				require.True(t, exists, "Expected key %s should exist in result", key)

				// For complex nested structures, just check type and presence
				if key == "users" || key == "settingsApplied" {
					assert.NotNil(t, actualValue, "Value for %s should not be nil", key)
				} else {
					assert.Equal(t, expectedValue, actualValue, "Value for %s should match", key)
				}
			}
		})
	}
}

func TestCodeNode_ErrorScenarios(t *testing.T) {
	def := NewCodeDefinition().(*codeDefinition)
	require.NoError(t, def.Initialize(nil))

	ctx := api.ExecutionContext{
		AgentID: "error-test-agent",
		RunID:   "error-test-run",
	}

	envelope := &api.Envelope[interface{}]{
		ID:       "error-test-envelope",
		IssuedAt: time.Now(),
		Version:  1,
		Data:     map[string]interface{}{"test": "data"},
		Trace:    newTestTrace(),
	}

	tests := []struct {
		name          string
		nodeData      map[string]interface{}
		expectedError string
	}{
		{
			name: "Missing code parameter",
			nodeData: map[string]interface{}{
				"language": "javascript",
				"timeout":  30.0,
			},
			expectedError: "code parameter is required",
		},
		{
			name: "Empty code parameter",
			nodeData: map[string]interface{}{
				"language": "javascript",
				"code":     "",
				"timeout":  30.0,
			},
			expectedError: "code parameter is required",
		},
		{
			name: "Unsupported language",
			nodeData: map[string]interface{}{
				"language": "cobol",
				"code":     "DISPLAY 'Hello World'.",
				"timeout":  30.0,
			},
			expectedError: "unsupported language: cobol",
		},
		{
			name: "JavaScript syntax error",
			nodeData: map[string]interface{}{
				"language": "javascript",
				"code":     "return {invalid: syntax error;;;",
				"timeout":  30.0,
			},
			expectedError: "execution error:",
		},
		{
			name: "JavaScript runtime error",
			nodeData: map[string]interface{}{
				"language": "javascript",
				"code":     "throw new Error('Something went wrong');",
				"timeout":  30.0,
			},
			expectedError: "execution error:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := api.Node{
				ID:   "error-test-node",
				Type: "code",
				Data: tt.nodeData,
			}

			result, err := def.ExecuteEnvelope(ctx, node, envelope)
			assert.Nil(t, result, "Result should be nil on error")
			assert.Error(t, err, "Should return an error")
			assert.Contains(t, err.Error(), tt.expectedError, "Error message should contain expected text")
		})
	}
}

func TestCodeNode_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	def := NewCodeDefinition().(*codeDefinition)
	require.NoError(t, def.Initialize(nil))

	ctx := api.ExecutionContext{
		AgentID: "perf-test-agent",
		RunID:   "perf-test-run",
	}

	envelope := &api.Envelope[interface{}]{
		ID:       "perf-test-envelope",
		IssuedAt: time.Now(),
		Version:  1,
		Data:     map[string]interface{}{"iterations": 1000},
		Trace:    newTestTrace(),
	}

	node := api.Node{
		ID:   "perf-test-node",
		Type: "code",
		Data: map[string]interface{}{
			"language": "javascript",
			"code": `
				const iterations = input.data.iterations;
				let sum = 0;
				
				for (let i = 0; i < iterations; i++) {
					sum += Math.sqrt(i) * Math.sin(i);
				}
				
				return {
					result: sum,
					iterations: iterations,
					computedAt: new Date().toISOString()
				};
			`,
			"timeout": 10.0,
		},
	}

	start := time.Now()
	result, err := def.ExecuteEnvelope(ctx, node, envelope)
	duration := time.Since(start)

	require.NoError(t, err, "Performance test should not fail")
	require.NotNil(t, result, "Result should not be nil")

	// Should complete well under the timeout
	assert.Less(t, duration, 5*time.Second, "Should complete in reasonable time")

	resultData, ok := result.Data.(map[string]interface{})
	require.True(t, ok, "Result should be a map")
	assert.Contains(t, resultData, "result", "Should contain result")
	assert.Contains(t, resultData, "iterations", "Should contain iterations")

	t.Logf("Performance test completed in %v", duration)
}
