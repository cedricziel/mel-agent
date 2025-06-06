# Code Editor Integration Guide

## Overview

This document outlines the integration of Monaco Editor (VS Code's editor) into the MEL Agent platform for providing rich code editing capabilities in Code nodes that support multiple programming languages.

## Monaco Editor Integration

### 1. Package Installation

```json
{
  "dependencies": {
    "@monaco-editor/react": "^4.6.0",
    "monaco-editor": "^0.45.0"
  }
}
```

### 2. Component Architecture

#### Base Code Editor Component
```jsx
// src/components/CodeEditor.jsx
import { useState, useRef, useEffect } from 'react';
import Editor from '@monaco-editor/react';

export default function CodeEditor({
  value = '',
  onChange,
  language = 'javascript',
  theme = 'vs-dark',
  height = '400px',
  options = {},
  onMount,
  className = '',
  readonly = false,
  showMinimap = true,
  autoComplete = true,
  typeDefinitions = null,
  completionProvider = null
}) {
  const editorRef = useRef(null);
  const monacoRef = useRef(null);

  const defaultOptions = {
    fontSize: 14,
    lineNumbers: 'on',
    minimap: { enabled: showMinimap },
    automaticLayout: true,
    wordWrap: 'on',
    folding: true,
    renderWhitespace: 'boundary',
    scrollBeyondLastLine: false,
    readOnly: readonly,
    tabSize: 2,
    insertSpaces: true,
    bracketPairColorization: { enabled: true },
    guides: {
      bracketPairs: true,
      indentation: true
    },
    suggest: {
      enabled: autoComplete,
      showMethods: true,
      showFunctions: true,
      showConstructors: true,
      showFields: true,
      showVariables: true,
      showClasses: true,
      showStructs: true,
      showInterfaces: true,
      showModules: true,
      showProperties: true,
      showEvents: true,
      showOperators: false,
      showUnits: false,
      showValues: true,
      showConstants: true,
      showEnums: true,
      showEnumMembers: true,
      showKeywords: true,
      showWords: true,
      showColors: false,
      showFiles: false,
      showReferences: false,
      showFolders: false,
      showTypeParameters: true,
      showIssues: true,
      showUsers: false,
      showSnippets: true
    }
  };

  const mergedOptions = { ...defaultOptions, ...options };

  function handleEditorDidMount(editor, monaco) {
    editorRef.current = editor;
    monacoRef.current = monaco;

    // Setup type definitions if provided
    if (typeDefinitions && language === 'javascript') {
      monaco.languages.typescript.javascriptDefaults.addExtraLib(
        typeDefinitions,
        'mel-agent-types.d.ts'
      );
    }

    // Setup custom completion provider if provided
    if (completionProvider && autoComplete) {
      monaco.languages.registerCompletionItemProvider(language, completionProvider);
    }

    // Setup error markers and diagnostics
    monaco.languages.onLanguage(language, () => {
      monaco.languages.setLanguageConfiguration(language, {
        brackets: [
          ['{', '}'],
          ['[', ']'],
          ['(', ')']
        ],
        autoClosingPairs: [
          { open: '{', close: '}' },
          { open: '[', close: ']' },
          { open: '(', close: ')' },
          { open: '"', close: '"' },
          { open: "'", close: "'" },
          { open: '`', close: '`' }
        ],
        surroundingPairs: [
          { open: '{', close: '}' },
          { open: '[', close: ']' },
          { open: '(', close: ')' },
          { open: '"', close: '"' },
          { open: "'", close: "'" },
          { open: '`', close: '`' }
        ]
      });
    });

    // Call user's onMount callback
    if (onMount) {
      onMount(editor, monaco);
    }
  }

  // Expose editor instance for external control
  useEffect(() => {
    if (editorRef.current && window) {
      window.melAgentEditor = editorRef.current;
    }
  }, []);

  return (
    <div className={`border rounded-lg overflow-hidden ${className}`}>
      <Editor
        height={height}
        language={language}
        theme={theme}
        value={value}
        onChange={onChange}
        options={mergedOptions}
        onMount={handleEditorDidMount}
        loading={
          <div className="flex items-center justify-center h-full">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
          </div>
        }
      />
    </div>
  );
}
```

#### Multi-Language Code Editor Component
```jsx
// src/components/MultiLanguageCodeEditor.jsx
import CodeEditor from './CodeEditor';
import { 
  getMELAgentTypeDefinitions, 
  getLanguageCompletionProvider,
  getLanguagePlaceholder 
} from '../utils/editorHelpers';

export default function MultiLanguageCodeEditor({
  value = '',
  onChange,
  language = 'javascript',
  height = '400px',
  readonly = false,
  showMinimap = true
}) {
  const typeDefinitions = getMELAgentTypeDefinitions(language);
  const completionProvider = getLanguageCompletionProvider(language);
  const placeholder = getLanguagePlaceholder(language);

  const handleEditorMount = (editor, monaco) => {
    // Set placeholder if no value
    if (!value) {
      editor.setValue(placeholder);
      editor.setSelection(new monaco.Selection(1, 1, 1, 1));
    }

    // Add custom commands
    editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyS, () => {
      // Trigger save (will be handled by parent component)
      const saveEvent = new CustomEvent('editorSave', { detail: { value: editor.getValue() } });
      window.dispatchEvent(saveEvent);
    });

    // Add run command
    editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyMod.Shift | monaco.KeyCode.Enter, () => {
      // Trigger run (will be handled by parent component)
      const runEvent = new CustomEvent('editorRun', { detail: { value: editor.getValue() } });
      window.dispatchEvent(runEvent);
    });
  };

  const getLanguageOptions = (lang) => {
    const baseOptions = {
      formatOnPaste: true,
      formatOnType: true,
      autoIndent: 'full',
      contextmenu: true,
      mouseWheelZoom: true
    };

    switch (lang) {
      case 'python':
        return {
          ...baseOptions,
          tabSize: 4,
          insertSpaces: true,
          detectIndentation: false
        };
      case 'typescript':
        return {
          ...baseOptions,
          tabSize: 2,
          insertSpaces: true
        };
      default: // javascript
        return baseOptions;
    }
  };

  return (
    <CodeEditor
      value={value}
      onChange={onChange}
      language={language}
      height={height}
      readonly={readonly}
      showMinimap={showMinimap}
      typeDefinitions={typeDefinitions}
      completionProvider={completionProvider}
      onMount={handleEditorMount}
      options={getLanguageOptions(language)}
    />
  );
}
```

### 3. Editor Helper Utilities

#### Type Definitions Provider
```javascript
// src/utils/editorHelpers.js

export function getMELAgentTypeDefinitions(language = 'javascript') {
  if (language === 'python') {
    return getPythonTypeDefinitions();
  }
  return getJavaScriptTypeDefinitions();
}

function getJavaScriptTypeDefinitions() {
  return `
// MEL Agent Type Definitions for Code Node

interface InputContext {
  /** Data from the workflow envelope */
  data: any;
  
  /** Workflow variables accessible across nodes */
  variables: Record<string, any>;
  
  /** Configuration data from this node */
  nodeData: Record<string, any>;
  
  /** Unique identifier for this node */
  nodeId: string;
  
  /** Unique identifier for the current agent/workflow */
  agentId: string;
}

interface HTTPOptions {
  /** HTTP headers to send with the request */
  headers?: Record<string, string>;
  
  /** Request timeout in milliseconds */
  timeout?: number;
  
  /** Whether to follow redirects */
  followRedirects?: boolean;
}

interface HTTPResponse {
  /** HTTP status code */
  statusCode: number;
  
  /** Response headers */
  headers: Record<string, string>;
  
  /** Response body (automatically parsed if JSON) */
  data: any;
  
  /** Raw response text */
  text: string;
}

interface HTTPClient {
  /** Make a GET request */
  get(url: string, options?: HTTPOptions): Promise<HTTPResponse>;
  
  /** Make a POST request */
  post(url: string, body: any, options?: HTTPOptions): Promise<HTTPResponse>;
  
  /** Make a PUT request */
  put(url: string, body: any, options?: HTTPOptions): Promise<HTTPResponse>;
  
  /** Make a DELETE request */
  delete(url: string, options?: HTTPOptions): Promise<HTTPResponse>;
  
  /** Make a PATCH request */
  patch(url: string, body: any, options?: HTTPOptions): Promise<HTTPResponse>;
}

interface Utils {
  /** Parse JSON string safely, returns null if invalid */
  parseJSON(str: string): any;
  
  /** Convert object to JSON string */
  stringifyJSON(obj: any, pretty?: boolean): string;
  
  /** Generate MD5 hash of string */
  md5(str: string): string;
  
  /** Generate SHA256 hash of string */
  sha256(str: string): string;
  
  /** Generate a new UUID v4 */
  generateUUID(): string;
  
  /** Base64 encode a string */
  base64Encode(str: string): string;
  
  /** Base64 decode a string */
  base64Decode(str: string): string;
  
  /** URL encode a string */
  urlEncode(str: string): string;
  
  /** URL decode a string */
  urlDecode(str: string): string;
  
  /** Generate random number between min and max */
  randomInt(min: number, max: number): number;
  
  /** Generate random float between 0 and 1 */
  randomFloat(): number;
  
  /** Sleep for specified milliseconds */
  sleep(ms: number): Promise<void>;
  
  /** Deep clone an object */
  deepClone(obj: any): any;
  
  /** Check if value is empty (null, undefined, empty string, empty array, empty object) */
  isEmpty(value: any): boolean;
  
  /** Validate email address format */
  isValidEmail(email: string): boolean;
  
  /** Validate URL format */
  isValidURL(url: string): boolean;
}

interface Console {
  /** Log a message (for debugging) */
  log(...args: any[]): void;
  
  /** Log an error message */
  error(...args: any[]): void;
  
  /** Log a warning message */
  warn(...args: any[]): void;
  
  /** Log an info message */
  info(...args: any[]): void;
  
  /** Log a debug message */
  debug(...args: any[]): void;
}

// Global objects available in your code
declare const input: InputContext;
declare const utils: Utils;
declare const http: HTTPClient;
declare const console: Console;

// Return type for your code - you can return any structure
type CodeResult = any;
`;
}

function getPythonTypeDefinitions() {
  return `
# MEL Agent Type Definitions for Code Node (Python)

from typing import Dict, Any, Optional, Union

class InputContext:
    """Input context object containing data, variables, and node information"""
    data: Any  # Data from the workflow envelope
    variables: Dict[str, Any]  # Workflow variables accessible across nodes
    node_data: Dict[str, Any]  # Configuration data from this node
    node_id: str  # Unique identifier for this node
    agent_id: str  # Unique identifier for the current agent/workflow

class HTTPResponse:
    """HTTP response object"""
    status_code: int  # HTTP status code
    headers: Dict[str, str]  # Response headers
    data: Any  # Response body (automatically parsed if JSON)
    text: str  # Raw response text

class HTTPClient:
    """HTTP client for making requests"""
    def get(self, url: str, options: Optional[Dict[str, Any]] = None) -> HTTPResponse: ...
    def post(self, url: str, body: Any, options: Optional[Dict[str, Any]] = None) -> HTTPResponse: ...
    def put(self, url: str, body: Any, options: Optional[Dict[str, Any]] = None) -> HTTPResponse: ...
    def delete(self, url: str, options: Optional[Dict[str, Any]] = None) -> HTTPResponse: ...

class Utils:
    """Utility functions"""
    def parse_json(self, json_str: str) -> Any: ...
    def stringify_json(self, obj: Any, pretty: bool = False) -> str: ...
    def md5(self, text: str) -> str: ...
    def sha256(self, text: str) -> str: ...
    def generate_uuid(self) -> str: ...
    def base64_encode(self, text: str) -> str: ...
    def base64_decode(self, text: str) -> str: ...
    def url_encode(self, text: str) -> str: ...
    def url_decode(self, text: str) -> str: ...

# Global objects available in your code
input: InputContext
utils: Utils
http: HTTPClient
# Console methods: print() for logging
`;
}

export function getLanguageCompletionProvider(language = 'javascript') {
  if (language === 'python') {
    return getPythonCompletionProvider();
  }
  return getJavaScriptCompletionProvider();
}

export function getLanguagePlaceholder(language = 'javascript') {
  switch (language) {
    case 'python':
      return `# Write your Python code here
# Available objects: input, utils, http
# Use print() for logging

data = input.data

# Process your data here

return {
    "processed": True,
    "result": data
}`;
    case 'typescript':
      return `// Write your TypeScript code here
// Available objects: input, utils, http, console

const data = input.data;

// Process your data here

return {
  processed: true,
  result: data
};`;
    default: // javascript
      return `// Write your JavaScript code here
// Available objects: input, utils, http, console

const data = input.data;

// Process your data here

return {
  processed: true,
  result: data
};`;
  }
}

function getJavaScriptCompletionProvider() {
  return {
    provideCompletionItems: (model, position) => {
      const word = model.getWordUntilPosition(position);
      const range = {
        startLineNumber: position.lineNumber,
        endLineNumber: position.lineNumber,
        startColumn: word.startColumn,
        endColumn: word.endColumn
      };

      const suggestions = [
        // Input object completions
        {
          label: 'input',
          kind: 5, // Variable
          documentation: 'Input context object containing data, variables, and node information',
          insertText: 'input',
          range
        },
        {
          label: 'input.data',
          kind: 6, // Property
          documentation: 'Data from the workflow envelope',
          insertText: 'input.data',
          range
        },
        {
          label: 'input.variables',
          kind: 6, // Property
          documentation: 'Workflow variables accessible across nodes',
          insertText: 'input.variables',
          range
        },
        {
          label: 'input.nodeData',
          kind: 6, // Property
          documentation: 'Configuration data from this node',
          insertText: 'input.nodeData',
          range
        },

        // Utils completions
        {
          label: 'utils.parseJSON',
          kind: 2, // Method
          documentation: 'Parse JSON string safely, returns null if invalid',
          insertText: 'utils.parseJSON(${1:jsonString})',
          insertTextRules: 4, // InsertAsSnippet
          range
        },
        {
          label: 'utils.stringifyJSON',
          kind: 2, // Method
          documentation: 'Convert object to JSON string',
          insertText: 'utils.stringifyJSON(${1:object})',
          insertTextRules: 4,
          range
        },
        {
          label: 'utils.generateUUID',
          kind: 2, // Method
          documentation: 'Generate a new UUID v4',
          insertText: 'utils.generateUUID()',
          range
        },
        {
          label: 'utils.md5',
          kind: 2, // Method
          documentation: 'Generate MD5 hash of string',
          insertText: 'utils.md5(${1:string})',
          insertTextRules: 4,
          range
        },

        // HTTP client completions
        {
          label: 'http.get',
          kind: 2, // Method
          documentation: 'Make HTTP GET request',
          insertText: 'await http.get(${1:url}, ${2:options})',
          insertTextRules: 4,
          range
        },
        {
          label: 'http.post',
          kind: 2, // Method
          documentation: 'Make HTTP POST request',
          insertText: 'await http.post(${1:url}, ${2:body}, ${3:options})',
          insertTextRules: 4,
          range
        },

        // Console completions
        {
          label: 'console.log',
          kind: 2, // Method
          documentation: 'Log a message for debugging',
          insertText: 'console.log(${1:message})',
          insertTextRules: 4,
          range
        },
        {
          label: 'console.error',
          kind: 2, // Method
          documentation: 'Log an error message',
          insertText: 'console.error(${1:message})',
          insertTextRules: 4,
          range
        },

        // Common patterns
        {
          label: 'return-object',
          kind: 14, // Snippet
          documentation: 'Return result object',
          insertText: 'return {\n  ${1:success}: ${2:true},\n  ${3:data}: ${4:result}\n};',
          insertTextRules: 4,
          range
        },
        {
          label: 'try-catch',
          kind: 14, // Snippet
          documentation: 'Try-catch block for error handling',
          insertText: 'try {\n  ${1:// Your code here}\n} catch (error) {\n  console.error(${2:"Error:"}, error.message);\n  return { error: error.message };\n}',
          insertTextRules: 4,
          range
        },
        {
          label: 'async-function',
          kind: 14, // Snippet
          documentation: 'Async function pattern',
          insertText: 'async function ${1:processData}() {\n  ${2:// Your async code here}\n}\n\n${3:return await processData();}',
          insertTextRules: 4,
          range
        }
      ];

      return { suggestions };
    }
  };
}
```

### 4. Enhanced Parameter Rendering

#### Updated NodeModal Integration
```jsx
// Enhancement to src/components/NodeModal.jsx

import JavaScriptCodeEditor from './JavaScriptCodeEditor';

// In the renderParameterField function, add a new case:
case 'javascript':
  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <span className="text-sm text-gray-600">
          JavaScript Code Editor
        </span>
        <div className="flex items-center gap-2 text-xs text-gray-500">
          <span>Ctrl+S: Save</span>
          <span>Ctrl+Shift+Enter: Test</span>
        </div>
      </div>
      <JavaScriptCodeEditor
        value={val || ''}
        onChange={(value) => handleChange(param.name, value)}
        height="500px"
        readonly={viewMode === 'executions'}
      />
    </div>
  );

// Also handle the new format in the parameter type detection
const paramType = param.parameterType || param.type;
const isCodeParameter = param.jsonSchema?.format === 'javascript';

if (isCodeParameter || paramType === 'javascript') {
  // Render code editor
  return renderCodeEditor(param);
}
```

### 4. Enhanced Parameter Rendering

#### Updated NodeModal Integration
```jsx
// Enhancement to src/components/NodeModal.jsx

import MultiLanguageCodeEditor from './MultiLanguageCodeEditor';

// In the renderParameterField function, add a new case:
case 'code':
  const selectedLanguage = currentFormData.language || 'javascript';
  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <span className="text-sm text-gray-600">
          Code Editor ({selectedLanguage})
        </span>
        <div className="flex items-center gap-2 text-xs text-gray-500">
          <span>Ctrl+S: Save</span>
          <span>Ctrl+Shift+Enter: Test</span>
        </div>
      </div>
      <MultiLanguageCodeEditor
        value={val || ''}
        onChange={(value) => handleChange(param.name, value)}
        language={selectedLanguage}
        height="500px"
        readonly={viewMode === 'executions'}
      />
    </div>
  );

// Also handle the new format in the parameter type detection
const paramType = param.parameterType || param.type;
const isCodeParameter = param.jsonSchema?.format === 'code';

if (isCodeParameter || paramType === 'code') {
  // Render code editor
  return renderCodeEditor(param);
}
```

### 5. Event Handling

#### Editor Events Setup
```jsx
// In the parent component using the code editor

useEffect(() => {
  const handleEditorSave = (event) => {
    const code = event.detail.value;
    handleChange('code', code);
    onSave?.();
  };

  const handleEditorRun = (event) => {
    const code = event.detail.value;
    handleChange('code', code);
    onExecute?.({ code });
  };

  window.addEventListener('editorSave', handleEditorSave);
  window.addEventListener('editorRun', handleEditorRun);

  return () => {
    window.removeEventListener('editorSave', handleEditorSave);
    window.removeEventListener('editorRun', handleEditorRun);
  };
}, [handleChange, onSave, onExecute]);
```

### 6. Theme and Styling

#### Custom Theme Configuration
```javascript
// src/utils/editorThemes.js

export function setupMELAgentTheme(monaco) {
  monaco.editor.defineTheme('mel-agent-dark', {
    base: 'vs-dark',
    inherit: true,
    rules: [
      { token: 'comment', foreground: '6A9955' },
      { token: 'keyword', foreground: '569CD6' },
      { token: 'string', foreground: 'CE9178' },
      { token: 'number', foreground: 'B5CEA8' },
      { token: 'regexp', foreground: 'D16969' },
      { token: 'type', foreground: '4EC9B0' },
      { token: 'class', foreground: '4EC9B0' },
      { token: 'function', foreground: 'DCDCAA' },
      { token: 'variable', foreground: '9CDCFE' }
    ],
    colors: {
      'editor.background': '#1E1E1E',
      'editor.foreground': '#D4D4D4',
      'editor.lineHighlightBackground': '#2D2D30',
      'editor.selectionBackground': '#264F78',
      'editor.inactiveSelectionBackground': '#3A3D41'
    }
  });

  monaco.editor.defineTheme('mel-agent-light', {
    base: 'vs',
    inherit: true,
    rules: [
      { token: 'comment', foreground: '008000' },
      { token: 'keyword', foreground: '0000FF' },
      { token: 'string', foreground: 'A31515' },
      { token: 'number', foreground: '098658' },
      { token: 'regexp', foreground: 'D16969' },
      { token: 'type', foreground: '267F99' },
      { token: 'class', foreground: '267F99' },
      { token: 'function', foreground: '795E26' },
      { token: 'variable', foreground: '001080' }
    ],
    colors: {
      'editor.background': '#FFFFFF',
      'editor.foreground': '#000000',
      'editor.lineHighlightBackground': '#F5F5F5',
      'editor.selectionBackground': '#ADD6FF',
      'editor.inactiveSelectionBackground': '#E5EBF1'
    }
  });
}
```

### 7. Error Handling and Validation

#### Code Validation Component
```jsx
// src/components/CodeValidator.jsx

import { useState, useEffect } from 'react';

export default function CodeValidator({ code, onValidationChange }) {
  const [errors, setErrors] = useState([]);
  const [warnings, setWarnings] = useState([]);

  useEffect(() => {
    validateCode(code);
  }, [code]);

  const validateCode = async (codeString) => {
    const newErrors = [];
    const newWarnings = [];

    // Basic syntax validation
    try {
      new Function(codeString);
    } catch (error) {
      newErrors.push({
        line: 1,
        column: 1,
        message: `Syntax Error: ${error.message}`,
        severity: 'error'
      });
    }

    // MEL Agent specific validations
    if (!codeString.includes('return')) {
      newWarnings.push({
        line: 1,
        column: 1,
        message: 'Consider adding a return statement to output data',
        severity: 'warning'
      });
    }

    if (codeString.includes('require(') || codeString.includes('import(')) {
      newErrors.push({
        line: 1,
        column: 1,
        message: 'require() and import() are not allowed in the sandbox',
        severity: 'error'
      });
    }

    setErrors(newErrors);
    setWarnings(newWarnings);
    
    onValidationChange?.({
      isValid: newErrors.length === 0,
      errors: newErrors,
      warnings: newWarnings
    });
  };

  if (errors.length === 0 && warnings.length === 0) {
    return null;
  }

  return (
    <div className="mt-2 space-y-1">
      {errors.map((error, index) => (
        <div key={`error-${index}`} className="flex items-center gap-2 text-sm text-red-600">
          <span className="font-mono">×</span>
          <span>{error.message}</span>
        </div>
      ))}
      {warnings.map((warning, index) => (
        <div key={`warning-${index}`} className="flex items-center gap-2 text-sm text-yellow-600">
          <span className="font-mono">⚠</span>
          <span>{warning.message}</span>
        </div>
      ))}
    </div>
  );
}
```

### 8. Performance Optimizations

#### Code Editor Optimizations
```jsx
// Debounced change handler
import { useMemo, useCallback } from 'react';
import { debounce } from 'lodash';

function useCodeEditor(initialValue, onChange) {
  const debouncedOnChange = useMemo(
    () => debounce(onChange, 300),
    [onChange]
  );

  const handleChange = useCallback((value) => {
    debouncedOnChange(value);
  }, [debouncedOnChange]);

  return { handleChange };
}
```

## Testing Strategy

### 1. Unit Tests for Components
```javascript
// tests/components/CodeEditor.test.jsx
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import CodeEditor from '../src/components/CodeEditor';

describe('CodeEditor', () => {
  test('renders editor with default options', () => {
    render(<CodeEditor value="console.log('hello');" />);
    expect(screen.getByRole('textbox')).toBeInTheDocument();
  });

  test('calls onChange when code is modified', async () => {
    const mockOnChange = jest.fn();
    render(<CodeEditor value="" onChange={mockOnChange} />);
    
    const editor = screen.getByRole('textbox');
    await userEvent.type(editor, 'const x = 1;');
    
    expect(mockOnChange).toHaveBeenCalled();
  });
});
```

### 2. Integration Tests
```javascript
// tests/integration/JSCodeNode.test.jsx
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import NodeModal from '../src/components/NodeModal';

describe('Code Node Integration', () => {
  test('renders code editor for code node type', () => {
    const nodeDef = {
      type: 'code',
      label: 'Code',
      parameters: [
        {
          name: 'language',
          type: 'enum',
          options: ['javascript', 'python', 'typescript']
        },
        {
          name: 'code',
          type: 'string',
          jsonSchema: { format: 'code' }
        }
      ]
    };

    render(
      <NodeModal
        node={{ id: 'test', data: { language: 'javascript' } }}
        nodeDef={nodeDef}
        isOpen={true}
        onClose={() => {}}
        onChange={() => {}}
      />
    );

    expect(screen.getByText('Code Editor')).toBeInTheDocument();
  });
  
  test('switches language when selection changes', async () => {
    // Test language switching functionality
  });
});
```

## Deployment Considerations

### 1. Bundle Size Optimization
- Monaco Editor is large (~3MB). Consider code-splitting:
```javascript
// Lazy load Monaco Editor
const CodeEditor = lazy(() => import('./components/CodeEditor'));

// In component
<Suspense fallback={<div>Loading editor...</div>}>
  <CodeEditor />
</Suspense>
```

### 2. CDN Integration
```javascript
// Alternative: Load Monaco from CDN
import { loader } from '@monaco-editor/react';

loader.config({
  paths: { vs: 'https://cdn.jsdelivr.net/npm/monaco-editor@0.45.0/min/vs' }
});
```

This integration provides a robust, user-friendly code editing experience that enhances the MEL Agent platform's flexibility while maintaining security and performance standards.