package credentials

import (
	"fmt"

	"github.com/cedricziel/mel-agent/pkg/api"
)

type apiKeyCredential struct{}

func (apiKeyCredential) Type() string {
	return "api_key"
}

func (apiKeyCredential) Name() string {
	return "API Key"
}

func (apiKeyCredential) Description() string {
	return "Simple API key authentication"
}

func (apiKeyCredential) Parameters() []api.ParameterDefinition {
	return []api.ParameterDefinition{
		api.NewStringParameter("api_key", "API Key", true).
			WithDescription("Your API key for authentication").
			WithValidators(api.ValidatorSpec{
				Type: "notEmpty",
			}),
	}
}

func (apiKeyCredential) Validate(data map[string]interface{}) error {
	apiKey, ok := data["api_key"].(string)
	if !ok || apiKey == "" {
		return fmt.Errorf("api_key is required and must be a non-empty string")
	}
	return nil
}

func (apiKeyCredential) Transform(data map[string]interface{}) (map[string]interface{}, error) {
	// No transformation needed for API keys
	return data, nil
}

func (apiKeyCredential) Test(data map[string]interface{}) error {
	// Basic validation - API keys generally can't be tested without knowing the specific service
	apiKey, ok := data["api_key"].(string)
	if !ok || apiKey == "" {
		return fmt.Errorf("api_key is required and must be a non-empty string")
	}
	
	// For generic API keys, we can't test without knowing the endpoint
	// This would be overridden by specific API key implementations
	return nil
}

func init() {
	api.RegisterCredentialDefinition(apiKeyCredential{})
}