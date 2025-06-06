# Code Node Design

## Overview

The Code Node allows users to write custom code in supported languages to process data within workflows. This node provides a secure, sandboxed environment for executing user-provided code with access to workflow data, variables, and platform utilities.

**Initially supported language:** JavaScript (with plans for Python, TypeScript, and others)

## Requirements

### Functional Requirements

1. **Multi-Language Support**: Execute code in multiple supported languages (starting with JavaScript)
2. **Language Selection**: Users can choose their preferred language from supported options
3. **Data Access**: Provide access to envelope data, variables, and node parameters
4. **Editor Experience**: Rich code editor with syntax highlighting and auto-completion per language
5. **Security**: Sandboxed execution environment with limited API access
6. **Error Handling**: Clear error reporting and debugging information
7. **Performance**: Reasonable execution limits and timeout handling

### Non-Functional Requirements

1. **Security**: No access to filesystem, network, or system APIs
2. **Performance**: Execution timeout of 30 seconds maximum
3. **Memory**: Limited memory allocation per execution
4. **Compatibility**: ES2020+ JavaScript support (initial), Python 3.9+ support (planned)
5. **Debugging**: Source maps and stack traces for errors
6. **Extensibility**: Plugin architecture for adding new language support

## Architecture

### Backend Components

#### 1. Multi-Language Runtime Engine

**JavaScript Runtime**
- **Technology**: Goja (Pure Go JavaScript engine)
- **Rationale**: No CGO dependencies, better security isolation
- **Alternatives Considered**: V8 (too complex), Otto (outdated ES support)

**Future Language Support**
- **Python**: Embedded Python interpreter or Starlark (Go-native Python subset)
- **TypeScript**: Transpile to JavaScript, execute with Goja
- **Lua**: GopherLua for lightweight scripting
- **WebAssembly**: WASM runtime for compiled languages

#### 2. Execution Context
```go
type CodeExecutionContext struct {
    // Input data from envelope
    Data      interface{}            `js:"data"`
    Variables map[string]interface{} `js:"variables"`
    
    // Platform utilities (limited subset)
    HTTP      JSHTTPClient           `js:"http"`
    Utils     JSUtilities            `js:"utils"`
    Console   JSConsole             `js:"console"`
    
    // Configuration
    NodeData  map[string]interface{} `js:"nodeData"`
    NodeID    string                 `js:"nodeId"`
    AgentID   string                 `js:"agentId"`
}
```

#### 3. Security Sandbox
- Disabled: `require()`, `import()`, `eval()`, `Function()`
- Limited: `setTimeout()`, `setInterval()` (with execution timeout)
- Allowed: JSON, Math, Date, String, Array methods

#### 4. API Surface
```javascript
// Available in user code
const data = input.data;          // Envelope data
const variables = input.variables; // Workflow variables
const nodeData = input.nodeData;   // Node configuration

// Utilities
const result = utils.parseJSON(str);
const hash = utils.md5(str);
const uuid = utils.generateUUID();

// HTTP (limited)
const response = await http.get(url, { headers: {...} });
const data = await http.post(url, body, { headers: {...} });

// Console (for debugging)
console.log("Debug message");
console.error("Error message");

// Return data
return { 
  processedData: data,
  success: true 
};
```

### Frontend Components

#### 1. Code Editor Integration
- **Technology**: Monaco Editor (VS Code editor)
- **Features**: 
  - JavaScript syntax highlighting
  - Auto-completion for available APIs
  - Error squiggles and inline diagnostics
  - Code folding and minimap
  - Bracket matching and auto-indentation

