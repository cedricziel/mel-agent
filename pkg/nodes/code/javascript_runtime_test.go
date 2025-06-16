package code

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJavaScriptRuntime_GetLanguage(t *testing.T) {
	runtime := NewJavaScriptRuntime()
	assert.Equal(t, "javascript", runtime.GetLanguage())
}

func TestJavaScriptRuntime_Initialize(t *testing.T) {
	runtime := NewJavaScriptRuntime()
	err := runtime.Initialize()
	assert.NoError(t, err)
}

func TestJavaScriptRuntime_Cleanup(t *testing.T) {
	runtime := NewJavaScriptRuntime()
	err := runtime.Cleanup()
	assert.NoError(t, err)
}

func TestJavaScriptRuntime_Execute_BasicReturn(t *testing.T) {
	runtime := NewJavaScriptRuntime()
	require.NoError(t, runtime.Initialize())

	execContext := CodeExecutionContext{
		Data:      map[string]interface{}{"value": 42},
		Variables: map[string]interface{}{"name": "test"},
		NodeData:  map[string]interface{}{"setting": "enabled"},
		NodeID:    "node-123",
		AgentID:   "agent-456",
	}

	code := `return {processed: true, value: input.data.value * 2};`

	result, err := runtime.Execute(context.Background(), code, execContext)
	require.NoError(t, err)

	expected := map[string]interface{}{
		"processed": true,
		"value":     int64(84), // Goja exports as int64 for this operation
	}
	assert.Equal(t, expected, result)
}

func TestJavaScriptRuntime_Execute_AccessInputData(t *testing.T) {
	runtime := NewJavaScriptRuntime()
	require.NoError(t, runtime.Initialize())

	execContext := CodeExecutionContext{
		Data: map[string]interface{}{
			"user": map[string]interface{}{
				"name":  "John Doe",
				"email": "john@example.com",
				"age":   30,
			},
		},
		Variables: map[string]interface{}{},
		NodeData:  map[string]interface{}{},
		NodeID:    "node-123",
		AgentID:   "agent-456",
	}

	code := `
		const user = input.data.user;
		return {
			greeting: "Hello, " + user.name,
			isAdult: user.age >= 18,
			contact: user.email.toLowerCase()
		};
	`

	result, err := runtime.Execute(context.Background(), code, execContext)
	require.NoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)

	assert.Equal(t, "Hello, John Doe", resultMap["greeting"])
	assert.Equal(t, true, resultMap["isAdult"])
	assert.Equal(t, "john@example.com", resultMap["contact"])
}

func TestJavaScriptRuntime_Execute_AccessVariables(t *testing.T) {
	runtime := NewJavaScriptRuntime()
	require.NoError(t, runtime.Initialize())

	execContext := CodeExecutionContext{
		Data: map[string]interface{}{"input": "test"},
		Variables: map[string]interface{}{
			"apiKey":     "secret-key",
			"baseURL":    "https://api.example.com",
			"retryCount": 3,
		},
		NodeData: map[string]interface{}{},
		NodeID:   "node-123",
		AgentID:  "agent-456",
	}

	code := `
		return {
			url: input.variables.baseURL + "/data",
			auth: "Bearer " + input.variables.apiKey,
			config: {
				retries: input.variables.retryCount,
				hasKey: !!input.variables.apiKey
			}
		};
	`

	result, err := runtime.Execute(context.Background(), code, execContext)
	require.NoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)

	assert.Equal(t, "https://api.example.com/data", resultMap["url"])
	assert.Equal(t, "Bearer secret-key", resultMap["auth"])

	config, ok := resultMap["config"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, int64(3), config["retries"])
	assert.Equal(t, true, config["hasKey"])
}

