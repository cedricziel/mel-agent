package httprequest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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

// ExecuteEnvelope performs HTTP request using envelopes
func (d httpRequestDefinition) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[interface{}]) (*api.Envelope[interface{}], error) {
	// Perform existing logic
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

	// Prepare request body
	var bodyReader io.Reader
	if body != "" && method != "GET" && method != "HEAD" {
		bodyReader = strings.NewReader(body)
	}

	// Prepare headers
	headers := make(map[string]string)
	
	// Set Content-Type header if we have a body
	if bodyReader != nil {
		headers["Content-Type"] = contentType
	}

	// Add custom headers
	if headersVal, exists := config["headers"]; exists {
		if headersMap, ok := headersVal.(map[string]interface{}); ok {
			for key, value := range headersMap {
				if strVal, ok := value.(string); ok {
					headers[key] = strVal
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
			headers["Authorization"] = "Bearer " + authValue
		}
	case "basic":
		if authValue != "" {
			headers["Authorization"] = "Basic " + authValue
		}
	case "apikey":
		if authValue != "" {
			authHeader, _ := config["authHeader"].(string)
			if authHeader == "" {
				authHeader = "Authorization"
			}
			headers[authHeader] = authValue
		}
	}

	// Use platform HTTP client
	httpReq := api.HTTPRequest{
		Method:  method,
		URL:     url,
		Headers: headers,
		Body:    bodyReader,
		Timeout: time.Duration(timeout) * time.Second,
	}

	httpResp, err := ctx.Mel.HTTPRequest(context.Background(), httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// Try to parse response as JSON, fallback to string
	var responseData interface{}
	if err := json.Unmarshal(httpResp.Body, &responseData); err != nil {
		// Not valid JSON, return as string
		responseData = string(httpResp.Body)
	}

	// Return structured response
	resultData := map[string]interface{}{
		"status":     httpResp.StatusCode,
		"statusText": fmt.Sprintf("%d", httpResp.StatusCode),
		"headers":    httpResp.Headers,
		"data":       responseData,
		"url":        url,
		"method":     method,
		"duration":   httpResp.Duration.Milliseconds(),
	}

	// Create result envelope
	result := envelope.Clone()
	result.Trace = envelope.Trace.Next(node.ID)
	result.Data = resultData
	return result, nil
}

func (httpRequestDefinition) Initialize(mel api.Mel) error {
	return nil
}

func init() {
	api.RegisterNodeDefinition(httpRequestDefinition{})
}

// assert that httpRequestDefinition implements both interfaces
var _ api.NodeDefinition = (*httpRequestDefinition)(nil)
