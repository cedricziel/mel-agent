package credentials

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/cedricziel/mel-agent/pkg/api"
)

type baserowTokenCredential struct{}

func (baserowTokenCredential) Type() string {
	return "baserow_token"
}

func (baserowTokenCredential) Name() string {
	return "Baserow (API Token)"
}

func (baserowTokenCredential) Description() string {
	return "Baserow authentication using an API token"
}

func (baserowTokenCredential) Parameters() []api.ParameterDefinition {
	return []api.ParameterDefinition{
		api.NewStringParameter("baseUrl", "Base URL", true).
			WithDescription("Your Baserow instance URL (e.g., https://api.baserow.io)").
			WithValidators(api.ValidatorSpec{
				Type: "url",
			}),
		api.NewStringParameter("token", "API Token", true).
			WithDescription("Your Baserow API token").
			WithValidators(api.ValidatorSpec{
				Type: "notEmpty",
			}),
	}
}

func (baserowTokenCredential) Validate(data map[string]interface{}) error {
	baseURL, ok := data["baseUrl"].(string)
	if !ok || baseURL == "" {
		return fmt.Errorf("baseUrl is required and must be a non-empty string")
	}
	
	token, ok := data["token"].(string)
	if !ok || token == "" {
		return fmt.Errorf("token is required and must be a non-empty string")
	}
	
	return nil
}

func (baserowTokenCredential) Transform(data map[string]interface{}) (map[string]interface{}, error) {
	// Add authType to distinguish from JWT
	result := make(map[string]interface{})
	for k, v := range data {
		result[k] = v
	}
	result["authType"] = "token"
	return result, nil
}

func (baserowTokenCredential) Test(data map[string]interface{}) error {
	baseURL := data["baseUrl"].(string)
	token := data["token"].(string)
	
	// Test by making a simple API call to list applications
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", baseURL+"/api/applications/", nil)
	if err != nil {
		return fmt.Errorf("failed to create test request: %w", err)
	}
	
	req.Header.Set("Authorization", "Token "+token)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("test request failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("authentication test failed with status %d: %s", resp.StatusCode, string(body))
	}
	
	return nil
}

func init() {
	api.RegisterCredentialDefinition(baserowTokenCredential{})
}