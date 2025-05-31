package credentials

import (
	"testing"

	"github.com/cedricziel/mel-agent/pkg/api"
)

func TestCredentialRegistration(t *testing.T) {
	// Test that credentials are properly registered
	types := api.ListCredentialDefinitions()
	
	expectedTypes := []string{"api_key", "baserow_jwt", "baserow_token"}
	found := make(map[string]bool)
	
	for _, credType := range types {
		found[credType.Type] = true
	}
	
	for _, expected := range expectedTypes {
		if !found[expected] {
			t.Errorf("Expected credential type %s to be registered", expected)
		}
	}
}

func TestAPIKeyCredential(t *testing.T) {
	def := api.FindCredentialDefinition("api_key")
	if def == nil {
		t.Fatal("API key credential definition not found")
	}
	
	// Test validation
	validData := map[string]interface{}{
		"api_key": "test-key",
	}
	
	if err := def.Validate(validData); err != nil {
		t.Errorf("Valid data should not produce error: %v", err)
	}
	
	invalidData := map[string]interface{}{}
	if err := def.Validate(invalidData); err == nil {
		t.Error("Invalid data should produce error")
	}
}

func TestBaserowJWTCredential(t *testing.T) {
	def := api.FindCredentialDefinition("baserow_jwt")
	if def == nil {
		t.Fatal("Baserow JWT credential definition not found")
	}
	
	// Test validation
	validData := map[string]interface{}{
		"baseUrl":  "https://api.baserow.io",
		"username": "testuser",
		"password": "testpass",
	}
	
	if err := def.Validate(validData); err != nil {
		t.Errorf("Valid data should not produce error: %v", err)
	}
	
	invalidData := map[string]interface{}{
		"baseUrl": "https://api.baserow.io",
		// missing username and password
	}
	if err := def.Validate(invalidData); err == nil {
		t.Error("Invalid data should produce error")
	}
}