#### 2. Editor Configuration
```javascript
// Monaco editor setup
const editorConfig = {
  language: 'javascript',
  theme: 'vs-dark',
  automaticLayout: true,
  fontSize: 14,
  lineNumbers: 'on',
  minimap: { enabled: true },
  scrollBeyondLastLine: false,
  wordWrap: 'on',
  folding: true,
  renderWhitespace: 'boundary'
};

// Auto-completion configuration
const completionProvider = {
  provideCompletionItems: (model, position) => ({
    suggestions: [
      {
        label: 'input.data',
        kind: monaco.languages.CompletionItemKind.Property,
        documentation: 'Access to envelope data'
      },
      {
        label: 'utils.parseJSON',
        kind: monaco.languages.CompletionItemKind.Function,
        documentation: 'Parse JSON string safely'
      }
      // ... more completions
    ]
  })
};
```

#### 3. TypeScript Definitions
Provide TypeScript definitions for auto-completion:

```typescript
interface InputContext {
  data: any;
  variables: Record<string, any>;
  nodeData: Record<string, any>;
  nodeId: string;
  agentId: string;
}

interface Utils {
  parseJSON(str: string): any;
  stringifyJSON(obj: any): string;
  md5(str: string): string;
  sha256(str: string): string;
  generateUUID(): string;
  base64Encode(str: string): string;
  base64Decode(str: string): string;
}

interface HTTPClient {
  get(url: string, options?: HTTPOptions): Promise<HTTPResponse>;
  post(url: string, body: any, options?: HTTPOptions): Promise<HTTPResponse>;
  put(url: string, body: any, options?: HTTPOptions): Promise<HTTPResponse>;
  delete(url: string, options?: HTTPOptions): Promise<HTTPResponse>;
}

declare const input: InputContext;
declare const utils: Utils;
declare const http: HTTPClient;
declare const console: Console;
```

## Implementation Details

### Backend Implementation

#### 1. Node Definition
```go
package code

import (
    "context"
    "fmt"
    "time"
    
    "github.com/dop251/goja"
    "github.com/cedricziel/mel-agent/pkg/api"
)

type codeDefinition struct {
    runtimes map[string]Runtime // Runtime engines by language
}

type Runtime interface {
    Execute(code string, context CodeExecutionContext) (interface{}, error)
    GetLanguage() string
    Initialize() error
    Cleanup() error
}

func (codeDefinition) Meta() api.NodeType {
    return api.NodeType{
        Type:     "code",
        Label:    "Code",
        Category: "Code",
        Parameters: []api.ParameterDefinition{
            api.NewEnumParameter("language", "Language", 
                []string{"javascript", "python", "typescript"}, true).
                WithDefault("javascript").
                WithGroup("Settings").
                WithDescription("Programming language to use"),
            api.NewStringParameter("code", "Code", true).
                WithDescription("Code to execute").
                WithGroup("Code").
                WithJSONSchema(&api.JSONSchema{
                    Type:   "string",
                    Format: "code", // Custom format for UI that adapts to language
                }),
            api.NewNumberParameter("timeout", "Timeout (seconds)", false).
                WithDefault(30).
                WithGroup("Settings").
                WithDescription("Maximum execution time"),
            api.NewBooleanParameter("strict_mode", "Strict Mode", false).
                WithDefault(true).
                WithGroup("Settings").
                WithDescription("Enable strict mode (where applicable)"),
        },
    }
}
```

