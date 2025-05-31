package core

import (
	"encoding/json"
	"fmt"

	"github.com/cedricziel/mel-agent/pkg/api"
)

// PayloadConverter defines the interface for serializing and deserializing envelopes
type PayloadConverter interface {
	Marshal(envelope *api.Envelope[interface{}]) ([]byte, error)
	Unmarshal(data []byte) (*api.Envelope[interface{}], error)
	ContentType() string
}

// JSONConverter implements PayloadConverter using JSON serialization
type JSONConverter struct{}

// NewJSONConverter creates a new JSON payload converter
func NewJSONConverter() *JSONConverter {
	return &JSONConverter{}
}

// Marshal serializes an envelope to JSON bytes
func (c *JSONConverter) Marshal(envelope *api.Envelope[interface{}]) ([]byte, error) {
	if envelope == nil {
		return nil, fmt.Errorf("envelope is nil")
	}
	
	return json.Marshal(envelope)
}

// Unmarshal deserializes JSON bytes to an envelope
func (c *JSONConverter) Unmarshal(data []byte) (*api.Envelope[interface{}], error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("data is empty")
	}
	
	var envelope api.Envelope[interface{}]
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, fmt.Errorf("failed to unmarshal envelope: %w", err)
	}
	
	return &envelope, nil
}

// ContentType returns the MIME type for JSON
func (c *JSONConverter) ContentType() string {
	return "application/json"
}

// CompactJSONConverter implements PayloadConverter using compact JSON (no indentation)
type CompactJSONConverter struct{}

// NewCompactJSONConverter creates a new compact JSON payload converter
func NewCompactJSONConverter() *CompactJSONConverter {
	return &CompactJSONConverter{}
}

// Marshal serializes an envelope to compact JSON bytes
func (c *CompactJSONConverter) Marshal(envelope *api.Envelope[interface{}]) ([]byte, error) {
	if envelope == nil {
		return nil, fmt.Errorf("envelope is nil")
	}
	
	return json.Marshal(envelope)
}

// Unmarshal deserializes JSON bytes to an envelope
func (c *CompactJSONConverter) Unmarshal(data []byte) (*api.Envelope[interface{}], error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("data is empty")
	}
	
	var envelope api.Envelope[interface{}]
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, fmt.Errorf("failed to unmarshal envelope: %w", err)
	}
	
	return &envelope, nil
}

// ContentType returns the MIME type for JSON
func (c *CompactJSONConverter) ContentType() string {
	return "application/json"
}

// PrettyJSONConverter implements PayloadConverter using pretty-printed JSON
type PrettyJSONConverter struct{}

// NewPrettyJSONConverter creates a new pretty JSON payload converter
func NewPrettyJSONConverter() *PrettyJSONConverter {
	return &PrettyJSONConverter{}
}

// Marshal serializes an envelope to pretty-printed JSON bytes
func (c *PrettyJSONConverter) Marshal(envelope *api.Envelope[interface{}]) ([]byte, error) {
	if envelope == nil {
		return nil, fmt.Errorf("envelope is nil")
	}
	
	return json.MarshalIndent(envelope, "", "  ")
}

// Unmarshal deserializes JSON bytes to an envelope
func (c *PrettyJSONConverter) Unmarshal(data []byte) (*api.Envelope[interface{}], error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("data is empty")
	}
	
	var envelope api.Envelope[interface{}]
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, fmt.Errorf("failed to unmarshal envelope: %w", err)
	}
	
	return &envelope, nil
}

// ContentType returns the MIME type for JSON
func (c *PrettyJSONConverter) ContentType() string {
	return "application/json"
}

// ConverterRegistry manages different payload converters
type ConverterRegistry struct {
	converters map[string]PayloadConverter
	default_   PayloadConverter
}

// NewConverterRegistry creates a new converter registry with common converters
func NewConverterRegistry() *ConverterRegistry {
	registry := &ConverterRegistry{
		converters: make(map[string]PayloadConverter),
		default_:   NewJSONConverter(),
	}
	
	// Register built-in converters
	registry.Register("json", NewJSONConverter())
	registry.Register("compact", NewCompactJSONConverter())
	registry.Register("pretty", NewPrettyJSONConverter())
	
	return registry
}

// Register adds a converter to the registry
func (r *ConverterRegistry) Register(name string, converter PayloadConverter) {
	r.converters[name] = converter
}

// Get retrieves a converter by name
func (r *ConverterRegistry) Get(name string) PayloadConverter {
	if converter, exists := r.converters[name]; exists {
		return converter
	}
	return r.default_
}

// GetByContentType retrieves a converter by content type
func (r *ConverterRegistry) GetByContentType(contentType string) PayloadConverter {
	for _, converter := range r.converters {
		if converter.ContentType() == contentType {
			return converter
		}
	}
	return r.default_
}

// SetDefault sets the default converter
func (r *ConverterRegistry) SetDefault(converter PayloadConverter) {
	r.default_ = converter
}

// List returns all registered converter names
func (r *ConverterRegistry) List() []string {
	var names []string
	for name := range r.converters {
		names = append(names, name)
	}
	return names
}

// Global converter registry
var globalRegistry = NewConverterRegistry()

// GetConverter retrieves a converter by name from the global registry
func GetConverter(name string) PayloadConverter {
	return globalRegistry.Get(name)
}

// GetConverterByContentType retrieves a converter by content type from the global registry
func GetConverterByContentType(contentType string) PayloadConverter {
	return globalRegistry.GetByContentType(contentType)
}

// RegisterConverter adds a converter to the global registry
func RegisterConverter(name string, converter PayloadConverter) {
	globalRegistry.Register(name, converter)
}

// SetDefaultConverter sets the default converter in the global registry
func SetDefaultConverter(converter PayloadConverter) {
	globalRegistry.SetDefault(converter)
}

// ListConverters returns all registered converter names from the global registry
func ListConverters() []string {
	return globalRegistry.List()
}

// MarshalEnvelope serializes an envelope using the specified converter
func MarshalEnvelope(envelope *api.Envelope[interface{}], converterName string) ([]byte, error) {
	converter := GetConverter(converterName)
	return converter.Marshal(envelope)
}

// UnmarshalEnvelope deserializes data to an envelope using the specified converter
func UnmarshalEnvelope(data []byte, converterName string) (*api.Envelope[interface{}], error) {
	converter := GetConverter(converterName)
	return converter.Unmarshal(data)
}

// MarshalEnvelopeDefault serializes an envelope using the default converter
func MarshalEnvelopeDefault(envelope *api.Envelope[interface{}]) ([]byte, error) {
	return globalRegistry.default_.Marshal(envelope)
}

// UnmarshalEnvelopeDefault deserializes data to an envelope using the default converter
func UnmarshalEnvelopeDefault(data []byte) (*api.Envelope[interface{}], error) {
	return globalRegistry.default_.Unmarshal(data)
}