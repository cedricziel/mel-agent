package api

import "sync"

// Mel is the interface for the MEL (MEL Agent) API, providing methods to manage node definitions and types.
type Mel interface {
	// RegisterNodeDefinition registers a node definition.
	RegisterNodeDefinition(def NodeDefinition)
	// ListNodeDefinitions returns all registered node definitions.
	ListNodeDefinitions() []NodeDefinition
	// FindDefinition retrieves the NodeDefinition for a given type.
	FindDefinition(typ string) NodeDefinition
	// AllCoreDefinitions returns the built-in core trigger and utility node definitions.
	AllCoreDefinitions() []NodeDefinition
	// ListNodeTypes returns all registered node type metadata.
	ListNodeTypes() []NodeType
}

// melImpl is the concrete implementation of the Mel interface.
type melImpl struct {
	mu          sync.RWMutex
	definitions []NodeDefinition
}

// NewMel creates a new Mel instance.
func NewMel() Mel {
	return &melImpl{
		definitions: make([]NodeDefinition, 0),
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
		result = append(result, def.Meta())
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
