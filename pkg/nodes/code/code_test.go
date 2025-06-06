package code

import (
	"testing"
	"time"

	"github.com/cedricziel/mel-agent/pkg/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCodeDefinition_Meta(t *testing.T) {
	def := codeDefinition{}
	meta := def.Meta()

	assert.Equal(t, "code", meta.Type)
	assert.Equal(t, "Code", meta.Label)
	assert.Equal(t, "Code", meta.Category)

	// Check parameters
	require.Len(t, meta.Parameters, 4)

	// Language parameter
	langParam := meta.Parameters[0]
	assert.Equal(t, "language", langParam.Name)
	assert.Equal(t, "Language", langParam.Label)
	assert.Equal(t, string(api.TypeEnum), langParam.Type)
	assert.Equal(t, "javascript", langParam.Default)
	assert.Contains(t, langParam.Options, "javascript")
	assert.Contains(t, langParam.Options, "python")
	assert.Contains(t, langParam.Options, "typescript")

	// Code parameter
	codeParam := meta.Parameters[1]
	assert.Equal(t, "code", codeParam.Name)
	assert.Equal(t, "Code", codeParam.Label)
	assert.Equal(t, string(api.TypeString), codeParam.Type)
	assert.True(t, codeParam.Required)
	assert.Equal(t, "code", codeParam.JSONSchema.Format)

	// Timeout parameter
	timeoutParam := meta.Parameters[2]
	assert.Equal(t, "timeout", timeoutParam.Name)
	assert.Equal(t, "Timeout (seconds)", timeoutParam.Label)
	assert.Equal(t, string(api.TypeNumber), timeoutParam.Type)
	assert.False(t, timeoutParam.Required)
	assert.Equal(t, 30, timeoutParam.Default)

	// Strict mode parameter
	strictParam := meta.Parameters[3]
	assert.Equal(t, "strict_mode", strictParam.Name)
	assert.Equal(t, "Strict Mode", strictParam.Label)
	assert.Equal(t, string(api.TypeBoolean), strictParam.Type)
	assert.False(t, strictParam.Required)
	assert.Equal(t, true, strictParam.Default)
}

func TestCodeDefinition_ExecuteEnvelope_RequiredParameters(t *testing.T) {
	def := &codeDefinition{
		runtimes: map[string]Runtime{
			"javascript": &mockRuntime{language: "javascript"},
		},
	}

	ctx := api.ExecutionContext{
		AgentID: "test-agent",
		RunID:   "test-run",
	}

	envelope := &api.Envelope[interface{}]{
		ID:       "test-envelope",
		IssuedAt: time.Now(),
		Version:  1,
		Data:     map[string]interface{}{"input": "test"},
		Trace:    newTestTrace(),
	}

	t.Run("missing code parameter", func(t *testing.T) {
		node := api.Node{
			ID:   "test-node",
			Type: "code",
			Data: map[string]interface{}{
				"language": "javascript",
			},
		}

		result, err := def.ExecuteEnvelope(ctx, node, envelope)
		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "code parameter is required")
	})

	t.Run("empty code parameter", func(t *testing.T) {
		node := api.Node{
			ID:   "test-node",
			Type: "code",
			Data: map[string]interface{}{
				"language": "javascript",
				"code":     "",
			},
		}

		result, err := def.ExecuteEnvelope(ctx, node, envelope)
		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "code parameter is required")
	})
}

func TestCodeDefinition_ExecuteEnvelope_UnsupportedLanguage(t *testing.T) {
	def := &codeDefinition{
		runtimes: map[string]Runtime{
			"javascript": &mockRuntime{language: "javascript"},
		},
	}

	ctx := api.ExecutionContext{
		AgentID: "test-agent",
		RunID:   "test-run",
	}

	envelope := &api.Envelope[interface{}]{
		ID:       "test-envelope",
		IssuedAt: time.Now(),
		Version:  1,
		Data:     map[string]interface{}{"input": "test"},
		Trace:    newTestTrace(),
	}

	node := api.Node{
		ID:   "test-node",
		Type: "code",
		Data: map[string]interface{}{
			"language": "unsupported",
			"code":     "console.log('test');",
		},
	}

	result, err := def.ExecuteEnvelope(ctx, node, envelope)
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported language: unsupported")
}

func TestCodeDefinition_ExecuteEnvelope_DefaultLanguage(t *testing.T) {
	mockRT := &mockRuntime{
		language: "javascript",
		result:   map[string]interface{}{"success": true},
	}

	def := &codeDefinition{
		runtimes: map[string]Runtime{
			"javascript": mockRT,
		},
	}

	ctx := api.ExecutionContext{
		AgentID: "test-agent",
		RunID:   "test-run",
	}

	envelope := &api.Envelope[interface{}]{
		ID:       "test-envelope",
		IssuedAt: time.Now(),
		Version:  1,
		Data:     map[string]interface{}{"input": "test"},
		Trace:    newTestTrace(),
	}

	node := api.Node{
		ID:   "test-node",
		Type: "code",
		Data: map[string]interface{}{
			// No language specified - should default to javascript
			"code": "return {success: true};",
		},
	}

	result, err := def.ExecuteEnvelope(ctx, node, envelope)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEqual(t, envelope.ID, result.ID)
	assert.Equal(t, envelope.Version, result.Version)
	assert.Equal(t, mockRT.result, result.Data)

	// Verify runtime was called with correct context
	assert.Equal(t, "return {success: true};", mockRT.lastCode)
	assert.Equal(t, envelope.Data, mockRT.lastContext.Data)
	assert.Equal(t, envelope.Variables, mockRT.lastContext.Variables)
	assert.Equal(t, node.Data, mockRT.lastContext.NodeData)
	assert.Equal(t, node.ID, mockRT.lastContext.NodeID)
	assert.Equal(t, ctx.AgentID, mockRT.lastContext.AgentID)
}

