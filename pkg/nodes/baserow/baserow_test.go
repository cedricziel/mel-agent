package baserow

import (
	"testing"

	"github.com/cedricziel/mel-agent/pkg/api"
)

func TestBaserowNodeMeta(t *testing.T) {
	def := baserowDefinition{}
	meta := def.Meta()

	if meta.Type != "baserow" {
		t.Errorf("Expected type 'baserow', got '%s'", meta.Type)
	}

	if meta.Label != "Baserow" {
		t.Errorf("Expected label 'Baserow', got '%s'", meta.Label)
	}

	if meta.Category != "Database" {
		t.Errorf("Expected category 'Database', got '%s'", meta.Category)
	}

	// Check that we have the required parameters
	paramNames := make(map[string]bool)
	for _, param := range meta.Parameters {
		paramNames[param.Name] = true
	}

	requiredParams := []string{"credentialId", "resource", "operation"}
	for _, required := range requiredParams {
		if !paramNames[required] {
			t.Errorf("Missing required parameter: %s", required)
		}
	}
	
	// Check that optional dynamic parameters exist
	optionalParams := []string{"databaseId", "tableId", "rowId"}
	for _, optional := range optionalParams {
		if !paramNames[optional] {
			t.Errorf("Missing optional parameter: %s", optional)
		}
	}
}

func TestGetIntParameter(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]interface{}
		key      string
		expected int
		hasError bool
	}{
		{
			name:     "valid int",
			data:     map[string]interface{}{"test": 42},
			key:      "test",
			expected: 42,
			hasError: false,
		},
		{
			name:     "valid float64",
			data:     map[string]interface{}{"test": 42.0},
			key:      "test",
			expected: 42,
			hasError: false,
		},
		{
			name:     "valid string",
			data:     map[string]interface{}{"test": "42"},
			key:      "test",
			expected: 42,
			hasError: false,
		},
		{
			name:     "missing parameter",
			data:     map[string]interface{}{},
			key:      "test",
			expected: 0,
			hasError: true,
		},
		{
			name:     "invalid string",
			data:     map[string]interface{}{"test": "not-a-number"},
			key:      "test",
			expected: 0,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getIntParameter(tt.data, tt.key)
			
			if tt.hasError && err == nil {
				t.Errorf("Expected error but got none")
			}
			
			if !tt.hasError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestBaserowClientCreation(t *testing.T) {
	client := NewBaserowClient("https://example.baserow.io", "test-token")
	
	if client.BaseURL != "https://example.baserow.io" {
		t.Errorf("Expected BaseURL 'https://example.baserow.io', got '%s'", client.BaseURL)
	}
	
	if client.Token != "test-token" {
		t.Errorf("Expected Token 'test-token', got '%s'", client.Token)
	}
	
	if client.Client == nil {
		t.Error("Expected HTTP client to be initialized")
	}
}

func TestBaserowConnectionStructure(t *testing.T) {
	// Test token-based connection
	tokenConn := BaserowConnection{
		BaseURL: "https://example.baserow.io",
		Token:   "test-token",
	}
	
	if tokenConn.BaseURL != "https://example.baserow.io" {
		t.Errorf("Expected BaseURL to be set")
	}
	
	if tokenConn.Token != "test-token" {
		t.Errorf("Expected Token to be set")
	}
	
	// Test JWT-based connection
	jwtConn := BaserowConnection{
		BaseURL:  "https://example.baserow.io",
		Username: "testuser",
		Password: "testpass",
	}
	
	if jwtConn.BaseURL != "https://example.baserow.io" {
		t.Errorf("Expected BaseURL to be set")
	}
	
	if jwtConn.Username != "testuser" {
		t.Errorf("Expected Username to be set")
	}
	
	if jwtConn.Password != "testpass" {
		t.Errorf("Expected Password to be set")
	}
}

func TestNodeImplementsInterface(t *testing.T) {
	var _ api.NodeDefinition = (*baserowDefinition)(nil)
}