func TestJavaScriptRuntime_Execute_AccessNodeData(t *testing.T) {
	runtime := NewJavaScriptRuntime()
	require.NoError(t, runtime.Initialize())

	execContext := CodeExecutionContext{
		Data:      map[string]interface{}{"input": "test"},
		Variables: map[string]interface{}{},
		NodeData: map[string]interface{}{
			"operation": "multiply",
			"factor":    5,
			"format":    "json",
		},
		NodeID:  "node-123",
		AgentID: "agent-456",
	}

	code := `
		const operation = input.nodeData.operation;
		const factor = input.nodeData.factor;
		
		return {
			operation: operation,
			factor: factor,
			nodeId: input.nodeId,
			agentId: input.agentId
		};
	`

	result, err := runtime.Execute(context.Background(), code, execContext)
	require.NoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)

	assert.Equal(t, "multiply", resultMap["operation"])
	assert.Equal(t, int64(5), resultMap["factor"])
	assert.Equal(t, "node-123", resultMap["nodeId"])
	assert.Equal(t, "agent-456", resultMap["agentId"])
}

func TestJavaScriptRuntime_Execute_UtilityFunctions(t *testing.T) {
	runtime := NewJavaScriptRuntime()
	require.NoError(t, runtime.Initialize())

	execContext := CodeExecutionContext{
		Data:      map[string]interface{}{"json": `{"name": "test", "value": 42}`},
		Variables: map[string]interface{}{},
		NodeData:  map[string]interface{}{},
		NodeID:    "node-123",
		AgentID:   "agent-456",
	}

	code := `
		const parsed = utils.parseJSON(input.data.json);
		const uuid = utils.generateUUID();
		const hash = utils.md5("test");
		
		return {
			parsed: parsed,
			uuidLength: uuid.length,
			hash: hash,
			stringified: utils.stringifyJSON({test: true})
		};
	`

	result, err := runtime.Execute(context.Background(), code, execContext)
	require.NoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)

	// Check parsed JSON
	parsed, ok := resultMap["parsed"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "test", parsed["name"])
	assert.Equal(t, 42.0, parsed["value"])

	// Check UUID (should be 36 characters with dashes)
	assert.Equal(t, int64(36), resultMap["uuidLength"])

	// Check MD5 hash (should be 32 hex characters)
	hash, ok := resultMap["hash"].(string)
	require.True(t, ok)
	assert.Len(t, hash, 32)

	// Check stringified JSON
	assert.Equal(t, `{"test":true}`, resultMap["stringified"])
}

func TestJavaScriptRuntime_Execute_Console(t *testing.T) {
	runtime := NewJavaScriptRuntime()
	require.NoError(t, runtime.Initialize())

	execContext := CodeExecutionContext{
		Data:      map[string]interface{}{"debug": true},
		Variables: map[string]interface{}{},
		NodeData:  map[string]interface{}{},
		NodeID:    "node-123",
		AgentID:   "agent-456",
	}

	code := `
		console.log("Starting execution");
		console.error("This is a test error");
		console.warn("This is a warning");
		
		return {success: true};
	`

	result, err := runtime.Execute(context.Background(), code, execContext)
	require.NoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, true, resultMap["success"])

	// Note: In a real implementation, we'd want to capture console output
	// For now, we just verify the code executes without error
}

func TestJavaScriptRuntime_Execute_SyntaxError(t *testing.T) {
	runtime := NewJavaScriptRuntime()
	require.NoError(t, runtime.Initialize())

	execContext := CodeExecutionContext{
		Data:      map[string]interface{}{},
		Variables: map[string]interface{}{},
		NodeData:  map[string]interface{}{},
		NodeID:    "node-123",
		AgentID:   "agent-456",
	}

	code := `invalid javascript syntax {{{`

	result, err := runtime.Execute(context.Background(), code, execContext)
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SyntaxError") // Should contain syntax error info
}

func TestJavaScriptRuntime_Execute_RuntimeError(t *testing.T) {
	runtime := NewJavaScriptRuntime()
	require.NoError(t, runtime.Initialize())

	execContext := CodeExecutionContext{
		Data:      map[string]interface{}{},
		Variables: map[string]interface{}{},
		NodeData:  map[string]interface{}{},
		NodeID:    "node-123",
		AgentID:   "agent-456",
	}

	code := `throw new Error("Custom runtime error");`

	result, err := runtime.Execute(context.Background(), code, execContext)
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Custom runtime error")
}

