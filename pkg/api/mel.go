package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// WorkflowCallRequest represents a request to call another workflow
type WorkflowCallRequest struct {
	TargetWorkflowID string                 `json:"target_workflow_id"`
	CallData         map[string]interface{} `json:"call_data"`
	CallMode         string                 `json:"call_mode"` // "async" or "sync"
	TimeoutSeconds   int                    `json:"timeout_seconds"`
	SourceContext    ExecutionContext       `json:"source_context"`
}

// WorkflowCallResponse represents the response from a workflow call
type WorkflowCallResponse struct {
	CallID      string                 `json:"call_id"`
	Status      string                 `json:"status"` // "success", "error", "timeout"
	Data        map[string]interface{} `json:"data"`
	Message     string                 `json:"message"`
	CompletedAt time.Time              `json:"completed_at"`
}

// HTTPRequest represents an HTTP request to be made by the platform
type HTTPRequest struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Body    io.Reader         `json:"-"`
	Timeout time.Duration     `json:"timeout"`
}

// HTTPResponse represents an HTTP response from the platform
type HTTPResponse struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Body       []byte            `json:"body"`
	Duration   time.Duration     `json:"duration"`
}

// Mel is the interface for the MEL (MEL Agent) API, providing methods to manage node definitions and platform utilities.
type Mel interface {
	// Node Definition Management
	RegisterNodeDefinition(def NodeDefinition)
	ListNodeDefinitions() []NodeDefinition
	FindDefinition(typ string) NodeDefinition
	AllCoreDefinitions() []NodeDefinition
	ListNodeTypes() []NodeType

	// Platform Utilities
	// HTTP client for making external requests
	HTTPRequest(ctx context.Context, req HTTPRequest) (*HTTPResponse, error)

	// Workflow management
	CallWorkflow(ctx context.Context, req WorkflowCallRequest) (*WorkflowCallResponse, error)
	ReturnToWorkflow(ctx context.Context, callID string, data map[string]interface{}, status string) error

	// Data storage for cross-workflow communication
	StoreData(ctx context.Context, key string, data interface{}, ttl time.Duration) error
	RetrieveData(ctx context.Context, key string) (interface{}, error)
	DeleteData(ctx context.Context, key string) error
}

// melImpl is the concrete implementation of the Mel interface.
type melImpl struct {
	mu               sync.RWMutex
	definitions      []NodeDefinition
	httpClient       *http.Client
	workflowEndpoint string
	dataStore        map[string]dataStoreEntry
	dataStoreMu      sync.RWMutex
	pendingCalls     map[string]pendingCallEntry
	pendingCallsMu   sync.RWMutex
}

// dataStoreEntry represents a stored data entry with TTL
type dataStoreEntry struct {
	Data      interface{}
	ExpiresAt time.Time
}

// pendingCallEntry tracks calls waiting for responses
type pendingCallEntry struct {
	CallID        string
	SourceContext ExecutionContext
	ResponseChan  chan WorkflowCallResponse
	ExpiresAt     time.Time
}

// NewMel creates a new Mel instance.
func NewMel() Mel {
	return &melImpl{
		definitions:      make([]NodeDefinition, 0),
		httpClient:       &http.Client{Timeout: 30 * time.Second},
		workflowEndpoint: "http://localhost:8080/api", // Default to local API
		dataStore:        make(map[string]dataStoreEntry),
		pendingCalls:     make(map[string]pendingCallEntry),
	}
}

// NewMelWithConfig creates a new Mel instance with custom configuration.
func NewMelWithConfig(httpTimeout time.Duration, workflowEndpoint string) Mel {
	return &melImpl{
		definitions:      make([]NodeDefinition, 0),
		httpClient:       &http.Client{Timeout: httpTimeout},
		workflowEndpoint: workflowEndpoint,
		dataStore:        make(map[string]dataStoreEntry),
		pendingCalls:     make(map[string]pendingCallEntry),
	}
}

// RegisterNodeDefinition registers a node definition.
func (m *melImpl) RegisterNodeDefinition(def NodeDefinition) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.definitions = append(m.definitions, def)
}

// ListNodeDefinitions returns all registered node definitions.
func (m *melImpl) ListNodeDefinitions() []NodeDefinition {
	m.mu.RLock()
	defer m.mu.RUnlock()
	// Return a copy to avoid race conditions
	result := make([]NodeDefinition, len(m.definitions))
	copy(result, m.definitions)
	return result
}

// FindDefinition retrieves the NodeDefinition for a given type.
func (m *melImpl) FindDefinition(typ string) NodeDefinition {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, def := range m.definitions {
		if def.Meta().Type == typ {
			return def
		}
	}
	return nil
}

