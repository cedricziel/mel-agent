package credentials

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/cedricziel/mel-agent/pkg/api"
)

type baserowJWTCredential struct{}

func (baserowJWTCredential) Type() string {
	return "baserow_jwt"
}

func (baserowJWTCredential) Name() string {
	return "Baserow"
}

func (baserowJWTCredential) Description() string {
	return "Baserow authentication using email and password (generates JWT token)"
}

func (baserowJWTCredential) Parameters() []api.ParameterDefinition {
	return []api.ParameterDefinition{
		api.NewStringParameter("baseUrl", "Base URL", true).
			WithDescription("Your Baserow instance URL").
			WithDefault("https://api.baserow.io").
			WithValidators(api.ValidatorSpec{
				Type: "url",
			}),
		api.NewStringParameter("username", "Email", true).
			WithDescription("Your Baserow email address").
			WithValidators(api.ValidatorSpec{
				Type: "notEmpty",
			}),
		api.NewStringParameter("password", "Password", true).
			WithDescription("Your Baserow password").
			WithValidators(api.ValidatorSpec{
				Type: "notEmpty",
			}),
	}
}

func (baserowJWTCredential) Validate(data map[string]interface{}) error {
	baseURL, ok := data["baseUrl"].(string)
	if !ok || baseURL == "" {
		return fmt.Errorf("baseUrl is required and must be a non-empty string")
	}

	username, ok := data["username"].(string)
	if !ok || username == "" {
		return fmt.Errorf("username is required and must be a non-empty string")
	}

	password, ok := data["password"].(string)
	if !ok || password == "" {
		return fmt.Errorf("password is required and must be a non-empty string")
	}

	return nil
}

func (baserowJWTCredential) Transform(data map[string]interface{}) (map[string]interface{}, error) {
	baseURL := data["baseUrl"].(string)
	username := data["username"].(string)
	password := data["password"].(string)

	// Test authentication but don't store the token (it will expire)
	_, err := authenticateWithBaserow(baseURL, username, password)
	if err != nil {
		return nil, fmt.Errorf("authentication test failed: %w", err)
	}

	// Store credentials for fresh authentication each time
	return map[string]interface{}{
		"baseUrl":  baseURL,
		"username": username,
		"password": password,
		"authType": "jwt", // Mark this as JWT auth for the client
	}, nil
}

// JWTResponse represents a JWT authentication response from Baserow
type JWTResponse struct {
	Token string `json:"token"`
}

// authenticateWithBaserow performs JWT authentication with Baserow
func authenticateWithBaserow(baseURL, username, password string) (string, error) {
	authPayload := map[string]string{
		"email":    username, // Baserow uses "email" not "username"
		"password": password,
	}

	jsonBody, err := json.Marshal(authPayload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal auth payload: %w", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("POST", baseURL+"/api/user/token-auth/", bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create auth request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("auth request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("authentication failed with status %d: %s", resp.StatusCode, string(body))
	}

	var jwtResp struct {
		AccessToken string `json:"access_token"` // /api/user/token-auth/ now returns "access_token"
	}
	if err := json.NewDecoder(resp.Body).Decode(&jwtResp); err != nil {
		return "", fmt.Errorf("failed to decode JWT response: %w", err)
	}

	return jwtResp.AccessToken, nil
}

func (baserowJWTCredential) Test(data map[string]interface{}) error {
	baseURL := data["baseUrl"].(string)
	username := data["username"].(string)
	password := data["password"].(string)

	// Test by attempting to authenticate
	_, err := authenticateWithBaserow(baseURL, username, password)
	if err != nil {
		return fmt.Errorf("authentication test failed: %w", err)
	}

	return nil
}

func init() {
	api.RegisterCredentialDefinition(baserowJWTCredential{})
}