func TestCodeDefinition_ExecuteEnvelope_SuccessfulExecution(t *testing.T) {
	mockRT := &mockRuntime{
		language: "javascript",
		result:   map[string]interface{}{"processed": true, "value": 42},
	}

	def := &codeDefinition{
		runtimes: map[string]Runtime{
			"javascript": mockRT,
		},
	}

	ctx := api.ExecutionContext{
		AgentID: "test-agent",
		RunID:   "test-run",
	}

	envelope := &api.Envelope[interface{}]{
		ID:        "test-envelope",
		IssuedAt:  time.Now(),
		Version:   1,
		Data:      map[string]interface{}{"input": "test"},
		Variables: map[string]interface{}{"var1": "value1"},
		Trace:     newTestTrace(),
	}

	node := api.Node{
		ID:   "test-node",
		Type: "code",
		Data: map[string]interface{}{
			"language": "javascript",
			"code":     "const result = input.data; return {processed: true, value: 42};",
			"timeout":  60.0,
		},
	}

	result, err := def.ExecuteEnvelope(ctx, node, envelope)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Check envelope properties (Clone creates new ID)
	assert.NotEqual(t, envelope.ID, result.ID)
	assert.Equal(t, envelope.Version, result.Version)
	assert.Equal(t, mockRT.result, result.Data)

	// Check trace was updated
	assert.NotEqual(t, envelope.Trace, result.Trace)

	// Verify runtime was called correctly
	assert.Equal(t, "const result = input.data; return {processed: true, value: 42};", mockRT.lastCode)
	assert.Equal(t, envelope.Data, mockRT.lastContext.Data)
	assert.Equal(t, envelope.Variables, mockRT.lastContext.Variables)
	assert.Equal(t, node.Data, mockRT.lastContext.NodeData)
	assert.Equal(t, node.ID, mockRT.lastContext.NodeID)
	assert.Equal(t, ctx.AgentID, mockRT.lastContext.AgentID)
}

func TestCodeDefinition_ExecuteEnvelope_RuntimeError(t *testing.T) {
	mockRT := &mockRuntime{
		language: "javascript",
		err:      assert.AnError,
	}

	def := &codeDefinition{
		runtimes: map[string]Runtime{
			"javascript": mockRT,
		},
	}

	ctx := api.ExecutionContext{
		AgentID: "test-agent",
		RunID:   "test-run",
	}

	envelope := &api.Envelope[interface{}]{
		ID:       "test-envelope",
		IssuedAt: time.Now(),
		Version:  1,
		Data:     map[string]interface{}{"input": "test"},
		Trace:    newTestTrace(),
	}

	node := api.Node{
		ID:   "test-node",
		Type: "code",
		Data: map[string]interface{}{
			"language": "javascript",
			"code":     "invalid code;",
		},
	}

	result, err := def.ExecuteEnvelope(ctx, node, envelope)
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "execution error:")
}

func TestCodeDefinition_ExecuteEnvelope_Timeout(t *testing.T) {
	// This test would need a runtime that can simulate long execution
	// For now, we'll test the timeout logic with a mock that sleeps
	mockRT := &mockRuntime{
		language:     "javascript",
		simulateDelay: 2 * time.Second, // Longer than timeout
		result:       map[string]interface{}{"result": "slow"},
	}

	def := &codeDefinition{
		runtimes: map[string]Runtime{
			"javascript": mockRT,
		},
	}

	ctx := api.ExecutionContext{
		AgentID: "test-agent",
		RunID:   "test-run",
	}

	envelope := &api.Envelope[interface{}]{
		ID:       "test-envelope",
		IssuedAt: time.Now(),
		Version:  1,
		Data:     map[string]interface{}{"input": "test"},
		Trace:    newTestTrace(),
	}

	node := api.Node{
		ID:   "test-node",
		Type: "code",
		Data: map[string]interface{}{
			"language": "javascript",
			"code":     "while(true) { /* infinite loop */ }",
			"timeout":  1.0, // 1 second timeout
		},
	}

	result, err := def.ExecuteEnvelope(ctx, node, envelope)
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "execution timeout exceeded")
}

func TestCodeDefinition_Initialize(t *testing.T) {
	def := &codeDefinition{}

	err := def.Initialize(nil)
	assert.NoError(t, err)
	
	// Verify JavaScript runtime was initialized
	jsRuntime, exists := def.runtimes["javascript"]
	assert.True(t, exists)
	assert.Equal(t, "javascript", jsRuntime.GetLanguage())
}

// Mock runtime for testing
type mockRuntime struct {
	language      string
	result        interface{}
	err           error
	simulateDelay time.Duration
	initialized   bool

	// Capture last execution for verification
	lastCode    string
	lastContext CodeExecutionContext
}

func (m *mockRuntime) Execute(code string, context CodeExecutionContext) (interface{}, error) {
	m.lastCode = code
	m.lastContext = context

	if m.simulateDelay > 0 {
		time.Sleep(m.simulateDelay)
	}

	if m.err != nil {
		return nil, m.err
	}

	return m.result, nil
}

func (m *mockRuntime) GetLanguage() string {
	return m.language
}

func (m *mockRuntime) Initialize() error {
	m.initialized = true
	return nil
}

func (m *mockRuntime) Cleanup() error {
	return nil
}

// Helper function to create a test trace
func newTestTrace() api.Trace {
	return api.Trace{
		AgentID: "test-agent",
		RunID:   "test-run",
		NodeID:  "test-node",
		Step:    "test-step",
		Attempt: 1,
	}
}