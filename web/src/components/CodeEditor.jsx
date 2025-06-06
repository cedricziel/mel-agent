import { Editor } from '@monaco-editor/react';
import { useState, useEffect } from 'react';

/**
 * CodeEditor component that provides syntax highlighting and auto-completion
 * for the Code node. Adapts language highlighting based on the node's language parameter.
 */
export default function CodeEditor({ 
  value = '', 
  onChange, 
  language = 'javascript',
  height = '200px',
  className = '',
  placeholder = 'Enter your code here...',
  readOnly = false 
}) {
  const [editorLanguage, setEditorLanguage] = useState('javascript');

  // Map MEL Agent language names to Monaco Editor language identifiers
  useEffect(() => {
    const languageMap = {
      'javascript': 'javascript',
      'typescript': 'typescript', 
      'python': 'python',
      'json': 'json',
      'yaml': 'yaml',
      'sql': 'sql',
      'html': 'html',
      'css': 'css'
    };
    
    setEditorLanguage(languageMap[language] || 'javascript');
  }, [language]);

  const editorOptions = {
    minimap: { enabled: false },
    scrollBeyondLastLine: false,
    fontSize: 14,
    lineNumbers: 'on',
    roundedSelection: false,
    readOnly,
    cursorStyle: 'line',
    automaticLayout: true,
    wordWrap: 'on',
    tabSize: 2,
    insertSpaces: true,
    folding: true,
    foldingStrategy: 'indentation',
    showFoldingControls: 'always',
    unfoldOnClickAfterEndOfLine: false,
    contextmenu: true,
    selectOnLineNumbers: true,
    lineDecorationsWidth: 0,
    lineNumbersMinChars: 3,
    glyphMargin: false,
    fixedOverflowWidgets: true
  };

  const handleEditorChange = (value) => {
    if (onChange) {
      onChange(value || '');
    }
  };

  const handleEditorDidMount = (editor, monaco) => {
    // Configure JavaScript/TypeScript auto-completion with MEL Agent context
    if (editorLanguage === 'javascript' || editorLanguage === 'typescript') {
      // Enhanced compiler options for better IntelliSense
      monaco.languages.typescript.javascriptDefaults.setCompilerOptions({
        target: monaco.languages.typescript.ScriptTarget.ES2020,
        allowNonTsExtensions: true,
        moduleResolution: monaco.languages.typescript.ModuleResolutionKind.NodeJs,
        module: monaco.languages.typescript.ModuleKind.CommonJS,
        noEmit: true,
        esModuleInterop: true,
        allowJs: true,
        checkJs: false,
        allowSyntheticDefaultImports: true,
        strict: false,
        noImplicitAny: false,
        noImplicitReturns: false,
        noImplicitThis: false
      });

      // More comprehensive MEL Agent type definitions
      const melAgentTypes = `
// MEL Agent Runtime API
declare const input: {
  /** Input data from the previous node or trigger */
  data: any;
  /** Workflow variables available across nodes */
  variables: { [key: string]: any };
  /** Configuration data from this node */
  nodeData: { [key: string]: any };
  /** Unique identifier of this node */
  nodeId: string;
  /** Unique identifier of the current agent/workflow */
  agentId: string;
};

declare const utils: {
  /** Parse JSON string into object */
  parseJSON(str: string): any;
  /** Convert object to JSON string */
  stringifyJSON(obj: any): string;
  /** Generate MD5 hash of string */
  md5(str: string): string;
  /** Generate UUID v4 string */
  generateUUID(): string;
  /** Encode string to base64 */
  base64Encode(str: string): string;
  /** Decode base64 string */
  base64Decode(str: string): string;
};

declare const console: {
  /** Log information message */
  log(...args: any[]): void;
  /** Log error message */
  error(...args: any[]): void;
  /** Log warning message */
  warn(...args: any[]): void;
  /** Log info message */
  info(...args: any[]): void;
  /** Log debug message */
  debug(...args: any[]): void;
};

// Common JavaScript globals
declare const Date: DateConstructor;
declare const JSON: JSON;
declare const Math: Math;
declare const Number: NumberConstructor;
declare const String: StringConstructor;
declare const Array: ArrayConstructor;
declare const Object: ObjectConstructor;
`;

      // Add the type definitions
      monaco.languages.typescript.javascriptDefaults.addExtraLib(melAgentTypes, 'file:///mel-agent-globals.d.ts');
      
      // Configure diagnostics to be more permissive
      monaco.languages.typescript.javascriptDefaults.setDiagnosticsOptions({
        noSemanticValidation: false,
        noSyntaxValidation: false,
        noSuggestionDiagnostics: true,
        diagnosticCodesToIgnore: [1108, 1005, 1109, 1056]
      });

      // Configure IntelliSense options
      monaco.languages.typescript.javascriptDefaults.setEagerModelSync(true);
      
      // Register completion provider for better suggestions
      monaco.languages.registerCompletionItemProvider('javascript', {
        provideCompletionItems: (model, position) => {
          const word = model.getWordUntilPosition(position);
          const range = {
            startLineNumber: position.lineNumber,
            endLineNumber: position.lineNumber,
            startColumn: word.startColumn,
            endColumn: word.endColumn
          };

          const suggestions = [
            {
              label: 'input',
              kind: monaco.languages.CompletionItemKind.Variable,
              insertText: 'input',
              documentation: 'MEL Agent input context with data, variables, nodeData, nodeId, and agentId',
              range: range
            },
            {
              label: 'input.data',
              kind: monaco.languages.CompletionItemKind.Property,
              insertText: 'input.data',
              documentation: 'Input data from the previous node',
              range: range
            },
            {
              label: 'input.variables',
              kind: monaco.languages.CompletionItemKind.Property,
              insertText: 'input.variables',
              documentation: 'Workflow variables available across nodes',
              range: range
            },
            {
              label: 'utils',
              kind: monaco.languages.CompletionItemKind.Module,
              insertText: 'utils',
              documentation: 'MEL Agent utility functions',
              range: range
            },
            {
              label: 'utils.parseJSON',
              kind: monaco.languages.CompletionItemKind.Function,
              insertText: 'utils.parseJSON(${1:jsonString})',
              insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
              documentation: 'Parse JSON string into object',
              range: range
            },
            {
              label: 'utils.generateUUID',
              kind: monaco.languages.CompletionItemKind.Function,
              insertText: 'utils.generateUUID()',
              documentation: 'Generate UUID v4 string',
              range: range
            }
          ];

          return { suggestions: suggestions };
        }
      });
    }

    // Set theme to match the UI
    monaco.editor.setTheme('vs');
    
    // Focus the editor
    editor.focus();
  };

  return (
    <div className={`border rounded ${className}`}>
      <div className="bg-gray-50 px-3 py-2 border-b flex items-center justify-between">
        <div className="flex items-center space-x-2">
          <span className="text-xs font-medium text-gray-600">Language:</span>
          <span className="text-xs bg-blue-100 text-blue-800 px-2 py-1 rounded">
            {language}
          </span>
        </div>
        <div className="flex items-center space-x-2">
          <span className="text-xs text-gray-500">
            Ctrl+Space for autocomplete
          </span>
        </div>
      </div>
      <Editor
        height={height}
        language={editorLanguage}
        value={value}
        onChange={handleEditorChange}
        onMount={handleEditorDidMount}
        options={editorOptions}
        loading={
          <div className="flex items-center justify-center h-full">
            <div className="text-gray-500">Loading editor...</div>
          </div>
        }
      />
      {!value && (
        <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
          <div className="text-gray-400 text-sm">{placeholder}</div>
        </div>
      )}
    </div>
  );
}