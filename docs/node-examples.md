# Node Implementation Examples

This document provides practical examples of different types of nodes you can build for the MEL Agent platform. Each example demonstrates different patterns and capabilities.

## Table of Contents

1. [Simple Utility Node](#simple-utility-node)
2. [HTTP API Integration Node](#http-api-integration-node)
3. [Data Transformation Node](#data-transformation-node)
4. [Conditional Logic Node](#conditional-logic-node)
5. [External Service Integration](#external-service-integration)
6. [Workflow Communication Node](#workflow-communication-node)
7. [File Processing Node](#file-processing-node)
8. [Database Integration Node](#database-integration-node)

## Simple Utility Node

A basic text processing node that demonstrates fundamental concepts:

```go
package text_processor

import (
    "strings"
    "github.com/cedricziel/mel-agent/pkg/api"
)

type textProcessorDefinition struct{}

func (textProcessorDefinition) Meta() api.NodeType {
    return api.NodeType{
        Type:     "text_processor",
        Label:    "Text Processor",
        Category: "Utility",
        Parameters: []api.ParameterDefinition{
            api.NewStringParameter("text", "Text", true).
                WithDescription("The text to process").
                WithGroup("Input"),
            api.NewEnumParameter("operation", "Operation", 
                []string{"uppercase", "lowercase", "trim", "reverse"}, true).
                WithDefault("uppercase").
                WithGroup("Settings"),
            api.NewBooleanParameter("preserve_spaces", "Preserve Spaces", false).
                WithDefault(true).
                WithGroup("Settings"),
        },
    }
}

func (d textProcessorDefinition) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[interface{}]) (*api.Envelope[interface{}], error) {
    // Extract parameters
    text, ok := node.Data["text"].(string)
    if !ok {
        return nil, api.NewNodeError(node.ID, "text_processor", "text parameter is required")
    }
    
    operation, _ := node.Data["operation"].(string)
    if operation == "" {
        operation = "uppercase"
    }
    
    preserveSpaces, _ := node.Data["preserve_spaces"].(bool)
    
    // Process text based on operation
    var result string
    switch operation {
    case "uppercase":
        result = strings.ToUpper(text)
    case "lowercase":
        result = strings.ToLower(text)
    case "trim":
        result = strings.TrimSpace(text)
    case "reverse":
        runes := []rune(text)
        for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
            runes[i], runes[j] = runes[j], runes[i]
        }
        result = string(runes)
    default:
        return nil, api.NewNodeError(node.ID, "text_processor", "unknown operation: "+operation)
    }
    
    // Handle space preservation
    if !preserveSpaces && operation != "trim" {
        result = strings.ReplaceAll(result, " ", "")
    }
    
    // Create result envelope
    outputEnvelope := envelope.Clone()
    outputEnvelope.Trace = envelope.Trace.Next(node.ID)
    outputEnvelope.Data = map[string]interface{}{
        "processed_text": result,
        "original_text":  text,
        "operation":      operation,
        "length":         len(result),
    }
    
    return outputEnvelope, nil
}

func (textProcessorDefinition) Initialize(mel api.Mel) error {
    return nil
}

func init() {
    api.RegisterNodeDefinition(textProcessorDefinition{})
}

var _ api.NodeDefinition = (*textProcessorDefinition)(nil)
```

## HTTP API Integration Node

A node that calls external APIs with authentication and error handling:

```go
package api_caller

import (
    "context"
    "encoding/json"
    "fmt"
    "strings"
    "time"
    
    "github.com/cedricziel/mel-agent/pkg/api"
)

type apiCallerDefinition struct{}

func (apiCallerDefinition) Meta() api.NodeType {
    return api.NodeType{
        Type:     "api_caller",
        Label:    "API Caller",
        Category: "Integration",
        Parameters: []api.ParameterDefinition{
            api.NewStringParameter("url", "API URL", true).
                WithDescription("The API endpoint to call").
                WithGroup("Request").
                WithValidators(api.ValidatorSpec{Type: "url"}),
            api.NewEnumParameter("method", "HTTP Method", 
                []string{"GET", "POST", "PUT", "DELETE", "PATCH"}, true).
                WithDefault("GET").
                WithGroup("Request"),
            api.NewObjectParameter("headers", "Headers", false).
                WithDescription("HTTP headers as JSON object").
                WithGroup("Request"),
            api.NewObjectParameter("body", "Request Body", false).
                WithDescription("Request body as JSON object").
                WithGroup("Request").
                WithVisibilityCondition("method != 'GET'"),
            api.NewNumberParameter("timeout", "Timeout (seconds)", false).
                WithDefault(30).
                WithGroup("Settings"),
            api.NewCredentialParameter("auth", "Authentication", "api_key", false).
                WithGroup("Authentication"),
            api.NewBooleanParameter("retry_on_failure", "Retry on Failure", false).
                WithDefault(true).
                WithGroup("Settings"),
            api.NewIntegerParameter("max_retries", "Max Retries", false).
                WithDefault(3).
                WithGroup("Settings").
                WithVisibilityCondition("retry_on_failure == true"),
        },
    }
}

func (d apiCallerDefinition) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[interface{}]) (*api.Envelope[interface{}], error) {
    url, ok := node.Data["url"].(string)
    if !ok || url == "" {
        return nil, api.NewNodeError(node.ID, "api_caller", "url parameter is required")
    }
    
    method, _ := node.Data["method"].(string)
    if method == "" {
        method = "GET"
    }
    
    timeout, _ := node.Data["timeout"].(float64)
    if timeout <= 0 {
        timeout = 30
    }
    
    retryOnFailure, _ := node.Data["retry_on_failure"].(bool)
    maxRetries, _ := node.Data["max_retries"].(float64)
    if maxRetries <= 0 {
        maxRetries = 3
    }
    
    // Build headers
    headers := make(map[string]string)
    headers["Content-Type"] = "application/json"
    
    if headerData, ok := node.Data["headers"].(map[string]interface{}); ok {
        for k, v := range headerData {
            if str, ok := v.(string); ok {
                headers[k] = str
            }
        }
    }
    
    // Add authentication if provided
    if authID, ok := node.Data["auth"].(string); ok && authID != "" {
        // In a real implementation, you'd resolve the credential
        headers["Authorization"] = "Bearer " + authID
    }
    
    // Build request body
    var bodyReader *strings.Reader
    if bodyData, ok := node.Data["body"].(map[string]interface{}); ok && method != "GET" {
        bodyBytes, err := json.Marshal(bodyData)
        if err != nil {
            return nil, api.NewNodeError(node.ID, "api_caller", "failed to marshal request body: "+err.Error())
        }
        bodyReader = strings.NewReader(string(bodyBytes))
    }
    
    // Execute request with retry logic
    var response *api.HTTPResponse
    var err error
    
    for attempt := 0; attempt <= int(maxRetries); attempt++ {
        if bodyReader != nil {
            bodyReader.Seek(0, 0) // Reset reader for retries
        }
        
        httpReq := api.HTTPRequest{
            Method:  method,
            URL:     url,
            Headers: headers,
            Body:    bodyReader,
            Timeout: time.Duration(timeout) * time.Second,
        }
        
        response, err = ctx.Mel.HTTPRequest(context.Background(), httpReq)
        
        if err == nil && response.StatusCode < 500 {
            break // Success or client error (don't retry)
        }
        
        if !retryOnFailure || attempt == int(maxRetries) {
            break // No retry or max attempts reached
        }
        
        // Wait before retry with exponential backoff
        time.Sleep(time.Duration(attempt+1) * time.Second)
    }
    
    if err != nil {
        return nil, api.NewNodeError(node.ID, "api_caller", "HTTP request failed: "+err.Error())
    }
    
    // Parse response body as JSON if possible
    var responseData interface{}
    if len(response.Body) > 0 {
        if err := json.Unmarshal(response.Body, &responseData); err != nil {
            // If JSON parsing fails, use raw string
            responseData = string(response.Body)
        }
    }
    
    // Create result envelope
    result := envelope.Clone()
    result.Trace = envelope.Trace.Next(node.ID)
    result.Data = map[string]interface{}{
        "status_code": response.StatusCode,
        "headers":     response.Headers,
        "body":        responseData,
        "url":         url,
        "method":      method,
        "success":     response.StatusCode >= 200 && response.StatusCode < 300,
    }
    
    return result, nil
}

func (apiCallerDefinition) Initialize(mel api.Mel) error {
    return nil
}

func init() {
    api.RegisterNodeDefinition(apiCallerDefinition{})
}

var _ api.NodeDefinition = (*apiCallerDefinition)(nil)
```

## Data Transformation Node

A node that transforms data structures using JSONPath and templates:

```go
package data_transformer

import (
    "encoding/json"
    "fmt"
    "regexp"
    "strconv"
    "strings"
    
    "github.com/cedricziel/mel-agent/pkg/api"
)

type dataTransformerDefinition struct{}

func (dataTransformerDefinition) Meta() api.NodeType {
    return api.NodeType{
        Type:     "data_transformer",
        Label:    "Data Transformer",
        Category: "Data",
        Parameters: []api.ParameterDefinition{
            api.NewObjectParameter("mapping", "Field Mapping", true).
                WithDescription("JSON object defining output field mappings").
                WithGroup("Transformation"),
            api.NewEnumParameter("source", "Data Source", 
                []string{"envelope", "variables", "parameters"}, true).
                WithDefault("envelope").
                WithGroup("Input"),
            api.NewBooleanParameter("preserve_original", "Preserve Original Data", false).
                WithDefault(false).
                WithGroup("Output"),
            api.NewStringParameter("output_format", "Output Format", false).
                WithDefault("json").
                WithGroup("Output"),
        },
    }
}

func (d dataTransformerDefinition) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[interface{}]) (*api.Envelope[interface{}], error) {
    mapping, ok := node.Data["mapping"].(map[string]interface{})
    if !ok {
        return nil, api.NewNodeError(node.ID, "data_transformer", "mapping parameter is required")
    }
    
    source, _ := node.Data["source"].(string)
    if source == "" {
        source = "envelope"
    }
    
    preserveOriginal, _ := node.Data["preserve_original"].(bool)
    
    // Get source data
    var sourceData map[string]interface{}
    switch source {
    case "envelope":
        if envData, ok := envelope.Data.(map[string]interface{}); ok {
            sourceData = envData
        } else {
            sourceData = map[string]interface{}{"data": envelope.Data}
        }
    case "variables":
        sourceData = envelope.Variables
    case "parameters":
        sourceData = node.Data
    default:
        return nil, api.NewNodeError(node.ID, "data_transformer", "invalid source: "+source)
    }
    
    if sourceData == nil {
        sourceData = make(map[string]interface{})
    }
    
    // Transform data according to mapping
    transformed := make(map[string]interface{})
    
    for outputField, rule := range mapping {
        value, err := d.extractValue(sourceData, rule)
        if err != nil {
            return nil, api.NewNodeError(node.ID, "data_transformer", 
                fmt.Sprintf("error transforming field %s: %v", outputField, err))
        }
        transformed[outputField] = value
    }
    
    // Build result data
    var resultData interface{}
    if preserveOriginal {
        resultData = map[string]interface{}{
            "original":    sourceData,
            "transformed": transformed,
        }
    } else {
        resultData = transformed
    }
    
    result := envelope.Clone()
    result.Trace = envelope.Trace.Next(node.ID)
    result.Data = resultData
    
    return result, nil
}

func (d dataTransformerDefinition) extractValue(data map[string]interface{}, rule interface{}) (interface{}, error) {
    switch r := rule.(type) {
    case string:
        // Simple field reference or JSONPath
        return d.extractByPath(data, r)
    case map[string]interface{}:
        // Complex transformation rule
        if template, ok := r["template"].(string); ok {
            return d.applyTemplate(data, template)
        }
        if path, ok := r["path"].(string); ok {
            value, err := d.extractByPath(data, path)
            if err != nil {
                return nil, err
            }
            return d.applyTransformations(value, r)
        }
        return r, nil // Return as literal object
    default:
        return r, nil // Return as literal value
    }
}

func (d dataTransformerDefinition) extractByPath(data map[string]interface{}, path string) (interface{}, error) {
    if path == "." {
        return data, nil
    }
    
    // Simple dot notation (e.g., "user.name")
    parts := strings.Split(path, ".")
    current := interface{}(data)
    
    for _, part := range parts {
        if part == "" {
            continue
        }
        
        switch v := current.(type) {
        case map[string]interface{}:
            current = v[part]
        default:
            return nil, fmt.Errorf("cannot access field %s in non-object", part)
        }
        
        if current == nil {
            return nil, nil
        }
    }
    
    return current, nil
}

func (d dataTransformerDefinition) applyTemplate(data map[string]interface{}, template string) (interface{}, error) {
    // Simple template substitution using {{field}} syntax
    re := regexp.MustCompile(`\{\{([^}]+)\}\}`)
    
    result := re.ReplaceAllStringFunc(template, func(match string) string {
        field := strings.Trim(match, "{}")
        field = strings.TrimSpace(field)
        
        if value, err := d.extractByPath(data, field); err == nil && value != nil {
            return fmt.Sprintf("%v", value)
        }
        return match // Leave unchanged if field not found
    })
    
    return result, nil
}

func (d dataTransformerDefinition) applyTransformations(value interface{}, rule map[string]interface{}) (interface{}, error) {
    // Apply various transformations
    if cast, ok := rule["cast"].(string); ok {
        return d.castValue(value, cast)
    }
    
    if defaultVal, ok := rule["default"]; ok && value == nil {
        return defaultVal, nil
    }
    
    return value, nil
}

func (d dataTransformerDefinition) castValue(value interface{}, targetType string) (interface{}, error) {
    if value == nil {
        return nil, nil
    }
    
    switch targetType {
    case "string":
        return fmt.Sprintf("%v", value), nil
    case "number":
        if str, ok := value.(string); ok {
            return strconv.ParseFloat(str, 64)
        }
        return value, nil
    case "boolean":
        if str, ok := value.(string); ok {
            return strconv.ParseBool(str)
        }
        return value, nil
    default:
        return value, nil
    }
}

func (dataTransformerDefinition) Initialize(mel api.Mel) error {
    return nil
}

func init() {
    api.RegisterNodeDefinition(dataTransformerDefinition{})
}

var _ api.NodeDefinition = (*dataTransformerDefinition)(nil)
```

## Conditional Logic Node

A node that implements conditional logic with multiple branches:

```go
package conditional

import (
    "fmt"
    "reflect"
    "strconv"
    "strings"
    
    "github.com/cedricziel/mel-agent/pkg/api"
)

type conditionalDefinition struct{}

func (conditionalDefinition) Meta() api.NodeType {
    return api.NodeType{
        Type:      "conditional",
        Label:     "Conditional",
        Category:  "Control",
        Branching: true, // Indicates this node can have multiple outputs
        Parameters: []api.ParameterDefinition{
            api.NewArrayParameter("conditions", "Conditions", true).
                WithDescription("Array of condition objects").
                WithGroup("Logic"),
            api.NewStringParameter("default_branch", "Default Branch", false).
                WithDescription("Branch to take if no conditions match").
                WithDefault("else").
                WithGroup("Logic"),
            api.NewBooleanParameter("evaluate_all", "Evaluate All Conditions", false).
                WithDescription("Continue evaluating after first match").
                WithDefault(false).
                WithGroup("Behavior"),
        },
    }
}

type Condition struct {
    Field    string      `json:"field"`
    Operator string      `json:"operator"`
    Value    interface{} `json:"value"`
    Branch   string      `json:"branch"`
}

func (d conditionalDefinition) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[interface{}]) (*api.Envelope[interface{}], error) {
    conditionsData, ok := node.Data["conditions"].([]interface{})
    if !ok {
        return nil, api.NewNodeError(node.ID, "conditional", "conditions parameter is required")
    }
    
    defaultBranch, _ := node.Data["default_branch"].(string)
    if defaultBranch == "" {
        defaultBranch = "else"
    }
    
    evaluateAll, _ := node.Data["evaluate_all"].(bool)
    
    // Parse conditions
    var conditions []Condition
    for i, condData := range conditionsData {
        condMap, ok := condData.(map[string]interface{})
        if !ok {
            return nil, api.NewNodeError(node.ID, "conditional", 
                fmt.Sprintf("condition %d must be an object", i))
        }
        
        field, _ := condMap["field"].(string)
        operator, _ := condMap["operator"].(string)
        value := condMap["value"]
        branch, _ := condMap["branch"].(string)
        
        if field == "" || operator == "" || branch == "" {
            return nil, api.NewNodeError(node.ID, "conditional", 
                fmt.Sprintf("condition %d missing required fields", i))
        }
        
        conditions = append(conditions, Condition{
            Field:    field,
            Operator: operator,
            Value:    value,
            Branch:   branch,
        })
    }
    
    // Get data for evaluation
    var evalData map[string]interface{}
    if envData, ok := envelope.Data.(map[string]interface{}); ok {
        evalData = envData
    } else {
        evalData = map[string]interface{}{"data": envelope.Data}
    }
    
    // Add variables to evaluation context
    if envelope.Variables != nil {
        for k, v := range envelope.Variables {
            evalData["var_"+k] = v
        }
    }
    
    // Evaluate conditions
    var matchedBranches []string
    var evaluationResults []map[string]interface{}
    
    for _, condition := range conditions {
        matched, err := d.evaluateCondition(evalData, condition)
        if err != nil {
            return nil, api.NewNodeError(node.ID, "conditional", 
                fmt.Sprintf("error evaluating condition for field %s: %v", condition.Field, err))
        }
        
        evaluationResults = append(evaluationResults, map[string]interface{}{
            "field":    condition.Field,
            "operator": condition.Operator,
            "value":    condition.Value,
            "branch":   condition.Branch,
            "matched":  matched,
        })
        
        if matched {
            matchedBranches = append(matchedBranches, condition.Branch)
            if !evaluateAll {
                break // Stop on first match
            }
        }
    }
    
    // Determine which branch to take
    var selectedBranch string
    if len(matchedBranches) > 0 {
        selectedBranch = matchedBranches[0] // Take first match
    } else {
        selectedBranch = defaultBranch
    }
    
    result := envelope.Clone()
    result.Trace = envelope.Trace.Next(node.ID)
    result.Data = map[string]interface{}{
        "branch":              selectedBranch,
        "matched_branches":    matchedBranches,
        "evaluation_results":  evaluationResults,
        "original_data":       envelope.Data,
    }
    
    return result, nil
}

func (d conditionalDefinition) evaluateCondition(data map[string]interface{}, condition Condition) (bool, error) {
    // Extract field value
    fieldValue := d.getFieldValue(data, condition.Field)
    
    // Evaluate based on operator
    switch condition.Operator {
    case "eq", "equals":
        return d.compareValues(fieldValue, condition.Value, "eq"), nil
    case "ne", "not_equals":
        return !d.compareValues(fieldValue, condition.Value, "eq"), nil
    case "gt", "greater_than":
        return d.compareValues(fieldValue, condition.Value, "gt"), nil
    case "gte", "greater_than_or_equal":
        return d.compareValues(fieldValue, condition.Value, "gte"), nil
    case "lt", "less_than":
        return d.compareValues(fieldValue, condition.Value, "lt"), nil
    case "lte", "less_than_or_equal":
        return d.compareValues(fieldValue, condition.Value, "lte"), nil
    case "contains":
        return d.stringContains(fieldValue, condition.Value), nil
    case "starts_with":
        return d.stringStartsWith(fieldValue, condition.Value), nil
    case "ends_with":
        return d.stringEndsWith(fieldValue, condition.Value), nil
    case "in":
        return d.valueInList(fieldValue, condition.Value), nil
    case "exists":
        return fieldValue != nil, nil
    case "is_empty":
        return d.isEmpty(fieldValue), nil
    default:
        return false, fmt.Errorf("unknown operator: %s", condition.Operator)
    }
}

func (d conditionalDefinition) getFieldValue(data map[string]interface{}, field string) interface{} {
    // Support dot notation for nested fields
    parts := strings.Split(field, ".")
    current := interface{}(data)
    
    for _, part := range parts {
        if part == "" {
            continue
        }
        
        if m, ok := current.(map[string]interface{}); ok {
            current = m[part]
        } else {
            return nil
        }
    }
    
    return current
}

func (d conditionalDefinition) compareValues(a, b interface{}, op string) bool {
    // Try numeric comparison first
    if aNum, aOk := d.toNumber(a); aOk {
        if bNum, bOk := d.toNumber(b); bOk {
            switch op {
            case "eq":
                return aNum == bNum
            case "gt":
                return aNum > bNum
            case "gte":
                return aNum >= bNum
            case "lt":
                return aNum < bNum
            case "lte":
                return aNum <= bNum
            }
        }
    }
    
    // Fall back to string comparison
    aStr := fmt.Sprintf("%v", a)
    bStr := fmt.Sprintf("%v", b)
    
    switch op {
    case "eq":
        return aStr == bStr
    case "gt":
        return aStr > bStr
    case "gte":
        return aStr >= bStr
    case "lt":
        return aStr < bStr
    case "lte":
        return aStr <= bStr
    default:
        return false
    }
}

func (d conditionalDefinition) toNumber(v interface{}) (float64, bool) {
    switch val := v.(type) {
    case float64:
        return val, true
    case float32:
        return float64(val), true
    case int:
        return float64(val), true
    case int64:
        return float64(val), true
    case string:
        if num, err := strconv.ParseFloat(val, 64); err == nil {
            return num, true
        }
    }
    return 0, false
}

func (d conditionalDefinition) stringContains(haystack, needle interface{}) bool {
    h := fmt.Sprintf("%v", haystack)
    n := fmt.Sprintf("%v", needle)
    return strings.Contains(h, n)
}

func (d conditionalDefinition) stringStartsWith(str, prefix interface{}) bool {
    s := fmt.Sprintf("%v", str)
    p := fmt.Sprintf("%v", prefix)
    return strings.HasPrefix(s, p)
}

func (d conditionalDefinition) stringEndsWith(str, suffix interface{}) bool {
    s := fmt.Sprintf("%v", str)
    suf := fmt.Sprintf("%v", suffix)
    return strings.HasSuffix(s, suf)
}

func (d conditionalDefinition) valueInList(value, list interface{}) bool {
    listSlice, ok := list.([]interface{})
    if !ok {
        return false
    }
    
    for _, item := range listSlice {
        if reflect.DeepEqual(value, item) {
            return true
        }
    }
    return false
}

func (d conditionalDefinition) isEmpty(value interface{}) bool {
    if value == nil {
        return true
    }
    
    switch v := value.(type) {
    case string:
        return strings.TrimSpace(v) == ""
    case []interface{}:
        return len(v) == 0
    case map[string]interface{}:
        return len(v) == 0
    default:
        return false
    }
}

func (conditionalDefinition) Initialize(mel api.Mel) error {
    return nil
}

func init() {
    api.RegisterNodeDefinition(conditionalDefinition{})
}

var _ api.NodeDefinition = (*conditionalDefinition)(nil)
```

## External Service Integration

A node that integrates with a third-party service (example: Slack):

```go
package slack_integration

import (
    "context"
    "encoding/json"
    "fmt"
    "strings"
    "time"
    
    "github.com/cedricziel/mel-agent/pkg/api"
)

type slackIntegrationDefinition struct{}

func (slackIntegrationDefinition) Meta() api.NodeType {
    return api.NodeType{
        Type:     "slack_integration",
        Label:    "Slack Integration",
        Category: "Communication",
        Parameters: []api.ParameterDefinition{
            api.NewCredentialParameter("webhook_url", "Slack Webhook URL", "slack_webhook", true).
                WithDescription("Slack incoming webhook URL").
                WithGroup("Authentication"),
            api.NewStringParameter("channel", "Channel", false).
                WithDescription("Channel to post to (overrides webhook default)").
                WithGroup("Message"),
            api.NewStringParameter("username", "Username", false).
                WithDescription("Bot username for the message").
                WithGroup("Message"),
            api.NewStringParameter("text", "Message Text", true).
                WithDescription("The message content").
                WithGroup("Message"),
            api.NewStringParameter("emoji", "Emoji", false).
                WithDescription("Emoji for bot avatar (e.g., :robot_face:)").
                WithGroup("Message"),
            api.NewBooleanParameter("unfurl_links", "Unfurl Links", false).
                WithDefault(true).
                WithGroup("Formatting"),
            api.NewBooleanParameter("unfurl_media", "Unfurl Media", false).
                WithDefault(true).
                WithGroup("Formatting"),
            api.NewArrayParameter("attachments", "Attachments", false).
                WithDescription("Rich message attachments").
                WithGroup("Advanced"),
        },
    }
}

type SlackMessage struct {
    Text        string                   `json:"text"`
    Channel     string                   `json:"channel,omitempty"`
    Username    string                   `json:"username,omitempty"`
    IconEmoji   string                   `json:"icon_emoji,omitempty"`
    UnfurlLinks bool                     `json:"unfurl_links"`
    UnfurlMedia bool                     `json:"unfurl_media"`
    Attachments []map[string]interface{} `json:"attachments,omitempty"`
}

func (d slackIntegrationDefinition) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[interface{}]) (*api.Envelope[interface{}], error) {
    webhookURL, ok := node.Data["webhook_url"].(string)
    if !ok || webhookURL == "" {
        return nil, api.NewNodeError(node.ID, "slack_integration", "webhook_url parameter is required")
    }
    
    text, ok := node.Data["text"].(string)
    if !ok || text == "" {
        return nil, api.NewNodeError(node.ID, "slack_integration", "text parameter is required")
    }
    
    // Build Slack message
    message := SlackMessage{
        Text:        d.interpolateText(text, envelope),
        UnfurlLinks: true,
        UnfurlMedia: true,
    }
    
    if channel, ok := node.Data["channel"].(string); ok && channel != "" {
        message.Channel = channel
    }
    
    if username, ok := node.Data["username"].(string); ok && username != "" {
        message.Username = username
    }
    
    if emoji, ok := node.Data["emoji"].(string); ok && emoji != "" {
        if !strings.HasPrefix(emoji, ":") {
            emoji = ":" + emoji
        }
        if !strings.HasSuffix(emoji, ":") {
            emoji = emoji + ":"
        }
        message.IconEmoji = emoji
    }
    
    if unfurlLinks, ok := node.Data["unfurl_links"].(bool); ok {
        message.UnfurlLinks = unfurlLinks
    }
    
    if unfurlMedia, ok := node.Data["unfurl_media"].(bool); ok {
        message.UnfurlMedia = unfurlMedia
    }
    
    // Handle attachments
    if attachmentsData, ok := node.Data["attachments"].([]interface{}); ok {
        attachments := make([]map[string]interface{}, 0, len(attachmentsData))
        for _, attachmentData := range attachmentsData {
            if attachment, ok := attachmentData.(map[string]interface{}); ok {
                attachments = append(attachments, attachment)
            }
        }
        message.Attachments = attachments
    }
    
    // Send message to Slack
    response, err := d.sendSlackMessage(ctx, webhookURL, message)
    if err != nil {
        return nil, api.NewNodeError(node.ID, "slack_integration", "failed to send Slack message: "+err.Error())
    }
    
    result := envelope.Clone()
    result.Trace = envelope.Trace.Next(node.ID)
    result.Data = map[string]interface{}{
        "sent":         true,
        "message":      message,
        "response":     response,
        "timestamp":    time.Now().Unix(),
        "channel":      message.Channel,
    }
    
    return result, nil
}

func (d slackIntegrationDefinition) interpolateText(text string, envelope *api.Envelope[interface{}]) string {
    // Simple variable substitution
    result := text
    
    // Replace data fields
    if data, ok := envelope.Data.(map[string]interface{}); ok {
        for key, value := range data {
            placeholder := fmt.Sprintf("{{%s}}", key)
            replacement := fmt.Sprintf("%v", value)
            result = strings.ReplaceAll(result, placeholder, replacement)
        }
    }
    
    // Replace variables
    if envelope.Variables != nil {
        for key, value := range envelope.Variables {
            placeholder := fmt.Sprintf("{{var.%s}}", key)
            replacement := fmt.Sprintf("%v", value)
            result = strings.ReplaceAll(result, placeholder, replacement)
        }
    }
    
    return result
}

func (d slackIntegrationDefinition) sendSlackMessage(ctx api.ExecutionContext, webhookURL string, message SlackMessage) (map[string]interface{}, error) {
    // Marshal message to JSON
    messageBytes, err := json.Marshal(message)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal message: %w", err)
    }
    
    // Send HTTP request
    httpReq := api.HTTPRequest{
        Method: "POST",
        URL:    webhookURL,
        Headers: map[string]string{
            "Content-Type": "application/json",
        },
        Body:    strings.NewReader(string(messageBytes)),
        Timeout: 30 * time.Second,
    }
    
    response, err := ctx.Mel.HTTPRequest(context.Background(), httpReq)
    if err != nil {
        return nil, fmt.Errorf("HTTP request failed: %w", err)
    }
    
    if response.StatusCode != 200 {
        return nil, fmt.Errorf("Slack returned status %d: %s", response.StatusCode, string(response.Body))
    }
    
    return map[string]interface{}{
        "status_code": response.StatusCode,
        "body":        string(response.Body),
    }, nil
}

func (slackIntegrationDefinition) Initialize(mel api.Mel) error {
    // Could validate Slack API connectivity here
    return nil
}

func init() {
    api.RegisterNodeDefinition(slackIntegrationDefinition{})
}

var _ api.NodeDefinition = (*slackIntegrationDefinition)(nil)
```

## Summary

These examples demonstrate:

1. **Parameter handling** - Different types, validation, grouping
2. **Envelope processing** - Cloning, tracing, data transformation
3. **Error handling** - Descriptive errors with context
4. **External integrations** - HTTP calls, authentication, retries
5. **Data manipulation** - JSON processing, templating, type conversion
6. **Control flow** - Conditional logic, branching
7. **Platform utilities** - Using the Mel interface for HTTP requests

Each example follows the established patterns and can serve as a starting point for your own custom nodes. Remember to:

- Always clone input envelopes
- Update the trace correctly
- Handle errors gracefully
- Validate input parameters
- Use the `api.NewNodeError()` helpers
- Register your nodes in the `init()` function
- Include compile-time interface checks

For more complex scenarios, you can combine these patterns and add additional features like caching, background processing, or advanced data validation.