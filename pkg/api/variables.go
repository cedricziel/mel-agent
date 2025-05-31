package api

import (
	"context"
	"fmt"
	"sync"
)

// VariableScope defines the scope of a variable
type VariableScope string

const (
	// GlobalScope - variables shared across all workflows and runs
	GlobalScope VariableScope = "global"
	// WorkflowScope - variables shared across all runs of a workflow/agent
	WorkflowScope VariableScope = "workflow"
	// RunScope - variables specific to a single run
	RunScope VariableScope = "run"
)

// VariableStore manages variables at different scopes
type VariableStore interface {
	// Set stores a variable at the specified scope
	Set(ctx context.Context, scope VariableScope, key string, value interface{}) error

	// Get retrieves a variable from the specified scope
	Get(ctx context.Context, scope VariableScope, key string) (interface{}, bool, error)

	// Delete removes a variable from the specified scope
	Delete(ctx context.Context, scope VariableScope, key string) error

	// List returns all variables in the specified scope
	List(ctx context.Context, scope VariableScope) (map[string]interface{}, error)

	// Clear removes all variables from the specified scope
	Clear(ctx context.Context, scope VariableScope) error
}

// MemoryVariableStore is an in-memory implementation of VariableStore
type MemoryVariableStore struct {
	mu           sync.RWMutex
	globalVars   map[string]interface{}            // global variables
	workflowVars map[string]map[string]interface{} // agentID -> variables
	runVars      map[string]map[string]interface{} // runID -> variables
}

// NewMemoryVariableStore creates a new in-memory variable store
func NewMemoryVariableStore() *MemoryVariableStore {
	return &MemoryVariableStore{
		globalVars:   make(map[string]interface{}),
		workflowVars: make(map[string]map[string]interface{}),
		runVars:      make(map[string]map[string]interface{}),
	}
}