#### 2. Execution Engine
```go
func (d codeDefinition) ExecuteEnvelope(ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[interface{}]) (*api.Envelope[interface{}], error) {
    code, ok := node.Data["code"].(string)
    if !ok || code == "" {
        return nil, api.NewNodeError(node.ID, "code", "code parameter is required")
    }
    
    language, ok := node.Data["language"].(string)
    if !ok || language == "" {
        language = "javascript" // Default language
    }
    
    timeout, _ := node.Data["timeout"].(float64)
    if timeout <= 0 {
        timeout = 30
    }
    
    // Get runtime for selected language
    runtime, exists := d.runtimes[language]
    if !exists {
        return nil, api.NewNodeError(node.ID, "code", "unsupported language: "+language)
    }
    
    // Prepare execution context
    execContext := CodeExecutionContext{
        Data:      envelope.Data,
        Variables: envelope.Variables,
        NodeData:  node.Data,
        NodeID:    node.ID,
        AgentID:   ctx.AgentID,
        Mel:       ctx.Mel,
    }
    
    // Execute with timeout
    resultChan := make(chan codeExecutionResult, 1)
    go func() {
        defer func() {
            if r := recover(); r != nil {
                resultChan <- codeExecutionResult{
                    Error: fmt.Errorf("panic during execution: %v", r),
                }
            }
        }()
        
        result, err := runtime.Execute(code, execContext)
        resultChan <- codeExecutionResult{
            Value: result,
            Error: err,
        }
    }()
    
    // Wait for result or timeout
    timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
    defer cancel()
    
    select {
    case result := <-resultChan:
        if result.Error != nil {
            return nil, api.NewNodeError(node.ID, "code", "execution error: "+result.Error.Error())
        }
        
        resultEnvelope := envelope.Clone()
        resultEnvelope.Trace = envelope.Trace.Next(node.ID)
        resultEnvelope.Data = result.Value
        
        return resultEnvelope, nil
        
    case <-timeoutCtx.Done():
        return nil, api.NewNodeError(node.ID, "code", "execution timeout exceeded")
    }
}

type codeExecutionResult struct {
    Value interface{}
    Error error
}
```

#### 3. Sandbox Setup
```go
func (d jsCodeDefinition) setupSandbox(runtime *goja.Runtime, ctx api.ExecutionContext, node api.Node, envelope *api.Envelope[interface{}], strictMode bool) error {
    // Disable dangerous globals
    runtime.Set("require", goja.Undefined())
    runtime.Set("import", goja.Undefined())
    runtime.Set("eval", goja.Undefined())
    runtime.Set("Function", goja.Undefined())
    
    // Setup input context
    inputObj := runtime.NewObject()
    inputObj.Set("data", envelope.Data)
    inputObj.Set("variables", envelope.Variables)
    inputObj.Set("nodeData", node.Data)
    inputObj.Set("nodeId", node.ID)
    inputObj.Set("agentId", ctx.AgentID)
    runtime.Set("input", inputObj)
    
    // Setup utilities
    utils := &JSUtilities{}
    runtime.Set("utils", utils)
    
    // Setup HTTP client (limited)
    httpClient := &JSHTTPClient{mel: ctx.Mel}
    runtime.Set("http", httpClient)
    
    // Setup console
    consoleObj := &JSConsole{}
    runtime.Set("console", consoleObj)
    
    if strictMode {
        runtime.RunString(`"use strict";`)
    }
    
    return nil
}
```

#### 4. Utility Objects
```go
type JSUtilities struct{}

func (u *JSUtilities) ParseJSON(str string) interface{} {
    var result interface{}
    if err := json.Unmarshal([]byte(str), &result); err != nil {
        return nil
    }
    return result
}

func (u *JSUtilities) StringifyJSON(obj interface{}) string {
    bytes, err := json.Marshal(obj)
    if err != nil {
        return ""
    }
    return string(bytes)
}

func (u *JSUtilities) MD5(str string) string {
    h := md5.Sum([]byte(str))
    return hex.EncodeToString(h[:])
}

func (u *JSUtilities) GenerateUUID() string {
    return uuid.New().String()
}

type JSHTTPClient struct {
    mel api.Mel
}

func (h *JSHTTPClient) Get(url string, options map[string]interface{}) (*JSHTTPResponse, error) {
    // Implementation with mel.HTTPRequest
    return h.makeRequest("GET", url, nil, options)
}

func (h *JSHTTPClient) Post(url string, body interface{}, options map[string]interface{}) (*JSHTTPResponse, error) {
    return h.makeRequest("POST", url, body, options)
}

type JSConsole struct {
    logs []string
}

func (c *JSConsole) Log(args ...interface{}) {
    c.logs = append(c.logs, fmt.Sprint(args...))
}

func (c *JSConsole) Error(args ...interface{}) {
    c.logs = append(c.logs, "ERROR: "+fmt.Sprint(args...))
}
```