// AllCoreDefinitions returns the built-in core trigger and utility node definitions.
func (m *melImpl) AllCoreDefinitions() []NodeDefinition {
	// TODO: Implement core definitions (webhook, schedule, etc.)
	return []NodeDefinition{}
}

// ListNodeTypes returns all registered node type metadata.
func (m *melImpl) ListNodeTypes() []NodeType {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]NodeType, 0, len(m.definitions))
	for _, def := range m.definitions {
		meta := def.Meta()
		meta.Kinds = GetNodeKinds(def)
		result = append(result, meta)
	}
	return result
}

// Global instance for backward compatibility
var globalMel = NewMel()

// RegisterNodeDefinition registers a node definition with the global instance.
func RegisterNodeDefinition(def NodeDefinition) {
	globalMel.RegisterNodeDefinition(def)
}

// ListNodeDefinitions returns all registered node definitions from the global instance.
func ListNodeDefinitions() []NodeDefinition {
	return globalMel.ListNodeDefinitions()
}

// FindDefinition retrieves the NodeDefinition for a given type from the global instance.
func FindDefinition(typ string) NodeDefinition {
	return globalMel.FindDefinition(typ)
}

// AllCoreDefinitions returns the built-in core trigger and utility node definitions from the global instance.
func AllCoreDefinitions() []NodeDefinition {
	return globalMel.AllCoreDefinitions()
}

// ListNodeTypes returns all registered node type metadata from the global instance.
func ListNodeTypes() []NodeType {
	return globalMel.ListNodeTypes()
}

// Platform utility method implementations

// HTTPRequest makes an HTTP request using the platform's HTTP client
func (m *melImpl) HTTPRequest(ctx context.Context, req HTTPRequest) (*HTTPResponse, error) {
	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, req.URL, req.Body)
	if err != nil {
		return nil, err
	}

	// Add headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Use custom timeout if specified
	client := m.httpClient
	if req.Timeout > 0 {
		client = &http.Client{Timeout: req.Timeout}
	}

	// Make the request and measure duration
	start := time.Now()
	resp, err := client.Do(httpReq)
	duration := time.Since(start)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Convert headers to map
	headers := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	return &HTTPResponse{
		StatusCode: resp.StatusCode,
		Headers:    headers,
		Body:       body,
		Duration:   duration,
	}, nil
}

// CallWorkflow calls another workflow via the API
func (m *melImpl) CallWorkflow(ctx context.Context, req WorkflowCallRequest) (*WorkflowCallResponse, error) {
	callID := fmt.Sprintf("call-%d", time.Now().UnixNano())

	// Prepare the payload for the target workflow
	payload := map[string]interface{}{
		"callId":           callID,
		"sourceWorkflowId": req.SourceContext.AgentID,
		"sourceRunId":      req.SourceContext.RunID,
		"sourceNodeId":     req.SourceContext.AgentID, // We don't have the node ID in the context yet
		"callData":         req.CallData,
		"callMode":         req.CallMode,
		"calledAt":         time.Now().Format(time.RFC3339),
	}

	if req.CallMode == "sync" {
		// For synchronous calls, register a pending call and wait for workflow_return

		// Step 1: Register pending call before triggering workflow
		responseChan := make(chan WorkflowCallResponse, 1)
		m.pendingCallsMu.Lock()
		m.pendingCalls[callID] = pendingCallEntry{
			CallID:        callID,
			SourceContext: req.SourceContext,
			ResponseChan:  responseChan,
			ExpiresAt:     time.Now().Add(time.Duration(req.TimeoutSeconds) * time.Second),
		}
		m.pendingCallsMu.Unlock()

		// Cleanup function to remove pending call if we exit early
		defer func() {
			m.pendingCallsMu.Lock()
			delete(m.pendingCalls, callID)
			close(responseChan)
			m.pendingCallsMu.Unlock()
		}()

		// Step 2: Trigger the workflow asynchronously
		triggerURL := fmt.Sprintf("%s/agents/%s/runs/test", m.workflowEndpoint, req.TargetWorkflowID)

		// Create the HTTP request to trigger the workflow
		httpReq := HTTPRequest{
			Method: "POST",
			URL:    triggerURL,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body:    strings.NewReader(marshalJSON(payload)),
			Timeout: 30 * time.Second, // Use shorter timeout for triggering
		}

		resp, err := m.HTTPRequest(ctx, httpReq)
		if err != nil {
			return nil, fmt.Errorf("failed to trigger workflow: %w", err)
		}

		if resp.StatusCode >= 400 {
			return nil, fmt.Errorf("workflow trigger failed with status %d: %s", resp.StatusCode, string(resp.Body))
		}

		// Step 3: Wait for workflow_return node to call ReturnToWorkflow
		timeoutDuration := time.Duration(req.TimeoutSeconds) * time.Second
		select {
		case response := <-responseChan:
			// Received response from workflow_return node
			return &response, nil

		case <-time.After(timeoutDuration):
			// Timeout waiting for response
			return nil, fmt.Errorf("timeout waiting for workflow response after %v", timeoutDuration)

		case <-ctx.Done():
			// Context cancelled
			return nil, ctx.Err()
		}

	} else {
		// For asynchronous calls, use the test endpoint for immediate execution
		// This is more reliable than the async runs endpoint for now
		triggerURL := fmt.Sprintf("%s/agents/%s/runs/test", m.workflowEndpoint, req.TargetWorkflowID)

		httpReq := HTTPRequest{
			Method: "POST",
			URL:    triggerURL,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body:    strings.NewReader(marshalJSON(payload)),
			Timeout: 30 * time.Second, // Shorter timeout for async triggers
		}

		resp, err := m.HTTPRequest(ctx, httpReq)
		if err != nil {
			return nil, fmt.Errorf("failed to trigger async workflow: %w", err)
		}

		if resp.StatusCode >= 400 {
			return nil, fmt.Errorf("async workflow trigger failed with status %d: %s", resp.StatusCode, string(resp.Body))
		}

		return &WorkflowCallResponse{
			CallID:      callID,
			Status:      "sent",
			Data:        map[string]interface{}{"message": "Async workflow triggered"},
			Message:     "Asynchronous workflow call sent",
			CompletedAt: time.Now(),
		}, nil
	}
}