// Set implements VariableStore.Set
func (m *MemoryVariableStore) Set(ctx context.Context, scope VariableScope, key string, value interface{}) error {
	if key == "" {
		return fmt.Errorf("variable key cannot be empty")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	switch scope {
	case GlobalScope:
		m.globalVars[key] = value
	case WorkflowScope, RunScope:
		scopeKey := m.getScopeKey(ctx, scope)
		if scopeKey == "" {
			return fmt.Errorf("invalid scope context for %s scope", scope)
		}

		var targetMap map[string]map[string]interface{}
		if scope == WorkflowScope {
			targetMap = m.workflowVars
		} else {
			targetMap = m.runVars
		}

		if targetMap[scopeKey] == nil {
			targetMap[scopeKey] = make(map[string]interface{})
		}
		targetMap[scopeKey][key] = value
	default:
		return fmt.Errorf("unknown variable scope: %s", scope)
	}

	return nil
}

// Get implements VariableStore.Get
func (m *MemoryVariableStore) Get(ctx context.Context, scope VariableScope, key string) (interface{}, bool, error) {
	if key == "" {
		return nil, false, fmt.Errorf("variable key cannot be empty")
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	switch scope {
	case GlobalScope:
		value, exists := m.globalVars[key]
		return value, exists, nil
	case WorkflowScope, RunScope:
		scopeKey := m.getScopeKey(ctx, scope)
		if scopeKey == "" {
			return nil, false, fmt.Errorf("invalid scope context for %s scope", scope)
		}

		var targetMap map[string]map[string]interface{}
		if scope == WorkflowScope {
			targetMap = m.workflowVars
		} else {
			targetMap = m.runVars
		}

		vars, exists := targetMap[scopeKey]
		if !exists {
			return nil, false, nil
		}

		value, exists := vars[key]
		return value, exists, nil
	default:
		return nil, false, fmt.Errorf("unknown variable scope: %s", scope)
	}
}

// Delete implements VariableStore.Delete
func (m *MemoryVariableStore) Delete(ctx context.Context, scope VariableScope, key string) error {
	if key == "" {
		return fmt.Errorf("variable key cannot be empty")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	switch scope {
	case GlobalScope:
		delete(m.globalVars, key)
	case WorkflowScope, RunScope:
		scopeKey := m.getScopeKey(ctx, scope)
		if scopeKey == "" {
			return fmt.Errorf("invalid scope context for %s scope", scope)
		}

		var targetMap map[string]map[string]interface{}
		if scope == WorkflowScope {
			targetMap = m.workflowVars
		} else {
			targetMap = m.runVars
		}

		if vars, exists := targetMap[scopeKey]; exists {
			delete(vars, key)
		}
	default:
		return fmt.Errorf("unknown variable scope: %s", scope)
	}

	return nil
}

// List implements VariableStore.List
func (m *MemoryVariableStore) List(ctx context.Context, scope VariableScope) (map[string]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	switch scope {
	case GlobalScope:
		// Return a copy to avoid external modifications
		result := make(map[string]interface{})
		for k, v := range m.globalVars {
			result[k] = v
		}
		return result, nil
	case WorkflowScope, RunScope:
		scopeKey := m.getScopeKey(ctx, scope)
		if scopeKey == "" {
			return nil, fmt.Errorf("invalid scope context for %s scope", scope)
		}

		var targetMap map[string]map[string]interface{}
		if scope == WorkflowScope {
			targetMap = m.workflowVars
		} else {
			targetMap = m.runVars
		}

		vars, exists := targetMap[scopeKey]
		if !exists {
			return make(map[string]interface{}), nil
		}

		// Return a copy to avoid external modifications
		result := make(map[string]interface{})
		for k, v := range vars {
			result[k] = v
		}
		return result, nil
	default:
		return nil, fmt.Errorf("unknown variable scope: %s", scope)
	}
}

// Clear implements VariableStore.Clear
func (m *MemoryVariableStore) Clear(ctx context.Context, scope VariableScope) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	switch scope {
	case GlobalScope:
		m.globalVars = make(map[string]interface{})
	case WorkflowScope, RunScope:
		scopeKey := m.getScopeKey(ctx, scope)
		if scopeKey == "" {
			return fmt.Errorf("invalid scope context for %s scope", scope)
		}

		var targetMap map[string]map[string]interface{}
		if scope == WorkflowScope {
			targetMap = m.workflowVars
		} else {
			targetMap = m.runVars
		}

		delete(targetMap, scopeKey)
	default:
		return fmt.Errorf("unknown variable scope: %s", scope)
	}
	return nil
}

// getScopeKey generates the appropriate key for the given scope and context
func (m *MemoryVariableStore) getScopeKey(ctx context.Context, scope VariableScope) string {
	switch scope {
	case WorkflowScope:
		if agentID := ctx.Value("agentID"); agentID != nil {
			return agentID.(string)
		}
	case RunScope:
		if runID := ctx.Value("runID"); runID != nil {
			return runID.(string)
		}
	}
	return ""
}

// Global variable store instance
var globalVariableStore VariableStore = NewMemoryVariableStore()

// SetVariableStore sets the global variable store
func SetVariableStore(store VariableStore) {
	globalVariableStore = store
}

// GetVariableStore returns the global variable store
func GetVariableStore() VariableStore {
	return globalVariableStore
}

// Helper functions for easier variable access

// SetVariable sets a variable in the given context and scope
func SetVariable(ctx context.Context, scope VariableScope, key string, value interface{}) error {
	return globalVariableStore.Set(ctx, scope, key, value)
}

// GetVariable gets a variable from the given context and scope
func GetVariable(ctx context.Context, scope VariableScope, key string) (interface{}, bool, error) {
	return globalVariableStore.Get(ctx, scope, key)
}

// DeleteVariable deletes a variable from the given context and scope
func DeleteVariable(ctx context.Context, scope VariableScope, key string) error {
	return globalVariableStore.Delete(ctx, scope, key)
}

// ListVariables lists all variables in the given context and scope
func ListVariables(ctx context.Context, scope VariableScope) (map[string]interface{}, error) {
	return globalVariableStore.List(ctx, scope)
}

// CreateVariableContext creates a context with variable-related values
func CreateVariableContext(agentID, runID, nodeID string) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "agentID", agentID)
	if runID != "" {
		ctx = context.WithValue(ctx, "runID", runID)
	}
	if nodeID != "" {
		ctx = context.WithValue(ctx, "nodeID", nodeID)
	}
	return ctx
}