func TestJavaScriptRuntime_Execute_NoReturn(t *testing.T) {
	runtime := NewJavaScriptRuntime()
	require.NoError(t, runtime.Initialize())

	execContext := CodeExecutionContext{
		Data:      map[string]interface{}{"value": 42},
		Variables: map[string]interface{}{},
		NodeData:  map[string]interface{}{},
		NodeID:    "node-123",
		AgentID:   "agent-456",
	}

	code := `
		const x = input.data.value * 2;
		console.log("Calculated:", x);
		// No return statement
	`

	result, err := runtime.Execute(context.Background(), code, execContext)
	require.NoError(t, err)
	// Should return undefined when no explicit return
	assert.Nil(t, result)
}

func TestJavaScriptRuntime_Execute_ComplexDataTransformation(t *testing.T) {
	runtime := NewJavaScriptRuntime()
	require.NoError(t, runtime.Initialize())

	execContext := CodeExecutionContext{
		Data: map[string]interface{}{
			"users": []interface{}{
				map[string]interface{}{"name": "Alice", "age": 25, "active": true},
				map[string]interface{}{"name": "Bob", "age": 30, "active": false},
				map[string]interface{}{"name": "Charlie", "age": 35, "active": true},
			},
		},
		Variables: map[string]interface{}{
			"minAge": 18,
		},
		NodeData: map[string]interface{}{},
		NodeID:   "node-123",
		AgentID:  "agent-456",
	}

	code := `
		const users = input.data.users;
		const minAge = input.variables.minAge;
		
		const activeUsers = users.filter(user => user.active && user.age >= minAge);
		const userNames = activeUsers.map(user => user.name.toUpperCase());
		const averageAge = activeUsers.reduce((sum, user) => sum + user.age, 0) / activeUsers.length;
		
		return {
			totalUsers: users.length,
			activeUsers: activeUsers.length,
			userNames: userNames,
			averageAge: averageAge,
			summary: "Processed " + users.length + " users"
		};
	`

	result, err := runtime.Execute(context.Background(), code, execContext)
	require.NoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)

	assert.Equal(t, int64(3), resultMap["totalUsers"])
	assert.Equal(t, int64(2), resultMap["activeUsers"])

	userNames, ok := resultMap["userNames"].([]interface{})
	require.True(t, ok)
	assert.Equal(t, []interface{}{"ALICE", "CHARLIE"}, userNames)

	assert.Equal(t, int64(30), resultMap["averageAge"]) // (25 + 35) / 2
	assert.Equal(t, "Processed 3 users", resultMap["summary"])
}

func TestJavaScriptRuntime_Execute_SecurityRestrictions(t *testing.T) {
	runtime := NewJavaScriptRuntime()
	require.NoError(t, runtime.Initialize())

	execContext := CodeExecutionContext{
		Data:      map[string]interface{}{},
		Variables: map[string]interface{}{},
		NodeData:  map[string]interface{}{},
		NodeID:    "node-123",
		AgentID:   "agent-456",
	}

	tests := []struct {
		name        string
		code        string
		shouldError bool
	}{
		{
			name:        "require should be undefined",
			code:        `return typeof require;`,
			shouldError: false, // Should return "undefined"
		},
		{
			name:        "import should be undefined",
			code:        `var importVar = (typeof window !== 'undefined' && window.import) || this.import; return typeof importVar;`,
			shouldError: false, // Should return "undefined"
		},
		{
			name:        "eval should be undefined",
			code:        `return typeof eval;`,
			shouldError: false, // Should return "undefined"
		},
		{
			name:        "Function constructor should be undefined",
			code:        `return typeof Function;`,
			shouldError: false, // Should return "undefined"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := runtime.Execute(context.Background(), tt.code, execContext)
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, "undefined", result)
			}
		})
	}
}