### Frontend Implementation

#### 1. Parameter Rendering Enhancement
```jsx
// Enhanced NodeModal for code parameters
case 'javascript':
  return (
    <div className="h-96 border rounded">
      <MonacoEditor
        language="javascript"
        theme="vs-dark"
        value={val || ''}
        onChange={(value) => handleChange(param.name, value)}
        options={{
          fontSize: 14,
          lineNumbers: 'on',
          minimap: { enabled: true },
          automaticLayout: true,
          wordWrap: 'on'
        }}
        editorDidMount={(editor, monaco) => {
          // Setup auto-completion
          monaco.languages.registerCompletionItemProvider('javascript', {
            provideCompletionItems: (model, position) => ({
              suggestions: getJSCompletions()
            })
          });
          
          // Setup type definitions
          monaco.languages.typescript.javascriptDefaults.addExtraLib(
            getTypeDefinitions(),
            'mel-agent-types.d.ts'
          );
        }}
      />
    </div>
  );
```

#### 2. Auto-completion Data
```javascript
function getJSCompletions() {
  return [
    {
      label: 'input',
      kind: monaco.languages.CompletionItemKind.Variable,
      documentation: 'Input context object containing data, variables, and node information',
      insertText: 'input'
    },
    {
      label: 'input.data',
      kind: monaco.languages.CompletionItemKind.Property,
      documentation: 'Data from the envelope',
      insertText: 'input.data'
    },
    {
      label: 'utils.parseJSON',
      kind: monaco.languages.CompletionItemKind.Function,
      documentation: 'Parse JSON string safely',
      insertText: 'utils.parseJSON(${1:jsonString})',
      insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet
    },
    {
      label: 'http.get',
      kind: monaco.languages.CompletionItemKind.Function,
      documentation: 'Make HTTP GET request',
      insertText: 'await http.get(${1:url}, ${2:options})',
      insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet
    },
    {
      label: 'console.log',
      kind: monaco.languages.CompletionItemKind.Function,
      documentation: 'Log message for debugging',
      insertText: 'console.log(${1:message})',
      insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet
    }
  ];
}
```

#### 3. Package Dependencies
```json
{
  "dependencies": {
    "@monaco-editor/react": "^4.6.0",
    "monaco-editor": "^0.45.0"
  }
}
```

## Security Considerations

### 1. Sandbox Restrictions
- **No File System Access**: Prevent reading/writing files
- **No Network Access**: Except through controlled HTTP client
- **No Process Access**: Prevent spawning processes or accessing system info
- **No Eval/Function**: Disable dynamic code execution
- **Memory Limits**: Prevent memory exhaustion
- **CPU Limits**: Timeout and interrupt long-running code

### 2. HTTP Client Restrictions
- **Whitelist/Blacklist**: Configure allowed domains
- **Rate Limiting**: Prevent abuse of external services
- **Timeout Enforcement**: Prevent hanging requests
- **Size Limits**: Limit request/response sizes

### 3. Error Handling
- **Stack Trace Sanitization**: Remove internal paths
- **Error Context**: Provide helpful debugging info
- **Security Information**: Don't leak sensitive system details

## Usage Examples

### 1. Basic Data Transformation
```javascript
// Transform input data
const data = input.data;

return {
  transformed: {
    fullName: `${data.firstName} ${data.lastName}`,
    email: data.email.toLowerCase(),
    timestamp: new Date().toISOString()
  }
};
```

### 2. API Integration
```javascript
// Call external API with input data
const apiKey = input.nodeData.apiKey;
const userId = input.data.userId;

try {
  const response = await http.get(`https://api.example.com/users/${userId}`, {
    headers: {
      'Authorization': `Bearer ${apiKey}`,
      'Content-Type': 'application/json'
    }
  });
  
  return {
    user: response.data,
    enriched: true
  };
} catch (error) {
  console.error('API call failed:', error.message);
  return {
    error: error.message,
    enriched: false
  };
}
```

### 3. Data Validation
```javascript
// Validate and clean input data
const data = input.data;
const errors = [];

