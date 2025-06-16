package code

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dop251/goja"
	"github.com/google/uuid"
)

// JavaScriptRuntime implements the Runtime interface for JavaScript execution
type JavaScriptRuntime struct {
	language string
}

// NewJavaScriptRuntime creates a new JavaScript runtime
func NewJavaScriptRuntime() Runtime {
	return &JavaScriptRuntime{
		language: "javascript",
	}
}

// GetLanguage returns the language this runtime supports
func (js *JavaScriptRuntime) GetLanguage() string {
	return js.language
}

// Initialize sets up the JavaScript runtime
func (js *JavaScriptRuntime) Initialize() error {
	// No global initialization needed for JavaScript runtime
	// Each execution gets its own VM instance
	return nil
}

// Cleanup cleans up the JavaScript runtime
func (js *JavaScriptRuntime) Cleanup() error {
	// No cleanup needed
	return nil
}

// Execute runs JavaScript code with the provided context
func (js *JavaScriptRuntime) Execute(ctx context.Context, code string, execContext CodeExecutionContext) (interface{}, error) {
	// Create new VM for this execution
	vm := goja.New()

	// Setup sandbox environment
	if err := js.setupSandbox(vm, execContext); err != nil {
		return nil, fmt.Errorf("failed to setup sandbox: %w", err)
	}

	// Wrap code in a function to allow return statements
	wrappedCode := fmt.Sprintf("(function() {\n%s\n})()", code)

	// Create a channel to handle the execution with context cancellation
	resultChan := make(chan goja.Value, 1)
	errorChan := make(chan error, 1)

	// Execute in a separate goroutine to handle cancellation
	go func() {
		defer func() {
			if r := recover(); r != nil {
				errorChan <- fmt.Errorf("panic during JavaScript execution: %v", r)
			}
		}()

		// Execute the code
		result, err := vm.RunString(wrappedCode)
		if err != nil {
			errorChan <- fmt.Errorf("execution error: %w", err)
			return
		}
		resultChan <- result
	}()

	// Wait for execution or cancellation
	select {
	case result := <-resultChan:
		// Convert result to Go value
		if result == nil || goja.IsUndefined(result) {
			return nil, nil
		}
		return result.Export(), nil
	case err := <-errorChan:
		return nil, err
	case <-ctx.Done():
		return nil, fmt.Errorf("JavaScript execution cancelled: %w", ctx.Err())
	}
}

// setupSandbox configures the JavaScript execution environment
func (js *JavaScriptRuntime) setupSandbox(vm *goja.Runtime, context CodeExecutionContext) error {
	// Disable dangerous globals
	vm.Set("require", goja.Undefined())
	vm.Set("import", goja.Undefined())
	vm.Set("eval", goja.Undefined())
	vm.Set("Function", goja.Undefined())

	// Setup input context
	inputObj := vm.NewObject()
	inputObj.Set("data", context.Data)
	inputObj.Set("variables", context.Variables)
	inputObj.Set("nodeData", context.NodeData)
	inputObj.Set("nodeId", context.NodeID)
	inputObj.Set("agentId", context.AgentID)
	vm.Set("input", inputObj)

	// Setup utility functions
	utils := js.createUtilities(vm)
	vm.Set("utils", utils)

	// Setup console for debugging
	console := js.createConsole(vm)
	vm.Set("console", console)

	// TODO: Setup HTTP client when needed
	// httpClient := js.createHTTPClient(vm, context.Mel)
	// vm.Set("http", httpClient)

	return nil
}

// createUtilities creates utility functions available to JavaScript code
func (js *JavaScriptRuntime) createUtilities(vm *goja.Runtime) *goja.Object {
	utils := vm.NewObject()

	// JSON utilities
	utils.Set("parseJSON", func(str string) interface{} {
		var result interface{}
		if err := json.Unmarshal([]byte(str), &result); err != nil {
			return nil
		}
		return result
	})

	utils.Set("stringifyJSON", func(obj interface{}) string {
		bytes, err := json.Marshal(obj)
		if err != nil {
			return ""
		}
		return string(bytes)
	})

	// Hash utilities
	utils.Set("md5", func(str string) string {
		h := md5.Sum([]byte(str))
		return hex.EncodeToString(h[:])
	})

	// UUID generation
	utils.Set("generateUUID", func() string {
		return uuid.New().String()
	})

	// Base64 utilities
	utils.Set("base64Encode", func(str string) string {
		// TODO: Implement base64 encoding
		return str
	})

	utils.Set("base64Decode", func(str string) string {
		// TODO: Implement base64 decoding
		return str
	})

	return utils
}

// createConsole creates console functions for debugging
func (js *JavaScriptRuntime) createConsole(vm *goja.Runtime) *goja.Object {
	console := vm.NewObject()

	// For now, console functions are no-ops
	// In a production implementation, these would integrate with the logging system
	console.Set("log", func(args ...interface{}) {
		// TODO: Integrate with logging system
		fmt.Println("CONSOLE LOG:", formatConsoleArgs(args...))
	})

	console.Set("error", func(args ...interface{}) {
		// TODO: Integrate with logging system
		fmt.Println("CONSOLE ERROR:", formatConsoleArgs(args...))
	})

	console.Set("warn", func(args ...interface{}) {
		// TODO: Integrate with logging system
		fmt.Println("CONSOLE WARN:", formatConsoleArgs(args...))
	})

	console.Set("info", func(args ...interface{}) {
		// TODO: Integrate with logging system
		fmt.Println("CONSOLE INFO:", formatConsoleArgs(args...))
	})

	console.Set("debug", func(args ...interface{}) {
		// TODO: Integrate with logging system
		fmt.Println("CONSOLE DEBUG:", formatConsoleArgs(args...))
	})

	return console
}

// formatConsoleArgs formats console arguments for output
func formatConsoleArgs(args ...interface{}) string {
	var parts []string
	for _, arg := range args {
		parts = append(parts, fmt.Sprintf("%v", arg))
	}
	return strings.Join(parts, " ")
}

// Compile-time interface check
var _ Runtime = (*JavaScriptRuntime)(nil)