// Helper function to marshal JSON without error handling for simplicity
func marshalJSON(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}

// ReturnToWorkflow returns data to a calling workflow
func (m *melImpl) ReturnToWorkflow(ctx context.Context, callID string, data map[string]interface{}, status string) error {
	m.pendingCallsMu.Lock()
	defer m.pendingCallsMu.Unlock()

	// Find the pending call
	pendingCall, exists := m.pendingCalls[callID]
	if !exists {
		// If no pending call found, store the response in case it's retrieved later
		return m.StoreData(ctx, "workflow_return:"+callID, map[string]interface{}{
			"data":       data,
			"status":     status,
			"returnedAt": time.Now().Format(time.RFC3339),
		}, 1*time.Hour) // Store for 1 hour
	}

	// Check if the call has expired
	if time.Now().After(pendingCall.ExpiresAt) {
		delete(m.pendingCalls, callID)
		return fmt.Errorf("workflow call %s has expired", callID)
	}

	// Send the response to the waiting workflow call
	response := WorkflowCallResponse{
		CallID:      callID,
		Status:      status,
		Data:        data,
		Message:     fmt.Sprintf("Workflow return received with status: %s", status),
		CompletedAt: time.Now(),
	}

	// Try to send the response without blocking
	select {
	case pendingCall.ResponseChan <- response:
		// Successfully sent response
		delete(m.pendingCalls, callID)
		return nil
	default:
		// Response channel is full or closed, remove the pending call
		delete(m.pendingCalls, callID)
		return fmt.Errorf("failed to deliver workflow return for call %s", callID)
	}
}

// StoreData stores data with TTL for cross-workflow communication
func (m *melImpl) StoreData(ctx context.Context, key string, data interface{}, ttl time.Duration) error {
	m.dataStoreMu.Lock()
	defer m.dataStoreMu.Unlock()

	expiresAt := time.Now().Add(ttl)
	if ttl <= 0 {
		expiresAt = time.Now().Add(24 * time.Hour) // Default 24h TTL
	}

	m.dataStore[key] = dataStoreEntry{
		Data:      data,
		ExpiresAt: expiresAt,
	}

	return nil
}

// RetrieveData retrieves stored data
func (m *melImpl) RetrieveData(ctx context.Context, key string) (interface{}, error) {
	m.dataStoreMu.RLock()
	defer m.dataStoreMu.RUnlock()

	entry, exists := m.dataStore[key]
	if !exists {
		return nil, fmt.Errorf("data not found for key: %s", key)
	}

	// Check if expired
	if time.Now().After(entry.ExpiresAt) {
		// Clean up expired entry (we do this in read to avoid needing a background cleanup goroutine)
		delete(m.dataStore, key)
		return nil, fmt.Errorf("data expired for key: %s", key)
	}

	return entry.Data, nil
}

// DeleteData deletes stored data
func (m *melImpl) DeleteData(ctx context.Context, key string) error {
	m.dataStoreMu.Lock()
	defer m.dataStoreMu.Unlock()

	delete(m.dataStore, key)
	return nil
}
