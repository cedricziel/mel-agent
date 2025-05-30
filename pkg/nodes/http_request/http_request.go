package httprequest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cedricziel/mel-agent/pkg/api"
)

type httpRequestDefinition struct{}

func (httpRequestDefinition) Meta() api.NodeType {
	return api.NodeType{
		Type:     "http_request",
		Label:    "HTTP Request",
		Icon:     "🌐",
		Category: "Integration",
		Parameters: []api.ParameterDefinition{
			api.NewStringParameter("url", "URL", true).
				WithGroup("Request").
				WithDescription("Endpoint to call").
				WithValidators(api.ValidatorSpec{Type: "notEmpty"}, api.ValidatorSpec{Type: "url"}),
			api.NewEnumParameter("method", "Method", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}, true).
				WithDefault("GET").
				WithGroup("Request"),
			api.NewObjectParameter("headers", "Headers", false).
				WithDefault("{}").
				WithGroup("Request").
				WithDescription("Additional headers as JSON object").
				WithValidators(api.ValidatorSpec{Type: "json"}),
			api.NewStringParameter("body", "Body", false).
				WithGroup("Request").
				WithDescription("Request body (JSON, text, etc.)").
				WithVisibilityCondition("method!='GET' && method!='HEAD'"),
			api.NewStringParameter("contentType", "Content-Type", false).
				WithDefault("application/json").
				WithGroup("Request").
				WithDescription("Content-Type header").
				WithVisibilityCondition("method!='GET' && method!='HEAD'"),
			api.NewEnumParameter("authType", "Authentication", []string{"none", "bearer", "apikey", "basic"}, false).
				WithDefault("none").
				WithGroup("Authentication"),
			api.NewStringParameter("authValue", "Auth Value", false).
				WithGroup("Authentication").
				WithDescription("Token, API key, or username:password").
				WithVisibilityCondition("authType!='none'"),
			api.NewStringParameter("authHeader", "Auth Header", false).
				WithDefault("Authorization").
				WithGroup("Authentication").
				WithDescription("Header name for API key auth").
				WithVisibilityCondition("authType=='apikey'"),
			api.NewNumberParameter("timeout", "Timeout", false).
				WithDefault(30).
				WithGroup("Advanced").
				WithDescription("Timeout in seconds"),
			api.NewBooleanParameter("followRedirects", "Follow Redirects", false).
				WithDefault(true).
				WithGroup("Advanced"),
			api.NewBooleanParameter("verifySSL", "Verify SSL", false).
				WithDefault(true).
				WithGroup("Advanced"),
		},
	}
}

func (httpRequestDefinition) Execute(ctx api.ExecutionContext, node api.Node, input interface{}) (interface{}, error) {
	// Extract configuration from node data
	config := node.Data
	
	// Get required parameters
	url, ok := config["url"].(string)
	if !ok || url == "" {
		return nil, fmt.Errorf("url is required")
	}
	
	method, ok := config["method"].(string)
	if !ok {
		method = "GET"
	}
	
	// Get optional parameters
	body := ""
	if bodyVal, exists := config["body"]; exists {
		if bodyStr, ok := bodyVal.(string); ok {
			body = bodyStr
		}
	}
	
	contentType := "application/json"
	if ctVal, exists := config["contentType"]; exists {
		if ctStr, ok := ctVal.(string); ok && ctStr != "" {
			contentType = ctStr
		}
	}
	
	timeout := 30
	if timeoutVal, exists := config["timeout"]; exists {
		switch t := timeoutVal.(type) {
		case float64:
			timeout = int(t)
		case int:
			timeout = t
		case string:
			if parsed, err := strconv.Atoi(t); err == nil {
				timeout = parsed
			}
		}
	}
	
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}
	
	// Prepare request body
	var bodyReader io.Reader
	if body != "" && method != "GET" && method != "HEAD" {
		bodyReader = strings.NewReader(body)
	}
	
	// Create request
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Set Content-Type header if we have a body
	if bodyReader != nil {
		req.Header.Set("Content-Type", contentType)
	}
	
	// Add custom headers
	if headersVal, exists := config["headers"]; exists {
		if headersMap, ok := headersVal.(map[string]interface{}); ok {
			for key, value := range headersMap {
				if strVal, ok := value.(string); ok {
					req.Header.Set(key, strVal)
				}
			}
		}
	}
	
	// Handle authentication
	authType, _ := config["authType"].(string)
	authValue, _ := config["authValue"].(string)
	
	switch authType {
	case "bearer":
		if authValue != "" {
			req.Header.Set("Authorization", "Bearer "+authValue)
		}
	case "basic":
		if authValue != "" {
			req.Header.Set("Authorization", "Basic "+authValue)
		}
	case "apikey":
		if authValue != "" {
			authHeader, _ := config["authHeader"].(string)
			if authHeader == "" {
				authHeader = "Authorization"
			}
			req.Header.Set(authHeader, authValue)
		}
	}
	
	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	
	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	
	// Try to parse response as JSON, fallback to string
	var responseData interface{}
	if err := json.Unmarshal(respBody, &responseData); err != nil {
		// Not valid JSON, return as string
		responseData = string(respBody)
	}
	
	// Return structured response
	result := map[string]interface{}{
		"status":     resp.StatusCode,
		"statusText": resp.Status,
		"headers":    resp.Header,
		"data":       responseData,
		"url":        url,
		"method":     method,
	}
	
	return result, nil
}

func (httpRequestDefinition) Initialize(mel api.Mel) error {
	return nil
}

func init() {
	api.RegisterNodeDefinition(httpRequestDefinition{})
}

// assert that httpRequestDefinition implements the NodeDefinition interface
var _ api.NodeDefinition = (*httpRequestDefinition)(nil)