// Validate email
if (!data.email || !data.email.includes('@')) {
  errors.push('Invalid email address');
}

// Validate age
if (!data.age || data.age < 0 || data.age > 150) {
  errors.push('Invalid age');
}

// Clean phone number
const cleanPhone = data.phone?.replace(/[^\d]/g, '') || '';

return {
  isValid: errors.length === 0,
  errors: errors,
  cleanedData: {
    ...data,
    email: data.email?.toLowerCase(),
    phone: cleanPhone
  }
};
```

### 4. Complex Logic
```javascript
// Business logic with multiple conditions
const order = input.data;
const settings = input.variables.settings || {};

let discount = 0;
let shippingCost = 0;

// Calculate discount
if (order.amount > 100) {
  discount = 0.1; // 10% discount for orders over $100
}

if (order.customerType === 'premium') {
  discount += 0.05; // Additional 5% for premium customers
}

// Calculate shipping
if (order.amount > 50) {
  shippingCost = 0; // Free shipping over $50
} else {
  shippingCost = settings.standardShipping || 9.99;
}

const finalAmount = order.amount * (1 - discount) + shippingCost;

return {
  originalAmount: order.amount,
  discount: discount,
  discountAmount: order.amount * discount,
  shippingCost: shippingCost,
  finalAmount: finalAmount,
  summary: `Order total: $${finalAmount.toFixed(2)} (${Math.round(discount * 100)}% discount applied)`
};
```

## Implementation Plan

### Phase 1: Backend Foundation
1. Implement JavaScript runtime with Goja
2. Create sandbox environment
3. Build utility objects (utils, console)
4. Implement basic HTTP client
5. Add timeout and security controls

### Phase 2: Frontend Integration
1. Add Monaco Editor dependency
2. Implement code parameter renderer
3. Create auto-completion provider
4. Add TypeScript definitions
5. Enhance error display

### Phase 3: Advanced Features
1. Add more utility functions
2. Implement variable scoping
3. Add debugging capabilities
4. Performance optimizations
5. Advanced security features

### Phase 4: Documentation & Testing
1. Create comprehensive documentation
2. Add usage examples
3. Implement thorough test suite
4. Performance benchmarking
5. Security audit

## Testing Strategy

### 1. Unit Tests
```go
func TestJSCodeNode_BasicExecution(t *testing.T) {
    def := jsCodeDefinition{}
    
    node := api.Node{
        ID:   "test-js-node",
        Type: "javascript_code",
        Data: map[string]interface{}{
            "code": `
                const result = input.data.value * 2;
                return { doubled: result };
            `,
        },
    }
    
    envelope := &api.Envelope[interface{}]{
        Data: map[string]interface{}{
            "value": 5,
        },
    }
    
    result, err := def.ExecuteEnvelope(api.ExecutionContext{}, node, envelope)
    
    assert.NoError(t, err)
    assert.Equal(t, 10, result.Data.(map[string]interface{})["doubled"])
}
```

### 2. Security Tests
- Test sandbox restrictions
- Verify timeout enforcement
- Test memory limits
- Validate error handling

### 3. Performance Tests
- Benchmark execution times
- Test memory usage
- Stress test with complex scripts
- Concurrent execution tests

## Future Enhancements

1. **TypeScript Support**: Full TypeScript compilation
2. **NPM Modules**: Safe subset of NPM packages
3. **Debugging Tools**: Step-through debugging
4. **Code Templates**: Pre-built code snippets
5. **Performance Metrics**: Execution time tracking
6. **Code Sharing**: Reusable code libraries
7. **Version Control**: Code history and rollback
8. **Collaborative Editing**: Real-time collaboration

This design provides a comprehensive foundation for implementing a secure, user-friendly JavaScript code node that enhances the MEL Agent platform's flexibility while maintaining security and performance standards.