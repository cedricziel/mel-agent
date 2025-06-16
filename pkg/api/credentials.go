package api

import (
	"fmt"
	"sync"
)

// CredentialDefinition defines how to collect and validate credentials for an integration
type CredentialDefinition interface {
	// Type returns the unique identifier for this credential type
	Type() string

	// Name returns the human-readable name for this credential type
	Name() string

	// Description returns a description of what this credential type is for
	Description() string

	// Parameters returns the parameter definitions for collecting credentials
	Parameters() []ParameterDefinition

	// Validate validates the provided credential data
	Validate(data map[string]interface{}) error

	// Transform allows the credential to transform the data before storage
	// This can be used for things like exchanging username/password for tokens
	Transform(data map[string]interface{}) (map[string]interface{}, error)

	// Test tests the credentials by attempting to authenticate/connect
	// Returns nil if successful, error if failed
	Test(data map[string]interface{}) error
}

// CredentialType represents metadata about a credential type
type CredentialType struct {
	Type        string                `json:"type"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Parameters  []ParameterDefinition `json:"parameters"`
}

// credentialRegistry holds all registered credential definitions
var credentialRegistry = struct {
	mu          sync.RWMutex
	definitions map[string]CredentialDefinition
}{
	definitions: make(map[string]CredentialDefinition),
}

// RegisterCredentialDefinition registers a credential definition globally
func RegisterCredentialDefinition(def CredentialDefinition) {
	credentialRegistry.mu.Lock()
	defer credentialRegistry.mu.Unlock()
	credentialRegistry.definitions[def.Type()] = def
}

// ListCredentialDefinitions returns all registered credential definitions
func ListCredentialDefinitions() []CredentialType {
	credentialRegistry.mu.RLock()
	defer credentialRegistry.mu.RUnlock()

	types := make([]CredentialType, 0, len(credentialRegistry.definitions))
	for _, def := range credentialRegistry.definitions {
		types = append(types, CredentialType{
			Type:        def.Type(),
			Name:        def.Name(),
			Description: def.Description(),
			Parameters:  def.Parameters(),
		})
	}
	return types
}

// FindCredentialDefinition finds a credential definition by type
func FindCredentialDefinition(credentialType string) CredentialDefinition {
	credentialRegistry.mu.RLock()
	defer credentialRegistry.mu.RUnlock()
	return credentialRegistry.definitions[credentialType]
}

// ValidateCredentials validates credential data using the appropriate definition
func ValidateCredentials(credentialType string, data map[string]interface{}) error {
	def := FindCredentialDefinition(credentialType)
	if def == nil {
		return fmt.Errorf("unknown credential type: %s", credentialType)
	}
	return def.Validate(data)
}

// TransformCredentials transforms credential data using the appropriate definition
func TransformCredentials(credentialType string, data map[string]interface{}) (map[string]interface{}, error) {
	def := FindCredentialDefinition(credentialType)
	if def == nil {
		return nil, fmt.Errorf("unknown credential type: %s", credentialType)
	}
	return def.Transform(data)
}

// TestCredentials tests credential data using the appropriate definition
func TestCredentials(credentialType string, data map[string]interface{}) error {
	def := FindCredentialDefinition(credentialType)
	if def == nil {
		return fmt.Errorf("unknown credential type: %s", credentialType)
	}
	return def.Test(data)
}
